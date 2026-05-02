package terminal

import (
	"fmt"
	"io"
	"rgt-server/buffer"
	"rgt-server/config"
	"rgt-server/log"
	"rgt-server/option"
	"rgt-server/server"
	"rgt-server/service"
	"rgt-server/util"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

type TerminalSession struct {
	id                   int64
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
	commandLine          string
	Process              *process.Process
	Options              *option.Options
	Mode                 option.TypedOption[server.SessionMode]
	TimeoutEnabled       option.TypedOption[bool]
	status               atomic.Uint32
	SessionType          server.SessionType
	closing              atomic.Bool
	statusListeners      []server.SessionListener
}

var _ server.Session = &TerminalSession{}

func newSession(teHandler service.TerminalConnectionHandler, sessionType server.SessionType, teAddr string, username string, osUser string, commandLine string) *TerminalSession {
	now := time.Now()
	s := &TerminalSession{
		id:                   server.NextSessionId(),
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
		Mode:                 option.NewUint(server.SESS_MODE_NORMAL, "mode"),
		Options:              option.NewOptions(),
		SessionType:          sessionType,
		statusListeners:      make([]server.SessionListener, 0),
	}
	s.closing.Store(false)
	s.Options.Add(s.Mode)
	s.Options.Add(s.TimeoutEnabled)
	log.Infof("terminal.newSession(). session=%d type=%v addr=%s user=%s cmd=[%s]", s.id, sessionType, teAddr, osUser, commandLine)
	return s
}

func (s *TerminalSession) Id() int64 {
	return s.id
}

func (s *TerminalSession) SetAppHandler(appHandler service.TerminalConnectionHandler) {
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

func (s *TerminalSession) notifyStatusListeners(oldStatus, newStatus server.SessionStatus) {
	for _, listener := range s.statusListeners {
		listener.StatusChange(s, oldStatus, newStatus)
	}
}

func (s *TerminalSession) ChangeStatus(oldStatus, newStatus server.SessionStatus) error {
	if s.closing.Load() {
		return nil
	}
	if oldStatus == newStatus {
		return fmt.Errorf("Session %d already in status %s", s.id, server.SessionStatusName(oldStatus))
	}
	previousStatus := server.SessionStatus(s.status.Swap(uint32(newStatus)))
	if previousStatus != oldStatus {
		return fmt.Errorf("Session %d with unexpected status %s. Expected %s to change to %s", s.id,
			server.SessionStatusName(previousStatus), server.SessionStatusName(oldStatus), server.SessionStatusName(newStatus))
	}
	s.notifyStatusListeners(oldStatus, newStatus)
	return nil
}

func (s *TerminalSession) SetStatus(status server.SessionStatus) {
	if !s.closing.Load() {
		oldStatus := server.SessionStatus(s.status.Swap(uint32(status)))
		if oldStatus != status {
			s.notifyStatusListeners(oldStatus, status)
		}
	}
}

func (s *TerminalSession) GetType() server.SessionType {
	return s.SessionType
}

func (s *TerminalSession) GetStatus() server.SessionStatus {
	return server.SessionStatus(s.status.Load())
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

func (s *TerminalSession) GetMode() server.SessionMode {
	return s.Mode.Get()
}

func (s *TerminalSession) SetMode(mode server.SessionMode) {
	s.Mode.Set(mode)
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
		log.Infof("TerminalSession.killAppProcess(). Killed process. session=%d reason=%s cmd=%s", s.id, reason)
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
	if !s.closing.CompareAndSwap(false, true) {
		return false
	}
	log.Debugf("TerminalSession.Close(). closing session %d", s.id)
	if s.GetStatus() != server.SESS_CLOSE_REQUEST && s.InTransctionMode() {
		log.Infof("TerminalSession.Close(). Not closed. Session %d in transaction mode.", s.id)
		s.TransactionStartTime = time.Now()
		return false
	}
	oldStatus := server.SessionStatus(s.status.Swap(uint32(server.SESS_CLOSING)))
	if oldStatus != server.SESS_CLOSING {
		s.notifyStatusListeners(oldStatus, server.SESS_CLOSING)
	}
	if message != "" {
		if s.TeHandler != nil {
			s.TeHandler.SendLogout(message)
		} else {
			log.Debugf("TerminalSession.closeWithMessage(). Unknown error. Session %d closed without terminal handler. message '%s' not sent.", s.id, message)
		}
	}
	s.closeTE()
	s.closeApp(killProcess)
	s.status.Store(uint32(server.SESS_CLOSED))
	s.notifyStatusListeners(server.SESS_CLOSING, server.SESS_CLOSED)
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
	return s.GetMode() == server.SESS_MODE_TRANSACTION
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
	varSessId := s.GetEnvVar(server.ENV_VAR_AUTH_TOKEN)
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

func (s *TerminalSession) AddStatusListener(listener server.SessionListener) {
	s.statusListeners = append(s.statusListeners, listener)
}

func (s *TerminalSession) RemoveStatusListener(listener server.SessionListener) {
	s.statusListeners = slices.DeleteFunc(s.statusListeners, func(l server.SessionListener) bool {
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

func (s *TerminalSession) timeoutAppLaunch(conf *config.ServerConfig) bool {
	if s.GetStatus() != server.SESS_NEW {
		return false
	}
	if s.AppHandler != nil {
		return false
	}
	if conf.AppLaunchTimeout().Get() == 0 {
		return false
	}
	if !s.StartTime.Equal(s.AppLaunchTime) {
		return false
	}
	return time.Since(s.StartTime) >= conf.AppLaunchTimeout().Get()
}

func (s *TerminalSession) timeoutAppLogin(conf *config.ServerConfig) bool {
	if s.GetStatus() != server.SESS_CONNECTING {
		return false
	}
	if s.AppHandler != nil {
		return false
	}
	if s.SessionType == server.SESS_TYPE_STANDALONE {
		return false
	}
	if conf.AppLoginTimeout().Get() == 0 {
		return false
	}
	if s.StartTime.Equal(s.AppLaunchTime) {
		return false
	}
	if !s.StartTime.Equal(s.AppLoginTime) {
		return false
	}
	return time.Since(s.AppLaunchTime) >= conf.AppLoginTimeout().Get()
}

func (s *TerminalSession) appIsRunning() bool {
	if s.GetStatus() != server.SESS_READY {
		return true
	}
	if s.Process == nil {
		return false
	}
	running, err := s.Process.IsRunning()
	if err != nil {
		log.Debugf("TerminalSession.AppIsRunning(). Error checking app is running: %v", err)
	}
	return running
}

func (s *TerminalSession) idleTimeout(conf *config.ServerConfig) bool {
	if conf.SessionIdleTimeout().Get() == 0 || !s.TimeoutEnabled.Get() || s.InTransctionMode() {
		return false
	}
	return (s.AppHandler != nil && time.Since(s.AppHandler.GetLastAppOperationTime()) > conf.SessionIdleTimeout().Get()) ||
		(s.TeHandler != nil && time.Since(s.TeHandler.GetLastAppOperationTime()) > conf.SessionIdleTimeout().Get())
}

func (s *TerminalSession) communicationLackTimeout(conf *config.ServerConfig) bool {
	if conf.AppLackTimeout().Get() == 0 {
		return false
	}
	return s.AppHandler != nil && time.Since(s.AppHandler.GetLastDataReadTime()) > conf.AppLackTimeout().Get()
}

func (s *TerminalSession) timeoutLostTransactionSession(conf *config.ServerConfig) bool {
	return s.InTransctionMode() && s.GetStatus() != server.SESS_READY && time.Since(s.TransactionStartTime) > conf.AppTransactionTimeout().Get()
}

func (s *TerminalSession) CloseConditionally(conf *config.ServerConfig) bool {
	if s.timeoutAppLaunch(conf) {
		s.sendLogoutToTerminal("session closed because application was not launched")
		return true
	} else if s.timeoutAppLogin(conf) {
		s.sendLogoutToTerminal("application killed because did not respond")
		return true
	} else if !s.appIsRunning() {
		s.sendLogoutToTerminal("application closed")
		return true
	} else if s.idleTimeout(conf) {
		s.sendLogoutToTerminal("application closed by inactivity")
		return true
	} else if s.communicationLackTimeout(conf) {
		s.sendLogoutToTerminal("application killed by communication lack")
		return true
	} else if s.timeoutLostTransactionSession(conf) {
		s.killAppProcess("lost transaction session")
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
	log.Infof("TerminalSession.sendLogoutToTerminal() id=%d message='%s'", s.Id(), msg)
}
