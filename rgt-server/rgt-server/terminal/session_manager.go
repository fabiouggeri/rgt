package terminal

import (
	"fmt"
	"rgt-server/log"
	"rgt-server/option"
	"rgt-server/service"
	"rgt-server/util"
	"sync"
	"time"
)

type TerminalServiceCallbacks interface {
	Pause()
	Resume()
	GetStatus() service.ServiceStatus
}

type SessionManagerConfig interface {
	SessionsCheckInterval() option.TypedOption[time.Duration]
	HealthMaxLoginTime() option.TypedOption[time.Duration]
	HealthLoginsTimeoutThreshold() option.TypedOption[uint16]
	HealthLoginsTimeoutResumeThreshold() option.TypedOption[uint16]
	HealthMaxLoginsTimeoutAlerts() option.TypedOption[uint16]
	AppLaunchTimeout() option.TypedOption[time.Duration]
	AppLoginTimeout() option.TypedOption[time.Duration]
	SessionIdleTimeout() option.TypedOption[time.Duration]
	AppLackTimeout() option.TypedOption[time.Duration]
	AppTransactionTimeout() option.TypedOption[time.Duration]
}

type SessionManager struct {
	sessions                   map[int64]*TerminalSession
	sessionsLock               sync.RWMutex
	monitorSessionsTimer       *time.Ticker
	serviceCallback            TerminalServiceCallbacks
	config                     SessionManagerConfig
	maxLoginsTimeoutAlertCount uint16
}

func NewSessionManager(callbacks TerminalServiceCallbacks, config SessionManagerConfig) *SessionManager {
	return &SessionManager{
		sessions:        make(map[int64]*TerminalSession),
		serviceCallback: callbacks,
		config:          config,
	}
}

func (s *SessionManager) GetSession(id int64) *TerminalSession {
	s.sessionsLock.Lock()
	defer s.sessionsLock.Unlock()
	session := s.sessions[id]
	return session
}

func (s *SessionManager) GetSessions() []*TerminalSession {
	sessions := make([]*TerminalSession, 0, len(s.sessions))
	s.sessionsLock.RLock()
	defer s.sessionsLock.RUnlock()
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func (s *SessionManager) AddSession(session *TerminalSession) error {
	s.sessionsLock.Lock()
	defer s.sessionsLock.Unlock()
	if _, exists := s.sessions[session.Id()]; exists {
		return fmt.Errorf("session %d already exists", session.Id())
	}
	s.sessions[session.Id()] = session
	log.Debugf("SessionManager.AddSession(). Session %d added", session.Id())
	return nil
}

func (s *SessionManager) DeleteSession(sessionId int64) *TerminalSession {
	s.sessionsLock.Lock()
	defer s.sessionsLock.Unlock()
	session := s.sessions[sessionId]
	if session != nil {
		delete(s.sessions, sessionId)
		log.Debugf("SessionManager.DeleteSession(). Session %d deleted", sessionId)
	}
	return session
}

func (s *SessionManager) CloseSession(id int64) {
	session := s.GetSession(id)
	if session == nil {
		log.Tracef("SessionManager.CloseSession(). session %d not found.", id)
		return
	}
	if session.Close(false, "") {
		log.Infof("SessionManager.CloseSession(). session %d closed", id)
		s.DeleteSession(session.Id())
	}
}

func (s *SessionManager) handlePanic(message string) {
	if err := recover(); err != nil {
		log.Errorf("%s: %v\n%s", message, err, util.FullStack())
	}
}

func (s *SessionManager) KillSession(id int64, reason string) *TerminalSession {
	defer s.handlePanic("unknown error in session manager(SessionManager.KillSession)")
	session := s.DeleteSession(id)
	if session != nil {
		session.Close(true, "")
		log.Infof("SessionManager.KillSession(). session=%d reason='%s'", id, reason)
	} else {
		log.Errorf("ServerManager.KillSession(). session %d not found.", id)
	}
	return session
}

func (s *SessionManager) KillAllSessions(reason string) int32 {
	sessionsToKil := s.GetSessions()
	killedSessions := int32(0)
	for _, sess := range sessionsToKil {
		if s.KillSession(sess.Id(), reason) != nil {
			killedSessions++
		}
	}
	log.Debugf("ServerManager.KillAllSessions(). %d sessions killed", killedSessions)
	return killedSessions
}

func (s *SessionManager) GetSessionsCount() int32 {
	s.sessionsLock.RLock()
	defer s.sessionsLock.RUnlock()
	return int32(len(s.sessions))
}

func (s *SessionManager) GetSessionsStatus(status SessionStatus, returnPrevious bool) []*TerminalSession {
	sessions := make([]*TerminalSession, 0)
	s.sessionsLock.RLock()
	defer s.sessionsLock.RUnlock()
	if returnPrevious {
		for _, session := range s.sessions {
			if session.GetStatus() <= status {
				sessions = append(sessions, session)
			}
		}
	} else {
		for _, session := range s.sessions {
			if session.GetStatus() == status {
				sessions = append(sessions, session)
			}
		}
	}
	return sessions
}

func (s *SessionManager) StartSessionsMonitorJob() {
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

func (s *SessionManager) StopSessionsMonitorJob(killSessions bool) {
	if s.monitorSessionsTimer != nil {
		m := s.monitorSessionsTimer
		s.monitorSessionsTimer = nil
		m.Stop()
		log.Info("Session monitor stopped.")
	}
	if killSessions {
		s.KillAllSessions("service stopped")
	}
}

func (s *SessionManager) sessionsMonitorJob() {
	log.Debug("SessionManager.sessionsMonitor(). started.")
	defer s.handlePanic("unknown error in service (SessionManager.sessionsMonitor)")
	for range s.monitorSessionsTimer.C {
		loginsTimeoutCount := uint16(0)
		maxLoginTime := s.config.HealthMaxLoginTime().Get()
		maxPendingLoginsAlerts := s.config.HealthMaxLoginsTimeoutAlerts().Get()
		sessions := s.GetSessions()
		log.Debugf("SessionManager.sessionsMonitor(). Checking %d sessions", len(sessions))
		for _, session := range sessions {
			loginsTimeoutCount = s.checkSession(session, maxLoginTime, loginsTimeoutCount)
		}
		if loginsTimeoutCount > 0 {
			s.checkPendingLoginsExceededCount(loginsTimeoutCount, maxPendingLoginsAlerts)
		}
	}
	log.Debug("SessionManager.sessionsMonitor(). stopped.")
}

func (s *SessionManager) checkSession(session *TerminalSession, maxLoginTime time.Duration, loginsTimeoutCount uint16) uint16 {
	switch session.GetStatus() {
	case SESS_NEW:
		if session.timeoutAppLaunch(s.config.AppLaunchTimeout().Get()) {
			session.sendLogoutToTerminal("session closed because application was not launched")
			s.DeleteSession(session.Id())
		} else if session.loginTimeExceeded(maxLoginTime) {
			return loginsTimeoutCount + 1
		}
	case SESS_LAUNCHING_APP, SESS_CONNECTING:
		if session.timeoutAppLogin(s.config.AppLoginTimeout().Get()) {
			giveBackLaunchAppSlot(session)
			session.sendLogoutToTerminal("application killed because did not respond")
			s.DeleteSession(session.Id())
		} else if session.loginTimeExceeded(maxLoginTime) {
			return loginsTimeoutCount + 1
		}
	case SESS_READY:
		if !session.appIsRunning() {
			session.sendLogoutToTerminal("application closed")
			s.DeleteSession(session.Id())
		} else if session.idleTimeout(s.config.SessionIdleTimeout().Get()) {
			session.sendLogoutToTerminal("application closed by inactivity")
			s.DeleteSession(session.Id())
		} else if session.communicationLackTimeout(s.config.AppLackTimeout().Get()) {
			session.sendLogoutToTerminal("application killed by communication lack")
			s.DeleteSession(session.Id())
		} else if session.timeoutLostTransactionSession(s.config.AppTransactionTimeout().Get()) {
			session.killAppProcess("lost transaction session")
			s.DeleteSession(session.Id())
		}
	case SESS_CLOSE_REQUEST, SESS_CLOSING:
		break
	case SESS_CLOSED:
		s.DeleteSession(session.Id())
	}
	return loginsTimeoutCount
}

func (s *SessionManager) checkPendingLoginsExceededCount(loginsTimeoutCount uint16, maxLoginsTimeoutAlerts uint16) {
	threshold := s.config.HealthLoginsTimeoutThreshold().Get()
	resumeThreshold := s.config.HealthLoginsTimeoutResumeThreshold().Get()
	if loginsTimeoutCount >= threshold {
		if s.serviceCallback.GetStatus() == service.STARTED {
			s.maxLoginsTimeoutAlertCount++
			if s.maxLoginsTimeoutAlertCount >= maxLoginsTimeoutAlerts {
				log.Infof("Login Timeout Checker. Server unhealthy. Pausing new connections. Alerts: %d", s.maxLoginsTimeoutAlertCount)
				s.serviceCallback.Pause()
			} else {
				log.Infof("Login Timeout Checker. %d logins timeouts exceeds threshold %d. Alert increased to %d", loginsTimeoutCount, threshold, s.maxLoginsTimeoutAlertCount)
			}
		}
	} else if loginsTimeoutCount <= resumeThreshold {
		if s.maxLoginsTimeoutAlertCount > 0 {
			log.Infof("Login Timeout Checker. Alert cleared")
		}
		s.maxLoginsTimeoutAlertCount = 0
		if s.serviceCallback.GetStatus() == service.PAUSED {
			log.Infof("Login Timeout Checker. Server healthy. Resuming new connections.")
			s.serviceCallback.Resume()
		}
	}
}
