package run

import (
	"os/exec"
	"rgt-server/log"
	"rgt-server/server"
	"strconv"

	"github.com/shirou/gopsutil/v3/process"
)

func StartTrmApp(srv *server.Server, sess *server.Session, exePathName string, workingDir string, arguments []string) (*process.Process, error) {
	cmd := exec.Command(exePathName, arguments...)
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
