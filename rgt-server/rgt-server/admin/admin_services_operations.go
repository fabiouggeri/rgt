package admin

import (
	"rgt-server/buffer"
	"rgt-server/log"
	"rgt-server/protocol"
	"rgt-server/server"
	"rgt-server/service"
)

type ServerStatusResponse struct {
	protocol.BaseResponse
	serverStatus server.ServerStatus
}

func init() {
	registerOperation(ADM_STOP_SERVICE, stopService)
	registerOperation(ADM_START_SERVICE, startService)
	registerProtocol(ADM_STOP_SERVICE, 0, protocol.New(protocol.BufferToBaseRequest, protocol.BaseRequestToBuffer, bufferToServerStatusResponse, serverStatusResponseToBuffer))
	registerProtocol(ADM_START_SERVICE, 0, protocol.New(protocol.BufferToBaseRequest, protocol.BaseRequestToBuffer, bufferToServerStatusResponse, serverStatusResponseToBuffer))
}

func bufferToServerStatusResponse(buf *buffer.ByteBuffer) *ServerStatusResponse {
	return &ServerStatusResponse{serverStatus: server.ServerStatus(buf.GetString())}
}

func serverStatusResponseToBuffer(resp *ServerStatusResponse, buf *buffer.ByteBuffer) {
	buf.PutString(string(resp.serverStatus))
}

func stopService(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_services_operations.stopService()")
	if pack.handler.readOnly {
		return nil, NewError(NOT_ALLOWED_OPERATION, "Operation not allowed in read only session")
	}
	proto, err := findProtocol[*protocol.BaseRequest, *ServerStatusResponse](ADM_STOP_SERVICE, pack.handler.protocolVersion)
	if err != nil {
		return nil, err
	}
	srv := pack.handler.service.server
	if srv.GetStatus() != server.SERVER_RUNNING {
		return nil, NewError(INVALID_STATUS, "Sever is not running")
	}
	stopErr := srv.Stop(service.SERVICE_EMULATION)
	if stopErr != nil {
		return nil, NewError(SERVER_ERROR, "Error stopping service: ", stopErr)
	}
	resp := &ServerStatusResponse{serverStatus: srv.GetStatus()}
	bufResp := buffer.NewCapacity(uint32(len(resp.serverStatus) + 4))
	proto.PutResponse(resp, bufResp)
	return bufResp, nil
}

func startService(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_services_operations.startService()")
	if pack.handler.readOnly {
		return nil, NewError(NOT_ALLOWED_OPERATION, "Operation not allowed in read only session")
	}
	proto, err := findProtocol[*protocol.BaseRequest, *ServerStatusResponse](ADM_STOP_SERVICE, pack.handler.protocolVersion)
	if err != nil {
		return nil, err
	}
	srv := pack.handler.service.server
	if srv.GetStatus() != server.SERVER_STOPPED {
		return nil, NewError(INVALID_STATUS, "Sever is not stopped")
	}
	startErr := srv.Start(service.SERVICE_EMULATION)
	if startErr != nil {
		return nil, NewError(SERVER_ERROR, "Error starting service: ", startErr)
	}
	resp := &ServerStatusResponse{serverStatus: srv.GetStatus()}
	bufResp := buffer.NewCapacity(uint32(len(resp.serverStatus) + 4))
	proto.PutResponse(resp, bufResp)
	return bufResp, nil
}
