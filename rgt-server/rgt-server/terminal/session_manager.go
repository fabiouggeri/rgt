package terminal

import (
	"fmt"
	"rgt-server/log"
	"rgt-server/util"
	"sync"
)

type SessionManager struct {
	sessions     map[int64]*TerminalSession
	sessionsLock sync.RWMutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[int64]*TerminalSession),
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
