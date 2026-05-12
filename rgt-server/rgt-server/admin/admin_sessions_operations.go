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
	sessionsCount   int32
	protocolVersion int16
	readOnly        bool
}
type AdminLoginResponseV4 struct {
	AdminLoginResponse
	serverVersion string
	userEditing   string
}

const (
	MULTIPLES_CLIENTS_MINIMUM_VERSION int16 = 3
	ADMIN_LOGIN_RESPONSE_MIN_LEN      int   = 64
)

func init() {
	loginOp := adminProtocol.Operation(ADM_LOGIN, "Admin login")
	loginOp.Version(0).Executor(loginAdmin)
	loginOp.Version(4).Executor(loginAdminV4)
	adminProtocol.Operation(ADM_LOGOFF, "Admin logoff").Version(0).Executor(logoffAdmin)
	adminProtocol.Operation(ADM_KILL_ADMIN_SESSIONS, "Kill admin sessions").Version(0).Executor(killAdminSessions)

}

func (r *AdminLoginRequest) FromBuffer(buf *buffer.ByteBuffer) {
	r.protocolVersion = buf.GetInt16()
	r.username = buf.GetString()
	r.password = buf.GetString()
}

func (r *AdminLoginRequest) ToBuffer(buf *buffer.ByteBuffer) {
	buf.PutInt16(r.protocolVersion)
	buf.PutString(r.username)
	buf.PutString(r.password)
}

func (r *AdminLoginResponse) FromBuffer(buf *buffer.ByteBuffer) {
	r.serverStatus = server.ServerStatus(buf.GetString())
	r.sessionsCount = buf.GetInt32()
	r.startTime = buf.GetInt64()
	r.readOnly = buf.GetBool()
	r.protocolVersion = 0
}

func (r *AdminLoginResponse) ToBuffer(buf *buffer.ByteBuffer) {
	buf.PutString(string(r.serverStatus))
	buf.PutInt32(r.sessionsCount)
	buf.PutInt64(r.startTime)
	buf.PutBool(r.readOnly)
	buf.PutInt16(r.protocolVersion)
}

func (r *AdminLoginResponseV4) FromBuffer(buf *buffer.ByteBuffer) {
	r.AdminLoginResponse.FromBuffer(buf)
	r.serverVersion = buf.GetString()
	r.userEditing = buf.GetString()
}

func (r *AdminLoginResponseV4) ToBuffer(buf *buffer.ByteBuffer) {
	r.AdminLoginResponse.ToBuffer(buf)
	buf.PutString(r.serverVersion)
	buf.PutString(r.userEditing)
}

func login(handler *AdminHandler, req *AdminLoginRequest) (*AdminLoginResponseV4, protocol.ErrorResponse) {
	if !handler.service.AuthenticateUser(req.username, req.password) {
		return nil, NewError(INVALID_CREDENTIAL, "invalid credential for user ", req.username)
	}
	if handler.readOnly && req.protocolVersion < MULTIPLES_CLIENTS_MINIMUM_VERSION {
		return nil, NewError(ADMIN_SESSION_ALREADY_OPEN, "admin session already open")
	}
	srv := handler.service.server
	handler.protocolVersion = req.protocolVersion
	response := &AdminLoginResponseV4{
		AdminLoginResponse: AdminLoginResponse{
			protocolVersion: req.protocolVersion,
			serverStatus:    srv.GetStatus(),
			sessionsCount:   handler.service.terminalService.GetSessionsCount(),
			startTime:       srv.GetStartTime(),
			readOnly:        handler.readOnly,
		},
		serverVersion: srv.Version()}
	if handler.service.GetHandlerEditing() != nil {
		response.userEditing = handler.service.GetHandlerEditing().username
	}
	log.Infof("ADMIN: Administration login. User: '%s' Address: '%s'", req.username, handler.GetRemoteAddr())
	return response, nil
}

func loginAdmin(proto *protocol.OperationVersion[*RequestPack], pack *RequestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_sessions_operations.loginAdmin()")
	buf := buffer.Wrap(pack.body)
	if len(pack.handler.GetUsername()) > 0 {
		return nil, NewError(ADMIN_SESSION_ALREADY_OPEN, "user already logged: ", pack.handler.username)
	} else if buf.Remaining() <= 1 {
		return nil, NewError(PROTOCOL_ERROR, "client and server incompatibility")
	}
	req := &AdminLoginRequest{}
	req.FromBuffer(buf)
	if req.protocolVersion > ADMIN_PROTOCOL_VERSION {
		req.protocolVersion = ADMIN_PROTOCOL_VERSION
	}
	resp, err := login(pack.handler, req)
	if err != nil {
		return nil, err
	}
	respBuf := buffer.NewCapacity(uint32(protocol.RESPONSE_HEADER_SIZE + ADMIN_LOGIN_RESPONSE_MIN_LEN))
	protocol.PutResponse(&resp.AdminLoginResponse, respBuf)
	pack.handler.SetUsername(req.username)
	return respBuf, nil
}

func loginAdminV4(proto *protocol.OperationVersion[*RequestPack], pack *RequestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_sessions_operations.loginAdmin()")
	buf := buffer.Wrap(pack.body)
	if len(pack.handler.GetUsername()) > 0 {
		return nil, NewError(ADMIN_SESSION_ALREADY_OPEN, "user already logged: ", pack.handler.username)
	} else if buf.Remaining() <= 1 {
		return nil, NewError(PROTOCOL_ERROR, "client and server incompatibility")
	}
	req := &AdminLoginRequest{}
	req.FromBuffer(buf)
	if req.protocolVersion > ADMIN_PROTOCOL_VERSION {
		req.protocolVersion = ADMIN_PROTOCOL_VERSION
	}
	resp, err := login(pack.handler, req)
	if err != nil {
		return nil, err
	}
	respBuf := buffer.NewCapacity(uint32(protocol.RESPONSE_HEADER_SIZE + ADMIN_LOGIN_RESPONSE_MIN_LEN + buffer.STRING_HEADER_SIZE + len(resp.serverVersion) + buffer.STRING_HEADER_SIZE + len(resp.userEditing)))
	protocol.PutResponse(resp, respBuf)
	pack.handler.SetUsername(req.username)
	return respBuf, nil
}

func logoffAdmin(proto *protocol.OperationVersion[*RequestPack], pack *RequestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Infof("Admin logoff. User: '%s' Address: '%s'", pack.handler.username, pack.handler.GetRemoteAddr())
	return protocol.SuccessResponse(), nil
}

func killAdminSessions(proto *protocol.OperationVersion[*RequestPack], pack *RequestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_sessions_operations.killAdminSessions()")
	for _, handler := range pack.handler.service.handlers {
		if handler.id != pack.handler.id {
			err := handler.Close()
			if err != nil {
				return nil, NewError(ERROR_KILLING_ADMIN_SESSION, "error killing admin session ", handler.GetRemoteAddr(), ". Cause: ", err)
			}
			log.Info("Admin connection ", handler.GetRemoteAddr(), " killed by ", pack.handler.GetRemoteAddr())
		}
	}
	return protocol.SuccessResponse(), nil
}
