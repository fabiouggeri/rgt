package terminal

import (
	"rgt-server/buffer"
	"rgt-server/protocol"
	"rgt-server/server"
	"strings"
)

type SessionConfigRequest struct {
	protocol.BaseRequest
	config map[string]string
}

func init() {
	registerOperation(TRM_TE_APP_RESPONSE, trmSendToEndpoint)
	registerTEOperations()
	registerAppOperations()
	registerStandaloneAppOperations()
	registerAdminOperations()
}

func registerTEOperations() {
	registerOperation(TRM_TE_LOGIN, trmTELogin)
	// registerOperation(TRM_TE_LOGOUT, nil)
	// registerOperation(TRM_TE_RECONNECT, nil) --> TODO
}

func registerAppOperations() {
	registerOperation(TRM_APP_LOGIN, trmAppLogin)
	registerOperation(TRM_APP_LOGOUT, trmSendLogoutToTE)
	registerOperation(TRM_APP_SET_ENV, trmSendToEndpoint)
	registerOperation(TRM_APP_UPDATE, trmSendToEndpoint)
	// registerOperation(TRM_APP_READ_KEY, nil) --> Deprecated
	registerOperation(TRM_APP_RPC, trmSendToEndpoint)
	registerOperation(TRM_APP_PUT_FILE, trmSendToEndpoint)
	registerOperation(TRM_APP_GET_FILE, trmSendToEndpoint)
	registerOperation(TRM_APP_KEY_BUFFER_LEN, trmSendToEndpoint)
	// registerOperation(TRM_APP_RECONNECT, nil)  --> TODO
	registerOperation(TRM_APP_KEEP_ALIVE, trmSendToEndpoint)
	registerOperation(TRM_APP_SESSION_CONFIG, sessionConfig)
	registerProtocol(TRM_APP_SESSION_CONFIG, 1, protocol.New(bufferToSessionConfigRequest, sessionConfigRequestToBuffer, protocol.BufferToBaseResponse, protocol.BaseResponseToBuffer))
}

func registerStandaloneAppOperations() {
	registerOperation(TRM_STANDALONE_APP_EXEC, trmStandAloneAppExec)
}

func registerAdminOperations() {
	registerOperation(ADMIN_REQUEST_RESP_OP_CODE, trmSendResponseToAdminClient)
}

func bufferToSessionConfigRequest(buf *buffer.ByteBuffer) *SessionConfigRequest {
	req := &SessionConfigRequest{config: make(map[string]string)}
	count := int(buf.GetInt32())
	for i := 0; i < count; i++ {
		req.config[strings.ToLower(buf.GetString())] = buf.GetString()
	}
	return req
}

func sessionConfigRequestToBuffer(req *SessionConfigRequest, buf *buffer.ByteBuffer) {
	buf.PutInt32(int32(len(req.config)))
	for k, v := range req.config {
		buf.PutString(k)
		buf.PutString(v)
	}
}

func sessionConfig(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	proto, err := findProtocol[*SessionConfigRequest, *protocol.BaseResponse](TRM_APP_SESSION_CONFIG, pack.handler.protocolVersion)
	if err != nil {
		return nil, err
	}
	bufReq := buffer.Wrap(pack.packet.RemainingSlice())
	req := proto.GetRequest(bufReq)
	session := pack.handler.session
	for k := range req.config {
		op := session.Options.Get(k)
		if op == nil {
			return nil, NewError(INVALID_SESSION_OPTION_ERROR, "Invalid session option: ", k)
		}
	}
	for k, v := range req.config {
		op := session.Options.Get(k)
		if op != nil {
			op.SetString(v)
		}
	}
	return TrmAppSuccessResponse(), nil
}

func trmTELogin(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	return pack.handler.processTeLogin(pack.packet.RemainingBuffer())
}

func trmAppLogin(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	return pack.handler.processAppLogin(pack.packet.RemainingBuffer())
}

func trmSendToEndpoint(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	pack.packet.Rewind()
	return nil, pack.handler.sendToEndpoint(pack.packet)
}

func trmSendLogoutToTE(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	if pack.handler != nil && pack.handler.session != nil {
		pack.handler.session.SetStatus(server.SESS_CLOSE_REQUEST)
	}
	return trmSendToEndpoint(pack)
}

func trmSendResponseToAdminClient(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	adminId := pack.packet.GetUInt64()
	respLen := pack.packet.Remaining()
	lenPos := pack.packet.Position() - 4
	pack.packet.SetPosition(lenPos)
	pack.packet.PutUInt32(uint32(respLen))
	return nil, pack.handler.sendToAdminClient(adminId, pack.packet.GetBufferFrom(lenPos))
}

func trmStandAloneAppExec(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	return pack.handler.processExecStandaloneApp(pack.packet.RemainingBuffer())
}
