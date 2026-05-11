package run

import (
	"os/exec"
	"rgt-server/log"
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

func StartTrmApp(config RunAppConfig, exePathName string, workingDir string, arguments []string, envVars []string) (*process.Process, error) {
	var flags uint32 = 0
	cmd := exec.Command(exePathName, arguments...)
	showConsole := config.ShowConsole().Get()
	if showConsole {
		flags = CREATE_NEW_CONSOLE
	} else {
		flags = CREATE_NO_WINDOW
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: flags, HideWindow: !showConsole}
	cmd.Dir = workingDir
	cmd.Env = envVars
	log.Debugf("run.StartTrmApp() cmd=[%v]. env=[%v]", cmd, cmd.Env)
	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	appProcess, err := process.NewProcess(int32(cmd.Process.Pid))
	if err != nil {
		log.Errorf("run.StartTrmApp().Error creating terminal process data. PId=%d Error=%v", cmd.Process.Pid, err)
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
