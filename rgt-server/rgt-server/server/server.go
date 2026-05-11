package server

import (
	"path/filepath"
	"regexp"
	"rgt-server/auth"
	"rgt-server/config"
	"rgt-server/health"
	"rgt-server/log"
	"rgt-server/service"
	"rgt-server/stats"
	"rgt-server/util"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type ServerStatus string

const (
	SERVER_STOPPED       ServerStatus = "STOPPED"
	SERVER_STARTING      ServerStatus = "STARTING"
	SERVER_RUNNING       ServerStatus = "RUNNING"
	SERVER_STOPPING      ServerStatus = "STOPPING"
	SERVER_PAUSED        ServerStatus = "PAUSED"
	SERVER_DISCONNECTED  ServerStatus = "DISCONNECTED"
	SERVER_CONNECTING    ServerStatus = "CONNECTING"
	SERVER_DISCONNECTING ServerStatus = "DISCONNECTING"
)

type Server struct {
	services                  map[string]service.Service
	authenticatorManager      *auth.AuthenticatorManager
	config                    *config.ServerConfig
	waitGroup                 sync.WaitGroup
	startTime                 time.Time
	version                   string
	userRepository            UserRepository
	removeAppLogsTimer        *time.Ticker
	lastAppLogRemoveExecution time.Time
	status                    atomic.Value // stores ServerStatus
	stats                     *stats.ServerStats
	healthChecker             *health.HealthChecker
}

func New(config *config.ServerConfig, version string) *Server {
	srv := &Server{config: config,
		services:             make(map[string]service.Service),
		authenticatorManager: auth.NewAuthenticatorManager(),
		version:              version,
		stats:                stats.NewServerStats(),
	}
	srv.status.Store(SERVER_STOPPED)
	return srv
}

func (s *Server) Version() string {
	return s.version
}

func (s *Server) Finalize() {
	s.setStatus(SERVER_STOPPED)
	log.Debugf("Server.Finalize().")
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

func (s *Server) AuthenticatorManager() *auth.AuthenticatorManager {
	return s.authenticatorManager
}

func (s *Server) stopEmulationServices() error {
	s.setStatus(SERVER_STOPPING)
	s.StopHealthChecker()
	s.StopRemoveAppLogsJob()
	for _, srv := range s.services {
		if srv.GetType() == service.SERVICE_EMULATION {
			err := srv.Stop()
			if err != nil {
				return err
			}
		}
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
	s.StopHealthChecker()
	s.StopRemoveAppLogsJob()
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

func (s *Server) handlePanic(message string) {
	if err := recover(); err != nil {
		log.Errorf("%s: %v\n%s", message, err, util.FullStack())
	}
}

func (s *Server) GetStatus() ServerStatus {
	return s.status.Load().(ServerStatus)
}

func (s *Server) setStatus(status ServerStatus) {
	s.status.Store(status)
}

func (s *Server) GetStats() *stats.ServerStats {
	return s.stats
}

func (s *Server) Pause() {
	log.Info("Server.Pause(). Pausing services.")
	s.setStatus(SERVER_PAUSED)
	for _, srv := range s.services {
		srv.Pause()
	}
}

func (s *Server) Resume() {
	log.Info("Server.Resume(). Resuming services.")
	for _, srv := range s.services {
		srv.Resume()
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

func (s *Server) GetStartTime() int64 {
	return s.startTime.UnixMilli()
}

func (s *Server) GetUserRepository() UserRepository {
	return s.userRepository
}

func (s *Server) AwaitServices() {
	s.waitGroup.Wait()
}

func (s *Server) StartRemoveAppLogsJob() {
	if s.removeAppLogsTimer == nil {
		s.removeAppLogsTimer = time.NewTicker(time.Minute * 60)
		go s.removeAppLogsJob()
		log.Infof("Log cleaner started. Interval: %v", time.Minute*60)
	}
}

func (s *Server) StopRemoveAppLogsJob() {
	if s.removeAppLogsTimer != nil {
		m := s.removeAppLogsTimer
		s.removeAppLogsTimer = nil
		m.Stop()
		log.Info("Log cleaner stopped.")
	}
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
