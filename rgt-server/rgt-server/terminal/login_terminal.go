package terminal

import (
	"errors"
	"os"
	"os/exec"
	"rgt-server/buffer"
	"rgt-server/log"
	"rgt-server/protocol"
	"rgt-server/server"
	"strings"
)

type TeLoginRequest struct {
	protocol.BaseRequest
	Arguments       []string
	Username        string
	Password        string
	OsUser          string
	TerminalAddress string
	WorkingDir      string
	ExePathName     string
}

type TeLoginResponse struct {
	protocol.BaseResponse
	SessionId   int64
	LogPathName string
	LogLevel    log.LogLevel
}

func init() {
	registerProtocol(TRM_TE_LOGIN, 0, protocol.New(bufferToTeLoginRequest, teLoginRequestToBuffer, bufferToTeLoginResponse, teLoginResponseToBuffer))
	registerProtocol(TRM_TE_LOGIN, 3, protocol.New(bufferToTeLoginRequestV3, teLoginRequestToBufferV3, bufferToTeLoginResponse, teLoginResponseToBuffer))
}

func NewTeLoginResponse(sessionId int64, logLevel log.LogLevel, logPathName string) *TeLoginResponse {
	return &TeLoginResponse{SessionId: sessionId,
		LogLevel:    logLevel,
		LogPathName: logPathName}
}

func (req *TeLoginRequest) GetOperationCode() protocol.OperationCode {
	return req.OperationCode
}

func (resp *TeLoginResponse) GetCode() protocol.ResponseCode {
	return resp.Code
}

func (resp *TeLoginResponse) GetMessage() string {
	return resp.Message
}

func bufferToTeLoginRequest(buf *buffer.ByteBuffer) *TeLoginRequest {
	request := &TeLoginRequest{BaseRequest: protocol.BaseRequest{OperationCode: TRM_TE_LOGIN}}
	request.Username = buf.GetString()
	request.Password = buf.GetString()
	request.WorkingDir = buf.GetString()
	request.ExePathName = buf.GetString()
	argCount := buf.GetInt8()
	request.Arguments = make([]string, 0, argCount)
	for argCount > 0 {
		request.Arguments = append(request.Arguments, buf.GetString())
		argCount--
	}
	return request
}

func teLoginRequestToBuffer(req *TeLoginRequest, buf *buffer.ByteBuffer) {
	buf.PutString(req.Username)
	buf.PutString(req.Password)
	buf.PutString(req.WorkingDir)
	buf.PutString(req.ExePathName)
	buf.PutInt8(int8(len(req.Arguments)))
	for _, arg := range req.Arguments {
		buf.PutString(arg)
	}
}

func bufferToTeLoginRequestV3(buf *buffer.ByteBuffer) *TeLoginRequest {
	request := &TeLoginRequest{BaseRequest: protocol.BaseRequest{OperationCode: TRM_TE_LOGIN}}
	request.Username = buf.GetString()
	request.Password = buf.GetString()
	request.OsUser = buf.GetString()
	request.TerminalAddress = buf.GetString()
	request.WorkingDir = buf.GetString()
	request.ExePathName = buf.GetString()
	argCount := buf.GetInt8()
	request.Arguments = make([]string, 0, argCount)
	for argCount > 0 {
		request.Arguments = append(request.Arguments, buf.GetString())
		argCount--
	}
	return request
}

func teLoginRequestToBufferV3(req *TeLoginRequest, buf *buffer.ByteBuffer) {
	buf.PutString(req.Username)
	buf.PutString(req.Password)
	buf.PutString(req.OsUser)
	buf.PutString(req.TerminalAddress)
	buf.PutString(req.WorkingDir)
	buf.PutString(req.ExePathName)
	buf.PutInt8(int8(len(req.Arguments)))
	for _, arg := range req.Arguments {
		buf.PutString(arg)
	}
}

func bufferToTeLoginResponse(buf *buffer.ByteBuffer) *TeLoginResponse {
	var response *TeLoginResponse
	code := protocol.ResponseCode(buf.GetInt16())
	if code == SUCCESS {
		response = &TeLoginResponse{SessionId: buf.GetInt64(), LogLevel: log.LogLevel(buf.GetInt8()), LogPathName: buf.GetString()}
	} else {
		response = &TeLoginResponse{}
		response.Code = code
		response.Message = buf.GetString()
	}
	return response
}

func teLoginResponseToBuffer(resp *TeLoginResponse, buf *buffer.ByteBuffer) {
	buf.PutInt64(resp.SessionId)
	buf.PutUInt8(uint8(resp.LogLevel))
	buf.PutString(resp.LogPathName)
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

func teLogin(service *TerminalEmulationService, req *TeLoginRequest, teHandler *TerminalHandler) (*server.Session, protocol.ErrorResponse) {
	if !service.server.AuthenticateUser(service.GetName(), req.Username, req.Password) {
		return nil, NewError(TE_AUTH_ERROR, "Authentication failed. Invalid credential or not authorized.")
	}
	log.Infof("[TE] terminal.teLogin(). handler=%d auth-user=%s user=%s Client=%s", teHandler.id, req.Username, req.OsUser, req.TerminalAddress)
	exePathName, err := findExecutable(req.ExePathName, req.WorkingDir)
	if err != nil {
		return nil, err
	}
	if req.TerminalAddress != "" {
		teHandler.remoteAddres = req.TerminalAddress
	}
	session := service.server.NewSession(teHandler,
		server.SESS_TYPE_EMULATION,
		req.TerminalAddress,
		req.Username,
		req.OsUser,
		strings.Join(append(append(make([]string, 0, len(req.Arguments)+1), exePathName), req.Arguments...), " "))
	err = launchTrmApp(service.server, session, exePathName, req.WorkingDir, req.Arguments)
	if err != nil {
		return nil, NewError(TE_APP_LAUNCH_ERROR, "Error launching executable: ", err.Error())
	}
	return session, nil
}
