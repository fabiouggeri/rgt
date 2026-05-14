package terminal

import (
	"fmt"
	"os/exec"
	"rgt-server/log"
	"rgt-server/protocol"
	"rgt-server/run"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

type outputWriter struct {
	app         *standaloneApp
	errorOutput bool
}

type sessionStatusListener struct{}

const (
	ENV_VAR_SERVER_ADDR       string = "RGT_SERVER_ADDR"
	ENV_VAR_SERVER_PORT       string = "RGT_SERVER_PORT"
	ENV_VAR_STANDALONE_APP    string = "RGT_STANDALONE_APP"
	ENV_VAR_AUTH_TOKEN        string = "RGT_AUTH_TOKEN"
	AUTH_TOKEN_VAR_PREFIX     string = "RGT_AUTH_TOKEN="
	AUTH_TOKEN_VAR_PREFIX_LEN int    = len(AUTH_TOKEN_VAR_PREFIX)
)

var (
	startAppMutex         sync.Mutex
	launchingAppSemaphore atomic.Pointer[chan struct{}]

	lastTimeLaunchStandaloneApp                 = time.Now()
	sessionListener             SessionListener = &sessionStatusListener{}
)

func (w *outputWriter) Write(data []byte) (n int, err error) {
	return w.app.writeAppOutput(data, w.errorOutput)
}

func (l *sessionStatusListener) StatusChange(session *TerminalSession, oldStatus SessionStatus, newStatus SessionStatus) {
	if oldStatus == SESS_CONNECTING {
		giveBackLaunchAppSlot(session)
		session.RemoveStatusListener(sessionListener)
	}
}

func configureLaunchAppSemaphore(maxLaunchingApps uint32) {
	newChannel := make(chan struct{}, maxLaunchingApps)
	oldSemaphore := launchingAppSemaphore.Swap(&newChannel)
	if oldSemaphore != nil {
		close(*oldSemaphore)
	}
	log.Infof("App launch semaphore created with limit %d.", maxLaunchingApps)
}

func takeLaunchAppSlot(session *TerminalSession, timeout time.Duration) bool {
	sem := launchingAppSemaphore.Load()
	if sem == nil {
		log.Debugf("terminal.takeLaunchAppSlot(). Session %d, no semaphore set", session.Id())
		return true
	}
	if timeout > 0 {
		select {
		case *sem <- struct{}{}:
			log.Debugf("terminal.takeLaunchAppSlot(). Session %d, slot taken", session.Id())
			return true
		case <-session.Done():
			log.Debugf("terminal.takeLaunchAppSlot(). Session %d, aborted due to session close", session.Id())
			return false
		case <-time.After(timeout):
			log.Debugf("terminal.takeLaunchAppSlot(). Session %d, timeout waiting for slot", session.Id())
			return false
		}
	} else {
		select {
		case *sem <- struct{}{}:
			log.Debugf("terminal.takeLaunchAppSlot(). No timeout set. Session %d, slot taken", session.Id())
			return true
		case <-session.Done():
			log.Debugf("terminal.takeLaunchAppSlot(). Session %d, aborted due to session close", session.Id())
			return false
		}
	}
}

func giveBackLaunchAppSlot(session *TerminalSession) {
	sem := launchingAppSemaphore.Load()
	if sem == nil {
		log.Debugf("terminal.giveBackLaunchAppSlot(). Session %d, no semaphore set", session.Id())
		return
	}
	select {
	case <-*sem:
		log.Debugf("terminal.giveBackLaunchAppSlot(). Session %d, slot given back", session.Id())
	default:
		log.Errorf("terminal.giveBackLaunchAppSlot(). Session %d, no slot to give back", session.Id())
	}
}

func launchTrmApp(svc *TerminalEmulationService, session *TerminalSession, exePathName string, workingDir string, arguments []string) protocol.ErrorResponse {
	if !takeLaunchAppSlot(session, svc.Config().AppLaunchTimeout().Get()) {
		return NewError(TE_APP_LAUNCH_ERROR, "Timeout waiting for app launch slot")
	}
	if err := session.ChangeStatus(SESS_NEW, SESS_LAUNCHING_APP); err != nil {
		giveBackLaunchAppSlot(session)
		return NewError(TE_APP_LAUNCH_ERROR, "Error launching app: ", err)
	}
	envVars := make([]string, 0, 3)
	envVars = append(envVars, ENV_VAR_SERVER_ADDR+"="+svc.Config().Address().Get())
	envVars = append(envVars, ENV_VAR_SERVER_PORT+"="+strconv.FormatUint(uint64(svc.AppListeningPort()), 10))
	envVars = append(envVars, ENV_VAR_AUTH_TOKEN+"="+strconv.FormatInt(session.Id(), 10))
	envVars = append(svc.Config().EnvVars(), envVars...)

	session.AddStatusListener(sessionListener)
	process, err := run.StartTrmApp(svc.Config(), exePathName, workingDir, arguments, envVars)
	if err != nil {
		giveBackLaunchAppSlot(session)
		return NewError(TE_APP_LAUNCH_ERROR, "Error launching app: ", err)
	}
	if err := session.ChangeStatus(SESS_LAUNCHING_APP, SESS_CONNECTING); err != nil {
		giveBackLaunchAppSlot(session)
		return NewError(TE_APP_LAUNCH_ERROR, "Error launching app: ", err)
	}
	session.SetProcess(process)
	session.SetAppLaunchTime(time.Now())
	log.Infof("[TE;session=%d] terminal.launchTrmApp(). pid=%d app=[%s]", session.Id(), process.Pid, exePathName)
	return nil
}

func launchStandaloneApp(svc *TerminalEmulationService, sess *TerminalSession, req *AppExecRequest) protocol.ErrorResponse {
	var err error
	startAppMutex.Lock()
	defer startAppMutex.Unlock()
	if svc.sessionManager.GetSession(sess.Id()) == nil {
		return NewError(TE_APP_LAUNCH_ERROR, "Error launching standalone app: Session ", sess.Id(), " not found")
	}
	if sess.timeoutAppLaunch(svc.Config()) {
		return NewError(TE_APP_LAUNCH_ERROR, "Error launching standalone app: Timeout launching app for session ", sess.Id())
	}
	envVars := make([]string, 0, 32)
	envVars = append(envVars, req.EnvVars...)
	envVars = append(envVars, svc.Config().EnvVars()...)
	envVars = append(envVars, "HB_GT=gtstd")
	envVars = append(envVars, ENV_VAR_STANDALONE_APP+"="+fmt.Sprint(sess.Id()))
	cmd := exec.Command(req.ExePathName, req.Arguments...)
	cmd.Env = envVars
	cmd.Dir = req.WorkingDir

	app := &standaloneApp{
		service:               svc,
		session:               sess,
		cmd:                   cmd,
		killAppLostConnection: req.KillAppLostConnection,
		keepAliveInterval:     req.keepAliveInterval,
		lastDataSentTime:      time.Now()}
	if req.CaptureOutput {
		cmd.Stderr = &outputWriter{app: app, errorOutput: true}
		cmd.Stdout = &outputWriter{app: app, errorOutput: false}
	}
	if time.Since(lastTimeLaunchStandaloneApp) < svc.Config().AppMinLaunchIntervalStandalone().Get() {
		time.Sleep(svc.Config().AppMinLaunchIntervalStandalone().Get())
	}
	if err = sess.ChangeStatus(SESS_NEW, SESS_LAUNCHING_APP); err != nil {
		return NewError(TE_APP_LAUNCH_ERROR, "Error launching standalone app: ", err)
	}
	err = cmd.Start()
	if err != nil {
		return NewError(TE_APP_LAUNCH_ERROR, "Error launching standalone app: ", err)
	}
	app.running = true
	lastTimeLaunchStandaloneApp = time.Now()
	appProcess, err := process.NewProcess(int32(cmd.Process.Pid))
	if err != nil {
		log.Errorf("terminal.launchStandaloneApp(). Error creating standalone process data: %v", err)
		cmd.Process.Kill()
		return NewError(TE_APP_LAUNCH_ERROR, "Error getting process data: ", err)
	}
	if err = sess.ChangeStatus(SESS_LAUNCHING_APP, SESS_READY); err != nil {
		return NewError(TE_APP_LAUNCH_ERROR, "Error launching standalone app: ", err)
	}
	sess.SetProcess(appProcess)
	sess.SetAppPid(int64(appProcess.Pid))
	sess.SetAppLaunchTime(lastTimeLaunchStandaloneApp)
	go app.waitFinish()
	go app.sendKeepAlive()
	log.Infof("[APP;session=%d] terminal.launchStandaloneApp(). pid=%d app=[%s]", sess.Id(), appProcess.Pid, req.ExePathName)
	return nil
}
