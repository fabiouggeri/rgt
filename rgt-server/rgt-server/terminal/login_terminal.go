package terminal

import (
	"errors"
	"os"
	"os/exec"
	"rgt-server/buffer"
	"rgt-server/log"
	"rgt-server/protocol"
	"rgt-server/service"
	"strings"
)

type TeLoginRequest struct {
	protocol.BaseRequest
	Arguments   []string
	Username    string
	Password    string
	WorkingDir  string
	ExePathName string
}
type TeLoginRequestV3 struct {
	TeLoginRequest
	OsUser          string
	TerminalAddress string
}

type TeLoginResponse struct {
	protocol.BaseResponse
	SessionId   int64
	LogPathName string
	LogLevel    log.LogLevel
}

func init() {
	loginOp := terminalProtocol.Operation(TRM_TE_LOGIN, "TE login")
	loginOp.Version(0).Executor(trmTELogin)
	loginOp.Version(3).Executor(trmTELoginV3)
}

func NewTeLoginResponse(sessionId int64, logLevel log.LogLevel, logPathName string) *TeLoginResponse {
	return &TeLoginResponse{SessionId: sessionId,
		LogLevel:    logLevel,
		LogPathName: logPathName}
}

func (req *TeLoginRequest) GetOperationCode() protocol.OperationCode {
	return req.OperationCode
}

func (r *TeLoginResponse) GetCode() protocol.ResponseCode {
	return r.Code
}

func (r *TeLoginResponse) GetMessage() string {
	return r.Message
}

func (r *TeLoginRequest) FromBuffer(buf *buffer.ByteBuffer) {
	r.BaseRequest.OperationCode = TRM_TE_LOGIN
	r.Username = buf.GetString()
	r.Password = buf.GetString()
	r.WorkingDir = buf.GetString()
	r.ExePathName = buf.GetString()
	argCount := buf.GetInt8()
	r.Arguments = make([]string, 0, argCount)
	for argCount > 0 {
		r.Arguments = append(r.Arguments, buf.GetString())
		argCount--
	}
}

func (r *TeLoginRequest) ToBuffer(buf *buffer.ByteBuffer) {
	buf.PutString(r.Username)
	buf.PutString(r.Password)
	buf.PutString(r.WorkingDir)
	buf.PutString(r.ExePathName)
	buf.PutInt8(int8(len(r.Arguments)))
	for _, arg := range r.Arguments {
		buf.PutString(arg)
	}
}

func (r *TeLoginRequestV3) FromBuffer(buf *buffer.ByteBuffer) {
	r.BaseRequest.OperationCode = TRM_TE_LOGIN
	r.Username = buf.GetString()
	r.Password = buf.GetString()
	r.OsUser = buf.GetString()
	r.TerminalAddress = buf.GetString()
	r.WorkingDir = buf.GetString()
	r.ExePathName = buf.GetString()
	argCount := buf.GetInt8()
	r.Arguments = make([]string, 0, argCount)
	for argCount > 0 {
		r.Arguments = append(r.Arguments, buf.GetString())
		argCount--
	}
}

func (r *TeLoginRequestV3) ToBuffer(buf *buffer.ByteBuffer) {
	buf.PutString(r.Username)
	buf.PutString(r.Password)
	buf.PutString(r.OsUser)
	buf.PutString(r.TerminalAddress)
	buf.PutString(r.WorkingDir)
	buf.PutString(r.ExePathName)
	buf.PutInt8(int8(len(r.Arguments)))
	for _, arg := range r.Arguments {
		buf.PutString(arg)
	}
}

func (r *TeLoginResponse) FromBuffer(buf *buffer.ByteBuffer) {
	code := protocol.ResponseCode(buf.GetInt16())
	if code == SUCCESS {
		r.SessionId = buf.GetInt64()
		r.LogLevel = log.LogLevel(buf.GetInt8())
		r.LogPathName = buf.GetString()
	} else {
		r.Code = code
		r.Message = buf.GetString()
	}
}

func (r *TeLoginResponse) ToBuffer(buf *buffer.ByteBuffer) {
	buf.PutInt64(r.SessionId)
	buf.PutUInt8(uint8(r.LogLevel))
	buf.PutString(r.LogPathName)
}

func trmTELogin(proto *protocol.OperationVersion[*requestPack], pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	handler := pack.handler
	log.Debug("terminal.trmTELogin(). handler=", handler.id)
	handler.connectionType = service.TERMINAL
	packet := pack.packet.RemainingBuffer()
	req := &TeLoginRequestV3{}
	req.TeLoginRequest.FromBuffer(packet)
	session, err := teLogin(handler.service, req, handler)
	if err != nil {
		return nil, err
	}
	handler.session = session
	config := handler.service.Config()
	response := NewTeLoginResponse(session.Id(), config.TeLogLevel().Get(), config.TeLogPathName().Get())
	protocol.PutResponse(response, packet)
	return packet, nil
}

func trmTELoginV3(proto *protocol.OperationVersion[*requestPack], pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	handler := pack.handler
	log.Debug("terminal.trmTELoginV3(). handler=", handler.id)
	handler.connectionType = service.TERMINAL
	packet := pack.packet.RemainingBuffer()
	req := &TeLoginRequestV3{}
	req.FromBuffer(packet)
	session, err := teLogin(handler.service, req, handler)
	if err != nil {
		return nil, err
	}
	handler.session = session
	config := handler.service.Config()
	response := NewTeLoginResponse(session.Id(), config.TeLogLevel().Get(), config.TeLogPathName().Get())
	protocol.PutResponse(response, packet)
	return packet, nil
}

func findExecutable(exeFileName string, workingDir string) (string, protocol.ErrorResponse) {
	workingDirInfo, err := os.Stat(workingDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", NewError(TE_APP_LAUNCH_ERROR, "Working directory does not exists: ", workingDir)
		} else {
			return "", NewError(TE_APP_LAUNCH_ERROR, "Invalid working directory: ", err)
		}
	}
	if workingDirInfo == nil {
		return "", NewError(TE_APP_LAUNCH_ERROR, "Invalid working directory: ", workingDir)
	} else if !workingDirInfo.IsDir() {
		return "", NewError(TE_APP_LAUNCH_ERROR, "Working path is not a directory: ", workingDir)
	}
	foundFile, err := exec.LookPath(exeFileName)
	if err != nil {
		return "", NewError(TE_APP_LAUNCH_ERROR, "Executable not found: ", exeFileName)
	}
	return foundFile, nil
}

func teLogin(service *TerminalEmulationService, req *TeLoginRequestV3, teHandler *TerminalHandler) (*TerminalSession, protocol.ErrorResponse) {
	log.Infof("[TE] terminal.teLogin(). handler=%d auth-user=%s user=%s Client=%s", teHandler.id, req.Username, req.OsUser, req.TerminalAddress)
	if !service.AuthenticateUser(service.GetName(), req.Username, req.Password) {
		return nil, NewError(TE_AUTH_ERROR, "Authentication failed. Invalid credential or not authorized.")
	}
	exePathName, err := findExecutable(req.ExePathName, req.WorkingDir)
	if err != nil {
		return nil, err
	}
	if req.TerminalAddress != "" {
		teHandler.remoteAddres = req.TerminalAddress
	}
	session := newSession(teHandler,
		SESS_TYPE_EMULATION,
		req.TerminalAddress,
		req.Username,
		req.OsUser,
		strings.Join(append(append(make([]string, 0, len(req.Arguments)+1), exePathName), req.Arguments...), " "))
	if err := service.sessionManager.AddSession(session); err != nil {
		return nil, NewError(TE_APP_LAUNCH_ERROR, "Error adding session: ", err.Error())
	}
	if err := launchTrmApp(service, session, exePathName, req.WorkingDir, req.Arguments); err != nil {
		return nil, NewError(TE_APP_LAUNCH_ERROR, "Error launching executable: ", err.Error())
	}
	return session, nil
}
