package terminal

import (
	"rgt-server/buffer"
	"rgt-server/log"
	"rgt-server/protocol"
	"rgt-server/server"
	"rgt-server/service"
	"rgt-server/util"
	"strconv"
	"time"
)

type AppLoginRequest struct {
	protocol.BaseRequest
	SessionId int64
	Pid       int64
}

type AppLoginResponse struct {
	protocol.BaseResponse
	LogPathName string
	LogLevel    log.LogLevel
}

func init() {
	terminalProtocol.Operation(TRM_APP_LOGIN, "Application login").Version(0).Executor(trmAppLogin)
}

func (r *AppLoginRequest) GetOperationCode() protocol.OperationCode {
	return r.OperationCode
}

func (r *AppLoginResponse) GetCode() protocol.ResponseCode {
	return r.Code
}

func (r *AppLoginResponse) GetMessage() string {
	return r.Message
}

func (r *AppLoginRequest) FromBuffer(buf *buffer.ByteBuffer) {
	r.BaseRequest.OperationCode = TRM_APP_LOGIN
	r.SessionId = buf.GetInt64()
	r.Pid = buf.GetInt64()
}

func (r *AppLoginRequest) ToBuffer(buf *buffer.ByteBuffer) {
	buf.PutInt64(r.SessionId)
	buf.PutInt64(r.Pid)
}

func (r *AppLoginResponse) FromBuffer(buf *buffer.ByteBuffer) {
	respCode := protocol.ResponseCode(buf.GetInt16())
	if respCode == SUCCESS {
		r.LogLevel = log.LogLevel(buf.GetInt8())
		r.LogPathName = buf.GetString()
	} else {
		r.Code = respCode
		r.Message = buf.GetString()
	}
}

func (r *AppLoginResponse) ToBuffer(buf *buffer.ByteBuffer) {
	buf.PutInt8(int8(r.LogLevel))
	buf.PutString(r.LogPathName)

}

func trmAppLogin(proto *protocol.OperationVersion[*requestPack], pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	h := pack.handler
	log.Debug("TerminalHandler.processAppLogin(). handler=", h.id)
	h.connectionType = service.APPLICATION
	packet := pack.packet.RemainingBuffer()
	req := &AppLoginRequest{}
	req.FromBuffer(packet)
	session, err := appLogin(h.service.server, req, h)
	if err != nil {
		return nil, err
	}
	h.session = session
	response := &AppLoginResponse{
		LogLevel:    h.service.server.Config().AppLogLevel().Get(),
		LogPathName: util.RelativePathToAbsolute(h.service.server.Config().AppLogPathName().Get()),
	}
	protocol.PutResponse(response, packet)
	return packet, nil
}

func appLogin(srv *server.Server, req *AppLoginRequest, appHandler *TerminalHandler) (*TerminalSession, protocol.ErrorResponse) {
	var err protocol.ErrorResponse
	log.Infof("[APP;session=%d] terminal.appLogin(). handler=%d pid=%d", req.SessionId, appHandler.id, req.Pid)
	session := appHandler.service.sessionManager.GetSession(req.SessionId)
	if session == nil {
		return nil, NewError(APP_CONNECT_ERROR, "Session ", strconv.FormatInt(req.SessionId, 10), " not found.")
	}
	if session.AppHandler != nil {
		return nil, NewError(APP_CONNECT_ERROR, "Session ", strconv.FormatInt(session.Id(), 10), " already have an app connected.")
	}
	session.SetAppLoginTime(time.Now())
	session.SetAppHandler(appHandler)
	appHandler.SetEndpoint(session.TeHandler)
	session.TeHandler.SetEndpoint(appHandler)
	session.SetAppPid(req.Pid)
	if err := session.ChangeStatus(SESS_CONNECTING, SESS_READY); err != nil {
		return nil, NewError(APP_CONNECT_ERROR, "Error in app login for session ", strconv.FormatInt(session.Id(), 10), ": ", err)
	}
	return session, err
}
