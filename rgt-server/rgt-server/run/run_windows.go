package run

import (
	"os/exec"
	"rgt-server/log"
	"rgt-server/server"
	"strconv"
	"syscall"

	"github.com/shirou/gopsutil/v3/process"
)

const (
	CREATE_NEW_CONSOLE       uint32 = 0x00000010
	DETACHED_PROCESS         uint32 = 0x00000008
	CREATE_NO_WINDOW         uint32 = 0x08000000
	STANDARD_RIGHTS_REQUIRED uint32 = 0x000F0000
	SYNCHRONIZE              uint32 = 0x00100000
	PROCESS_ALL_ACCESS       uint32 = STANDARD_RIGHTS_REQUIRED | SYNCHRONIZE | 0xFFFF
)

func StartTrmApp(srv *server.Server, sess *server.Session, exePathName string, workingDir string, arguments []string) (*process.Process, error) {
	var flags uint32 = 0
	cmd := exec.Command(exePathName, arguments...)
	showConsole := srv.Config().ShowConsole().Get()
	if showConsole {
		flags = CREATE_NEW_CONSOLE
	} else {
		flags = CREATE_NO_WINDOW
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: flags, HideWindow: !showConsole}
	cmd.Dir = workingDir
	envVars := srv.EnvVars()
	envVars = append(envVars, server.ENV_VAR_SERVER_ADDR+"="+srv.Config().Address().Get())
	envVars = append(envVars, server.ENV_VAR_SERVER_PORT+"="+srv.Config().EmulationPort().GetString())
	envVars = append(envVars, server.ENV_VAR_AUTH_TOKEN+"="+strconv.FormatInt(sess.Id, 10))
	cmd.Env = envVars
	log.Debugf("run.StartTrmApp() cmd=[%v]. env=[%v]", cmd, envVars)
	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	appProcess, err := process.NewProcess(int32(cmd.Process.Pid))
	if err != nil {
		log.Errorf("run.StartTrmApp().Error creating standalone process data: %v", err)
		cmd.Process.Kill()
		return nil, err
	}
	return appProcess, err
}

func Create(exePathName string, workingDir string, arguments []string, envVars []string) *exec.Cmd {
	cmd := exec.Command(exePathName, arguments...)
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x00000010}
	cmd.Dir = workingDir
	cmd.Env = envVars
	log.Debugf("run.Create() cmd=[%v]. env=[%v]", cmd, envVars)
	return cmd
}
