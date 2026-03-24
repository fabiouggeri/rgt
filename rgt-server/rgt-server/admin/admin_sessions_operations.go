package admin

import (
	"rgt-server/buffer"
	"rgt-server/log"
	"rgt-server/protocol"
	"rgt-server/server"
)

type AdminLoginRequest struct {
	protocol.BaseRequest
	username        string
	password        string
	protocolVersion int16
}

type AdminLoginResponse struct {
	protocol.BaseResponse
	serverStatus    server.ServerStatus
	startTime       int64
	serverVersion   string
	userEditing     string
	sessionsCount   int32
	protocolVersion int16
	readOnly        bool
}

const MULTIPLES_CLIENTS_MINIMUM_VERSION int16 = 3

func init() {
	registerOperation(ADM_LOGIN, loginAdmin)
	registerOperation(ADM_LOGOFF, logoffAdmin)
	registerOperation(ADM_KILL_ADMIN_SESSIONS, killAdminSessions)
	registerProtocol(ADM_LOGIN, 0, protocol.New(bufferToAdminLoginRequest, adminLoginRequestToBuffer, bufferToAdminLoginResponse, adminLoginResponseToBuffer))
	registerProtocol(ADM_LOGIN, 4, protocol.New(bufferToAdminLoginRequest, adminLoginRequestToBuffer, bufferToAdminLoginResponseV4, adminLoginResponseToBuffer))
	registerProtocol(ADM_LOGOFF, 0, protocol.NewDefault())
	registerProtocol(ADM_KILL_ADMIN_SESSIONS, 0, protocol.NewDefault())
}

func bufferToAdminLoginRequest(buf *buffer.ByteBuffer) *AdminLoginRequest {
	return &AdminLoginRequest{
		BaseRequest:     protocol.BaseRequest{OperationCode: ADM_LOGIN},
		protocolVersion: buf.GetInt16(),
		username:        buf.GetString(),
		password:        buf.GetString()}
}

func adminLoginRequestToBuffer(req *AdminLoginRequest, buf *buffer.ByteBuffer) {
	buf.PutInt16(req.protocolVersion)
	buf.PutString(req.username)
	buf.PutString(req.password)
}

func adminLoginResponseToBuffer(response *AdminLoginResponse, buf *buffer.ByteBuffer) {
	buf.PutString(string(response.serverStatus))
	buf.PutInt32(response.sessionsCount)
	buf.PutInt64(response.startTime)
	buf.PutBool(response.readOnly)
	buf.PutInt16(response.protocolVersion)
	buf.PutString(response.serverVersion)
	buf.PutString(response.userEditing)
}

func bufferToAdminLoginResponse(buf *buffer.ByteBuffer) *AdminLoginResponse {
	return &AdminLoginResponse{
		serverStatus:    server.ServerStatus(buf.GetString()),
		sessionsCount:   buf.GetInt32(),
		startTime:       buf.GetInt64(),
		readOnly:        buf.GetBool(),
		protocolVersion: 0}
}

func bufferToAdminLoginResponseV4(buf *buffer.ByteBuffer) *AdminLoginResponse {
	return &AdminLoginResponse{
		serverStatus:    server.ServerStatus(buf.GetString()),
		sessionsCount:   buf.GetInt32(),
		startTime:       buf.GetInt64(),
		readOnly:        buf.GetBool(),
		protocolVersion: buf.GetInt16(),
		serverVersion:   buf.GetString(),
		userEditing:     buf.GetString()}
}

func login(handler *AdminHandler, req *AdminLoginRequest) (*AdminLoginResponse, protocol.ErrorResponse) {
	if !handler.service.server.AuthenticateUser(handler.service.GetName(), req.username, req.password) {
		return nil, NewError(INVALID_CREDENTIAL, "invalid credential for user ", req.username)
	}
	if handler.readOnly && req.protocolVersion < MULTIPLES_CLIENTS_MINIMUM_VERSION {
		return nil, NewError(ADMIN_SESSION_ALREADY_OPEN, "admin session already open")
	}
	srv := handler.service.server
	handler.protocolVersion = req.protocolVersion
	response := &AdminLoginResponse{
		protocolVersion: req.protocolVersion,
		serverStatus:    srv.GetStatus(),
		sessionsCount:   srv.GetSessionsCount(),
		startTime:       srv.GetStartTime(),
		readOnly:        handler.readOnly,
		serverVersion:   srv.Version()}
	if handler.service.GetHandlerEditing() != nil {
		response.userEditing = handler.service.GetHandlerEditing().username
	}
	log.Infof("ADMIN: Administration login. User: '%s' Address: '%s'", req.username, handler.GetRemoteAddr())
	return response, nil
}

func (r *AdminLoginRequest) GetOperationCode() protocol.OperationCode {
	return r.OperationCode
}

func (r *AdminLoginResponse) GetCode() protocol.ResponseCode {
	return r.Code
}

func (r *AdminLoginResponse) GetMessage() string {
	return r.Message
}

func loginAdmin(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_sessions_operations.loginAdmin()")
	respBuf := buffer.NewCapacity(128)
	buf := buffer.Wrap(pack.body)
	if len(pack.handler.GetUsername()) > 0 {
		return nil, NewError(ADMIN_SESSION_ALREADY_OPEN, "user already logged: ", pack.handler.username)
	} else if buf.Remaining() <= 1 {
		return nil, NewError(PROTOCOL_ERROR, "client and server incompatibility")
	} else {
		proto, err := findProtocol[*AdminLoginRequest, *AdminLoginResponse](ADM_LOGIN, pack.handler.protocolVersion)
		if err != nil {
			return nil, err
		}
		loginRequest := proto.GetRequest(buf)
		if loginRequest.protocolVersion > ADMIN_PROTOCOL_VERSION {
			loginRequest.protocolVersion = ADMIN_PROTOCOL_VERSION
		}
		resp, err := login(pack.handler, loginRequest)
		if err != nil {
			return nil, err
		}
		proto.PutResponse(resp, respBuf)
		pack.handler.SetUsername(loginRequest.username)
	}
	return respBuf, nil
}

func killAdminSessions(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_sessions_operations.killAdminSessions()")
	for _, handler := range pack.handler.service.handlers {
		if handler.id != pack.handler.id {
			err := handler.Close()
			if err != nil {
				return nil, NewError(ERROR_KILLING_ADMIN_SESSION, "error killing admin session ", handler.GetRemoteAddr(), ". Cause: ", err)
			}
			log.Info("ADMIN: Admin connection ", handler.GetRemoteAddr(), " killed by ", pack.handler.GetRemoteAddr())
		}
	}
	return SuccessAdminResponse(), nil
}

func logoffAdmin(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Infof("ADMIN: Administration logoff. User: '%s' Address: '%s'", pack.handler.username, pack.handler.GetRemoteAddr())
	return SuccessAdminResponse(), nil
}
