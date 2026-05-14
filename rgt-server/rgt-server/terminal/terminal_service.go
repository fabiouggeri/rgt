package terminal

import (
	"fmt"
	"net"
	"os"
	"rgt-server/auth"
	"rgt-server/log"
	"rgt-server/option"
	"rgt-server/protocol"
	"rgt-server/service"
	"rgt-server/util"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

type TerminalServiceConfig interface {
	MaxConcurrentLaunchingApps() option.TypedOption[uint32]
	AppEmulationPort() option.TypedOption[uint16]
	EmulationPort() option.TypedOption[uint16]
	Address() option.TypedOption[string]
	ShowConsole() option.TypedOption[bool]
	OrphanProcessCheckInterval() option.TypedOption[time.Duration]
	SessionsCheckInterval() option.TypedOption[time.Duration]
	HealthMaxLoginTime() option.TypedOption[time.Duration]
	HealthMaxLoginsTimeoutAlerts() option.TypedOption[uint16]
	HealthLoginsTimeoutIncAlert() option.TypedOption[uint16]
	SessionIdleTimeout() option.TypedOption[time.Duration]
	AppLackTimeout() option.TypedOption[time.Duration]
	AppTransactionTimeout() option.TypedOption[time.Duration]
	AppLaunchTimeout() option.TypedOption[time.Duration]
	AppLoginTimeout() option.TypedOption[time.Duration]
	TeLogLevel() option.TypedOption[log.LogLevel]
	TeLogPathName() option.TypedOption[string]
	AppLogLevel() option.TypedOption[log.LogLevel]
	AppLogPathName() option.TypedOption[string]
	StandaloneEnabled() option.TypedOption[bool]
	AppMinLaunchIntervalStandalone() option.TypedOption[time.Duration]
	EnvVars() []string
	TerminalTCPWriteBufferSize() option.TypedOption[uint32]
	TerminalTCPReadBufferSize() option.TypedOption[uint32]
	TeAuthConf() map[string]option.Option
	StandaloneAuthConf() map[string]option.Option
}

type TerminalEmulationService struct {
	name                    string
	teListener              atomic.Pointer[net.TCPListener]
	appListener             atomic.Pointer[net.TCPListener]
	currHandlerId           atomic.Uint64
	sessionManager          *SessionManager
	config                  TerminalServiceConfig
	status                  atomic.Value // stores service.ServiceStatus
	waitGroup               *sync.WaitGroup
	emulationAuthenticator  auth.UserAuthenticator
	standaloneAuthenticator auth.UserAuthenticator
	orphanProcessTimer      *time.Ticker
	appListeningPort        uint16
}

const (
	EMULATION_SERVICE_ID string = "emulation"
	STANDALONE_CONFIG_ID string = "standalone"
)

var (
	_ service.Service = &TerminalEmulationService{}

	protocols map[protocol.OperationCode]map[int]any = make(map[protocol.OperationCode]map[int]any)
)

func NewService(config TerminalServiceConfig) *TerminalEmulationService {
	s := &TerminalEmulationService{
		name:                    EMULATION_SERVICE_ID,
		config:                  config,
		emulationAuthenticator:  auth.NewAuthenticator(config.TeAuthConf()),
		standaloneAuthenticator: auth.NewAuthenticator(config.StandaloneAuthConf()),
	}
	s.sessionManager = NewSessionManager(s, s.config)
	s.status.Store(service.STOPPED)
	configureLaunchAppSemaphore(s.config.MaxConcurrentLaunchingApps().Get())
	s.config.MaxConcurrentLaunchingApps().SetHook(configureLaunchAppSemaphore)
	return s
}

func (s *TerminalEmulationService) Name() string {
	return s.name
}

func (s *TerminalEmulationService) Config() TerminalServiceConfig {
	return s.config
}

func (s *TerminalEmulationService) Start(wait *sync.WaitGroup) error {
	if s.GetStatus() == service.STOPPED {
		s.setStatus(service.STARTING)
		log.Infof("Starting service %s...", s.name)
		appPort := s.config.AppEmulationPort().Get()
		tePort := s.config.EmulationPort().Get()
		address := s.config.Address().Get()
		if appPort == tePort {
			return fmt.Errorf("app port (%d) and emulation port (%d) must be different", appPort, tePort)
		}
		appListener, err := s.createListener("APP", address, appPort)
		if err != nil {
			return err
		}
		s.appListener.Store(appListener)
		s.appListeningPort = listenerPort(appPort, appListener)
		teListener, err := s.createListener("TE", address, tePort)
		if err != nil {
			return err
		}
		s.teListener.Store(teListener)
		s.setStatus(service.STARTED)
		wait.Add(1)
		go s.listenAppConnections(appListener)
		wait.Add(1)
		go s.listenTeConnections(teListener)
		s.waitGroup = wait
		s.sessionManager.StartSessionsMonitorJob()
		s.startProcessMonitorJob()
		log.Infof("Service %s started.", s.name)
	} else {
		log.Warnf("Service %s already running", s.name)
	}
	return nil
}

func (s *TerminalEmulationService) Stop() error {
	if s.GetStatus() == service.STARTED {
		s.setStatus(service.STOPPING)
		log.Infof("Stopping service %s...", s.name)
		s.closeListener("TE", s.teListener.Swap(nil))
		s.closeListener("APP", s.appListener.Swap(nil))
		s.sessionManager.StopSessionsMonitorJob(true)
		s.stopProcessMonitorJob()
		s.setStatus(service.STOPPED)
		log.Infof("Service %s stopped.", s.name)
	} else {
		log.Warnf("Service %s is not running", s.name)
	}
	return nil
}

func (s *TerminalEmulationService) AppListeningPort() uint16 {
	return s.appListeningPort
}

func (s *TerminalEmulationService) GetStatus() service.ServiceStatus {
	return s.status.Load().(service.ServiceStatus)
}

func (s *TerminalEmulationService) setStatus(status service.ServiceStatus) {
	s.status.Store(status)
}

func (s *TerminalEmulationService) changeStatus(oldStatus service.ServiceStatus, newStatus service.ServiceStatus) bool {
	return s.status.CompareAndSwap(oldStatus, newStatus)
}

func (s *TerminalEmulationService) createListener(name string, address string, port uint16) (*net.TCPListener, error) {
	var err error
	var proto string
	if address == "" {
		proto = "tcp4"
	} else {
		proto = "tcp"
	}
	addr, err := net.ResolveTCPAddr(proto, address+":"+strconv.Itoa(int(port)))
	if err != nil {
		if proto == "tcp4" {
			proto = "tcp"
			addr, err = net.ResolveTCPAddr("tcp", address+":"+strconv.Itoa(int(port)))
			if err != nil {
				return nil, fmt.Errorf("error starting %s listener of %s service: %v", name, s.name, err)
			}
		} else {
			return nil, fmt.Errorf("error starting %s listener of %s service: %v", name, s.name, err)
		}
	}

	listener, err := net.ListenTCP(proto, addr)
	if err != nil {
		return nil, fmt.Errorf("error starting %s listener of %s service: %v", name, s.name, err)
	}
	log.Debugf("TerminalEmulationService.createListener(). listener %s created for service %s.", name, s.name)
	return listener, nil
}

func (s *TerminalEmulationService) closeListener(name string, listener *net.TCPListener) {
	if listener != nil {
		listener.Close()
		log.Debugf("TerminalEmulationService.closeListener(). Listener %s closed for service %s.", name, s.name)
	}
}

func (s *TerminalEmulationService) listenTeConnections(listener *net.TCPListener) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("unknown error in service(TerminalEmulationService.listenTeConnections): %v\n%s", err, util.FullStack())
		}
	}()
	defer s.waitGroup.Done()

	log.Infof("Listening for TE connections on %s.", s.teListener.Load().Addr().String())
	for s.GetStatus() == service.STARTED {
		c, err := listener.AcceptTCP()
		if err != nil {
			if s.GetStatus() == service.PAUSED {
				log.Infof("TerminalEmulationService. Service %s paused and not accepting new TE connections.", s.name)
			} else if s.GetStatus() == service.STARTED {
				log.Error("error listening TE client connections: ", err)
			}
			break
		}
		handlerId := s.currHandlerId.Add(1)
		h := newHandler(service.TERMINAL, handlerId, c, s)
		go h.Handle()
	}
}

func (s *TerminalEmulationService) listenAppConnections(listener *net.TCPListener) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("unknown error in service(TerminalEmulationService.listenAppConnections): %v\n%s", err, util.FullStack())
		}
	}()
	defer s.waitGroup.Done()
	log.Infof("Listening for app connections on %s.", s.appListener.Load().Addr().String())
	status := s.GetStatus()
	for status == service.STARTED || status == service.PAUSED {
		c, err := listener.AcceptTCP()
		if err != nil {
			status = s.GetStatus()
			if status == service.STARTED || status == service.PAUSED {
				log.Error("error listening APP client connections: ", err)
			}
			break
		}
		handlerId := s.currHandlerId.Add(1)
		h := newHandler(service.APPLICATION, handlerId, c, s)
		go h.Handle()
		status = s.GetStatus()
	}
}

func (s *TerminalEmulationService) GetType() service.ServiceType {
	return service.SERVICE_EMULATION
}

func (s *TerminalEmulationService) Pause() {
	if s.changeStatus(service.STARTED, service.PAUSED) {
		log.Infof("TerminalEmulationService.Pause(). Pausing service %s.", s.name)
		s.closeListener("TE", s.teListener.Swap(nil))
	}
}

func (s *TerminalEmulationService) Resume() {
	if s.changeStatus(service.PAUSED, service.STARTED) {
		log.Infof("TerminalEmulationService.Resume(). Resuming service %s.", s.name)
		teListener, err := s.createListener("TE", s.config.Address().Get(), s.config.EmulationPort().Get())
		if err != nil {
			log.Errorf("TerminalEmulationService.ResumeAccepting(). Error recreating listener: %v", err)
			s.setStatus(service.PAUSED)
			return
		}
		s.teListener.Store(teListener)
		s.waitGroup.Add(1)
		go s.listenTeConnections(teListener)
	}
}

func (s *TerminalEmulationService) IsAccepting() bool {
	return s.GetStatus() == service.STARTED && s.teListener.Load() != nil
}

func (s *TerminalEmulationService) AuthenticateEmulation(username, password string) bool {
	return s.emulationAuthenticator.Authenticate(username, password)
}

func (s *TerminalEmulationService) AuthenticateStandalone(username, password string) bool {
	return s.standaloneAuthenticator.Authenticate(username, password)
}

func (s *TerminalEmulationService) SessionManager() *SessionManager {
	return s.sessionManager
}

func (s *TerminalEmulationService) GetSessionsCount() int32 {
	return s.sessionManager.GetSessionsCount()
}

func (s *TerminalEmulationService) handlePanic(message string) {
	if err := recover(); err != nil {
		log.Errorf("%s: %v\n%s", message, err, util.FullStack())
	}
}

func (s *TerminalEmulationService) processMonitorJob(currentProcess *process.Process) {
	log.Debugf("TerminalEmulationService.orphanProcessMonitor(). started.")
	defer s.handlePanic("unknown error in service(terminal.TerminalEmulationService)")
	for range s.orphanProcessTimer.C {
		s.checkAndKillOrphanProcesses(currentProcess)
	}
	log.Debugf("TerminalEmulationService.orphanProcessMonitor(). stopped.")
}

func (s *TerminalEmulationService) checkAndKillOrphanProcesses(p *process.Process) {
	sessions := s.sessionManager.GetSessions()
	processList, err := p.Children()
	if err == nil {
		killOrphanProcesses(sessions, processList)
	} else {
		log.Debugf("TerminalEmulationService.orphanProcessMonitor(). Error getting child process: %v", err)
	}
}

func isOrphanProcess(sessions []*TerminalSession, proc *process.Process) bool {
	value, err := util.ProcessEnvVar(proc, ENV_VAR_AUTH_TOKEN)
	if err != nil {
		cmd, _ := proc.Cmdline()
		log.Debugf("terminal.orphanProcess(). Error getting process environment variables. PID: %d Cmd: '%s' Error: %v", proc.Pid, cmd, err)
		return true
	}
	sessionId, _ := strconv.ParseInt(value, 10, 64)
	return slices.IndexFunc(sessions, func(session *TerminalSession) bool { return session.Id() == sessionId }) < 0
}

func killOrphanProcesses(sessions []*TerminalSession, processes []*process.Process) {
	for _, proc := range processes {
		if isOrphanProcess(sessions, proc) {
			util.KillProcessRecursive(proc, "orphan process")
			log.Debugf("terminal.killOrphanProcesses(). proc=%v", proc)
		}
	}
}

func (s *TerminalEmulationService) startProcessMonitorJob() {
	if s.orphanProcessTimer == nil && s.config.OrphanProcessCheckInterval().Get() > 0 {
		currentProcess, err := process.NewProcess(int32(os.Getpid()))
		if err != nil {
			log.Errorf("TerminalEmulationService.startProcessMonitorJob(). Error getting own process: %v", err)
			return
		}
		interval := s.config.OrphanProcessCheckInterval().Get()
		s.orphanProcessTimer = time.NewTicker(interval)
		go s.processMonitorJob(currentProcess)
		log.Infof("Process monitor started. Interval: %v", interval)
	}
}

func (s *TerminalEmulationService) stopProcessMonitorJob() {
	if s.orphanProcessTimer != nil {
		m := s.orphanProcessTimer
		s.orphanProcessTimer = nil
		m.Stop()
		log.Info("Process monitor stopped.")
	}
}

func listenerPort(port uint16, listener *net.TCPListener) uint16 {
	if port > 0 {
		return port
	}
	if addr, ok := listener.Addr().(*net.TCPAddr); ok {
		return uint16(addr.Port)
	}
	return 0
}
