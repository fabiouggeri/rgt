package admin

import (
	"rgt-server/buffer"
	"rgt-server/log"
	"rgt-server/protocol"
	"rgt-server/stats"
	"rgt-server/terminal"
	"time"
)

type getSetStats interface {
	BytesReceived() uint64
	BytesSent() uint64
	PacketsReceived() uint64
	PacketsSent() uint64
	SetBytesReceived(uint64)
	SetBytesSent(uint64)
	SetPacketsReceived(uint64)
	SetPacketsSent(uint64)
}

type KillSessionRequest struct {
	protocol.BaseRequest
	sessionId int64
}

type GetSessionsResponse struct {
	protocol.BaseResponse
	sessions []*SessionInfo
}

type GetSessionsResponseV4 GetSessionsResponse

type KillAllSessionsResponse struct {
	protocol.BaseResponse
	killedSessions int32
}

type AdminTerminalRequest struct {
	protocol.BaseRequest
	sessionId   int64
	data        []byte
	requestCode protocol.OperationCode
}

type AdminTerminalResponse struct {
	protocol.BaseResponse
	data *buffer.ByteBuffer
}

type GetSessionStatsRequest struct {
	protocol.BaseRequest
	sessionId int64
}

type GetSessionStatsResponse struct {
	protocol.BaseResponse
	teStats  getSetStats
	appStats getSetStats
}

type SessionInfo struct {
	id              int64
	terminalAddress string
	osUser          string
	appPid          int64
	status          terminal.SessionStatus
	startTime       time.Time
	commandLine     string
}

func init() {
	getSessionsOp := NewOperation(ADM_GET_SESSIONS, "Get sessions")
	getSessionsOp.Version(0).Executor(getSessions)
	getSessionsOp.Version(4).Executor(getSessionsV4)
	NewOperation(ADM_KILL_SESSION, "Kill session").Version(0).Executor(killSession)
	NewOperation(ADM_KILL_ALL_SESSIONS, "Kill all sessions").Version(0).Executor(killAllSessions)
	NewOperation(ADM_GET_SESSION_STATS, "Get session stats").Version(7).Executor(getSessionStats)
	NewOperation(ADM_SEND_TERMINAL_REQUEST, "Send terminal request").Version(0).Executor(sendTerminalRequest)
}

func (r *GetSessionsResponse) FromBuffer(buf *buffer.ByteBuffer) {
	sessionsCount := int(buf.GetInt32())
	r.sessions = make([]*SessionInfo, 0, sessionsCount)
	for range sessionsCount {
		session := &SessionInfo{
			id:              buf.GetInt64(),
			terminalAddress: buf.GetString(),
			osUser:          buf.GetString(),
			appPid:          buf.GetInt64(),
			status:          terminal.SessionStatusFromName(buf.GetString()),
			startTime:       time.UnixMilli(buf.GetInt64()),
		}
		r.sessions = append(r.sessions, session)
	}
}

func (resp *GetSessionsResponseV4) FromBuffer(buf *buffer.ByteBuffer) {
	sessionsCount := int(buf.GetInt32())
	resp.sessions = make([]*SessionInfo, 0, sessionsCount)
	for range sessionsCount {
		session := &SessionInfo{
			id:              buf.GetInt64(),
			terminalAddress: buf.GetString(),
			osUser:          buf.GetString(),
			appPid:          buf.GetInt64(),
			status:          terminal.SessionStatusFromName(buf.GetString()),
			startTime:       time.UnixMilli(buf.GetInt64()),
			commandLine:     buf.GetString(),
		}
		resp.sessions = append(resp.sessions, session)
	}
}

func (resp *GetSessionsResponse) ToBuffer(buf *buffer.ByteBuffer) {
	sessionsCount := len(resp.sessions)
	buf.PutInt32(int32(sessionsCount))
	for _, s := range resp.sessions {
		buf.PutInt64(s.id)
		buf.PutString(s.terminalAddress)
		buf.PutString(s.osUser)
		buf.PutInt64(s.appPid)
		buf.PutString(terminal.SessionStatusName(s.status))
		buf.PutInt64(s.startTime.UnixMilli())
	}
}

func (resp *GetSessionsResponseV4) ToBuffer(buf *buffer.ByteBuffer) {
	sessionsCount := len(resp.sessions)
	buf.PutInt32(int32(sessionsCount))
	for _, s := range resp.sessions {
		buf.PutInt64(s.id)
		buf.PutString(s.terminalAddress)
		buf.PutString(s.osUser)
		buf.PutInt64(s.appPid)
		buf.PutString(terminal.SessionStatusName(s.status))
		buf.PutInt64(s.startTime.UnixMilli())
		buf.PutString(s.commandLine)
	}
}

func getSessions(proto *protocol.OperationVersion[*RequestPack], pack *RequestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("users_sessions_operations.getSessions()")
	sessions := listSessions(pack.handler.service.terminalService.SessionManager())
	resp := &GetSessionsResponse{
		sessions: sessions,
	}
	respBuf := buffer.NewCapacity(uint32(protocol.RESPONSE_HEADER_SIZE + buffer.INT32_FIELD_SIZE + (len(resp.sessions) * 192)))
	protocol.PutResponse(resp, respBuf)
	return respBuf, nil
}

func getSessionsV4(proto *protocol.OperationVersion[*RequestPack], pack *RequestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("users_sessions_operations.getSessions()")
	sessions := listSessions(pack.handler.service.terminalService.SessionManager())
	resp := &GetSessionsResponseV4{
		sessions: sessions,
	}
	respBuf := buffer.NewCapacity(uint32(protocol.RESPONSE_HEADER_SIZE + buffer.INT32_FIELD_SIZE + (len(resp.sessions) * 256)))
	protocol.PutResponse(resp, respBuf)
	return respBuf, nil
}

func listSessions(sessionManager *terminal.SessionManager) []*SessionInfo {
	serverSessions := sessionManager.GetSessions()
	sessions := make([]*SessionInfo, 0, len(serverSessions))
	for _, s := range serverSessions {
		sessions = append(sessions, &SessionInfo{
			id:              s.Id(),
			terminalAddress: s.GetAddress(),
			osUser:          s.GetOSUser(),
			appPid:          s.Pid(),
			status:          s.GetStatus(),
			startTime:       s.GetStartTime(),
			commandLine:     s.CommandLine(),
		})
	}
	return sessions
}

func (r *KillSessionRequest) FromBuffer(buf *buffer.ByteBuffer) {
	r.sessionId = buf.GetInt64()
}

func (r *KillSessionRequest) ToBuffer(buf *buffer.ByteBuffer) {
	buf.PutInt64(r.sessionId)
}

func killSession(proto *protocol.OperationVersion[*RequestPack], pack *RequestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("users_sessions_operations.killSession()")
	if pack.IsReadOnly() {
		return nil, NewError(NOT_ALLOWED_OPERATION, "Operation not allowed in read only session")
	}
	req := &KillSessionRequest{}
	req.FromBuffer(buffer.Wrap(pack.Body()))
	sessionManager := pack.handler.service.terminalService.SessionManager()
	sessionManager.KillSession(req.sessionId, "admin request")
	log.Debugf("Session %d killed")
	return protocol.SuccessResponse(), nil
}

func bufferToKillAllSessionsResponse(buf *buffer.ByteBuffer) *KillAllSessionsResponse {
	return &KillAllSessionsResponse{killedSessions: buf.GetInt32()}
}

func killAllSessionsResponseToBuffer(resp *KillAllSessionsResponse, buf *buffer.ByteBuffer) {
	buf.PutInt32(resp.killedSessions)
}

func killAllSessions(proto *protocol.OperationVersion[*RequestPack], pack *RequestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("users_sessions_operations.killAllSessions()")
	if pack.IsReadOnly() {
		return nil, NewError(NOT_ALLOWED_OPERATION, "Operation not allowed in read only session")
	}
	sessionManager := pack.handler.service.terminalService.SessionManager()
	killedSessions := sessionManager.KillAllSessions("admin request")
	resp := &KillAllSessionsResponse{
		killedSessions: killedSessions,
	}
	respBuf := buffer.NewCapacity(uint32(protocol.RESPONSE_HEADER_SIZE + buffer.INT32_FIELD_SIZE))
	protocol.PutResponse(resp, respBuf)
	return respBuf, nil
}

func (r *AdminTerminalRequest) FromBuffer(buf *buffer.ByteBuffer) {
	r.BaseRequest.OperationCode = ADM_SEND_TERMINAL_REQUEST
	r.sessionId = buf.GetInt64()
	r.requestCode = protocol.OperationCode(buf.GetUInt8())
	r.data = buf.RemainingSlice()
}

func (r *AdminTerminalRequest) ToBuffer(buf *buffer.ByteBuffer) {
	buf.PutInt64(r.sessionId)
	buf.PutUInt8(uint8(r.requestCode))
	buf.Put(r.data)
}

func (r *AdminTerminalResponse) FromBuffer(buf *buffer.ByteBuffer) {
	r.BaseResponse.Code = protocol.ResponseCode(buf.GetUInt16())
	r.data = buffer.Wrap(buf.RemainingSlice())
}

func (r *AdminTerminalResponse) ToBuffer(buf *buffer.ByteBuffer) {
	buf.PutBuffer(r.data)
}

func sendTerminalRequest(proto *protocol.OperationVersion[*RequestPack], pack *RequestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("users_sessions_operations.sendTerminalRequest()")
	req := &AdminTerminalRequest{}
	req.FromBuffer(buffer.Wrap(pack.Body()))
	session := pack.handler.service.terminalService.SessionManager().GetSession(req.sessionId)
	if session == nil {
		return nil, NewError(SESSION_NOT_FOUND, "Session not found")
	}
	if session.TeHandler == nil {
		return nil, NewError(SESSION_NOT_FOUND, "Terminal Connection is down.")
	}
	adminClient := session.TeHandler.RegisterAdminClient(pack.Handler())
	response, err := adminClient.SendRequest(req.requestCode, req.data)
	if err != nil {
		return nil, err
	}
	return buffer.Wrap(response.RemainingSlice()), nil
}

func (r *GetSessionStatsRequest) FromBuffer(buf *buffer.ByteBuffer) {
	r.sessionId = buf.GetInt64()
}

func (r *GetSessionStatsRequest) ToBuffer(buf *buffer.ByteBuffer) {
	buf.PutInt64(r.sessionId)
}

func (r *GetSessionStatsResponse) ToBuffer(buf *buffer.ByteBuffer) {
	// te stats
	buf.PutUInt64(r.teStats.BytesReceived())
	buf.PutUInt64(r.teStats.BytesSent())
	buf.PutUInt64(r.teStats.PacketsReceived())
	buf.PutUInt64(r.teStats.PacketsSent())
	// app stats
	buf.PutUInt64(r.appStats.BytesReceived())
	buf.PutUInt64(r.appStats.BytesSent())
	buf.PutUInt64(r.appStats.PacketsReceived())
	buf.PutUInt64(r.appStats.PacketsSent())
}

func (r *GetSessionStatsResponse) FromBuffer(buf *buffer.ByteBuffer) {
	r.teStats = stats.NewSessionStats()
	r.appStats = stats.NewSessionStats()
	// te stats
	r.teStats.SetBytesReceived(buf.GetUInt64())
	r.teStats.SetBytesSent(buf.GetUInt64())
	r.teStats.SetPacketsReceived(buf.GetUInt64())
	r.teStats.SetPacketsSent(buf.GetUInt64())
	// app stats
	r.appStats.SetBytesReceived(buf.GetUInt64())
	r.appStats.SetBytesSent(buf.GetUInt64())
	r.appStats.SetPacketsReceived(buf.GetUInt64())
	r.appStats.SetPacketsSent(buf.GetUInt64())
}

func getSessionStats(proto *protocol.OperationVersion[*RequestPack], pack *RequestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("users_sessions_operations.getSessionStats()")
	req := &GetSessionStatsRequest{}
	req.FromBuffer(buffer.Wrap(pack.Body()))
	session := pack.handler.service.terminalService.SessionManager().GetSession(req.sessionId)
	if session == nil {
		return nil, NewError(SESSION_NOT_FOUND, "Session not found")
	}
	resp := &GetSessionStatsResponse{}
	if session.TeHandler != nil {
		resp.teStats = session.TeHandler.GetStats()
	} else {
		resp.teStats = stats.NewSessionStats()
	}
	if session.AppHandler != nil {
		resp.appStats = session.AppHandler.GetStats()
	} else {
		resp.appStats = stats.NewSessionStats()
	}
	respBuf := buffer.NewCapacity(64)
	protocol.PutResponse(resp, respBuf)
	return respBuf, nil
}
