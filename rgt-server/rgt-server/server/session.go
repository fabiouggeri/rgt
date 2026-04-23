package server

import (
	"fmt"
	"io"
	"rgt-server/buffer"
	"rgt-server/log"
	"rgt-server/option"
	"rgt-server/service"
	"rgt-server/util"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

type SessionStatus uint8

type SessionMode uint8

type SessionType uint8

type SessionListener interface {
	StatusChange(session *Session, oldStatus SessionStatus, newStatus SessionStatus)
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

type Session struct {
	Id                   int64
	TeHandler            service.TerminalConnectionHandler
	AppHandler           service.TerminalConnectionHandler
	TerminalUser         string
	TerminalAddress      string
	AppAddress           string
	OsUser               string
	AppPid               int64
	StartTime            time.Time
	AppLaunchTime        time.Time
	AppLoginTime         time.Time
	TransactionStartTime time.Time
	CommandLine          string
	Process              *process.Process
	Options              *option.Options
	Mode                 option.TypedOption[SessionMode]
	TimeoutEnabled       option.TypedOption[bool]
	status               atomic.Uint32
	SessionType          SessionType
	closing              atomic.Bool
	statusListeners      []SessionListener
}

var sessionCount int64 = 0

func newSession(teHandler service.TerminalConnectionHandler, sessionType SessionType, teAddr string, username string, osUser string, commandLine string) *Session {
	now := time.Now()
	s := &Session{Id: atomic.AddInt64(&sessionCount, 1),
		TeHandler:            teHandler,
		AppHandler:           nil,
		StartTime:            now,
		AppLaunchTime:        now,
		AppLoginTime:         now,
		TransactionStartTime: now,
		Process:              nil,
		TerminalAddress:      teAddr,
		TerminalUser:         username,
		OsUser:               osUser,
		CommandLine:          commandLine,
		TimeoutEnabled:       option.NewBool(false, "timeoutenabled"),
		Mode:                 option.NewUint(SESS_MODE_NORMAL, "mode"),
		Options:              option.NewOptions(),
		SessionType:          sessionType,
		statusListeners:      make([]SessionListener, 0),
	}
	s.closing.Store(false)
	s.Options.Add(s.Mode)
	s.Options.Add(s.TimeoutEnabled)
	return s
}

func (s *Session) SetAppHandler(appHandler service.TerminalConnectionHandler) {
	s.AppHandler = appHandler
}

func (s *Session) SetTerminalAddress(addr string) {
	s.TerminalAddress = addr
}

func (s *Session) SetTerminalUser(usr string) {
	s.TerminalUser = usr
}

func (s *Session) SetAppAddress(addr string) {
	s.AppAddress = addr
}

func (s *Session) SetAppPid(pid int64) {
	s.AppPid = pid
}

func (s *Session) SetOsUser(user string) {
	s.OsUser = user
}

func (s *Session) notifyStatusListeners(oldStatus, newStatus SessionStatus) {
	for _, listener := range s.statusListeners {
		listener.StatusChange(s, oldStatus, newStatus)
	}
}

func (s *Session) ChangeStatus(oldStatus, newStatus SessionStatus) error {
	if s.closing.Load() {
		return nil
	}
	if oldStatus == newStatus {
		return fmt.Errorf("Session %d already in status %s", s.Id, SessionStatusName(oldStatus))
	}
	previousStatus := SessionStatus(s.status.Swap(uint32(newStatus)))
	if previousStatus != oldStatus {
		return fmt.Errorf("Session %d with unexpected status %s. Expected %s to change to %s", s.Id,
			SessionStatusName(previousStatus), SessionStatusName(oldStatus), SessionStatusName(newStatus))
	}
	s.notifyStatusListeners(oldStatus, newStatus)
	return nil
}

func (s *Session) SetStatus(status SessionStatus) {
	if !s.closing.Load() {
		oldStatus := SessionStatus(s.status.Swap(uint32(status)))
		if oldStatus != status {
			s.notifyStatusListeners(oldStatus, status)
		}
	}
}

func (s *Session) GetStatus() SessionStatus {
	return SessionStatus(s.status.Load())
}

func (s *Session) SetStartTime(startTime time.Time) {
	s.StartTime = startTime
}

func (s *Session) SetAppLaunchTime(appLaunchTime time.Time) {
	s.AppLaunchTime = appLaunchTime
}

func (s *Session) SetAppLoginTime(appLoginTime time.Time) {
	s.AppLoginTime = appLoginTime
}

func (s *Session) SetCommandLine(cmd string) {
	s.CommandLine = cmd
}

func (s *Session) SetProcess(process *process.Process) {
	s.Process = process
}

func (s *Session) GetMode() SessionMode {
	return s.Mode.Get()
}

func (s *Session) SetMode(mode SessionMode) {
	s.Mode.Set(mode)
}

func (s *Session) killAppProcess() {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("unknown error in server(Session.killAppProcess): %v\n%s", err, util.FullStack())
		}
	}()
	if !s.isAppRunning() {
		return
	}
	if err := util.KillProcessRecursive(s.Process, "session killed"); err != nil {
		log.Debugf("Session.killAppProcess(). Error killing process. session=%d error=%v cmd=%s", s.Id, err, s.CommandLine)
	}
}

func (s *Session) closeTE() {
	if s.TeHandler != nil {
		th := s.TeHandler
		s.TeHandler = nil
		th.Close()
	}
}

func (s *Session) closeApp(killProcess bool) {
	if s.AppHandler != nil {
		ah := s.AppHandler
		s.AppHandler = nil
		ah.Close()
		if killProcess && !s.InTransctionMode() {
			s.killAppProcess()
		}
	} else if killProcess && !s.InTransctionMode() {
		s.killAppProcess()
	}
}

func (s *Session) close(killProcess bool, message string) {
	if !s.closing.CompareAndSwap(false, true) {
		return
	}
	log.Debugf("Session.Close(). closing session %d", s.Id)
	oldStatus := SessionStatus(s.status.Swap(uint32(SESS_CLOSING)))
	if oldStatus != SESS_CLOSING && s.statusListeners != nil {
		s.notifyStatusListeners(oldStatus, SESS_CLOSING)
	}
	if message != "" {
		if s.TeHandler != nil {
			s.TeHandler.SendLogout(message)
		} else {
			log.Debugf("Session.closeWithMessage(). Unknown error. Session %d closed without terminal handler. message '%s' not sent.", s.Id, message)
		}
	}
	s.closeTE()
	s.closeApp(killProcess)
	s.status.Store(uint32(SESS_CLOSED))
	s.notifyStatusListeners(SESS_CLOSING, SESS_CLOSED)
	log.Debugf("Session.Close(). session %d closed", s.Id)
}

func (s *Session) IsTEConnected() bool {
	return s.TeHandler != nil && s.TeHandler.Connected()
}

func (s *Session) IsAppConnected() bool {
	return s.AppHandler != nil && s.AppHandler.Connected()
}

func (s *Session) InTransctionMode() bool {
	return s.GetMode() == SESS_MODE_TRANSACTION
}

func (s *Session) GetEnvVar(varName string) string {
	value, err := util.ProcessEnvVar(s.Process, varName)
	if err != nil {
		log.Debugf("Session.GetEnvVar(). error getting variable for session %d: %v", s.Id, err)
		return ""
	}
	return value
}

func (s *Session) isAppRunning() bool {
	p := s.Process
	if p == nil {
		return false
	}
	running, err := p.IsRunning()
	if err != nil {
		log.Debugf("Session.isAppRunning(). Error checking app is running. session=%d error=%v cmd=%s", s.Id, err, s.CommandLine)
		return false
	} else if !running {
		log.Debugf("Session.isAppRunning(). App is not running. session=%d cmd=%s", s.Id, s.CommandLine)
		return false
	}
	varSessId := s.GetEnvVar(ENV_VAR_AUTH_TOKEN)
	if varSessId == "" {
		log.Errorf("Session.isAppRunning(). Process not found for session %d", s.Id)
		return false
	}
	sessionId, _ := strconv.ParseInt(varSessId, 10, 64)
	if sessionId != s.Id {
		log.Errorf("Session.isAppRunning(). Process session id %d does not match session %d", sessionId, s.Id)
		return false
	}
	return true
}

func (s *Session) SendTE(buffer *buffer.ByteBuffer) error {
	if s.TeHandler != nil {
		return s.TeHandler.Send(buffer)
	}
	log.Debugf("session has no connection with TE. connection closed")
	return io.EOF
}

func (s *Session) SendApp(buffer *buffer.ByteBuffer) error {
	if s.AppHandler != nil {
		return s.AppHandler.Send(buffer)
	}
	log.Debugf("session has no connection with APP. connection closed")
	return io.EOF
}

func (s *Session) AddStatusListener(listener SessionListener) {
	s.statusListeners = append(s.statusListeners, listener)
}

func (s *Session) RemoveStatusListener(listener SessionListener) {
	s.statusListeners = slices.DeleteFunc(s.statusListeners, func(l SessionListener) bool {
		return l == listener
	})
}

func (s *Session) String() string {
	var str strings.Builder
	str.WriteString("session={")
	str.WriteString("id=")
	str.WriteString(strconv.FormatInt(s.Id, 10))
	str.WriteString(", pid=")
	str.WriteString(strconv.FormatInt(s.AppPid, 10))
	str.WriteString(", user='")
	str.WriteString(s.OsUser)
	str.WriteString("', cmd='")
	str.WriteString(s.CommandLine)
	str.WriteString("'}")
	return str.String()
}

func (s *Session) GoString() string {
	return s.String()
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
