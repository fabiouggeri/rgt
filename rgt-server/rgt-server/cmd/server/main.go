package main

import (
	"bufio"
	"fmt"
	"net/http"

	_ "net/http/pprof"
	"os"
	"path/filepath"
	"rgt-server/admin"
	"rgt-server/auth"
	"rgt-server/config"
	"rgt-server/daemon"
	"rgt-server/log"
	"rgt-server/server"
	"rgt-server/service"
	"rgt-server/terminal"
	"rgt-server/util"
	"runtime/debug"
	"strings"
)

type RgtService struct {
	name   string
	server *server.Server
	logger log.Logger
}

const (
	ACT_RUN                uint8  = 0x0
	ACT_INSTALL            uint8  = 0x1
	ACT_REMOVE             uint8  = 0x2
	ACT_START              uint8  = 0x3
	ACT_STOP               uint8  = 0x4
	ACT_PAUSE              uint8  = 0x5
	ACT_CONTINUE           uint8  = 0x6
	ACT_UNKNOWN            uint8  = 0xFF
	DEFAULT_EMULATION_PORT uint16 = 7654
	DEFAULT_ADMIN_PORT     uint16 = 7656
)

var (
	serverRunning bool = false
	Version            = "dev"
)

func commandLineOptions(conf *config.ServerConfig, args map[string]string) error {
	for key, value := range args {
		if !conf.Set(key, value) {
			return fmt.Errorf("unknown option: %s", key)
		}
	}
	return nil
}

func changeToExeDir() {
	exePathName, err := os.Executable()
	if err != nil {
		log.Error(err)
		return
	}
	os.Chdir(filepath.Dir(exePathName))
}

func findArg(args map[string]string, names []string, defaultValue string) string {
	for _, name := range names {
		val, foundArg := args[name]
		if foundArg {
			return val
		}
	}
	return defaultValue
}

func serviceName(args []string) string {
	conf := config.NewConfig()
	argsMap := util.ArgsToMap(args)
	middle := findArg(argsMap, conf.Address().Names(), "default")
	sufix := findArg(argsMap, conf.EmulationPort().Names(), "")
	if sufix != "" {
		return "RGT-SERVER-" + middle + "-" + sufix
	}
	return "RGT-SERVER-" + middle
}

func (srv *RgtService) GetName() string {
	return srv.name
}

func (srv *RgtService) Start(args []string) {
	log.Info("Starting server...")
	err := srv.server.Start(service.SERVICE_ALL)
	if err != nil {
		srv.server.Stop(service.SERVICE_ALL)
		srv.server.AwaitServices()
		log.Errorf("Error starting server: %s", err)
	} else {
		log.Info("Server started.")
	}
}

func (srv *RgtService) Stop() {
	log.Info("Stopping server...")
	err := srv.server.Stop(service.SERVICE_ALL)
	if err != nil {
		log.Errorf("Error stopping server: %s", err)
		return
	}
	srv.server.AwaitServices()
	log.Info("Server stopped.")
}

func createDaemon(args []string) daemon.Daemon {
	server, logger := createServer(args)
	if server == nil {
		return nil
	}
	daemon := &RgtService{
		name:   serviceName(args),
		server: server,
		logger: logger,
	}
	return daemon
}

func activeProfile(conf *config.ServerConfig) {
	if conf.ProfilePort().Get() > 0 {
		go func() {
			log.Infof("Activing profile on port %d...", conf.ProfilePort().Get())
			if err := http.ListenAndServe("localhost:"+conf.ProfilePort().GetString(), nil); err != nil {
				log.Errorf("Error activating profile: %s", err)
			}
		}()
	}
}

func createServer(args []string) (*server.Server, log.Logger) {
	var err error
	var conf *config.ServerConfig
	logger := createLoggerFile("rgt-server.log", log.INFO)
	if logger != nil {
		log.SetDefaultLogger(log.NewCompositeLogger(log.GetDefaultLogger(), logger))
	}
	log.Infof("Server version: %s", Version)
	log.Info("Creating server...")
	argsMap := util.ArgsToMap(args)
	value, found := argsMap["config"]
	if found {
		delete(argsMap, "config")
		conf, err = config.LoadConfigFromFile(value)
	} else {
		conf, err = config.LoadConfig()
	}
	if err != nil {
		log.Error(err)
		return nil, logger
	}
	if logger != nil {
		configLoggerFile(logger, conf)
	}
	err = commandLineOptions(conf, argsMap)
	if err != nil {
		log.Error(err)
		return nil, logger
	}
	activeProfile(conf)
	server := server.New(conf, Version)
	log.Info("Registering services...")
	server.AddService(terminal.NewService(config.EMULATION_SERVICE_ID, server))
	server.AddService(admin.NewService(config.ADMIN_SERVICE_ID, server))
	log.Info("Registering authenticators...")
	server.AddAuthenticator(config.EMULATION_SERVICE_ID, auth.NewAuthenticator(config.TERMINAL_AUTH_PREFIX, conf.TeAuthConf()))
	server.AddAuthenticator(config.ADMIN_SERVICE_ID, auth.NewAuthenticator(config.ADMIN_AUTH_PREFIX, conf.AdminAuthConf()))
	server.AddAuthenticator(config.STANDALONE_CONFIG_ID, auth.NewAuthenticator(config.STANDALONE_AUTH_PREFIX, conf.StandaloneAuthConf()))
	log.Info("Server created.")
	return server, logger
}

func runServer(args []string) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("unknown error in server(runServer): %v\n%s", err, util.FullStack())
		}
	}()
	server, logger := createServer(args)
	if server == nil {
		log.Error("Server not created!")
		return
	}
	log.Info("Starting server...")
	err := server.Start(service.SERVICE_ALL)
	if err != nil {
		server.Stop(service.SERVICE_ALL)
		server.AwaitServices()
		log.Infof("Error starting server: %s", err)
		return
	}
	serverRunning = true
	log.Info("Server started.")
	log.SetDefaultLogger(logger)
	userInteraction(server)
	server.AwaitServices()
	serverRunning = false
	server.Finalize()
	log.Info("Server stopped.")
	closeLoggerFile(logger)
}

func action() (uint8, error) {
	if len(os.Args) < 2 {
		return ACT_RUN, nil
	}
	act := strings.ToLower(strings.TrimSpace(os.Args[1]))
	switch act {
	case "install":
		return ACT_INSTALL, nil
	case "remove":
		return ACT_REMOVE, nil
	case "start":
		return ACT_START, nil
	case "stop":
		return ACT_STOP, nil
	case "pause":
		return ACT_PAUSE, nil
	case "continue":
		return ACT_CONTINUE, nil
	default:
		if !strings.ContainsRune(act, '=') {
			return ACT_UNKNOWN, fmt.Errorf("unknown action: %s", act)
		}
		return ACT_RUN, nil
	}
}

func userInteraction(server *server.Server) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Type quit + <ENTER> to shut down server!")
	for serverRunning {
		fmt.Print("> ")
		scanner.Scan()
		cmd := strings.ToLower(scanner.Text())
		switch cmd {
		case "version":
			fmt.Println("RGT version: ", Version)
		case "quit":
			log.SetDefaultLogger(log.NewCompositeLogger(log.NewLogger(log.INFO, os.Stdout), log.GetDefaultLogger()))
			server.Stop(service.SERVICE_ALL)
			serverRunning = false
		default:
			fmt.Println("Unknown command:", cmd)
		}
	}
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Error("stacktrace from panic: \n", string(debug.Stack()))
			fmt.Print("stacktrace from panic: \n", string(debug.Stack()))
		}
	}()
	act, err := action()
	if err != nil {
		fmt.Println(err)
		return
	}
	changeToExeDir()
	switch act {
	case ACT_RUN:
		inService, _ := daemon.IsDaemon()
		if inService {
			d := createDaemon(os.Args[1:])
			if d != nil {
				daemon.Run(d)
			}
		} else {
			runServer(os.Args[1:])
		}
	case ACT_INSTALL:
		daemon.Install(serviceName(os.Args[2:]), os.Args[2:])
	case ACT_REMOVE:
		daemon.Remove(serviceName(os.Args[2:]), os.Args[2:])
	case ACT_START:
		daemon.Start(serviceName(os.Args[2:]), os.Args[2:])
	case ACT_STOP:
		daemon.Stop(serviceName(os.Args[2:]), os.Args[2:])
	case ACT_PAUSE:
		daemon.Pause(serviceName(os.Args[2:]), os.Args[2:])
	case ACT_CONTINUE:
		daemon.Continue(serviceName(os.Args[2:]), os.Args[2:])
	}
}
