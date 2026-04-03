package server

import (
	"os"
	"path/filepath"
	"regexp"
	"rgt-server/auth"
	"rgt-server/config"
	"rgt-server/health"
	"rgt-server/log"
	"rgt-server/service"
	"rgt-server/stats"
	"rgt-server/util"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

type ServerStatus string

const (
	SERVER_STOPPED            ServerStatus = "STOPPED"
	SERVER_STARTING           ServerStatus = "STARTING"
	SERVER_RUNNING            ServerStatus = "RUNNING"
	SERVER_STOPPING           ServerStatus = "STOPPING"
	SERVER_PAUSED             ServerStatus = "PAUSED"
	SERVER_DISCONNECTED       ServerStatus = "DISCONNECTED"
	SERVER_CONNECTING         ServerStatus = "CONNECTING"
	SERVER_DISCONNECTING      ServerStatus = "DISCONNECTING"
	ENV_VAR_AUTH_TOKEN        string       = "RGT_AUTH_TOKEN"
	ENV_VAR_SERVER_ADDR       string       = "RGT_SERVER_ADDR"
	ENV_VAR_SERVER_PORT       string       = "RGT_SERVER_PORT"
	ENV_VAR_STANDALONE_APP    string       = "RGT_STANDALONE_APP"
	SERVER_LOG_ID             string       = "server"
	AUTH_TOKEN_VAR_PREFIX     string       = "RGT_AUTH_TOKEN="
	AUTH_TOKEN_VAR_PREFIX_LEN int          = len(AUTH_TOKEN_VAR_PREFIX)
)

type Server struct {
	sessions                  map[int64]*Session
	sessionsLostConnection    map[int64]*Session
	services                  map[string]service.Service
	authenticators            map[string]auth.UserAuthenticator
	config                    *config.ServerConfig
	serverProcess             *process.Process
	waitGroup                 sync.WaitGroup
	sessionsLock              sync.RWMutex
	lostSessionsLock          sync.RWMutex
	startTime                 time.Time
	version                   string
	userRepository            UserRepository
	logFile                   *os.File
	monitorSessionsTimer      *time.Ticker
	orphanProcessTimer        *time.Ticker
	removeAppLogsTimer        *time.Ticker
	lastAppLogRemoveExecution time.Time
	status                    atomic.Value // stores ServerStatus
	stats                     *stats.Stats
	healthChecker             *health.HealthChecker
}

func New(config *config.ServerConfig, version string) *Server {
	var err error
	server := &Server{config: config,
		sessions:               make(map[int64]*Session),
		sessionsLostConnection: make(map[int64]*Session),
		services:               make(map[string]service.Service),
		authenticators:         make(map[string]auth.UserAuthenticator),
		version:                version,
		stats:                  stats.New(),
	}
	server.status.Store(SERVER_STOPPED)
	server.initLog()
	server.serverProcess, err = process.NewProcess(int32(os.Getpid()))
	if err != nil {
		log.Errorf("Error getting server process: %v", err)
	}
	return server
}

func (s *Server) Version() string {
	return s.version
}

func (s *Server) Finalize() {
	s.setStatus(SERVER_STOPPED)
	log.Debugf("Server.Finalize().")
	s.closeLog()
}

func (s *Server) initLog() {
	if s.logFile != nil {
		return
	}
	util.TruncateFile(s.config.ServerLogPathName().Get(), 15*1024*1024, 10*1024*1024)
	logPathname := util.RelativePathToAbsolute(s.config.ServerLogPathName().Get())
	file, err := os.OpenFile(logPathname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Error("Error creating log file:", logPathname, "Error:", err)
		return
	}
	s.logFile = file
	s.registerLogger(file)
	s.config.ServerLogLevel().SetHook(func(val log.LogLevel) { log.SetLevel(SERVER_LOG_ID, val) })
	s.config.ServerLogPathName().SetHook(s.setLogPathName)
}

func (s *Server) registerLogger(file *os.File) {
	log.AddLogger(SERVER_LOG_ID, s.config.ServerLogLevel().Get(), file)
	log.SetFormatterf(SERVER_LOG_ID, log.TimestampLogPrintf)
	log.SetFormatter(SERVER_LOG_ID, log.TimestampLogPrintln)
}

func (s *Server) setLogPathName(newLogPathname string) {
	newLogInfo, errNew := os.Stat(newLogPathname)
	if errNew != nil {
		log.Errorf("invalid path name for log: %s", newLogPathname)
		return
	}
	currentLogInfo, _ := s.logFile.Stat()
	if currentLogInfo == nil || !os.SameFile(currentLogInfo, newLogInfo) {
		file, err := os.OpenFile(newLogPathname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Error("Error creating log file:", newLogPathname, "Error:", err)
			return
		}
		s.logFile.Close()
		log.RemoveLogger(SERVER_LOG_ID)
		s.registerLogger(file)
	}
}

func (s *Server) closeLog() {
	if s.logFile != nil {
		f := s.logFile
		s.logFile = nil
		log.RemoveLogger(SERVER_LOG_ID)
		f.Close()
		return
	}
}

func (s *Server) startEmulationServices() error {
	s.setStatus(SERVER_STARTING)
	for _, srv := range s.services {
		if srv.GetType() == service.SERVICE_EMULATION {
			err := srv.Start(&s.waitGroup)
			if err != nil {
				return err
			}
		}
	}
	s.startTime = time.Now().Local()
	s.setStatus(SERVER_RUNNING)
	s.StartSessionsMonitorJob()
	s.StartProcessMonitorJob()
	s.StartRemoveAppLogsJob()
	s.StartHealthChecker()
	return nil
}

func (s *Server) startAdminServices() error {
	for _, srv := range s.services {
		if srv.GetType() == service.SERVICE_ADMIN {
			err := srv.Start(&s.waitGroup)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Server) startAllServices() error {
	s.setStatus(SERVER_STARTING)
	for _, srv := range s.services {
		err := srv.Start(&s.waitGroup)
		if err != nil {
			return err
		}
	}
	s.startTime = time.Now().Local()
	s.setStatus(SERVER_RUNNING)
	s.StartSessionsMonitorJob()
	s.StartProcessMonitorJob()
	s.StartRemoveAppLogsJob()
	s.StartHealthChecker()
	return nil
}

func (s *Server) Start(serviceType service.ServiceType) error {
	log.Info("Starting services...")
	defer log.Info("Services started.")
	switch serviceType {
	case service.SERVICE_EMULATION:
		return s.startEmulationServices()
	case service.SERVICE_ADMIN:
		return s.startAdminServices()
	default:
		return s.startAllServices()
	}
}

func (s *Server) AddService(srv service.Service) {
	s.services[srv.GetName()] = srv
	log.Infof("Service %s registered.", srv.GetName())
}

func (s *Server) AddAuthenticator(authId string, authenticator auth.UserAuthenticator) {
	s.authenticators[authId] = authenticator
	log.Infof("Authenticator %s registered.", authId)
}

func (s *Server) stopEmulationServices() error {
	killSessions := false
	s.setStatus(SERVER_STOPPING)
	s.StopHealthChecker()
	s.StopRemoveAppLogsJob()
	s.StopProcessMonitorJob()
	s.StopSessionsMonitorJob()
	for _, srv := range s.services {
		if srv.GetType() == service.SERVICE_EMULATION {
			err := srv.Stop()
			if err != nil {
				return err
			}
			killSessions = true
		}
	}
	if killSessions {
		s.KillAllSessions("service stopped")
	}
	s.startTime = time.Time{}
	s.setStatus(SERVER_STOPPED)
	return nil
}

func (s *Server) stopAdminServices() error {
	for _, srv := range s.services {
		if srv.GetType() == service.SERVICE_ADMIN {
			err := srv.Stop()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Server) stopAllServices() error {
	s.setStatus(SERVER_STOPPING)
	for _, srv := range s.services {
		err := srv.Stop()
		if err != nil {
			return err
		}
	}
	s.startTime = time.Time{}
	s.setStatus(SERVER_STOPPED)
	return nil
}

func (s *Server) Stop(serviceType service.ServiceType) error {
	log.Info("Stopping services...")
	defer log.Info("Services stopped.")
	switch serviceType {
	case service.SERVICE_EMULATION:
		return s.stopEmulationServices()
	case service.SERVICE_ADMIN:
		return s.stopAdminServices()
	default:
		return s.stopAllServices()
	}
}

func (s *Server) Config() *config.ServerConfig {
	return s.config
}

func (s *Server) Services() []service.Service {
	values := make([]service.Service, 0, len(s.services))
	for _, v := range s.services {
		values = append(values, v)
	}
	return values
}

func (s *Server) NewSession(teHandler service.TerminalConnectionHandler, sessionType SessionType, teAddr string, username string, osUser string, commandLine string) *Session {
	session := newSession(teHandler, sessionType, teAddr, username, osUser, commandLine)
	s.addSession(session)
	log.Infof("Server.NewSession(). session=%d type=%v addr=%s user=%s cmd=[%s]", session.Id, sessionType, teAddr, osUser, commandLine)
	return session
}

func (s *Server) GetSession(id int64) *Session {
	s.sessionsLock.RLock()
	defer s.sessionsLock.RUnlock()
	session := s.sessions[id]
	return session
}

func (s *Server) addSession(session *Session) {
	s.sessionsLock.Lock()
	defer s.sessionsLock.Unlock()
	s.sessions[session.Id] = session
}

func (s *Server) deleteSession(id int64) *Session {
	s.sessionsLock.Lock()
	defer s.sessionsLock.Unlock()
	session := s.sessions[id]
	delete(s.sessions, id)
	return session
}

func (s *Server) addLostSession(session *Session) {
	s.lostSessionsLock.Lock()
	defer s.lostSessionsLock.Unlock()
	s.sessionsLostConnection[session.Id] = session
}

func (s *Server) deleteLostSession(id int64) *Session {
	s.lostSessionsLock.Lock()
	defer s.lostSessionsLock.Unlock()
	session := s.sessionsLostConnection[id]
	delete(s.sessionsLostConnection, id)
	return session
}

func (s *Server) CloseSession(id int64) {
	session := s.deleteSession(id)
	if session != nil {
		if session.GetStatus() != SESS_CLOSE_REQUEST && session.InTransctionMode() {
			log.Infof("Server.CloseSession(). Not closed. Session %d in transaction mode.", id)
			s.addLostSession(session)
			session.TransactionStartTime = time.Now()
		}
		log.Infof("Server.CloseSession(). session=%d", id)
		session.close(false)
	} else {
		log.Tracef("Server.CloseSession(). session %d not found.", id)
	}
}

func (s *Server) handlePanic(message string) {
	if err := recover(); err != nil {
		log.Errorf("%s: %v\n%s", message, err, util.FullStack())
	}
}

func (s *Server) KillSession(id int64, reason string) *Session {
	defer s.handlePanic("unknown error in server(Server.KillSession)")
	session := s.deleteSession(id)
	if session != nil {
		session.close(true)
		log.Debugf("Server.KillSession(). id=%d reason='%s'", id, reason)
	} else {
		log.Errorf("Server.KillSession(). session %d not found.", id)
	}
	return session
}

func (s *Server) KillLostSession(id int64, reason string) *Session {
	defer s.handlePanic("unknown error in server(Server.KillLostSession)")
	session := s.deleteLostSession(id)
	if session != nil {
		session.killAppProcess()
		log.Debugf("Server.KillLostSession(). id=%d reason='%s'", id, reason)
	} else {
		log.Errorf("Server.KillLostSession(). session %d not found.", id)
	}
	return session
}

func (s *Server) sendLogoutToTerminal(sessionId int64, msg string) {
	defer s.handlePanic("unknown error in server(Server.sendLogoutToTerminal)")
	session := s.deleteSession(sessionId)
	if session == nil {
		log.Errorf("Server.sendLogoutToTerminal(). session %d not found.", sessionId)
		return
	}
	if session.TeHandler != nil {
		session.TeHandler.SendLogout(msg)
	} else {
		log.Debugf("Server.sendLogoutToTerminal(). Unknown error. Session %d closed without terminal handler. message '%s' not sent.", sessionId, msg)
	}
	session.close(true)
	log.Infof("Server.sendLogoutToTerminal() id=%d message='%s'", sessionId, msg)
}

func (s *Server) KillAllSessions(reason string) int32 {
	sessionsToKil := s.GetSessions()
	killedSessions := int32(0)
	for _, sess := range sessionsToKil {
		if s.KillSession(sess.Id, reason) != nil {
			killedSessions++
		}
	}
	log.Debugf("Server.KillAllSessions(). %d sessions killed", killedSessions)
	return killedSessions
}

func (s *Server) AuthenticateUser(authId, username, password string) bool {
	authenticator, found := s.authenticators[authId]
	if found {
		return authenticator.Authenticate(username, password)
	}
	return true
}

func (s *Server) GetStatus() ServerStatus {
	return s.status.Load().(ServerStatus)
}

func (s *Server) setStatus(status ServerStatus) {
	s.status.Store(status)
}

func (s *Server) GetStats() *stats.Stats {
	return s.stats
}

func (s *Server) PauseConnections() {
	log.Info("Server.PauseConnections(). Pausing new connections.")
	s.setStatus(SERVER_PAUSED)
	for _, srv := range s.services {
		if srv.GetType() == service.SERVICE_EMULATION {
			srv.PauseAccepting()
		}
	}
}

func (s *Server) ResumeConnections() {
	log.Info("Server.ResumeConnections(). Resuming new connections.")
	for _, srv := range s.services {
		if srv.GetType() == service.SERVICE_EMULATION {
			srv.ResumeAccepting()
		}
	}
	s.setStatus(SERVER_RUNNING)
}

func (s *Server) IsHealthy() bool {
	if s.healthChecker != nil {
		return s.healthChecker.IsHealthy()
	}
	return true
}

func (s *Server) GetHealthAlerts() []health.AlertType {
	if s.healthChecker != nil {
		return s.healthChecker.GetAlerts()
	}
	return nil
}

func (s *Server) StartHealthChecker() {
	if s.config.HealthEnabled().Get() && s.healthChecker == nil {
		s.healthChecker = health.New(s.config, s)
		s.healthChecker.Start()
	}
}

func (s *Server) StopHealthChecker() {
	if s.healthChecker != nil {
		h := s.healthChecker
		s.healthChecker = nil
		h.Stop()
	}
}

func (s *Server) GetPendingLoginSessions() []health.PendingSession {
	sessions := s.GetSessions()
	pending := make([]health.PendingSession, 0)
	for _, session := range sessions {
		if session.GetStatus() < SESS_READY {
			pending = append(pending, health.PendingSession{
				Id:        session.Id,
				StartTime: session.StartTime,
			})
		}
	}
	return pending
}

func (s *Server) GetSessionsCount() int32 {
	s.sessionsLock.RLock()
	defer s.sessionsLock.RUnlock()
	return int32(len(s.sessions))
}

func (s *Server) GetSessions() []*Session {
	sessions := make([]*Session, 0, len(s.sessions))
	s.sessionsLock.RLock()
	defer s.sessionsLock.RUnlock()
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func (s *Server) GetStartTime() int64 {
	return s.startTime.UnixMilli()
}

func (s *Server) GetUserRepository() UserRepository {
	return s.userRepository
}

func (s *Server) AwaitServices() {
	s.waitGroup.Wait()
}

func (s *Server) idleTimeout(session *Session) bool {
	if s.config.SessionIdleTimeout().Get() == 0 || !session.TimeoutEnabled.Get() || session.InTransctionMode() {
		return false
	}
	return (session.AppHandler != nil && time.Since(session.AppHandler.GetLastAppOperationTime()) > s.config.SessionIdleTimeout().Get()) ||
		(session.TeHandler != nil && time.Since(session.TeHandler.GetLastAppOperationTime()) > s.config.SessionIdleTimeout().Get())
}

func (s *Server) communicationLackTimeout(session *Session) bool {
	if s.config.AppLackTimeout().Get() == 0 {
		return false
	}
	return session.AppHandler != nil && time.Since(session.AppHandler.GetLastDataReadTime()) > s.config.AppLackTimeout().Get()
}

func (s *Server) zombieSession(session *Session) bool {
	return session.GetStatus() < SESS_READY && time.Since(session.StartTime) > s.Config().AppStartupTimeout().Get()
}

func (s *Server) timeoutLostTransactionSession(session *Session) bool {
	return session.InTransctionMode() && session.GetStatus() != SESS_READY && time.Since(session.TransactionStartTime) > s.Config().AppTransactionTimeout().Get()
}

func (s *Server) getLostSessions() []*Session {
	s.lostSessionsLock.RLock()
	defer s.lostSessionsLock.RUnlock()
	sessions := make([]*Session, 0, len(s.sessionsLostConnection))
	for _, s := range s.sessionsLostConnection {
		sessions = append(sessions, s)
	}
	return sessions
}

func (s *Server) timeoutAppConnect(session *Session) bool {
	return session.AppHandler == nil &&
		session.GetStatus() < SESS_READY &&
		session.SessionType != SESS_TYPE_STANDALONE &&
		time.Since(session.StartTime) >= s.Config().AppConnectionTimeout().Get()
}

func (s *Server) GetServerProcess() *process.Process {
	return s.serverProcess
}

func (s *Server) processMonitorJob() {
	log.Debugf("Server.orphanProcessMonitor(). started.")
	defer s.handlePanic("unknown error in server(Server.orphanProcessMonitor)")
	p := s.GetServerProcess()
	errorCount := 0
	for range s.orphanProcessTimer.C {
		sessions := s.GetSessions()
		sessions = append(sessions, s.getLostSessions()...)
		processList, err := p.Children()
		if err == nil {
			killOrphanProcesses(sessions, processList)
		} else if errorCount < 10 {
			log.Errorf("Server.orphanProcessMonitor(). Error getting child process: %v", err)
			errorCount++
		} else {
			log.Debugf("Server.orphanProcessMonitor(). Error getting child process: %v", err)
		}
	}
	log.Debugf("Server.orphanProcessMonitor(). stopped.")
}

func orphanProcess(sessions []*Session, proc *process.Process) bool {
	value, err := util.ProcessEnvVar(proc, ENV_VAR_AUTH_TOKEN)
	if err != nil {
		cmd, _ := proc.Cmdline()
		log.Debugf("server.orphanProcess(). Error getting process environment variables. PID: %d Cmd: '%s' Error: %v", proc.Pid, cmd, err)
		return true
	}
	sessionId, _ := strconv.ParseInt(value, 10, 64)
	return slices.IndexFunc(sessions, func(session *Session) bool { return session.Id == sessionId }) < 0
}

func killOrphanProcesses(sessions []*Session, processes []*process.Process) {
	for _, proc := range processes {
		if orphanProcess(sessions, proc) {
			util.KillProcessRecursive(proc, "orphan process")
			log.Debugf("server.killOrphanProcesses(). proc=%v", proc)
		}
	}
}

func (s *Server) StartProcessMonitorJob() {
	if s.orphanProcessTimer == nil {
		interval := s.config.OrphanProcessCheckInterval().Get()
		s.orphanProcessTimer = time.NewTicker(interval)
		go s.processMonitorJob()
	}
}

func (s *Server) StopProcessMonitorJob() {
	if s.orphanProcessTimer != nil {
		m := s.orphanProcessTimer
		s.orphanProcessTimer = nil
		m.Stop()
	}
}

func (s *Server) StartRemoveAppLogsJob() {
	if s.removeAppLogsTimer == nil {
		s.removeAppLogsTimer = time.NewTicker(time.Minute * 60)
		go s.removeAppLogsJob()
	}
}

func (s *Server) StopRemoveAppLogsJob() {
	if s.removeAppLogsTimer != nil {
		m := s.removeAppLogsTimer
		s.removeAppLogsTimer = nil
		m.Stop()
	}
}

func (s *Server) sessionsMonitorJob() {
	log.Debugf("server.sessionsMonitor(). started.")
	defer s.handlePanic("unknown error in server(Server.sessionsMonitor)")
	for range s.monitorSessionsTimer.C {
		sessions := s.GetSessions()
		for _, session := range sessions {
			if s.timeoutAppConnect(session) {
				s.sendLogoutToTerminal(session.Id, "application killed because did not respond")
			} else if s.idleTimeout(session) {
				s.sendLogoutToTerminal(session.Id, "Application closed by inactivity")
			} else if s.communicationLackTimeout(session) {
				s.sendLogoutToTerminal(session.Id, "Application killed by communication lack")
			} else if s.zombieSession(session) {
				s.KillSession(session.Id, "zombie session")
			}
		}
		sessions = s.getLostSessions()
		for _, session := range sessions {
			if s.timeoutLostTransactionSession(session) {
				s.KillLostSession(session.Id, "lost transaction session")
			}
		}
	}
	log.Debugf("server.sessionsMonitor(). stopped.")
}

func (s *Server) StartSessionsMonitorJob() {
	if s.monitorSessionsTimer == nil {
		interval := s.config.SessionsCheckInterval().Get()
		if interval <= 10*time.Second {
			interval = 10 * time.Second
		}
		s.monitorSessionsTimer = time.NewTicker(interval)
		go s.sessionsMonitorJob()
	}
}

func (s *Server) StopSessionsMonitorJob() {
	if s.monitorSessionsTimer != nil {
		m := s.monitorSessionsTimer
		s.monitorSessionsTimer = nil
		m.Stop()
	}
}

func (s *Server) EnvVars() []string {
	envVars := make([]string, 0, 32)
	envVars = append(envVars, os.Environ()...)
	for k, v := range s.Config().GetEnvVars() {
		envVars = append(envVars, k+"="+v)
	}
	return envVars
}

func (s *Server) removeAppLogsJob() {
	log.Debugf("server.removeAppLogsJob(). started.")
	path, fileName := filepath.Split(s.config.AppLogPathName().Get())
	multiFilesLog := strings.Contains(fileName, "${")
	if multiFilesLog {
		exp, err := regexp.Compile(`\$\{[^}]+\}`)
		if err != nil {
			log.Errorf("Error creating mask to remove old app log files: %v", err)
			log.Debugf("server.removeAppLogsJob(). stopped.")
			return
		}
		fileName = exp.ReplaceAllString(fileName, "*")
	}
	days := s.config.AppLogDaysRetention().Get()
	for range s.removeAppLogsTimer.C {
		now := time.Now()
		if now.Hour() >= 0 && now.Hour() <= 5 && now.Sub(s.lastAppLogRemoveExecution).Hours() >= 23 {
			log.Infof("Searching old app logs.")
			s.lastAppLogRemoveExecution = time.Now()
			if multiFilesLog {
				util.RemoveFiles(path, fileName, int(days))
			} else {
				util.TruncateFile(filepath.Join(path, fileName), 20*1024*1024, 15*1024*1024)
			}
		}
	}
	log.Debugf("server.removeAppLogsJob(). stopped.")
}
