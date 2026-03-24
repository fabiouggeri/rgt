package terminal

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"rgt-server/buffer"
	"rgt-server/config"
	"rgt-server/log"
	"rgt-server/protocol"
	"rgt-server/server"
	"rgt-server/service"
	"rgt-server/util"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

/*
   TODO: bufferize output
*/

type AppExecRequest struct {
	protocol.BaseRequest
	Username              string
	Password              string
	OsUser                string
	TerminalAddress       string
	WorkingDir            string
	ExePathName           string
	EnvVars               []string
	Arguments             []string
	keepAliveInterval     uint16
	ProtocolVersion       int16
	CaptureOutput         bool
	KillAppLostConnection bool
}

type AppExecResponse struct {
	protocol.BaseResponse
	SessionId int64
	Pid       int64
}

type AppOutputRequest struct {
	protocol.BaseRequest
	Output []byte
	Error  bool
}

type AppStatusRequest struct {
	protocol.BaseRequest
	Message  string
	ExitCode int32
}

type standaloneApp struct {
	server                *server.Server
	session               *server.Session
	cmd                   *exec.Cmd
	protoOutput           *protocol.Protocol[*AppOutputRequest, *protocol.BaseResponse]
	protoStatus           *protocol.Protocol[*AppStatusRequest, *protocol.BaseResponse]
	lastDataSentTime      time.Time
	keepAliveInterval     uint16
	running               bool
	killAppLostConnection bool
}

type outputWriter struct {
	app         *standaloneApp
	errorOutput bool
}

func init() {
	registerProtocol(TRM_STANDALONE_APP_EXEC, 0, protocol.New(bufferToAppExecRequest, appExecRequestToBuffer, bufferToAppExecResponse, appExecResponseToBuffer))
	registerProtocol(TRM_STANDALONE_APP_SEND_OUTPUT, 0, protocol.New(bufferToAppOutputRequest, appOutputRequestToBuffer, protocol.BufferToBaseResponse, protocol.BaseResponseToBuffer))
	registerProtocol(TRM_STANDALONE_APP_SEND_STATUS, 0, protocol.New(bufferToAppStatusRequest, appStatusRequestToBuffer, protocol.BufferToBaseResponse, protocol.BaseResponseToBuffer))
}

func CreateAppExecProtocol() *protocol.Protocol[*AppExecRequest, *AppExecResponse] {
	return protocol.New(bufferToAppExecRequest, appExecRequestToBuffer, bufferToAppExecResponse, appExecResponseToBuffer)
}

func CreateSendOutputProtocol() *protocol.Protocol[*AppOutputRequest, *protocol.BaseResponse] {
	return protocol.New(bufferToAppOutputRequest, appOutputRequestToBuffer, protocol.BufferToBaseResponse, protocol.BaseResponseToBuffer)
}

func CreateSendStatusProtocol() *protocol.Protocol[*AppStatusRequest, *protocol.BaseResponse] {
	return protocol.New(bufferToAppStatusRequest, appStatusRequestToBuffer, protocol.BufferToBaseResponse, protocol.BaseResponseToBuffer)
}

func bufferToAppExecRequest(buf *buffer.ByteBuffer) *AppExecRequest {
	request := &AppExecRequest{BaseRequest: protocol.BaseRequest{OperationCode: TRM_STANDALONE_APP_EXEC}}
	// request.ProtocolVersion = 0 --->>>> Protocol version already read
	request.Username = buf.GetString()
	request.Password = buf.GetString()
	request.OsUser = buf.GetString()
	request.TerminalAddress = buf.GetString()
	request.CaptureOutput = buf.GetBool()
	request.KillAppLostConnection = buf.GetBool()
	request.keepAliveInterval = buf.GetUInt16()
	request.WorkingDir = buf.GetString()
	request.ExePathName = buf.GetString()
	envCount := buf.GetInt8()
	request.EnvVars = make([]string, 0, envCount)
	for envCount > 0 {
		request.EnvVars = append(request.EnvVars, buf.GetString())
		envCount--
	}
	argCount := buf.GetInt8()
	request.Arguments = make([]string, 0, argCount)
	for argCount > 0 {
		request.Arguments = append(request.Arguments, buf.GetString())
		argCount--
	}
	return request
}

func appExecRequestToBuffer(req *AppExecRequest, buf *buffer.ByteBuffer) {
	buf.PutInt16(req.ProtocolVersion)
	buf.PutString(req.Username)
	buf.PutString(req.Password)
	buf.PutString(req.OsUser)
	buf.PutString(req.TerminalAddress)
	buf.PutBool(req.CaptureOutput)
	buf.PutBool(req.KillAppLostConnection)
	buf.PutUInt16(req.keepAliveInterval)
	buf.PutString(req.WorkingDir)
	buf.PutString(req.ExePathName)
	buf.PutInt8(int8(len(req.EnvVars)))
	for _, envVar := range req.EnvVars {
		buf.PutString(envVar)
	}
	buf.PutInt8(int8(len(req.Arguments)))
	for _, arg := range req.Arguments {
		buf.PutString(arg)
	}
}

func bufferToAppExecResponse(buf *buffer.ByteBuffer) *AppExecResponse {
	respCode := protocol.ResponseCode(buf.GetInt16())
	response := &AppExecResponse{}
	response.Code = respCode
	if respCode == SUCCESS {
		response.SessionId = buf.GetInt64()
		response.Pid = buf.GetInt64()
	} else {
		response.Message = buf.GetString()
	}
	return response
}

func appExecResponseToBuffer(resp *AppExecResponse, buf *buffer.ByteBuffer) {
	buf.PutInt64(resp.SessionId)
	buf.PutInt64(resp.Pid)
}

func bufferToAppOutputRequest(buf *buffer.ByteBuffer) *AppOutputRequest {
	request := &AppOutputRequest{BaseRequest: protocol.BaseRequest{OperationCode: TRM_STANDALONE_APP_SEND_OUTPUT}}
	request.Error = buf.GetBool()
	request.Output = buf.GetSlice()
	return request
}

func appOutputRequestToBuffer(req *AppOutputRequest, buf *buffer.ByteBuffer) {
	buf.PutBool(req.Error)
	buf.PutSlice(req.Output)
}

func bufferToAppStatusRequest(buf *buffer.ByteBuffer) *AppStatusRequest {
	request := &AppStatusRequest{BaseRequest: protocol.BaseRequest{OperationCode: TRM_STANDALONE_APP_SEND_STATUS}}
	request.ExitCode = buf.GetInt32()
	request.Message = buf.GetString()
	return request
}

func appStatusRequestToBuffer(req *AppStatusRequest, buf *buffer.ByteBuffer) {
	buf.PutInt32(req.ExitCode)
	buf.PutString(req.Message)
}

func setWorkingDir(req *AppExecRequest) protocol.ErrorResponse {
	if req.WorkingDir == "" {
		curDir, err := os.Getwd()
		if err != nil {
			return NewError(UNKNOWN_ERROR, "error trying to get working directory: ", err)
		}
		req.WorkingDir = curDir
	} else {
		regex, err := regexp.Compile("%[^%]+%")
		if err != nil {
			return NewError(UNKNOWN_ERROR, "internal error trying to get working directory: ", err)
		}
		req.WorkingDir = regex.ReplaceAllStringFunc(req.WorkingDir, func(envVar string) string {
			return os.Getenv(envVar[1 : len(envVar)-1])
		})
	}
	return nil
}

func executeStandaloneApp(service *TerminalEmulationService, req *AppExecRequest, teHandler service.TerminalConnectionHandler, protocolVersion int16) (*server.Session, protocol.ErrorResponse) {
	if !service.server.Config().StandaloneEnabled().Get() {
		return nil, NewError(TE_APP_LAUNCH_ERROR, "Server not configured to execute standalone app.")
	}
	if !service.server.AuthenticateUser(config.STANDALONE_CONFIG_ID, req.Username, req.Password) {
		return nil, NewError(TE_AUTH_ERROR, "Authentication failed. Invalid credential or not authorized.")
	}
	log.Debugf("[LAUNCHER] terminal.executeStandaloneApp() handler=%d, user=%s user=%s addr=%s", teHandler.Id(), req.Username, req.OsUser, req.TerminalAddress)
	errWC := setWorkingDir(req)
	if errWC != nil {
		return nil, errWC
	}
	exePathName, err := findExecutable(req.ExePathName, req.WorkingDir)
	if err != nil {
		return nil, err
	}
	session := service.server.NewSession(teHandler,
		server.SESS_TYPE_STANDALONE,
		req.TerminalAddress,
		req.Username,
		req.OsUser,
		strings.Join(append(append(make([]string, 0, len(req.Arguments)+1), exePathName), req.Arguments...), " "))
	session.TimeoutEnabled.Set(false)
	err = launchStandaloneApp(service.server, session, req, protocolVersion)
	if err != nil {
		return nil, NewError(TE_APP_LAUNCH_ERROR, "Error launching executable: "+err.Error())
	}
	return session, nil
}

func (app *standaloneApp) sessionStatus() server.SessionStatus {
	if app.session != nil {
		return app.session.GetStatus()
	}
	return server.SESS_CLOSED
}

func (app *standaloneApp) waitSessionReady(interval time.Duration, attempts int) bool {
	tries := 0
	for app.sessionStatus() == server.SESS_NEW && tries < attempts {
		time.Sleep(interval)
		tries++
	}
	return app.sessionStatus() == server.SESS_READY
}

func (app *standaloneApp) isConnected() bool {
	return app.session != nil && app.session.IsTEConnected()
}

func (app *standaloneApp) writeAppOutput(data []byte, errOut bool) (n int, err error) {
	dataLen := len(data)
	if app.sessionStatus() == server.SESS_READY || app.waitSessionReady(3*time.Second, 12) {
		if !app.isConnected() {
			return 0, nil
		}
		req := &AppOutputRequest{BaseRequest: protocol.BaseRequest{OperationCode: TRM_STANDALONE_APP_SEND_OUTPUT}}
		req.Error = errOut
		req.Output = data
		buf := buffer.NewCapacity(uint32(protocol.HEADER_SIZE + 8 + dataLen))
		app.protoOutput.PutRequest(req, buf)
		err := app.sendData(buf)
		app.lastDataSentTime = time.Now()
		if err != nil && app.killAppLostConnection {
			app.killProcess()
		}
		return dataLen, err
	}
	return dataLen, nil
}

func (app *standaloneApp) sendStatusError(errorMessage string) {
	req := &AppStatusRequest{BaseRequest: protocol.BaseRequest{OperationCode: TRM_STANDALONE_APP_SEND_STATUS}}
	req.ExitCode = int32(app.cmd.ProcessState.ExitCode())
	req.Message = errorMessage
	buf := buffer.New()
	app.protoStatus.PutRequest(req, buf)
	app.sendData(buf)
}

func (app *standaloneApp) sendStatusSuccess() {
	req := &AppStatusRequest{BaseRequest: protocol.BaseRequest{OperationCode: TRM_STANDALONE_APP_SEND_STATUS}}
	req.ExitCode = int32(app.cmd.ProcessState.ExitCode())
	buf := buffer.New()
	app.protoStatus.PutRequest(req, buf)
	app.sendData(buf)
}

func (app *standaloneApp) waitFinish() {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("unknown error in server(standaloneApp.waitFinish): %v\n%s", err, util.FullStack())
		}
	}()
	err := app.cmd.Wait()
	app.running = false
	app.session.SetProcess(nil)
	if err != nil || !app.cmd.ProcessState.Success() {
		app.sendStatusError(err.Error())
	} else {
		app.sendStatusSuccess()
	}
	sessId := app.session.Id
	app.session = nil
	app.server.CloseSession(sessId)
}

func (app *standaloneApp) sendKeepAlive() {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("unknown error in server(standaloneApp.sendKeepAlive): %v\n%s", err, util.FullStack())
		}
	}()
	keepAliveProtocol := protocol.New(protocol.BufferToBaseRequest, protocol.BaseRequestToBuffer, protocol.BufferToBaseResponse, protocol.BaseResponseToBuffer)
	buf := buffer.New()
	req := &protocol.BaseRequest{OperationCode: TRM_APP_KEEP_ALIVE}
	sendKeepAliveInterval := time.Duration(app.keepAliveInterval) * time.Second
	for app.running {
		if time.Since(app.lastDataSentTime) >= sendKeepAliveInterval {
			keepAliveProtocol.PutRequest(req, buf)
			err := app.sendData(buf)
			if err != nil && app.killAppLostConnection {
				app.killProcess()
				return
			}
			app.lastDataSentTime = time.Now()
		}
		time.Sleep(3 * time.Second)
	}
}

func (app *standaloneApp) sendData(buf *buffer.ByteBuffer) error {
	if app.session != nil {
		return app.session.SendTE(buf)
	}
	return nil
}

func (app *standaloneApp) killProcess() {
	if app.cmd.Process != nil {
		app.cmd.Process.Kill()
	}
}

func (w *outputWriter) Write(data []byte) (n int, err error) {
	return w.app.writeAppOutput(data, w.errorOutput)
}

func launchStandaloneApp(srv *server.Server, sess *server.Session, req *AppExecRequest, protocolVersion int16) protocol.ErrorResponse {
	var err error
	envVars := make([]string, 0, 32)
	envVars = append(envVars, req.EnvVars...)
	envVars = append(envVars, srv.EnvVars()...)
	envVars = append(envVars, "HB_GT=gtstd")
	envVars = append(envVars, server.ENV_VAR_STANDALONE_APP+"="+fmt.Sprint(sess.Id))
	cmd := exec.Command(req.ExePathName, req.Arguments...)
	cmd.Env = envVars
	cmd.Dir = req.WorkingDir

	app := &standaloneApp{
		server:                srv,
		session:               sess,
		cmd:                   cmd,
		killAppLostConnection: req.KillAppLostConnection,
		keepAliveInterval:     req.keepAliveInterval,
		lastDataSentTime:      time.Now()}
	app.protoOutput, err = findProtocol[*AppOutputRequest, *protocol.BaseResponse](TRM_STANDALONE_APP_SEND_OUTPUT, protocolVersion)
	if err != nil {
		return NewError(TE_APP_LAUNCH_ERROR, "Error launching executable: ", err)
	}
	app.protoStatus, err = findProtocol[*AppStatusRequest, *protocol.BaseResponse](TRM_STANDALONE_APP_SEND_STATUS, protocolVersion)
	if err != nil {
		return NewError(TE_APP_LAUNCH_ERROR, "Error launching executable: ", err)
	}
	if req.CaptureOutput {
		cmd.Stderr = &outputWriter{app: app, errorOutput: true}
		cmd.Stdout = &outputWriter{app: app, errorOutput: false}
	}
	err = cmd.Start()
	if err != nil {
		return NewError(TE_APP_LAUNCH_ERROR, "Error launching executable: ", err)
	}
	app.running = true
	appProcess, err := process.NewProcess(int32(cmd.Process.Pid))
	if err != nil {
		log.Errorf("terminal.launchStandaloneApp(). Error creating standalone process data: %v", err)
		cmd.Process.Kill()
		return NewError(TE_APP_LAUNCH_ERROR, "Error getting process data: ", err)
	}
	sess.SetProcess(appProcess)
	sess.SetAppPid(int64(appProcess.Pid))
	sess.SetStatus(server.SESS_READY)
	go app.waitFinish()
	go app.sendKeepAlive()
	log.Debug("terminal.launchStandaloneApp(). App started: ", req.ExePathName)
	return nil
}
