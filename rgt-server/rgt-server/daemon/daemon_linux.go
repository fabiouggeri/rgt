package daemon

import (
	"rgt-server/log"

	svc "github.com/takama/daemon"
)

type Service struct {
	svc.Daemon
}

func Run(service Daemon) {
	service.Start([]string{})
}

func Install(serviceName string, args []string) {
	srv, err := svc.New(serviceName, "RGT - Remote Graphical Terminal Server", svc.SystemDaemon)
	if err != nil {
		log.Error("Error: ", err)
	}
	status, err := srv.Install()
	if err != nil {
		log.Error("status:", status, "\nerror: ", err)
	}
}

func Remove(serviceName string, args []string) {
	srv, err := svc.New(serviceName, "RGT - Remote Graphical Terminal Server", svc.SystemDaemon)
	if err != nil {
		log.Error("Error: ", err)
	}
	status, err := srv.Remove()
	if err != nil {
		log.Error("status:", status, "\nerror: ", err)
	}
}

func Start(serviceName string, args []string) {
	srv, err := svc.New(serviceName, "RGT - Remote Graphical Terminal Server", svc.SystemDaemon)
	if err != nil {
		log.Error("Error: ", err)
	}
	status, err := srv.Start()
	if err != nil {
		log.Error("status:", status, "\nerror: ", err)
	}
}

func Stop(serviceName string, args []string) {
	srv, err := svc.New(serviceName, "RGT - Remote Graphical Terminal Server", svc.SystemDaemon)
	if err != nil {
		log.Error("Error: ", err)
	}
	status, err := srv.Stop()
	if err != nil {
		log.Error("status:", status, "\nerror: ", err)
	}
}

func Pause(serviceName string, args []string) {
	log.Error("Unsupported command: pause")
}

func Continue(serviceName string, args []string) {
	log.Error("Unsupported command: continue")
}

func IsDaemon() (bool, error) {
	return false, nil
}
