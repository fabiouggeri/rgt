package admin

import (
	"rgt-server/buffer"
	"rgt-server/log"
	"rgt-server/protocol"
	"rgt-server/server"
)

type ServerInfoResponse struct {
	protocol.BaseResponse
	serverStatus  server.ServerStatus
	startTime     int64
	sessionsCount int32
}

func init() {
	adminProtocol.Operation(ADM_GET_STATUS, "Get server status").Version(0).Executor(getServerStatus)
}

func (s *ServerInfoResponse) FromBuffer(buf *buffer.ByteBuffer) {
	s.serverStatus = server.ServerStatus(buf.GetString())
	s.sessionsCount = buf.GetInt32()
	s.startTime = buf.GetInt64()
}

func (s *ServerInfoResponse) ToBuffer(buf *buffer.ByteBuffer) {
	buf.PutString(string(s.serverStatus))
	buf.PutInt32(s.sessionsCount)
	buf.PutInt64(s.startTime)
}

func getServerStatus(proto *protocol.OperationVersion[*RequestPack], pack *RequestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_server_operations.getServerStatus()")
	srv := pack.handler.service.server
	resp := &ServerInfoResponse{
		serverStatus:  srv.GetStatus(),
		sessionsCount: pack.handler.service.terminalService.GetSessionsCount(),
		startTime:     srv.GetStartTime(),
	}
	respBuf := buffer.NewCapacity(uint32(len(resp.serverStatus) + 4 + 4 + 8))
	protocol.PutResponse(resp, respBuf)
	return respBuf, nil
}
