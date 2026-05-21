package terminal

import (
	"context"
	"fmt"
	"os/exec"
	"rgt-server/log"
	"rgt-server/protocol"
	"rgt-server/run"
	"rgt-server/util"
	"strconv"
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
	launchingAppSemaphore *util.Set[*TerminalSession]

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
	if launchingAppSemaphore == nil {
		launchingAppSemaphore = util.NewSet[*TerminalSession](uint(maxLaunchingApps))
	} else {
		launchingAppSemaphore.SetLimit(uint(maxLaunchingApps))

	}
	log.Infof("App launch semaphore configured with limit %d.", maxLaunchingApps)
}

func takeLaunchAppSlot(session *TerminalSession, timeout time.Duration) bool {
	if launchingAppSemaphore == nil {
		log.Debugf("terminal.takeLaunchAppSlot(). Session %d, no semaphore set", session.Id())
		return true
	}
	if timeout > 0 {
		if err := launchingAppSemaphore.AddWait(context.Background(), session, timeout); err != nil {
			log.Errorf("terminal.takeLaunchAppSlot(). Session %d, error taking slot: %v", session.Id(), err)
			return false
		}
	} else {
		launchingAppSemaphore.Add(session)
	}
	return true
}

func giveBackLaunchAppSlot(session *TerminalSession) {
	launchingAppSemaphore.Remove(session)
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

func launchStandaloneApp(svc *TerminalEmulationService, sess *TerminalSession, req *AppExecRequest) (*standaloneApp, protocol.ErrorResponse) {
	var err error
	if time.Since(lastTimeLaunchStandaloneApp) < svc.Config().AppMinLaunchIntervalStandalone().Get() {
		sleep := time.Now().Add(svc.Config().AppMinLaunchIntervalStandalone().Get())
		time.Sleep(time.Until(sleep))
	}
	if !takeLaunchAppSlot(sess, svc.Config().AppLaunchTimeout().Get()) {
		return nil, NewError(TE_APP_LAUNCH_ERROR, "Timeout waiting for app launch slot")
	}
	defer giveBackLaunchAppSlot(sess)
	if svc.sessionManager.GetSession(sess.Id()) == nil {
		return nil, NewError(TE_APP_LAUNCH_ERROR, "Error launching standalone app: Session ", sess.Id(), " not found")
	}
	if err = sess.ChangeStatus(SESS_NEW, SESS_LAUNCHING_APP); err != nil {
		return nil, NewError(TE_APP_LAUNCH_ERROR, "Error launching standalone app: ", err)
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
		waitStarting:          make(chan struct{}),
		killAppLostConnection: req.KillAppLostConnection,
		keepAliveInterval:     req.keepAliveInterval,
		lastDataSentTime:      time.Now(),
		launchTimeout:         svc.Config().AppLaunchTimeout().Get(),
		outputBuffer:          make([]byte, 0, 32*1024),
	}
	if req.CaptureOutput {
		cmd.Stderr = &outputWriter{app: app, errorOutput: true}
		cmd.Stdout = &outputWriter{app: app, errorOutput: false}
	}
	go app.waitFinish()
	go app.sendKeepAlive()
	err = cmd.Start()
	if err != nil {
		return nil, NewError(TE_APP_LAUNCH_ERROR, "Error launching standalone app: ", err)
	}
	app.running = true
	lastTimeLaunchStandaloneApp = time.Now()
	appProcess, err := process.NewProcess(int32(cmd.Process.Pid))
	if err != nil {
		log.Errorf("terminal.launchStandaloneApp(). Error creating standalone process data: %v", err)
		cmd.Process.Kill()
		return nil, NewError(TE_APP_LAUNCH_ERROR, "Error getting process data: ", err)
	}
	sess.SetProcess(appProcess)
	sess.SetAppPid(int64(appProcess.Pid))
	sess.SetAppLaunchTime(lastTimeLaunchStandaloneApp)
	return app, nil
}
