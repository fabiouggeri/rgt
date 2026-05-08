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
	adminProtocol.Operation(ADM_STOP_SERVICE, "Stop service").Version(0).Executor(stopService)
	adminProtocol.Operation(ADM_START_SERVICE, "Start service").Version(0).Executor(startService)
}

func (r *ServerStatusResponse) FromBuffer(buf *buffer.ByteBuffer) {
	r.serverStatus = server.ServerStatus(buf.GetString())
}

func (r *ServerStatusResponse) ToBuffer(buf *buffer.ByteBuffer) {
	buf.PutString(string(r.serverStatus))
}

func stopService(proto *protocol.OperationVersion[*RequestPack], pack *RequestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_services_operations.stopService()")
	if pack.handler.readOnly {
		return nil, NewError(NOT_ALLOWED_OPERATION, "Operation not allowed in read only session")
	}
	srv := pack.handler.service.server
	if srv.GetStatus() != server.SERVER_RUNNING {
		return nil, NewError(INVALID_STATUS, "Sever is not running")
	}
	stopErr := srv.Stop(service.SERVICE_EMULATION)
	if stopErr != nil {
		return nil, NewError(SERVER_ERROR, "Error stopping service: ", stopErr)
	}
	resp := &ServerStatusResponse{
		serverStatus: srv.GetStatus(),
	}
	bufResp := buffer.NewCapacity(uint32(protocol.RESPONSE_HEADER_SIZE + buffer.STRING_HEADER_SIZE + len(resp.serverStatus)))
	protocol.PutResponse(resp, bufResp)
	return bufResp, nil
}

func startService(proto *protocol.OperationVersion[*RequestPack], pack *RequestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_services_operations.startService()")
	if pack.handler.readOnly {
		return nil, NewError(NOT_ALLOWED_OPERATION, "Operation not allowed in read only session")
	}
	srv := pack.handler.service.server
	if srv.GetStatus() != server.SERVER_STOPPED {
		return nil, NewError(INVALID_STATUS, "Sever is not stopped")
	}
	startErr := srv.Start(service.SERVICE_EMULATION)
	if startErr != nil {
		return nil, NewError(SERVER_ERROR, "Error starting service: ", startErr)
	}
	resp := &ServerStatusResponse{
		serverStatus: srv.GetStatus(),
	}
	bufResp := buffer.NewCapacity(uint32(protocol.RESPONSE_HEADER_SIZE + buffer.STRING_HEADER_SIZE + len(resp.serverStatus)))
	protocol.PutResponse(resp, bufResp)
	return bufResp, nil
}
