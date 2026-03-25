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
	registerOperation(ADM_GET_STATUS, getServerStatus)
	registerOperation(ADM_GET_STATS, getServerStats)
	registerProtocol(ADM_GET_STATUS, 0, protocol.New(protocol.BufferToBaseRequest, protocol.BaseRequestToBuffer, bufferToServerInfoResponse, serverInfoResponseToBuffer))
	registerProtocol(ADM_GET_STATS, 7, protocol.New(protocol.BufferToBaseRequest, protocol.BaseRequestToBuffer, bufferToServerStatsResponse, serverStatsResponseToBuffer))
}

func bufferToServerInfoResponse(buf *buffer.ByteBuffer) *ServerInfoResponse {
	return &ServerInfoResponse{
		serverStatus:  server.ServerStatus(buf.GetString()),
		sessionsCount: buf.GetInt32(),
		startTime:     buf.GetInt64()}
}

func serverInfoResponseToBuffer(resp *ServerInfoResponse, buf *buffer.ByteBuffer) {
	buf.PutString(string(resp.serverStatus))
	buf.PutInt32(resp.sessionsCount)
	buf.PutInt64(resp.startTime)
}

func getServerStatus(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_server_operations.getServerStatus()")
	proto, err := findProtocol[*protocol.BaseRequest, *ServerInfoResponse](ADM_GET_STATUS, pack.handler.protocolVersion)
	if err != nil {
		return nil, err
	}
	srv := pack.handler.service.server
	resp := &ServerInfoResponse{serverStatus: srv.GetStatus(),
		sessionsCount: srv.GetSessionsCount(),
		startTime:     srv.GetStartTime()}
	respBuf := buffer.NewCapacity(uint32(len(resp.serverStatus) + 4 + 4 + 8))
	proto.PutResponse(resp, respBuf)
	return respBuf, nil
}

func bufferToServerStatsResponse(buf *buffer.ByteBuffer) *ServerStatsResponse {
	return &ServerStatsResponse{
		bytesReceived:   buf.GetUInt64(),
		bytesSent:       buf.GetUInt64(),
		packetsReceived: buf.GetUInt64(),
		packetsSent:     buf.GetUInt64(),
	}
}

func serverStatsResponseToBuffer(resp *ServerStatsResponse, buf *buffer.ByteBuffer) {
	buf.PutUInt64(resp.bytesReceived)
	buf.PutUInt64(resp.bytesSent)
	buf.PutUInt64(resp.packetsReceived)
	buf.PutUInt64(resp.packetsSent)
}

func getServerStats(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_server_operations.getServerStats()")
	proto, err := findProtocol[*protocol.BaseRequest, *ServerStatsResponse](ADM_GET_STATS, pack.handler.protocolVersion)
	if err != nil {
		return nil, err
	}
	srv := pack.handler.service.server
	stats := srv.GetStats()
	resp := &ServerStatsResponse{bytesReceived: stats.BytesReceived(),
		bytesSent:       stats.BytesSent(),
		packetsReceived: stats.PacketsReceived(),
		packetsSent:     stats.PacketsSent(),
	}
	respBuf := buffer.NewCapacity(40)
	proto.PutResponse(resp, respBuf)
	return respBuf, nil
}
