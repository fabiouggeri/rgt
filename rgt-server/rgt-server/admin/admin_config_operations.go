package admin

import (
	"rgt-server/buffer"
	"rgt-server/log"
	"rgt-server/protocol"
	"rgt-server/server"
)

type SetConfigRequest struct {
	protocol.BaseRequest
	config               map[string]string
	removeMissingOptions bool
}

type SetLogLevelRequest struct {
	protocol.BaseRequest
	appLogLevel    log.LogLevel
	teLogLevel     log.LogLevel
	serverLogLevel log.LogLevel
}

type GetConfigResponse struct {
	protocol.BaseResponse
	config map[string]string
}

func init() {
	registerOperation(ADM_SET_CONFIG, setServerConfig)
	registerOperation(ADM_GET_CONFIG, getServerConfig)
	registerOperation(ADM_SAVE_CONFIG, saveServerConfig)
	registerOperation(ADM_LOAD_CONFIG, loadServerConfig)
	registerOperation(ADM_SET_LOG_LEVEL, setLogLevel)
	registerProtocol(ADM_SET_CONFIG, 0, protocol.New(bufferToSetConfigRequest, setConfigRequestToBuffer, protocol.BufferToBaseResponse, protocol.BaseResponseToBuffer))
	registerProtocol(ADM_SET_CONFIG, 6, protocol.New(bufferToSetConfigRequestV6, setConfigRequestToBufferV6, protocol.BufferToBaseResponse, protocol.BaseResponseToBuffer))
	registerProtocol(ADM_GET_CONFIG, 0, protocol.New(protocol.BufferToBaseRequest, protocol.BaseRequestToBuffer, bufferToGetConfigResponse, getConfigResponseToBuffer))
	registerProtocol(ADM_SAVE_CONFIG, 0, protocol.New(protocol.BufferToBaseRequest, protocol.BaseRequestToBuffer, protocol.BufferToBaseResponse, protocol.BaseResponseToBuffer))
	registerProtocol(ADM_LOAD_CONFIG, 0, protocol.New(protocol.BufferToBaseRequest, protocol.BaseRequestToBuffer, protocol.BufferToBaseResponse, protocol.BaseResponseToBuffer))
	registerProtocol(ADM_SET_LOG_LEVEL, 0, protocol.New(bufferToSetLogLevelRequest, setLogLevelRequestToBuffer, protocol.BufferToBaseResponse, protocol.BaseResponseToBuffer))
}

func bufferToSetConfigRequest(buf *buffer.ByteBuffer) *SetConfigRequest {
	req := &SetConfigRequest{config: make(map[string]string)}
	count := int(buf.GetInt32())
	for i := 0; i < count; i++ {
		req.config[buf.GetString()] = buf.GetString()
	}
	return req
}

func setConfigRequestToBuffer(req *SetConfigRequest, buf *buffer.ByteBuffer) {
	buf.PutInt32(int32(len(req.config)))
	for k, v := range req.config {
		buf.PutString(k)
		buf.PutString(v)
	}
}

func bufferToSetConfigRequestV6(buf *buffer.ByteBuffer) *SetConfigRequest {
	req := &SetConfigRequest{config: make(map[string]string)}
	req.removeMissingOptions = buf.GetBool()
	count := int(buf.GetInt32())
	for i := 0; i < count; i++ {
		req.config[buf.GetString()] = buf.GetString()
	}
	return req
}

func setConfigRequestToBufferV6(req *SetConfigRequest, buf *buffer.ByteBuffer) {
	buf.PutBool(req.removeMissingOptions)
	buf.PutInt32(int32(len(req.config)))
	for k, v := range req.config {
		buf.PutString(k)
		buf.PutString(v)
	}
}

func removeMissing(srv *server.Server, req *SetConfigRequest) protocol.ErrorResponse {
	if req.removeMissingOptions {
		for k := range srv.Config().ToMap() {
			_, found := req.config[k]
			if !found {
				_, errDel := srv.Config().Delete(k)
				if errDel != nil {
					return NewError(SERVER_ERROR, errDel)
				}
			}
		}
	}
	return nil
}

func setServerConfig(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_sessions_operations.setServerConfig()")
	if pack.handler.readOnly {
		return nil, NewError(NOT_ALLOWED_OPERATION, "Operation not allowed in read only session")
	}
	proto, err := findProtocol[*SetConfigRequest, *protocol.BaseResponse](ADM_SET_CONFIG, pack.handler.protocolVersion)
	if err != nil {
		return nil, err
	}
	srv := pack.handler.service.server
	bufReq := buffer.Wrap(pack.body)
	req := proto.GetRequest(bufReq)
	err = removeMissing(srv, req)
	if err != nil {
		return nil, err
	}
	for option, value := range req.config {
		oldValue := srv.Config().GetValue(option)
		if !srv.Config().Set(option, value) {
			return nil, NewError(SERVER_ERROR, "Error setting ", option, " option")
		}
		if oldValue != value {
			log.Infof("Server config '%s' changed from '%s' to '%s'", option, oldValue, value)
		}
	}
	return SuccessAdminResponse(), nil
}

func bufferToGetConfigResponse(buf *buffer.ByteBuffer) *GetConfigResponse {
	resp := &GetConfigResponse{config: make(map[string]string)}
	count := int(buf.GetInt32())
	for i := 0; i < count; i++ {
		resp.config[buf.GetString()] = buf.GetString()
	}
	return resp
}

func getConfigResponseToBuffer(resp *GetConfigResponse, buf *buffer.ByteBuffer) {
	buf.PutInt32(int32(len(resp.config)))
	for k, v := range resp.config {
		buf.PutString(k)
		buf.PutString(v)
	}
}

func getServerConfig(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	proto, err := findProtocol[*protocol.BaseRequest, *GetConfigResponse](ADM_GET_CONFIG, pack.handler.protocolVersion)
	if err != nil {
		return nil, err
	}
	srv := pack.handler.service.server
	resp := &GetConfigResponse{config: make(map[string]string)}
	bufSize := 0
	props := srv.Config().ToMap()
	for k, v := range props {
		resp.config[k] = v
		bufSize += len(k) + len(v) + 8
	}
	bufResp := buffer.NewCapacity(4 + uint32(bufSize))
	proto.PutResponse(resp, bufResp)
	return bufResp, nil
}

func saveServerConfig(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_sessions_operations.saveServerConfig()")
	if pack.handler.readOnly {
		return nil, NewError(NOT_ALLOWED_OPERATION, "Operation not allowed in read only session")
	}
	cfg := pack.handler.service.server.Config()
	err := cfg.Save()
	if err != nil {
		return nil, NewError(SERVER_ERROR, "Error saving configuration in file ", cfg.GetFilePathName(), " Error: ", err)
	}
	log.Infof("Server configuration saved in %s", cfg.GetFilePathName())
	return SuccessAdminResponse(), nil
}

func loadServerConfig(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_sessions_operations.loadServerConfig()")
	if pack.handler.readOnly {
		return nil, NewError(NOT_ALLOWED_OPERATION, "Operation not allowed in read only session")
	}
	cfg := pack.handler.service.server.Config()
	err := cfg.Reload()
	if err != nil {
		return nil, NewError(SERVER_ERROR, "Error loading configuration from file ", cfg.GetFilePathName(), " Error: ", err)
	}
	log.Infof("Server configuration reload from in %s", cfg.GetFilePathName())
	return SuccessAdminResponse(), nil
}

func bufferToSetLogLevelRequest(buf *buffer.ByteBuffer) *SetLogLevelRequest {
	return &SetLogLevelRequest{
		appLogLevel:    log.LogLevelFromName(buf.GetString()),
		serverLogLevel: log.LogLevelFromName(buf.GetString()),
		teLogLevel:     log.LogLevelFromName(buf.GetString())}
}

func setLogLevelRequestToBuffer(req *SetLogLevelRequest, buf *buffer.ByteBuffer) {
	buf.PutString(req.appLogLevel.Name())
	buf.PutString(req.serverLogLevel.Name())
	buf.PutString(req.teLogLevel.Name())
}

func setLogLevel(pack *requestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_sessions_operations.setLogLevel()")
	if pack.handler.readOnly {
		return nil, NewError(NOT_ALLOWED_OPERATION, "Operation not allowed in read only session")
	}
	proto, err := findProtocol[*SetLogLevelRequest, *protocol.BaseResponse](ADM_SET_LOG_LEVEL, pack.handler.protocolVersion)
	if err != nil {
		return nil, err
	}
	bufReq := buffer.Wrap(pack.body)
	req := proto.GetRequest(bufReq)
	server := pack.handler.service.server
	cfg := server.Config()
	oldTELevel := cfg.TeLogLevel().Get()
	oldServerLevel := cfg.ServerLogLevel().Get()
	oldAppLevel := cfg.AppLogLevel().Get()
	if req.teLogLevel != oldTELevel {
		cfg.TeLogLevel().Set(req.teLogLevel)
		log.Info("Terminal log level changed from %s to %s", oldTELevel, req.teLogLevel)
	}
	if req.serverLogLevel != oldServerLevel {
		cfg.ServerLogLevel().Set(req.serverLogLevel)
		log.SetLevel(req.serverLogLevel)
		log.Info("Server log level changed from %s to %s", oldServerLevel, req.serverLogLevel)
	}
	if req.appLogLevel != oldAppLevel {
		cfg.AppLogLevel().Set(req.appLogLevel)
		log.Info("App log level changed from %s to %s", oldAppLevel, req.appLogLevel)
	}
	return SuccessAdminResponse(), nil
}
