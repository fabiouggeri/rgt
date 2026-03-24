package admin

import (
	"rgt-server/buffer"
	"rgt-server/log"
	"rgt-server/protocol"
	"rgt-server/server"
	"time"
)

type KillSessionRequest struct {
	protocol.BaseRequest
	sessionId int64
}

type GetSessionsResponse struct {
	protocol.BaseResponse
	sessions []*server.Session
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

func init() {
	registerOperation(ADM_GET_SESSIONS, getSessions)
	registerOperation(ADM_KILL_SESSION, killSession)
	registerOperation(ADM_KILL_ALL_SESSIONS, killAllSessions)
	registerOperation(ADM_SEND_TERMINAL_REQUEST, sendTerminalRequest)
	registerProtocol(ADM_GET_SESSIONS, 0, protocol.New(protocol.BufferToBaseRequest, protocol.BaseRequestToBuffer, bufferToGetSessionsResponse, getSessionsResponseToBuffer))
	registerProtocol(ADM_GET_SESSIONS, 4, protocol.New(protocol.BufferToBaseRequest, protocol.BaseRequestToBuffer, bufferToGetSessionsResponseV4, getSessionsResponseToBufferV4))
	registerProtocol(ADM_KILL_SESSION, 0, protocol.New(bufferToKillSessionRequest, KillSessionRequestToBuffer, protocol.BufferToBaseResponse, protocol.BaseResponseToBuffer))
	registerProtocol(ADM_KILL_ALL_SESSIONS, 0, protocol.New(protocol.BufferToBaseRequest, protocol.BaseRequestToBuffer, bufferToKillAllSessionsResponse, killAllSessionsResponseToBuffer))
	registerProtocol(ADM_SEND_TERMINAL_REQUEST, 0, protocol.New(bufferToSendTerminalRequest, sendTerminalRequestToBuffer, bufferToSendTerminalResponse, sendTerminalResponseToBuffer))
}

func bufferToGetSessionsResponse(buf *buffer.ByteBuffer) *GetSessionsResponse {
	sessionsCount := int(buf.GetInt32())
	resp := &GetSessionsResponse{sessions: make([]*server.Session, 0, sessionsCount)}
	for i := 0; i < sessionsCount; i++ {
		session := &server.Session{
			Id:              buf.GetInt64(),
			TerminalAddress: buf.GetString(),
			OsUser:          buf.GetString(),
			AppPid:          buf.GetInt64()}
		session.SetStatus(server.SessionStatusFromName(buf.GetString()))
		session.StartTime = time.UnixMilli(buf.GetInt64())
		resp.sessions = append(resp.sessions, session)
	}
	return resp
}

func bufferToGetSessionsResponseV4(buf *buffer.ByteBuffer) *GetSessionsResponse {
	sessionsCount := int(buf.GetInt32())
	resp := &GetSessionsResponse{sessions: make([]*server.Session, 0, sessionsCount)}
	for i := 0; i < sessionsCount; i++ {
		session := &server.Session{Id: buf.GetInt64(),
			TerminalAddress: buf.GetString(),
			OsUser:          buf.GetString(),
			AppPid:          buf.GetInt64()}
		session.SetStatus(server.SessionStatusFromName(buf.GetString()))
		session.StartTime = time.UnixMilli(buf.GetInt64())
		session.CommandLine = buf.GetString()
		resp.sessions = append(resp.sessions, session)
	}
	return resp
}

func getSessionsResponseToBuffer(resp *GetSessionsResponse, buf *buffer.ByteBuffer) {
	sessionsCount := len(resp.sessions)
	buf.PutInt32(int32(sessionsCount))
	for _, s := range resp.sessions {
		buf.PutInt64(s.Id)
		buf.PutString(s.TerminalAddress)
		buf.PutString(s.OsUser)
		buf.PutInt64(s.AppPid)
		buf.PutString(server.SessionStatusName(s.GetStatus()))
		buf.PutInt64(s.StartTime.UnixMilli())
	}
}

func getSessionsResponseToBufferV4(resp *GetSessionsResponse, buf *buffer.ByteBuffer) {
	sessionsCount := len(resp.sessions)
	buf.PutInt32(int32(sessionsCount))
	for _, s := range resp.sessions {
		buf.PutInt64(s.Id)
		buf.PutString(s.TerminalAddress)
		buf.PutString(s.OsUser)
		buf.PutInt64(s.AppPid)
		buf.PutString(server.SessionStatusName(s.GetStatus()))
		buf.PutInt64(s.StartTime.UnixMilli())
		buf.PutString(s.CommandLine)
	}
}

func getSessions(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("users_sessions_operations.getSessions()")
	proto, err := findProtocol[*protocol.BaseRequest, *GetSessionsResponse](ADM_GET_SESSIONS, pack.handler.protocolVersion)
	if err != nil {
		return nil, err
	}
	srv := pack.handler.service.server
	resp := &GetSessionsResponse{sessions: srv.GetSessions()}
	respBuf := buffer.New()
	proto.PutResponse(resp, respBuf)
	return respBuf, nil
}

func bufferToKillSessionRequest(buf *buffer.ByteBuffer) *KillSessionRequest {
	return &KillSessionRequest{sessionId: buf.GetInt64()}
}

func KillSessionRequestToBuffer(resp *KillSessionRequest, buf *buffer.ByteBuffer) {
	buf.PutInt64(resp.sessionId)
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
	} else if session.TeHandler == nil {
		return nil, NewError(SESSION_NOT_FOUND, "Terminal Connection is down.")
	}
	admin := session.TeHandler.RegisterAdminClient(pack.handler)
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
