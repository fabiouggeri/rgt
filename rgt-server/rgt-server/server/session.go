package server

import (
	"rgt-server/config"
	"strings"
	"sync/atomic"
	"time"
)

type SessionStatus uint8

type SessionMode uint8

type SessionType uint8

type Session interface {
	Id() int64
	Close(kill bool, message string) bool
	GetStatus() SessionStatus
	GetType() SessionType
	GetAddress() string
	GetStartTime() time.Time
	GetUser() string
	GetOSUser() string
	CommandLine() string
	Pid() int64
	CloseConditionally(config *config.ServerConfig) bool
}
type SessionListener interface {
	StatusChange(session Session, oldStatus SessionStatus, newStatus SessionStatus)
}

const (
	// Sessions status
	SESS_NEW           SessionStatus = 0
	SESS_LAUNCHING_APP SessionStatus = 1
	SESS_CONNECTING    SessionStatus = 2
	SESS_READY         SessionStatus = 3
	SESS_CLOSE_REQUEST SessionStatus = 4
	SESS_CLOSING       SessionStatus = 5
	SESS_CLOSED        SessionStatus = 6

	// Sessions modes
	SESS_MODE_NORMAL      SessionMode = 0
	SESS_MODE_TRANSACTION SessionMode = 1

	// Sessions types
	SESS_TYPE_EMULATION  SessionType = 0
	SESS_TYPE_STANDALONE SessionType = 1
)

var sessionCount int64 = 0

func NextSessionId() int64 {
	return atomic.AddInt64(&sessionCount, 1)
}

func SessionStatusFromName(statusName string) SessionStatus {
	switch strings.ToUpper(statusName) {
	case "NEW":
		return SESS_NEW
	case "LAUNCHING APP":
		return SESS_LAUNCHING_APP
	case "CONNECTING":
		return SESS_CONNECTING
	case "READY":
		return SESS_READY
	case "CLOSE REQUEST":
		return SESS_CLOSE_REQUEST
	case "CLOSING":
		return SESS_CLOSING
	default:
		return SESS_CLOSED
	}
}

func SessionStatusName(status SessionStatus) string {
	switch status {
	case SESS_NEW:
		return "NEW"
	case SESS_LAUNCHING_APP:
		return "LAUNCHING APP"
	case SESS_CONNECTING:
		return "CONNECTING"
	case SESS_READY:
		return "READY"
	case SESS_CLOSE_REQUEST:
		return "CLOSE REQUEST"
	case SESS_CLOSING:
		return "CLOSING"
	default:
		return "CLOSED"
	}
}

func (status SessionStatus) String() string {
	return SessionStatusName(status)
}

func (status SessionStatus) GoString() string {
	return SessionStatusName(status)
}

func (t SessionType) String() string {
	if t == SESS_TYPE_STANDALONE {
		return "STANDALONE"
	}
	return "EMULATION"
}

func (t SessionType) GoString() string {
	return t.String()
}

func (m SessionMode) String() string {
	if m == SESS_MODE_TRANSACTION {
		return "TRANSACTION"
	}
	return "NORMAL"
}

func (m SessionMode) GoString() string {
	return m.String()
}
