package server

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"rgt-server/config"
	"rgt-server/health"
	"rgt-server/log"
	"rgt-server/service"
	"rgt-server/util"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type ServerStatus uint8

const (
	SERVER_STOPPED       ServerStatus = 0
	SERVER_STARTING      ServerStatus = 1
	SERVER_RUNNING       ServerStatus = 2
	SERVER_STOPPING      ServerStatus = 3
	SERVER_PAUSED        ServerStatus = 4
	SERVER_DISCONNECTED  ServerStatus = 5
	SERVER_CONNECTING    ServerStatus = 6
	SERVER_DISCONNECTING ServerStatus = 7
)

type Server struct {
	services                  map[string]service.Service
	config                    *config.ServerConfig
	waitGroup                 sync.WaitGroup
	startTime                 time.Time
	version                   string
	userRepository            UserRepository
	removeAppLogsTimer        *time.Ticker
	lastAppLogRemoveExecution time.Time
	status                    atomic.Uint32
	healthMonitor             *health.HealthMonitor
}

func New(config *config.ServerConfig, version string) *Server {
	srv := &Server{config: config,
		services: make(map[string]service.Service),
		version:  version,
	}
	srv.setStatus(SERVER_STOPPED)
	return srv
}

func (s *Server) GetStatus() ServerStatus {
	return ServerStatus(s.status.Load())
}

func (s *Server) setStatus(status ServerStatus) {
	s.status.Store(uint32(status))
}

func (s *Server) ChangeStatus(oldStatus, newStatus ServerStatus) error {
	if oldStatus == newStatus {
		return fmt.Errorf("New server status (%s) is the same of expected status (%s)", newStatus, oldStatus)
	}
	if !s.status.CompareAndSwap(uint32(oldStatus), uint32(newStatus)) {
		previousStatus := s.GetStatus()
		return fmt.Errorf("Server with unexpected status %s. Expected %s to change to %s", previousStatus, oldStatus, newStatus)
	}
	log.Debugf("Server changed status from %s to %s", oldStatus, newStatus)
	return nil
}

func (s *Server) Version() string {
	return s.version
}

func (s *Server) startServices(admin bool) error {
	log.Infof("Starting %s services...", util.IIf(admin, "all", "non-admin"))
	for _, srv := range s.services {
		if !srv.IsAdmin() || admin {
			err := srv.Start(&s.waitGroup)
			if err != nil {
				return err
			}
		}
	}
	s.startTime = time.Now().Local()
	log.Infof("%s services started.", util.IIf(admin, "All", "Non-admin"))
	return nil
}

func (s *Server) Startup() error {
	log.Infof("Starting up server...")
	s.setStatus(SERVER_STARTING)
	if err := s.startServices(true); err != nil {
		s.stopRunningServices(true)
		s.setStatus(SERVER_STOPPED)
		return err
	}
	s.StartRemoveAppLogsJob()
	s.StartHealthMonitor()
	s.setStatus(SERVER_RUNNING)
	log.Infof("Server started.")
	return nil
}

func (s *Server) Start() error {
	if err := s.ChangeStatus(SERVER_STOPPED, SERVER_STARTING); err != nil {
		return err
	}
	if err := s.startServices(false); err != nil {
		s.stopRunningServices(false)
		s.setStatus(SERVER_STOPPED)
		return err
	}
	s.setStatus(SERVER_RUNNING)
	return nil
}

func (s *Server) AddService(srv service.Service) {
	s.services[srv.Name()] = srv
	log.Infof("Service %s registered.", srv.Name())
}

func (s *Server) stopServices(admin bool) error {
	log.Infof("Stopping %s services...", util.IIf(admin, "all", "non-admin"))
	for _, srv := range s.services {
		if !srv.IsAdmin() || admin {
			if err := srv.Stop(); err != nil {
				return err
			}
		}
	}
	s.startTime = time.Time{}
	log.Infof("%s services stopped.", util.IIf(admin, "All", "Non-admin"))
	return nil
}

func (s *Server) stopRunningServices(admin bool) {
	for _, srv := range s.services {
		if srv.IsRunning() && (!srv.IsAdmin() || admin) {
			if err := srv.Stop(); err != nil {
				log.Errorf("Error stopping service %s: %v", srv.Name(), err)
			}
		}
	}
}

func (s *Server) Stop() error {
	if err := s.ChangeStatus(SERVER_RUNNING, SERVER_STOPPING); err != nil {
		return err
	}
	if err := s.stopServices(false); err != nil {
		return err
	}
	s.setStatus(SERVER_STOPPED)
	return nil
}

func (s *Server) Shutdown() {
	log.Info("Shutting down server...")
	s.setStatus(SERVER_STOPPING)
	s.StopHealthMonitor()
	s.StopRemoveAppLogsJob()
	s.stopRunningServices(true)
	s.waitGroup.Wait()
	s.setStatus(SERVER_STOPPED)
	log.Info("Server shutdown.")
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

func (s *Server) Pause() {
	log.Info("Server.Pause(). Pausing services.")
	if err := s.ChangeStatus(SERVER_RUNNING, SERVER_PAUSED); err != nil {
		log.Errorf("Error occurred while trying to pause server: %v", err)
		return
	}
	for _, srv := range s.services {
		srv.Pause()
	}
}

func (s *Server) Resume() {
	log.Info("Server.Resume(). Resuming services.")
	if err := s.ChangeStatus(SERVER_PAUSED, SERVER_RUNNING); err != nil {
		log.Errorf("Error occurred while trying to resume server: %v", err)
		return
	}
	for _, srv := range s.services {
		srv.Resume()
	}
	s.setStatus(SERVER_RUNNING)
}

func (s *Server) StartHealthMonitor() {
	if s.config.HealthEnabled().Get() && s.healthMonitor == nil {
		var err error
		log.Info("Creating health monitor...")
		if s.healthMonitor, err = health.NewDefault(s.config, s); err != nil {
			log.Errorf("Error initializing health monitor: %v", err)
			return
		}
		s.healthMonitor.Start(context.Background())
	}
}

func (s *Server) StopHealthMonitor() {
	if s.healthMonitor != nil {
		h := s.healthMonitor
		s.healthMonitor = nil
		h.Stop()
	}
}

func (s *Server) GetStartTime() int64 {
	return s.startTime.UnixMilli()
}

func (s *Server) GetUserRepository() UserRepository {
	return s.userRepository
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

func (s ServerStatus) String() string {
	switch s {
	case SERVER_STOPPED:
		return "STOPPED"
	case SERVER_STARTING:
		return "STARTING"
	case SERVER_RUNNING:
		return "RUNNING"
	case SERVER_STOPPING:
		return "STOPPING"
	case SERVER_PAUSED:
		return "PAUSED"
	case SERVER_DISCONNECTED:
		return "DISCONNECTED"
	case SERVER_CONNECTING:
		return "CONNECTING"
	case SERVER_DISCONNECTING:
		return "DISCONNECTING"
	}
	return "UNKNOWN"
}

func StatusFromName(statusName string) ServerStatus {
	switch statusName {
	case "STOPPED":
		return SERVER_STOPPED
	case "STARTING":
		return SERVER_STARTING
	case "RUNNING":
		return SERVER_RUNNING
	case "STOPPING":
		return SERVER_STOPPING
	case "PAUSED":
		return SERVER_PAUSED
	case "DISCONNECTED":
		return SERVER_DISCONNECTED
	case "CONNECTING":
		return SERVER_CONNECTING
	case "DISCONNECTING":
		return SERVER_DISCONNECTING
	}
	return SERVER_STOPPED
}
