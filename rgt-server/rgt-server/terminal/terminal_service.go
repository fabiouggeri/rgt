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
	HealthMaxLoginsTimeout() option.TypedOption[uint16]
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
	monitorSessionsTimer    *time.Ticker
	loginTimeoutCheckCount  uint16
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
		sessionManager:          NewSessionManager(),
		emulationAuthenticator:  auth.NewAuthenticator(config.TeAuthConf()),
		standaloneAuthenticator: auth.NewAuthenticator(config.StandaloneAuthConf()),
	}
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
			teListener, err := s.createListener("TE/APP", address, tePort)
			if err != nil {
				return err
			}
			s.teListener.Store(teListener)
			s.setStatus(service.STARTED)
			wait.Add(1)
			go s.listenConnections("TE/APP", teListener)
		} else {
			appListener, err := s.createListener("APP", address, appPort)
			if err != nil {
				return err
			}
			s.appListener.Store(appListener)
			teListener, err := s.createListener("TE", address, tePort)
			if err != nil {
				return err
			}
			s.teListener.Store(teListener)
			s.setStatus(service.STARTED)
			wait.Add(1)
			go s.listenConnections("APP", appListener)
			wait.Add(1)
			go s.listenConnections("TE", teListener)
		}
		s.waitGroup = wait
		s.StartSessionsMonitorJob()
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
		if s.appListener.Load() == nil {
			s.closeListener("TE/APP", s.teListener.Swap(nil))
		} else {
			s.closeListener("TE", s.teListener.Swap(nil))
			s.closeListener("APP", s.appListener.Swap(nil))
		}
		s.StopSessionsMonitorJob()
		s.sessionManager.KillAllSessions("service stopped")
		s.stopProcessMonitorJob()
		s.setStatus(service.STOPPED)
		log.Infof("Service %s stopped.", s.name)
	} else {
		log.Warnf("Service %s is not running", s.name)
	}
	return nil
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

func (s *TerminalEmulationService) listenConnections(name string, listener *net.TCPListener) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("unknown error in service(TerminalEmulationService.listenConnections): %v\n%s", err, util.FullStack())
		}
	}()
	defer s.waitGroup.Done()

	for s.GetStatus() == service.STARTED {
		c, err := listener.AcceptTCP()
		if err != nil {
			if s.GetStatus() == service.PAUSED {
				log.Infof("TerminalEmulationService. Listener %s paused for service %s.", name, s.name)
			} else if s.GetStatus() == service.STARTED {
				log.Error("error listening client connections: ", err)
			}
			break
		}
		handlerId := s.currHandlerId.Add(1)
		h := newHandler(handlerId, c, s)
		go h.Handle()
	}
}

func (s *TerminalEmulationService) GetType() service.ServiceType {
	return service.SERVICE_EMULATION
}

func (s *TerminalEmulationService) Pause() {
	if s.changeStatus(service.STARTED, service.PAUSED) {
		var listenerLabel string
		log.Infof("TerminalEmulationService.Pause(). Pausing service %s.", s.name)
		if s.appListener.Load() == nil {
			listenerLabel = "TE/APP"
		} else {
			listenerLabel = "TE"
		}
		s.closeListener(listenerLabel, s.teListener.Swap(nil))
	}
}

func (s *TerminalEmulationService) Resume() {
	if s.changeStatus(service.PAUSED, service.STARTED) {
		log.Infof("TerminalEmulationService.Resume(). Resuming service %s.", s.name)
		var listenerLabel string
		if s.appListener.Load() == nil {
			listenerLabel = "TE/APP"
		} else {
			listenerLabel = "TE"
		}
		teListener, err := s.createListener(listenerLabel, s.config.Address().Get(), s.config.EmulationPort().Get())
		if err != nil {
			log.Errorf("TerminalEmulationService.ResumeAccepting(). Error recreating listener: %v", err)
			s.setStatus(service.PAUSED)
			return
		}
		s.teListener.Store(teListener)
		s.waitGroup.Add(1)
		go s.listenConnections(listenerLabel, teListener)
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

func (s *TerminalEmulationService) StartSessionsMonitorJob() {
	if s.monitorSessionsTimer == nil && s.config.SessionsCheckInterval().Get() > 0 {
		interval := s.config.SessionsCheckInterval().Get()
		if interval <= 10*time.Second {
			interval = 10 * time.Second
		}
		s.monitorSessionsTimer = time.NewTicker(interval)
		go s.sessionsMonitorJob()
		log.Infof("Session monitor started. Interval: %v", interval)
	}
}

func (s *TerminalEmulationService) StopSessionsMonitorJob() {
	if s.monitorSessionsTimer != nil {
		m := s.monitorSessionsTimer
		s.monitorSessionsTimer = nil
		m.Stop()
		log.Info("Session monitor stopped.")
	}
}

func (s *TerminalEmulationService) sessionsMonitorJob() {
	log.Debugf("TerminalEmulationService.sessionsMonitor(). started.")
	defer s.handlePanic("unknown error in service (TerminalEmulationService.sessionsMonitor)")
	for range s.monitorSessionsTimer.C {
		loginExceededCount := uint16(0)
		maxLoginTime := s.config.HealthMaxLoginTime().Get()
		maxPendingLoginsAlerts := s.config.HealthMaxLoginsTimeoutAlerts().Get()
		sessions := s.sessionManager.GetSessions()
		for _, session := range sessions {
			if session.CloseConditionally(s.config) {
				s.sessionManager.DeleteSession(session.Id())
			} else if maxPendingLoginsAlerts > 0 && s.loginTimeExceeded(session, maxLoginTime) {
				loginExceededCount++
			}
		}
		s.checkPendingLoginsExceededCount(loginExceededCount, maxPendingLoginsAlerts)
	}
	log.Debugf("TerminalEmulationService.sessionsMonitor(). stopped.")
}

func (s *TerminalEmulationService) checkPendingLoginsExceededCount(loginExceededCount uint16, maxPendingLoginsAlerts uint16) {
	if loginExceededCount > s.config.HealthMaxLoginsTimeout().Get() {
		s.loginTimeoutCheckCount++
		if s.loginTimeoutCheckCount > maxPendingLoginsAlerts && s.GetStatus() == service.STARTED {
			log.Infof("TerminalEmulationService.sessionsMonitorJob(). %d exceeded pending logins. Pausing service.", loginExceededCount)
			s.Pause()
		}
	} else {
		s.loginTimeoutCheckCount = 0
		if s.GetStatus() == service.PAUSED {
			log.Infof("TerminalEmulationService.sessionsMonitorJob(). %d exceeded pending logins. Resuming service.", loginExceededCount)
			s.Resume()
		}
	}
}

func (s *TerminalEmulationService) loginTimeExceeded(session *TerminalSession, maxLoginTime time.Duration) bool {
	if session.GetStatus() >= SESS_READY {
		return false
	}
	if maxLoginTime <= 0 {
		return false
	}
	if timeout := time.Since(session.StartTime); timeout >= maxLoginTime {
		log.Debugf("TerminalEmulationService.loginTimeExceeded(). Session %d exceeded login time %v", session.Id, timeout)
		return true
	}
	return false
}
