package terminal

import (
	"fmt"
	"io"
	"rgt-server/buffer"
	"rgt-server/log"
	"rgt-server/option"
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

type TerminalSession struct {
	id                   int64
	TeHandler            *TerminalHandler
	AppHandler           *TerminalHandler
	TerminalUser         string
	TerminalAddress      string
	AppAddress           string
	OsUser               string
	AppPid               int64
	StartTime            time.Time
	AppLaunchTime        time.Time
	AppLoginTime         time.Time
	TransactionStartTime time.Time
	commandLine          string
	Process              *process.Process
	Options              *option.Options
	Mode                 option.TypedOption[SessionMode]
	TimeoutEnabled       option.TypedOption[bool]
	status               atomic.Uint32
	SessionType          SessionType
	statusListeners      []SessionListener
}

type SessionListener interface {
	StatusChange(session *TerminalSession, oldStatus SessionStatus, newStatus SessionStatus)
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

var (
	sessionCount int64 = 0
)

func newSession(teHandler *TerminalHandler, sessionType SessionType, teAddr string, username string, osUser string, commandLine string) *TerminalSession {
	now := time.Now()
	s := &TerminalSession{
		id:                   atomic.AddInt64(&sessionCount, 1),
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
		commandLine:          commandLine,
		TimeoutEnabled:       option.NewBool(false, "timeoutenabled"),
		Mode:                 option.NewUint(SESS_MODE_NORMAL, "mode"),
		Options:              option.NewOptions(),
		SessionType:          sessionType,
		statusListeners:      make([]SessionListener, 0),
	}
	s.Mode.SetHook(s.modeChange)
	s.Options.Add(s.Mode)
	s.Options.Add(s.TimeoutEnabled)
	log.Infof("terminal.newSession(). handler=%d session=%d type=%v addr=%s user=%s cmd=[%s]", teHandler.id, s.id, sessionType, teAddr, osUser, commandLine)
	return s
}

func (s *TerminalSession) Id() int64 {
	return s.id
}

func (s *TerminalSession) SetAppHandler(appHandler *TerminalHandler) {
	s.AppHandler = appHandler
}

func (s *TerminalSession) GetAddress() string {
	return s.TerminalAddress
}

func (s *TerminalSession) SetTerminalAddress(addr string) {
	s.TerminalAddress = addr
}

func (s *TerminalSession) GetUser() string {
	return s.TerminalUser
}

func (s *TerminalSession) SetTerminalUser(usr string) {
	s.TerminalUser = usr
}

func (s *TerminalSession) SetAppAddress(addr string) {
	s.AppAddress = addr
}

func (s *TerminalSession) SetAppPid(pid int64) {
	s.AppPid = pid
}

func (s *TerminalSession) GetOSUser() string {
	return s.OsUser
}

func (s *TerminalSession) SetOsUser(user string) {
	s.OsUser = user
}

func (s *TerminalSession) CommandLine() string {
	return s.commandLine
}

func (s *TerminalSession) Pid() int64 {
	return s.AppPid
}

func (s *TerminalSession) notifyStatusListeners(oldStatus, newStatus SessionStatus) {
	for _, listener := range s.statusListeners {
		listener.StatusChange(s, oldStatus, newStatus)
	}
}

func (s *TerminalSession) ChangeStatus(oldStatus, newStatus SessionStatus) error {
	if oldStatus == newStatus {
		return fmt.Errorf("New status (%s) is the same of expected status (%s) for session %d", newStatus, oldStatus, s.id)
	}
	if !s.status.CompareAndSwap(uint32(oldStatus), uint32(newStatus)) {
		previousStatus := SessionStatus(s.status.Load())
		return fmt.Errorf("Session %d with unexpected status %s. Expected %s to change to %s", s.id, previousStatus, oldStatus, newStatus)
	}
	log.Debugf("Session %d changed status from %s to %s", s.id, oldStatus, newStatus)
	s.notifyStatusListeners(oldStatus, newStatus)
	return nil
}

func (s *TerminalSession) TryChangeStatus(oldStatus, newStatus SessionStatus) bool {
	if oldStatus == newStatus {
		log.Debugf("New status (%s) is the same of expected status (%s) for session %d", newStatus, oldStatus, s.id)
		return false
	}
	if !s.status.CompareAndSwap(uint32(oldStatus), uint32(newStatus)) {
		previousStatus := SessionStatus(s.status.Load())
		log.Debugf("Session %d with unexpected status %s. Expected %s to change to %s", s.id, previousStatus, oldStatus, newStatus)
		return false
	}
	s.notifyStatusListeners(oldStatus, newStatus)
	return true
}

func (s *TerminalSession) TrySetStatus(newStatus SessionStatus) bool {
	previousStatus := SessionStatus(s.status.Swap(uint32(newStatus)))
	if previousStatus != newStatus {
		s.notifyStatusListeners(previousStatus, newStatus)
		return true
	} else {
		log.Debugf("Session %d already in status %s.", s.id, newStatus)
	}
	return false
}

func (s *TerminalSession) SetStatus(status SessionStatus) {
	// SESS_CLOSING can only be changed from Close() method
	if s.IsClosing() {
		return
	}
	oldStatus := SessionStatus(s.status.Swap(uint32(status)))
	if oldStatus != status {
		s.notifyStatusListeners(oldStatus, status)
	}
}

func (s *TerminalSession) GetType() SessionType {
	return s.SessionType
}

func (s *TerminalSession) GetStatus() SessionStatus {
	return SessionStatus(s.status.Load())
}

func (s *TerminalSession) IsClosing() bool {
	return SessionStatus(s.status.Load()) == SESS_CLOSING
}

func (s *TerminalSession) GetStartTime() time.Time {
	return s.StartTime
}

func (s *TerminalSession) SetStartTime(startTime time.Time) {
	s.StartTime = startTime
}

func (s *TerminalSession) SetAppLaunchTime(appLaunchTime time.Time) {
	s.AppLaunchTime = appLaunchTime
}

func (s *TerminalSession) SetAppLoginTime(appLoginTime time.Time) {
	s.AppLoginTime = appLoginTime
}

func (s *TerminalSession) SetCommandLine(cmd string) {
	s.commandLine = cmd
}

func (s *TerminalSession) SetProcess(process *process.Process) {
	s.Process = process
}

func (s *TerminalSession) GetMode() SessionMode {
	return s.Mode.Get()
}

func (s *TerminalSession) SetMode(mode SessionMode) {
	s.Mode.Set(mode)
}

func (s *TerminalSession) modeChange(newMode SessionMode) {
	if s.GetMode() != SESS_MODE_TRANSACTION && newMode == SESS_MODE_TRANSACTION {
		s.TransactionStartTime = time.Now()
	} else {
		s.TransactionStartTime = time.Time{}
	}
}

func (s *TerminalSession) killAppProcess(reason string) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("unknown error in server(TerminalSession.killAppProcess): %v\n%s", err, util.FullStack())
		}
	}()
	if !s.isAppRunning() {
		return
	}
	if err := util.KillProcessRecursive(s.Process, "session killed"); err != nil {
		log.Debugf("TerminalSession.killAppProcess(). Error killing process. session=%d error=%v cmd=%s", s.id, err, s.commandLine)
	} else {
		log.Infof("TerminalSession.killAppProcess(). Killed process. session=%d reason=%s cmd=%s", s.id, reason, s.commandLine)
	}
}

func (s *TerminalSession) closeTE() {
	if s.TeHandler != nil {
		th := s.TeHandler
		s.TeHandler = nil
		th.Close()
	}
}

func (s *TerminalSession) closeApp(killProcess bool) {
	if s.AppHandler != nil {
		ah := s.AppHandler
		s.AppHandler = nil
		ah.Close()
		if killProcess && !s.InTransctionMode() {
			s.killAppProcess("session closed")
		}
	} else if killProcess && !s.InTransctionMode() {
		s.killAppProcess("session closed")
	}
}

func (s *TerminalSession) Close(killProcess bool, message string) bool {
	// Only the first call will proceed to close the session.
	if s.IsClosing() {
		return false
	}
	log.Debugf("TerminalSession.Close(). closing session %d", s.id)
	if s.GetStatus() != SESS_CLOSE_REQUEST && s.InTransctionMode() {
		log.Infof("TerminalSession.Close(). Not closed. Session %d in transaction mode.", s.id)
		return false
	}
	oldStatus := SessionStatus(s.status.Swap(uint32(SESS_CLOSING)))
	if oldStatus == SESS_CLOSING {
		log.Errorf("TerminalSession.Close(). Session %d already closing.", s.id)
		return false
	}
	s.notifyStatusListeners(oldStatus, SESS_CLOSING)
	if message != "" {
		if s.TeHandler != nil {
			s.TeHandler.SendLogout(message)
		} else {
			log.Debugf("TerminalSession.Close(). Unknown error. Session %d closed without terminal handler. message '%s' not sent.", s.id, message)
		}
	}
	s.closeTE()
	s.closeApp(killProcess)
	s.status.Store(uint32(SESS_CLOSED))
	s.notifyStatusListeners(SESS_CLOSING, SESS_CLOSED)
	log.Debugf("TerminalSession.Close(). session %d closed", s.id)
	return true
}

func (s *TerminalSession) IsTEConnected() bool {
	return s.TeHandler != nil && s.TeHandler.Connected()
}

func (s *TerminalSession) IsAppConnected() bool {
	return s.AppHandler != nil && s.AppHandler.Connected()
}

func (s *TerminalSession) InTransctionMode() bool {
	return s.GetMode() == SESS_MODE_TRANSACTION
}

func (s *TerminalSession) GetEnvVar(varName string) string {
	value, err := util.ProcessEnvVar(s.Process, varName)
	if err != nil {
		log.Debugf("TerminalSession.GetEnvVar(). error getting variable for session %d: %v", s.id, err)
		return ""
	}
	return value
}

func (s *TerminalSession) isAppRunning() bool {
	p := s.Process
	if p == nil {
		return false
	}
	running, err := p.IsRunning()
	if err != nil {
		log.Debugf("TerminalSession.isAppRunning(). Error checking app is running. session=%d error=%v cmd=%s", s.id, err, s.commandLine)
		return false
	} else if !running {
		log.Debugf("TerminalSession.isAppRunning(). App is not running. session=%d cmd=%s", s.id, s.commandLine)
		return false
	}
	varSessId := s.GetEnvVar(ENV_VAR_AUTH_TOKEN)
	if varSessId == "" {
		log.Errorf("TerminalSession.isAppRunning(). Process not found for session %d", s.id)
		return false
	}
	sessionId, _ := strconv.ParseInt(varSessId, 10, 64)
	if sessionId != s.id {
		log.Errorf("TerminalSession.isAppRunning(). Process session id %d does not match session %d", sessionId, s.id)
		return false
	}
	return true
}

func (s *TerminalSession) SendTE(buffer *buffer.ByteBuffer) error {
	if s.TeHandler != nil {
		return s.TeHandler.Send(buffer)
	}
	log.Debugf("session has no connection with TE. connection closed")
	return io.EOF
}

func (s *TerminalSession) SendApp(buffer *buffer.ByteBuffer) error {
	if s.AppHandler != nil {
		return s.AppHandler.Send(buffer)
	}
	log.Debugf("session has no connection with APP. connection closed")
	return io.EOF
}

func (s *TerminalSession) AddStatusListener(listener SessionListener) {
	s.statusListeners = append(s.statusListeners, listener)
}

func (s *TerminalSession) RemoveStatusListener(listener SessionListener) {
	s.statusListeners = slices.DeleteFunc(s.statusListeners, func(l SessionListener) bool {
		return l == listener
	})
}

func (s *TerminalSession) String() string {
	var str strings.Builder
	str.WriteString("session={")
	str.WriteString("id=")
	str.WriteString(strconv.FormatInt(s.id, 10))
	str.WriteString(", pid=")
	str.WriteString(strconv.FormatInt(s.AppPid, 10))
	str.WriteString(", user='")
	str.WriteString(s.OsUser)
	str.WriteString("', cmd='")
	str.WriteString(s.commandLine)
	str.WriteString("'}")
	return str.String()
}

func (s *TerminalSession) GoString() string {
	return s.String()
}

func (s *TerminalSession) timeoutAppLaunch(appLaunchTimeout time.Duration) bool {
	if appLaunchTimeout == 0 {
		return false
	}
	if s.AppHandler != nil {
		return false
	}
	if s.GetStatus() != SESS_NEW {
		return false
	}
	if !s.StartTime.Equal(s.AppLaunchTime) {
		return false
	}
	return time.Since(s.StartTime) >= appLaunchTimeout
}

func (s *TerminalSession) timeoutAppLogin(appLoginTimeout time.Duration) bool {
	if s.AppHandler != nil {
		return false
	}
	if s.SessionType == SESS_TYPE_STANDALONE {
		return false
	}
	if appLoginTimeout == 0 {
		return false
	}
	status := s.GetStatus()
	if status != SESS_LAUNCHING_APP && status != SESS_CONNECTING {
		return false
	}
	if s.StartTime.Equal(s.AppLaunchTime) {
		return false
	}
	return time.Since(s.AppLaunchTime) >= appLoginTimeout
}

func (s *TerminalSession) appIsRunning() bool {
	if s.Process == nil {
		return false
	}
	if s.GetStatus() != SESS_READY {
		return true
	}
	running, err := s.Process.IsRunning()
	if err != nil {
		log.Debugf("TerminalSession.AppIsRunning(). Error checking app is running: %v", err)
	}
	return running
}

func (s *TerminalSession) idleTimeout(idleTimeout time.Duration) bool {
	if idleTimeout == 0 || !s.TimeoutEnabled.Get() || s.InTransctionMode() {
		return false
	}
	return (s.AppHandler != nil && time.Since(s.AppHandler.GetLastAppOperationTime()) > idleTimeout) ||
		(s.TeHandler != nil && time.Since(s.TeHandler.GetLastAppOperationTime()) > idleTimeout)
}

func (s *TerminalSession) communicationLackTimeout(appLackTimeout time.Duration) bool {
	if appLackTimeout == 0 {
		return false
	}
	return s.AppHandler != nil && time.Since(s.AppHandler.GetLastDataReadTime()) > appLackTimeout
}

func (s *TerminalSession) timeoutLostTransactionSession(appTransactionTimeout time.Duration) bool {
	if appTransactionTimeout == 0 {
		return false
	}
	return s.InTransctionMode() && s.GetStatus() != SESS_READY && time.Since(s.TransactionStartTime) > appTransactionTimeout
}

func (s *TerminalSession) loginTimeExceeded(maxLoginTime time.Duration) bool {
	if s.GetStatus() >= SESS_READY {
		return false
	}
	if maxLoginTime <= 0 {
		return false
	}
	if timeout := time.Since(s.StartTime); timeout >= maxLoginTime {
		log.Debugf("SessionManager.loginTimeExceeded(). Session %d exceeded login time %v", s.Id(), timeout)
		return true
	}
	return false
}

func (s *TerminalSession) handlePanic(message string) {
	if err := recover(); err != nil {
		log.Errorf("%s: %v\n%s", message, err, util.FullStack())
	}
}

func (s *TerminalSession) sendLogoutToTerminal(msg string) {
	defer s.handlePanic("unknown error in server(TerminalSession.sendLogoutToTerminal)")
	s.Close(true, msg)
	log.Infof("TerminalSession.sendLogoutToTerminal() session=%d message='%s'", s.Id(), msg)
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
	return status.String()
}

func (status SessionStatus) String() string {
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
