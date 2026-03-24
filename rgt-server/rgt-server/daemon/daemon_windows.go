package daemon

import (
	"rgt-server/log"
	"rgt-server/util"
	"strings"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

type ServiceExecutor struct {
	service Daemon
}

func (srv *ServiceExecutor) Execute(args []string, requests <-chan svc.ChangeRequest, changes chan<- svc.Status) (svcSpecificEC bool, exitCode uint32) {
	// const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}
	srv.service.Start(args)
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	srv.run(requests, changes, cmdsAccepted)
	return false, 0
}

func (srv *ServiceExecutor) run(requests <-chan svc.ChangeRequest, changes chan<- svc.Status, cmdsAccepted svc.Accepted) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("unknown error in server(ServiceExecutor.run): %v\n%s", err, util.FullStack())
		}
	}()
loop:
	for req := range requests {
		switch req.Cmd {
		case svc.Interrogate:
			changes <- req.CurrentStatus
		case svc.Stop, svc.Shutdown:
			changes <- svc.Status{State: svc.StopPending}
			srv.service.Stop()
			changes <- svc.Status{State: svc.Stopped}
			break loop
		case svc.Pause:
			changes <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
		case svc.Continue:
			changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
		default:
			log.Errorf("unexpected control request #%d", req)
		}
	}
}

func Run(service Daemon) {
	svc.Run(service.GetName(), &ServiceExecutor{service: service})
}

func findArg(args []string, arg string) (string, int) {
	for index, value := range args {
		option, optionValue, _ := strings.Cut(value, "=")
		if option == arg {
			return optionValue, index
		}
	}
	return "", -1
}

func removeArg(args []string, index int) []string {
	if index < 0 || index >= len(args) {
		return args
	}
	if index == len(args)-1 {
		return args[:index]
	} else {
		return append(args[:index], args[index+1:]...)
	}
}

func Install(serviceName string, args []string) {
	exepath, err := util.ExePath()
	if err != nil {
		log.Errorf("Error finding executable: %v\n", err)
		return
	}
	manager, err := mgr.Connect()
	if err != nil {
		log.Errorf("Error connecting to service manager: %v\n", err)
		return
	}
	defer manager.Disconnect()
	serviceHandle, err := manager.OpenService(serviceName)
	if err == nil {
		serviceHandle.Close()
		log.Errorf("service %s already exists", serviceName)
		return
	}
	conf := mgr.Config{
		DisplayName: serviceName,
		StartType:   mgr.StartAutomatic,
	}
	user, userIndex := findArg(args, "user")
	if userIndex >= 0 {
		conf.ServiceStartName = user
		args = removeArg(args, userIndex)
	}
	pswd, pswdIndex := findArg(args, "password")
	if pswdIndex >= 0 {
		conf.Password = pswd
		args = removeArg(args, pswdIndex)
	}

	serviceHandle, err = manager.CreateService(serviceName, exepath, conf, args...)
	if err != nil {
		log.Errorf("Error creating service %s:  %v", serviceName, err)
		return
	}
	defer serviceHandle.Close()
	err = eventlog.InstallAsEventCreate(serviceName, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		serviceHandle.Delete()
		log.Errorf("Setup event log source failed: %v", err)
	}
}

func Remove(serviceName string, args []string) {
	manager, err := mgr.Connect()
	if err != nil {
		log.Errorf("Error connecting to service manager: %v\n", err)
		return
	}
	defer manager.Disconnect()
	serviceHandle, err := manager.OpenService(serviceName)
	if err != nil {
		log.Errorf("Service %s is not installed", serviceName)
		return
	}
	defer serviceHandle.Close()
	err = serviceHandle.Delete()
	if err != nil {
		log.Errorf("Error deleting service %s: %v", serviceName, err)
		return
	}
	err = eventlog.Remove(serviceName)
	if err != nil {
		log.Errorf("Remove event log source failed: %v", err)
	}
}

func Start(serviceName string, args []string) {
	manager, err := mgr.Connect()
	if err != nil {
		log.Errorf("Error connecting to service manager: %v\n", err)
		return
	}
	defer manager.Disconnect()
	serviceHandle, err := manager.OpenService(serviceName)
	if err != nil {
		log.Errorf("could not access service %s: %v", serviceName, err)
		return
	}
	defer serviceHandle.Close()
	err = serviceHandle.Start("is", "manual-started")
	if err != nil {
		log.Errorf("Could not start service %s: %v", serviceName, err)
	}
}

func Stop(serviceName string, args []string) {
	control(serviceName, args, StopCmd, Stopped)
}

func Pause(serviceName string, args []string) {
	control(serviceName, args, PauseCmd, Paused)
}

func Continue(serviceName string, args []string) {
	control(serviceName, args, ContinueCmd, Running)
}

func control(serviceName string, _ []string, cmd Command, state State) {
	manager, err := mgr.Connect()
	if err != nil {
		log.Errorf("Error connecting to service manager: %v\n", err)
		return
	}
	defer manager.Disconnect()
	serviceHandle, err := manager.OpenService(serviceName)
	if err != nil {
		log.Errorf("could not access service: %v", err)
		return
	}
	defer serviceHandle.Close()
	status, err := serviceHandle.Control(windowsServiceCommand(cmd))
	if err != nil {
		log.Errorf("Could not send control=%d to service: %v", cmd, err)
		return
	}
	timeout := time.Now().Add(10 * time.Second)
	wsState := windowsServiceState(state)
	for status.State != wsState {
		if timeout.Before(time.Now()) {
			log.Errorf("Timeout waiting for service to go to state=%d", state)
			return
		}
		time.Sleep(300 * time.Millisecond)
		status, err = serviceHandle.Query()
		if err != nil {
			log.Errorf("Could not retrieve service status: %v", err)
			return
		}
	}
}

func IsDaemon() (bool, error) {
	return svc.IsWindowsService()
}

func windowsServiceCommand(cmd Command) svc.Cmd {
	switch cmd {
	case StopCmd:
		return svc.Stop
	case PauseCmd:
		return svc.Pause
	case ContinueCmd:
		return svc.Continue
	case InterrogateCmd:
		return svc.Interrogate
	case ShutdownCmd:
		return svc.Shutdown
	default:
		return svc.Stop
	}
}

func windowsServiceState(state State) svc.State {
	switch state {
	case Stopped:
		return svc.Stopped
	case StartPending:
		return svc.StartPending
	case StopPending:
		return svc.StopPending
	case Running:
		return svc.Running
	case ContinuePending:
		return svc.ContinuePending
	case PausePending:
		return svc.PausePending
	case Paused:
		return svc.Paused
	default:
		return svc.Stopped
	}
}
