package run

import (
	"os/exec"
	"rgt-server/log"
	"rgt-server/server"

	"github.com/shirou/gopsutil/v3/process"
)

func StartTrmApp(srv *server.Server, exePathName string, workingDir string, arguments []string, envVars []string) (*process.Process, error) {
	cmd := exec.Command(exePathName, arguments...)
	cmd.Dir = workingDir
	cmd.Env = append(srv.EnvVars(), envVars...)
	log.Debugf("run.StartTrmApp() cmd=[%v]. env=[%v]", cmd, cmd.Env)
	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	appProcess, err := process.NewProcess(int32(cmd.Process.Pid))
	if err != nil {
		log.Errorf("run.StartTrmApp(). Error creating standalone process data: %v", err)
		cmd.Process.Kill()
		return nil, err
	}
	return appProcess, err
}

func Create(exePathName string, workingDir string, arguments []string, envVars []string) *exec.Cmd {
	cmd := exec.Command(exePathName, arguments...)
	cmd.Dir = workingDir
	cmd.Env = envVars
	return cmd
}
