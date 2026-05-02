package admin

import (
	"rgt-server/buffer"
	"rgt-server/log"
	"rgt-server/protocol"
	"rgt-server/server"
	"rgt-server/service"
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
	status          server.SessionStatus
	startTime       time.Time
	commandLine     string
}

func init() {
	registerOperation(ADM_GET_SESSIONS, getSessions)
	registerOperation(ADM_KILL_SESSION, killSession)
	registerOperation(ADM_KILL_ALL_SESSIONS, killAllSessions)
	registerOperation(ADM_SEND_TERMINAL_REQUEST, sendTerminalRequest)
	registerOperation(ADM_GET_SESSION_STATS, getSessionStats)
	registerProtocol(ADM_GET_SESSIONS, 0, protocol.New(protocol.BufferToBaseRequest, protocol.BaseRequestToBuffer, bufferToGetSessionsResponse, getSessionsResponseToBuffer))
	registerProtocol(ADM_GET_SESSIONS, 4, protocol.New(protocol.BufferToBaseRequest, protocol.BaseRequestToBuffer, bufferToGetSessionsResponseV4, getSessionsResponseToBufferV4))
	registerProtocol(ADM_GET_SESSION_STATS, 7, protocol.New(bufferToGetSessionStatsRequest, getSessionStatsRequestToBuffer, bufferToGetSessionStatsResponse, getSessionStatsResponseToBuffer))
	registerProtocol(ADM_KILL_SESSION, 0, protocol.New(bufferToKillSessionRequest, KillSessionRequestToBuffer, protocol.BufferToBaseResponse, protocol.BaseResponseToBuffer))
	registerProtocol(ADM_KILL_ALL_SESSIONS, 0, protocol.New(protocol.BufferToBaseRequest, protocol.BaseRequestToBuffer, bufferToKillAllSessionsResponse, killAllSessionsResponseToBuffer))
	registerProtocol(ADM_SEND_TERMINAL_REQUEST, 0, protocol.New(bufferToSendTerminalRequest, sendTerminalRequestToBuffer, bufferToSendTerminalResponse, sendTerminalResponseToBuffer))
}

func bufferToGetSessionsResponse(buf *buffer.ByteBuffer) *GetSessionsResponse {
	sessionsCount := int(buf.GetInt32())
	resp := &GetSessionsResponse{
		sessions: make([]*SessionInfo, 0, sessionsCount),
	}
	for range sessionsCount {
		session := &SessionInfo{
			id:              buf.GetInt64(),
			terminalAddress: buf.GetString(),
			osUser:          buf.GetString(),
			appPid:          buf.GetInt64(),
			status:          server.SessionStatusFromName(buf.GetString()),
			startTime:       time.UnixMilli(buf.GetInt64()),
		}
		resp.sessions = append(resp.sessions, session)
	}
	return resp
}

func bufferToGetSessionsResponseV4(buf *buffer.ByteBuffer) *GetSessionsResponse {
	sessionsCount := int(buf.GetInt32())
	resp := &GetSessionsResponse{sessions: make([]*SessionInfo, 0, sessionsCount)}
	for range sessionsCount {
		session := &SessionInfo{
			id:              buf.GetInt64(),
			terminalAddress: buf.GetString(),
			osUser:          buf.GetString(),
			appPid:          buf.GetInt64(),
			status:          server.SessionStatusFromName(buf.GetString()),
			startTime:       time.UnixMilli(buf.GetInt64()),
			commandLine:     buf.GetString(),
		}
		resp.sessions = append(resp.sessions, session)
	}
	return resp
}

func getSessionsResponseToBuffer(resp *GetSessionsResponse, buf *buffer.ByteBuffer) {
	sessionsCount := len(resp.sessions)
	buf.PutInt32(int32(sessionsCount))
	for _, s := range resp.sessions {
		buf.PutInt64(s.id)
		buf.PutString(s.terminalAddress)
		buf.PutString(s.osUser)
		buf.PutInt64(s.appPid)
		buf.PutString(server.SessionStatusName(s.status))
		buf.PutInt64(s.startTime.UnixMilli())
	}
}

func getSessionsResponseToBufferV4(resp *GetSessionsResponse, buf *buffer.ByteBuffer) {
	sessionsCount := len(resp.sessions)
	buf.PutInt32(int32(sessionsCount))
	for _, s := range resp.sessions {
		buf.PutInt64(s.id)
		buf.PutString(s.terminalAddress)
		buf.PutString(s.osUser)
		buf.PutInt64(s.appPid)
		buf.PutString(server.SessionStatusName(s.status))
		buf.PutInt64(s.startTime.UnixMilli())
		buf.PutString(s.commandLine)
	}
}

func getSessions(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("users_sessions_operations.getSessions()")
	proto, err := findProtocol[*protocol.BaseRequest, *GetSessionsResponse](ADM_GET_SESSIONS, pack.handler.protocolVersion)
	if err != nil {
		return nil, err
	}
	srv := pack.handler.service.server
	sessions := make([]*SessionInfo, 0, 10)
	for _, s := range srv.GetSessions() {
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
	resp := &GetSessionsResponse{sessions: sessions}
	respBuf := buffer.New()
	proto.PutResponse(resp, respBuf)
	return respBuf, nil
}

func bufferToKillSessionRequest(buf *buffer.ByteBuffer) *KillSessionRequest {
	return &KillSessionRequest{sessionId: buf.GetInt64()}
}

func KillSessionRequestToBuffer(req *KillSessionRequest, buf *buffer.ByteBuffer) {
	buf.PutInt64(req.sessionId)
}

func killSession(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("users_sessions_operations.killSession()")
	if pack.handler.readOnly {
		return nil, NewError(NOT_ALLOWED_OPERATION, "Operation not allowed in read only session")
	}
	proto, err := findProtocol[*KillSessionRequest, *protocol.BaseResponse](ADM_KILL_SESSION, pack.handler.protocolVersion)
	if err != nil {
		return nil, err
	}
	bufReq := buffer.Wrap(pack.body)
	req := proto.GetRequest(bufReq)
	pack.handler.service.server.KillSession(req.sessionId, "admin request")
	log.Debugf("Session %d killed")
	return SuccessAdminResponse(), nil
}

func bufferToKillAllSessionsResponse(buf *buffer.ByteBuffer) *KillAllSessionsResponse {
	return &KillAllSessionsResponse{killedSessions: buf.GetInt32()}
}

func killAllSessionsResponseToBuffer(resp *KillAllSessionsResponse, buf *buffer.ByteBuffer) {
	buf.PutInt32(resp.killedSessions)
}

func killAllSessions(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("users_sessions_operations.killAllSessions()")
	if pack.handler.readOnly {
		return nil, NewError(NOT_ALLOWED_OPERATION, "Operation not allowed in read only session")
	}
	proto, err := findProtocol[*protocol.BaseRequest, *KillAllSessionsResponse](ADM_KILL_ALL_SESSIONS, pack.handler.protocolVersion)
	if err != nil {
		return nil, err
	}
	srv := pack.handler.service.server
	killedSessions := srv.KillAllSessions("admin request")
	resp := &KillAllSessionsResponse{killedSessions: killedSessions}
	respBuf := buffer.New()
	proto.PutResponse(resp, respBuf)
	return respBuf, nil
}

func sendTerminalRequest(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("users_sessions_operations.sendTerminalRequest()")
	proto, err := findProtocol[*AdminTerminalRequest, *AdminTerminalResponse](ADM_SEND_TERMINAL_REQUEST, pack.handler.protocolVersion)
	if err != nil {
		return nil, err
	}
	bufReq := buffer.Wrap(pack.body)
	req := proto.GetRequest(bufReq)
	srv := pack.handler.service.server
	session := srv.GetSession(req.sessionId)
	if session == nil {
		return nil, NewError(SESSION_NOT_FOUND, "Session not found")
	}
	terminalSession, ok := session.(*terminal.TerminalSession)
	if !ok {
		return nil, NewError(SESSION_NOT_FOUND, "Session is not a terminal session")
	}
	if terminalSession.TeHandler == nil {
		return nil, NewError(SESSION_NOT_FOUND, "Terminal Connection is down.")
	}
	admin := terminalSession.TeHandler.RegisterAdminClient(pack.handler)
	response, err := admin.SendRequest(req.requestCode, req.data)
	if err != nil {
		return nil, err
	}
	return buffer.Wrap(response.RemainingSlice()), nil
}

func bufferToSendTerminalRequest(buf *buffer.ByteBuffer) *AdminTerminalRequest {
	return &AdminTerminalRequest{
		BaseRequest: protocol.BaseRequest{OperationCode: ADM_SEND_TERMINAL_REQUEST},
		sessionId:   buf.GetInt64(),
		requestCode: protocol.OperationCode(buf.GetUInt8()),
		data:        buf.RemainingSlice()}
}

func sendTerminalRequestToBuffer(req *AdminTerminalRequest, buf *buffer.ByteBuffer) {
	buf.PutInt64(req.sessionId)
	buf.PutUInt8(uint8(req.requestCode))
	buf.Put(req.data)
}

func sendTerminalResponseToBuffer(response *AdminTerminalResponse, buf *buffer.ByteBuffer) {
	buf.PutBuffer(response.data)
}

func bufferToSendTerminalResponse(buf *buffer.ByteBuffer) *AdminTerminalResponse {
	return &AdminTerminalResponse{
		BaseResponse: protocol.BaseResponse{Code: protocol.ResponseCode(buf.GetUInt16())},
		data:         buffer.Wrap(buf.RemainingSlice()),
	}
}
func bufferToGetSessionStatsRequest(buf *buffer.ByteBuffer) *GetSessionStatsRequest {
	return &GetSessionStatsRequest{sessionId: buf.GetInt64()}
}

func getSessionStatsRequestToBuffer(req *GetSessionStatsRequest, buf *buffer.ByteBuffer) {
	buf.PutInt64(req.sessionId)
}

func getSessionStatsResponseToBuffer(resp *GetSessionStatsResponse, buf *buffer.ByteBuffer) {
	// te stats
	buf.PutUInt64(resp.teStats.BytesReceived())
	buf.PutUInt64(resp.teStats.BytesSent())
	buf.PutUInt64(resp.teStats.PacketsReceived())
	buf.PutUInt64(resp.teStats.PacketsSent())
	// app stats
	buf.PutUInt64(resp.appStats.BytesReceived())
	buf.PutUInt64(resp.appStats.BytesSent())
	buf.PutUInt64(resp.appStats.PacketsReceived())
	buf.PutUInt64(resp.appStats.PacketsSent())
}

func bufferToGetSessionStatsResponse(buf *buffer.ByteBuffer) *GetSessionStatsResponse {
	resp := &GetSessionStatsResponse{
		teStats:  stats.NewSessionStats(),
		appStats: stats.NewSessionStats(),
	}
	// te stats
	resp.teStats.SetBytesReceived(buf.GetUInt64())
	resp.teStats.SetBytesSent(buf.GetUInt64())
	resp.teStats.SetPacketsReceived(buf.GetUInt64())
	resp.teStats.SetPacketsSent(buf.GetUInt64())
	// app stats
	resp.appStats.SetBytesReceived(buf.GetUInt64())
	resp.appStats.SetBytesSent(buf.GetUInt64())
	resp.appStats.SetPacketsReceived(buf.GetUInt64())
	resp.appStats.SetPacketsSent(buf.GetUInt64())
	return resp
}
func handlerStats(handler service.TerminalConnectionHandler) *stats.SessionStats {
	if handler == nil {
		return stats.NewSessionStats()
	}
	return handler.GetStats()
}

func getSessionStats(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("users_sessions_operations.getSessionStats()")
	proto, err := findProtocol[*GetSessionStatsRequest, *GetSessionStatsResponse](ADM_GET_SESSION_STATS, pack.handler.protocolVersion)
	if err != nil {
		return nil, err
	}
	bufReq := buffer.Wrap(pack.body)
	req := proto.GetRequest(bufReq)
	session := pack.handler.service.server.GetSession(req.sessionId)
	if session == nil {
		return nil, NewError(SESSION_NOT_FOUND, "Session not found")
	}
	sessionTerminal, ok := session.(*terminal.TerminalSession)
	if !ok {
		return nil, NewError(SERVER_ERROR, "Session is not a terminal session")
	}
	resp := &GetSessionStatsResponse{
		teStats:  handlerStats(sessionTerminal.TeHandler),
		appStats: handlerStats(sessionTerminal.AppHandler),
	}
	respBuf := buffer.NewCapacity(64)
	proto.PutResponse(resp, respBuf)
	return respBuf, nil
}
