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

type ServerStatsResponse struct {
	protocol.BaseResponse
	bytesReceived   uint64
	bytesSent       uint64
	packetsReceived uint64
	packetsSent     uint64
}

func init() {
	adminProtocol.Operation(ADM_GET_STATUS, "Get server status").Version(0).Executor(getServerStatus)
	adminProtocol.Operation(ADM_GET_STATS, "Get server statistics").Version(7).Executor(getServerStats)
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
		sessionsCount: srv.GetSessionsCount(),
		startTime:     srv.GetStartTime(),
	}
	respBuf := buffer.NewCapacity(uint32(len(resp.serverStatus) + 4 + 4 + 8))
	protocol.PutResponse(resp, respBuf)
	return respBuf, nil
}

func (r *ServerStatsResponse) FromBuffer(buf *buffer.ByteBuffer) {
	r.bytesReceived = buf.GetUInt64()
	r.bytesSent = buf.GetUInt64()
	r.packetsReceived = buf.GetUInt64()
	r.packetsSent = buf.GetUInt64()
}

func (r *ServerStatsResponse) ToBuffer(buf *buffer.ByteBuffer) {
	buf.PutUInt64(r.bytesReceived)
	buf.PutUInt64(r.bytesSent)
	buf.PutUInt64(r.packetsReceived)
	buf.PutUInt64(r.packetsSent)
}

func getServerStats(proto *protocol.OperationVersion[*RequestPack], pack *RequestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_server_operations.getServerStats()")
	srv := pack.handler.service.server
	stats := srv.GetStats()
	resp := &ServerStatsResponse{
		bytesReceived:   stats.BytesReceived(),
		bytesSent:       stats.BytesSent(),
		packetsReceived: stats.PacketsReceived(),
		packetsSent:     stats.PacketsSent(),
	}
	respBuf := buffer.NewCapacity(40)
	protocol.PutResponse(resp, respBuf)
	return respBuf, nil
}
