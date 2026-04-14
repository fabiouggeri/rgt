package run

import (
	"os/exec"
	"rgt-server/log"
	"rgt-server/server"
	"strconv"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

var startAppMutex sync.Mutex
var lastTimeLaunchedApp time.Time = time.Now()

func StartTrmApp(srv *server.Server, sess *server.Session, exePathName string, workingDir string, arguments []string) (*process.Process, error) {
	startAppMutex.Lock()
	defer startAppMutex.Unlock()
	cmd := exec.Command(exePathName, arguments...)
	cmd.Dir = workingDir
	config := srv.Config()
	envVars := srv.EnvVars()
	envVars = append(envVars, server.ENV_VAR_SERVER_ADDR+"="+config.Address().Get())
	envVars = append(envVars, server.ENV_VAR_SERVER_PORT+"="+config.EmulationPort().GetString())
	envVars = append(envVars, server.ENV_VAR_AUTH_TOKEN+"="+strconv.FormatInt(sess.Id, 10))
	cmd.Env = envVars
	if time.Since(lastTimeLaunchedApp) < config.AppMinLaunchInterval().Get() {
		time.Sleep(config.AppMinLaunchInterval().Get())
	}
	log.Debugf("run.StartTrmApp() cmd=[%v]. env=[%v]", cmd, envVars)
	err := cmd.Start()
	lastTimeLaunchedApp = time.Now()
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
