package terminal

import (
	"rgt-server/buffer"
	"rgt-server/protocol"
	"rgt-server/service"
	"time"
)

type adminClient struct {
	id              uint64
	terminalHandler *TerminalHandler
	adminConnection service.ConnectionHandler
	responses       chan *buffer.ByteBuffer
}

func newAdminClient(handler *TerminalHandler, adminConn service.ConnectionHandler) *adminClient {
	return &adminClient{id: adminConn.Id(),
		terminalHandler: handler,
		adminConnection: adminConn,
		responses:       make(chan *buffer.ByteBuffer, 128)}
}

func (a *adminClient) SendRequest(requestCode protocol.OperationCode, data []byte) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	if a.terminalHandler.protocolVersion >= 5 {
		sendBuffer := protocol.NewBufferRequest(requestCode)
		sendBuffer.PutUInt64(a.id)
		sendBuffer.Put(data)
		protocol.FinalizeBufferRequest(sendBuffer)
		a.terminalHandler.Send(sendBuffer)
		select {
		case resp := <-a.responses:
			return resp, nil
		case <-time.After(30 * time.Second):
			return nil, NewError(PROTOCOL_ERROR, "admin request timeout")
		}
	} else {
		return nil, NewError(PROTOCOL_ERROR, "Client doesn't support this feature")
	}
}

func (a *adminClient) ProcessPacket(packet *buffer.ByteBuffer) {
	select {
	case a.responses <- packet:
	default:
	}
}
