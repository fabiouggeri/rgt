package terminal

import (
	"rgt-server/buffer"
	"rgt-server/log"
	"rgt-server/protocol"
	"rgt-server/server"
	"strconv"
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
	registerProtocol(TRM_APP_LOGIN, 0, protocol.New(bufferToAppLoginRequest, appLoginRequestToBuffer, bufferToAppLoginResponse, appLoginResponseToBuffer))
}

func bufferToAppLoginRequest(buf *buffer.ByteBuffer) *AppLoginRequest {
	return &AppLoginRequest{BaseRequest: protocol.BaseRequest{OperationCode: TRM_APP_LOGIN},
		SessionId: buf.GetInt64(),
		Pid:       buf.GetInt64()}
}

func (req *AppLoginRequest) GetOperationCode() protocol.OperationCode {
	return req.OperationCode
}

func (resp *AppLoginResponse) GetCode() protocol.ResponseCode {
	return resp.Code
}

func (resp *AppLoginResponse) GetMessage() string {
	return resp.Message
}

func appLoginRequestToBuffer(req *AppLoginRequest, buf *buffer.ByteBuffer) {
	buf.PutInt64(req.SessionId)
	buf.PutInt64(req.Pid)
}

func bufferToAppLoginResponse(buf *buffer.ByteBuffer) *AppLoginResponse {
	var response *AppLoginResponse
	respCode := protocol.ResponseCode(buf.GetInt16())
	if respCode == SUCCESS {
		response = &AppLoginResponse{LogLevel: log.LogLevel(buf.GetInt8()), LogPathName: buf.GetString()}
	} else {
		response = &AppLoginResponse{}
		response.Code = respCode
		response.Message = buf.GetString()
	}
	return response
}

func appLoginResponseToBuffer(resp *AppLoginResponse, buf *buffer.ByteBuffer) {
	buf.PutInt8(int8(resp.LogLevel))
	buf.PutString(resp.LogPathName)

}

func appLogin(srv *server.Server, req *AppLoginRequest, appHandler *TerminalHandler) (*server.Session, protocol.ErrorResponse) {
	var err protocol.ErrorResponse
	log.Debugf("[APP;session=%d] terminal.appLogin(). handler=%d pid=%d", req.SessionId, appHandler.id, req.Pid)
	session := srv.GetSession(req.SessionId)
	if session != nil {
		if session.AppHandler == nil {
			session.SetAppHandler(appHandler)
			appHandler.SetEndpoint(session.TeHandler)
			session.TeHandler.SetEndpoint(appHandler)
			session.SetAppPid(req.Pid)
		} else {
			err = NewError(APP_CONNECT_ERROR, "Session "+strconv.FormatInt(session.Id, 10)+" already have an app connected.")
		}
	} else {
		err = NewError(APP_CONNECT_ERROR, "Session "+strconv.FormatInt(req.SessionId, 10)+" not found.")
	}
	return session, err
}
