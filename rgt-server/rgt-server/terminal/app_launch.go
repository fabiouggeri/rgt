package terminal

import (
	"fmt"
	"os/exec"
	"rgt-server/log"
	"rgt-server/protocol"
	"rgt-server/run"
	"rgt-server/server"
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

var (
	startAppMutex         sync.Mutex
	launchingAppSemaphore atomic.Pointer[chan struct{}]

	lastTimeLaunchStandaloneApp                        = time.Now()
	sessionListener             server.SessionListener = &sessionStatusListener{}
)

func (w *outputWriter) Write(data []byte) (n int, err error) {
	return w.app.writeAppOutput(data, w.errorOutput)
}

func (s *sessionStatusListener) StatusChange(session *server.Session, oldStatus server.SessionStatus, newStatus server.SessionStatus) {
	if oldStatus == server.SESS_CONNECTING {
		appReady()
	}
}

func configureLaunchAppSemaphore(maxLaunchingApps uint32) {
	newChannel := make(chan struct{}, maxLaunchingApps)
	oldSemaphore := launchingAppSemaphore.Swap(&newChannel)
	if oldSemaphore != nil {
		close(*oldSemaphore)
	}
}

func launchingApp() {
	sem := launchingAppSemaphore.Load()
	if sem != nil {
		*sem <- struct{}{}
	}
}

func appReady() {
	sem := launchingAppSemaphore.Load()
	if sem != nil {
		<-*sem
	}
}

func launchTrmApp(srv *server.Server, sess *server.Session, exePathName string, workingDir string, arguments []string) protocol.ErrorResponse {
	sess.AddStatusListener(sessionListener)
	launchingApp()
	if err := sess.ChangeStatus(server.SESS_NEW, server.SESS_LAUNCHING_APP); err != nil {
		return NewError(TE_APP_LAUNCH_ERROR, "Error launching app: ", err)
	}
	process, err := run.StartTrmApp(srv, sess, exePathName, workingDir, arguments)
	if err != nil {
		return NewError(TE_APP_LAUNCH_ERROR, "Error launching app: ", err)
	}
	if err := sess.ChangeStatus(server.SESS_LAUNCHING_APP, server.SESS_CONNECTING); err != nil {
		return NewError(TE_APP_LAUNCH_ERROR, "Error launching app: ", err)
	}
	sess.SetProcess(process)
	sess.SetAppLaunchTime(time.Now())
	log.Infof("[TE;session=%d] terminal.launchApp(). pid=%d app=[%s]", sess.Id, process.Pid, exePathName)
	return nil
}

func launchStandaloneApp(srv *server.Server, sess *server.Session, req *AppExecRequest, protocolVersion int16) protocol.ErrorResponse {
	var err error
	startAppMutex.Lock()
	defer startAppMutex.Unlock()
	if srv.GetSession(sess.Id) == nil {
		return NewError(TE_APP_LAUNCH_ERROR, "Error launching standalone app: Session ", sess.Id, " not found")
	}
	if srv.TimeoutAppLaunch(sess) {
		return NewError(TE_APP_LAUNCH_ERROR, "Error launching standalone app: Timeout launching app for session ", sess.Id)
	}
	envVars := make([]string, 0, 32)
	envVars = append(envVars, req.EnvVars...)
	envVars = append(envVars, srv.EnvVars()...)
	envVars = append(envVars, "HB_GT=gtstd")
	envVars = append(envVars, server.ENV_VAR_STANDALONE_APP+"="+fmt.Sprint(sess.Id))
	cmd := exec.Command(req.ExePathName, req.Arguments...)
	cmd.Env = envVars
	cmd.Dir = req.WorkingDir

	app := &standaloneApp{
		server:                srv,
		session:               sess,
		cmd:                   cmd,
		killAppLostConnection: req.KillAppLostConnection,
		keepAliveInterval:     req.keepAliveInterval,
		lastDataSentTime:      time.Now()}
	app.protoOutput, err = findProtocol[*AppOutputRequest, *protocol.BaseResponse](TRM_STANDALONE_APP_SEND_OUTPUT, protocolVersion)
	if err != nil {
		return NewError(TE_APP_LAUNCH_ERROR, "Error launching standalone app: ", err)
	}
	app.protoStatus, err = findProtocol[*AppStatusRequest, *protocol.BaseResponse](TRM_STANDALONE_APP_SEND_STATUS, protocolVersion)
	if err != nil {
		return NewError(TE_APP_LAUNCH_ERROR, "Error launching standalone app: ", err)
	}
	if req.CaptureOutput {
		cmd.Stderr = &outputWriter{app: app, errorOutput: true}
		cmd.Stdout = &outputWriter{app: app, errorOutput: false}
	}
	if time.Since(lastTimeLaunchStandaloneApp) < srv.Config().AppMinLaunchIntervalStandalone().Get() {
		time.Sleep(srv.Config().AppMinLaunchIntervalStandalone().Get())
	}
	if err = sess.ChangeStatus(server.SESS_NEW, server.SESS_LAUNCHING_APP); err != nil {
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
	if err = sess.ChangeStatus(server.SESS_LAUNCHING_APP, server.SESS_READY); err != nil {
		return NewError(TE_APP_LAUNCH_ERROR, "Error launching standalone app: ", err)
	}
	sess.SetProcess(appProcess)
	sess.SetAppPid(int64(appProcess.Pid))
	sess.SetAppLaunchTime(lastTimeLaunchStandaloneApp)
	go app.waitFinish()
	go app.sendKeepAlive()
	log.Infof("[APP;session=%d] terminal.launchStandaloneApp(). pid=%d app=[%s]", sess.Id, appProcess.Pid, req.ExePathName)
	return nil
}
