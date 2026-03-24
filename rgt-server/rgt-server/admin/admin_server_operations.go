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
	registerOperation(ADM_GET_STATUS, getServerStatus)
	registerProtocol(ADM_GET_STATUS, 0, protocol.New(protocol.BufferToBaseRequest, protocol.BaseRequestToBuffer, bufferToServerInfoResponse, serverInfoResponseToBuffer))
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
