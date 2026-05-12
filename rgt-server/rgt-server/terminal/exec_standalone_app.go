package terminal

import (
	"os"
	"os/exec"
	"regexp"
	"rgt-server/buffer"
	"rgt-server/log"
	"rgt-server/protocol"
	"rgt-server/service"
	"rgt-server/util"
	"strings"
	"time"
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
	service               *TerminalEmulationService
	session               *TerminalSession
	cmd                   *exec.Cmd
	lastDataSentTime      time.Time
	keepAliveInterval     uint16
	running               bool
	killAppLostConnection bool
}

func init() {
	terminalProtocol.Operation(TRM_STANDALONE_APP_EXEC, "Standalone app exec").Version(0).Executor(trmStandAloneAppExec)
	// terminalProtocol.Operation(TRM_STANDALONE_APP_SEND_OUTPUT, "Send output from standalone app").Version(0).Executor(trmSendAppOutput)
	// terminalProtocol.Operation(TRM_STANDALONE_APP_SEND_STATUS, "Send status from standalone app").Version(0).Executor(trmSendAppStatus)
}

func (r *AppExecRequest) FromBuffer(buf *buffer.ByteBuffer) {
	r.BaseRequest.OperationCode = TRM_STANDALONE_APP_EXEC
	r.ProtocolVersion = buf.GetInt16()
	r.Username = buf.GetString()
	r.Password = buf.GetString()
	r.OsUser = buf.GetString()
	r.TerminalAddress = buf.GetString()
	r.CaptureOutput = buf.GetBool()
	r.KillAppLostConnection = buf.GetBool()
	r.keepAliveInterval = buf.GetUInt16()
	r.WorkingDir = buf.GetString()
	r.ExePathName = buf.GetString()
	envCount := buf.GetInt8()
	r.EnvVars = make([]string, 0, envCount)
	for envCount > 0 {
		r.EnvVars = append(r.EnvVars, buf.GetString())
		envCount--
	}
	argCount := buf.GetInt8()
	r.Arguments = make([]string, 0, argCount)
	for argCount > 0 {
		r.Arguments = append(r.Arguments, buf.GetString())
		argCount--
	}
}

func (r *AppExecRequest) ToBuffer(buf *buffer.ByteBuffer) {
	buf.PutInt16(r.ProtocolVersion)
	buf.PutString(r.Username)
	buf.PutString(r.Password)
	buf.PutString(r.OsUser)
	buf.PutString(r.TerminalAddress)
	buf.PutBool(r.CaptureOutput)
	buf.PutBool(r.KillAppLostConnection)
	buf.PutUInt16(r.keepAliveInterval)
	buf.PutString(r.WorkingDir)
	buf.PutString(r.ExePathName)
	buf.PutInt8(int8(len(r.EnvVars)))
	for _, envVar := range r.EnvVars {
		buf.PutString(envVar)
	}
	buf.PutInt8(int8(len(r.Arguments)))
	for _, arg := range r.Arguments {
		buf.PutString(arg)
	}
}

func (r *AppExecResponse) FromBuffer(buf *buffer.ByteBuffer) {
	r.Code = protocol.ResponseCode(buf.GetInt16())
	if r.Code == SUCCESS {
		r.SessionId = buf.GetInt64()
		r.Pid = buf.GetInt64()
	} else {
		r.Message = buf.GetString()
	}
}

func (r *AppExecResponse) ToBuffer(buf *buffer.ByteBuffer) {
	buf.PutInt64(r.SessionId)
	buf.PutInt64(r.Pid)
}

func trmStandAloneAppExec(proto *protocol.OperationVersion[*requestPack], pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	handler := pack.handler
	log.Debug("TerminalHandler.processExecStandaloneApp(). handler=", handler.id)
	handler.connectionType = service.LAUNCHER
	packet := pack.packet.RemainingBuffer()
	req := &AppExecRequest{}
	req.FromBuffer(packet)
	handler.protocolVersion = req.ProtocolVersion
	session, err := executeStandaloneApp(handler.service, req, handler, handler.protocolVersion)
	if err != nil {
		return nil, err
	}
	handler.session = session
	response := &AppExecResponse{
		SessionId: session.Id(),
		Pid:       session.AppPid,
	}
	protocol.PutResponse(response, packet)
	return packet, nil
}

func (r *AppOutputRequest) FromBuffer(buf *buffer.ByteBuffer) {
	r.BaseRequest.OperationCode = TRM_STANDALONE_APP_SEND_OUTPUT
	r.Error = buf.GetBool()
	r.Output = buf.GetSlice()
}

func (r *AppOutputRequest) ToBuffer(buf *buffer.ByteBuffer) {
	buf.PutBool(r.Error)
	buf.PutSlice(r.Output)
}

func (r *AppStatusRequest) FromBuffer(buf *buffer.ByteBuffer) {
	r.BaseRequest.OperationCode = TRM_STANDALONE_APP_SEND_STATUS
	r.ExitCode = buf.GetInt32()
	r.Message = buf.GetString()
}

func (r *AppStatusRequest) ToBuffer(buf *buffer.ByteBuffer) {
	buf.PutInt32(r.ExitCode)
	buf.PutString(r.Message)
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

func executeStandaloneApp(service *TerminalEmulationService, req *AppExecRequest, teHandler *TerminalHandler, protocolVersion int16) (*TerminalSession, protocol.ErrorResponse) {
	if !service.Config().StandaloneEnabled().Get() {
		return nil, NewError(TE_APP_LAUNCH_ERROR, "Server not configured to execute standalone app.")
	}
	if !service.AuthenticateStandalone(req.Username, req.Password) {
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
	session := newSession(
		teHandler,
		SESS_TYPE_STANDALONE,
		req.TerminalAddress,
		req.Username,
		req.OsUser,
		strings.Join(append(append(make([]string, 0, len(req.Arguments)+1), exePathName), req.Arguments...), " "))
	session.TimeoutEnabled.Set(false)
	if err := service.sessionManager.AddSession(session); err != nil {
		return nil, NewError(TE_APP_LAUNCH_ERROR, "Error adding session: "+err.Error())
	}
	if err := launchStandaloneApp(service, session, req, protocolVersion); err != nil {
		return nil, NewError(TE_APP_LAUNCH_ERROR, "Error launching executable: "+err.Error())
	}
	return session, nil
}

func (app *standaloneApp) sessionStatus() SessionStatus {
	if app.session != nil {
		return app.session.GetStatus()
	}
	return SESS_CLOSED
}

func (app *standaloneApp) waitSessionReady(interval time.Duration, attempts int) bool {
	tries := 0
	for app.sessionStatus() == SESS_NEW && tries < attempts {
		time.Sleep(interval)
		tries++
	}
	return app.sessionStatus() == SESS_READY
}

func (app *standaloneApp) isConnected() bool {
	return app.session != nil && app.session.IsTEConnected()
}

func (app *standaloneApp) writeAppOutput(data []byte, errOut bool) (n int, err error) {
	dataLen := len(data)
	if app.sessionStatus() == SESS_READY || app.waitSessionReady(3*time.Second, 12) {
		if !app.isConnected() {
			return 0, nil
		}
		req := &AppOutputRequest{
			BaseRequest: protocol.BaseRequest{
				OperationCode: TRM_STANDALONE_APP_SEND_OUTPUT,
			},
		}
		req.Error = errOut
		req.Output = data
		buf := buffer.NewCapacity(uint32(protocol.HEADER_SIZE + buffer.BOOLEAN_FIELD_SIZE + buffer.SLICE_HEADER_SIZE + dataLen))
		req.ToBuffer(buf)
		protocol.PutRequest(req, buf)
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
	req := &AppStatusRequest{
		BaseRequest: protocol.BaseRequest{
			OperationCode: TRM_STANDALONE_APP_SEND_STATUS,
		},
	}
	req.ExitCode = int32(app.cmd.ProcessState.ExitCode())
	req.Message = errorMessage
	buf := buffer.NewCapacity(uint32(protocol.HEADER_SIZE + buffer.INT32_FIELD_SIZE + buffer.STRING_HEADER_SIZE + len(errorMessage)))
	req.ToBuffer(buf)
	protocol.PutRequest(req, buf)
	app.sendData(buf)
}

func (app *standaloneApp) sendStatusSuccess() {
	req := &AppStatusRequest{
		BaseRequest: protocol.BaseRequest{
			OperationCode: TRM_STANDALONE_APP_SEND_STATUS,
		},
	}
	req.ExitCode = int32(app.cmd.ProcessState.ExitCode())
	buf := buffer.NewCapacity(uint32(protocol.HEADER_SIZE + buffer.INT32_FIELD_SIZE))
	req.ToBuffer(buf)
	protocol.PutRequest(req, buf)
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
	sessId := app.session.Id()
	app.session = nil
	app.service.sessionManager.CloseSession(sessId)
}

func (app *standaloneApp) sendKeepAlive() {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("unknown error in server(standaloneApp.sendKeepAlive): %v\n%s", err, util.FullStack())
		}
	}()
	buf := buffer.NewCapacity(uint32(protocol.HEADER_SIZE + buffer.UINT8_FIELD_SIZE))
	req := &protocol.BaseRequest{
		OperationCode: TRM_APP_KEEP_ALIVE,
	}
	req.ToBuffer(buf)
	protocol.PutRequest(req, buf)
	sendKeepAliveInterval := time.Duration(app.keepAliveInterval) * time.Second
	for app.running {
		if time.Since(app.lastDataSentTime) >= sendKeepAliveInterval {
			protocol.PutRequest(req, buf)
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
