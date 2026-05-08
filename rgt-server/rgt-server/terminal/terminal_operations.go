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
	// APP and TE operations
	terminalProtocol.Operation(TRM_TE_APP_RESPONSE, "TE/APP response").Version(0).Executor(trmSendToEndpoint)

	// terminalProtocol.Operation(TRM_TE_LOGOUT, "TE logout").Version(0).Executor(nil)
	// terminalProtocol.Operation(TRM_TE_RECONNECT, "TE reconnect").Version(0).Executor(nil)

	// App operations
	terminalProtocol.Operation(TRM_APP_LOGOUT, "App logout").Version(0).Executor(trmSendLogoutToTE)
	terminalProtocol.Operation(TRM_APP_SET_ENV, "App set env").Version(0).Executor(trmSendToEndpoint)
	terminalProtocol.Operation(TRM_APP_UPDATE, "App update").Version(0).Executor(trmSendToEndpoint)
	// terminalProtocol.Operation(TRM_APP_READ_KEY, "") --> Deprecated
	terminalProtocol.Operation(TRM_APP_RPC, "App RPC").Version(0).Executor(trmSendToEndpoint)
	terminalProtocol.Operation(TRM_APP_PUT_FILE, "App put file on TE").Version(0).Executor(trmSendToEndpoint)
	terminalProtocol.Operation(TRM_APP_GET_FILE, "App get file from TE").Version(0).Executor(trmSendToEndpoint)
	terminalProtocol.Operation(TRM_APP_KEY_BUFFER_LEN, "App key buffer len").Version(0).Executor(trmSendToEndpoint)
	// terminalProtocol.Operation(TRM_APP_RECONNECT, "")  --> TODO
	terminalProtocol.Operation(TRM_APP_KEEP_ALIVE, "App keep alive").Version(0).Executor(trmSendToEndpoint)
	terminalProtocol.Operation(TRM_APP_SESSION_CONFIG, "App session config").Version(1).Executor(sessionConfig)

	// Admin operations
	terminalProtocol.Operation(ADMIN_REQUEST_RESP_OP_CODE, "Admin request resp").Version(0).Executor(trmSendResponseToAdminClient)
}

func (req *SessionConfigRequest) FromBuffer(buf *buffer.ByteBuffer) {
	count := int(buf.GetInt32())
	req.config = make(map[string]string, count)
	for i := 0; i < count; i++ {
		req.config[strings.ToLower(buf.GetString())] = buf.GetString()
	}
}

func (req *SessionConfigRequest) ToBuffer(buf *buffer.ByteBuffer) {
	buf.PutInt32(int32(len(req.config)))
	for k, v := range req.config {
		buf.PutString(k)
		buf.PutString(v)
	}
}

func sessionConfig(proto *protocol.OperationVersion[*requestPack], pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	req := &SessionConfigRequest{}
	req.FromBuffer(buffer.Wrap(pack.packet.RemainingSlice()))
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
	return protocol.SuccessResponse(), nil
}

func trmSendToEndpoint(proto *protocol.OperationVersion[*requestPack], pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	pack.packet.Rewind()
	return nil, pack.handler.sendToEndpoint(pack.packet)
}

func trmSendLogoutToTE(proto *protocol.OperationVersion[*requestPack], pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	if pack.handler != nil && pack.handler.session != nil {
		pack.handler.session.SetStatus(server.SESS_CLOSE_REQUEST)
	}
	return trmSendToEndpoint(proto, pack)
}

func trmSendResponseToAdminClient(proto *protocol.OperationVersion[*requestPack], pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	adminId := pack.packet.GetUInt64()
	respLen := pack.packet.Remaining()
	lenPos := pack.packet.Position() - 4
	pack.packet.SetPosition(lenPos)
	pack.packet.PutUInt32(uint32(respLen))
	return nil, pack.handler.sendToAdminClient(adminId, pack.packet.GetBufferFrom(lenPos))
}
