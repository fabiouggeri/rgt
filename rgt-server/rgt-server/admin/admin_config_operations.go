package admin

import (
	"rgt-server/buffer"
	"rgt-server/log"
	"rgt-server/protocol"
	"rgt-server/server"
)

type SetConfigRequest struct {
	protocol.BaseRequest
	config map[string]string
}
type SetConfigRequestV6 struct {
	SetConfigRequest
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

var (
	_ protocol.RequestSerializerDeserializer  = &SetConfigRequest{}
	_ protocol.RequestSerializerDeserializer  = &SetConfigRequestV6{}
	_ protocol.RequestSerializerDeserializer  = &SetLogLevelRequest{}
	_ protocol.ResponseSerializerDeserializer = &GetConfigResponse{}
)

func init() {
	setConfigOp := adminProtocol.Operation(ADM_SET_CONFIG, "Set config")
	setConfigOp.Version(0).Executor(setServerConfig)
	setConfigOp.Version(6).Executor(setServerConfigV6)

	adminProtocol.Operation(ADM_GET_CONFIG, "Get config").Version(0).Executor(getServerConfig)
	adminProtocol.Operation(ADM_SAVE_CONFIG, "Save config").Version(0).Executor(saveServerConfig)
	adminProtocol.Operation(ADM_LOAD_CONFIG, "Load config").Version(0).Executor(loadServerConfig)
	adminProtocol.Operation(ADM_SET_LOG_LEVEL, "Set log level").Version(0).Executor(setLogLevel)
}

func (r *SetConfigRequest) FromBuffer(buf *buffer.ByteBuffer) {
	count := int(buf.GetInt32())
	r.config = make(map[string]string, count)
	for i := 0; i < count; i++ {
		r.config[buf.GetString()] = buf.GetString()
	}
}

func (r *SetConfigRequest) ToBuffer(buf *buffer.ByteBuffer) {
	buf.PutInt32(int32(len(r.config)))
	for k, v := range r.config {
		buf.PutString(k)
		buf.PutString(v)
	}
}

func (r *SetConfigRequestV6) FromBuffer(buf *buffer.ByteBuffer) {
	r.removeMissingOptions = buf.GetBool()
	r.SetConfigRequest.FromBuffer(buf)
}

func (r *SetConfigRequestV6) ToBuffer(buf *buffer.ByteBuffer) {
	buf.PutBool(r.removeMissingOptions)
	r.SetConfigRequest.ToBuffer(buf)
}

func removeMissing(srv *server.Server, req *SetConfigRequestV6) protocol.ErrorResponse {
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

func setServerConfig(proto *protocol.OperationVersion[*RequestPack], pack *RequestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_sessions_operations.setServerConfig()")
	if pack.handler.readOnly {
		return nil, NewError(NOT_ALLOWED_OPERATION, "Operation not allowed in read only session")
	}
	req := &SetConfigRequest{}
	req.FromBuffer(buffer.Wrap(pack.body))
	return setOptions(req.config, pack.handler.service.server)
}

func setServerConfigV6(proto *protocol.OperationVersion[*RequestPack], pack *RequestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_sessions_operations.setServerConfig()")
	if pack.handler.readOnly {
		return nil, NewError(NOT_ALLOWED_OPERATION, "Operation not allowed in read only session")
	}
	srv := pack.handler.service.server
	req := &SetConfigRequestV6{}
	req.FromBuffer(buffer.Wrap(pack.body))
	err := removeMissing(srv, req)
	if err != nil {
		return nil, err
	}
	return setOptions(req.config, srv)
}

func setOptions(config map[string]string, srv *server.Server) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	for option, value := range config {
		oldValue := srv.Config().GetValue(option)
		if !srv.Config().Set(option, value) {
			return nil, NewError(SERVER_ERROR, "Error setting ", option, " option")
		}
		if oldValue != value {
			log.Infof("Server config '%s' changed from '%s' to '%s'", option, oldValue, value)
		}
	}
	return protocol.SuccessResponse(), nil
}

func (r *GetConfigResponse) ToBuffer(buf *buffer.ByteBuffer) {
	buf.PutInt32(int32(len(r.config)))
	for k, v := range r.config {
		buf.PutString(k)
		buf.PutString(v)
	}
}

func (r *GetConfigResponse) FromBuffer(buf *buffer.ByteBuffer) {
	count := int(buf.GetInt32())
	r.config = make(map[string]string, count)
	for i := 0; i < count; i++ {
		r.config[buf.GetString()] = buf.GetString()
	}
}

func getServerConfig(proto *protocol.OperationVersion[*RequestPack], pack *RequestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	bufSize := 0
	props := pack.handler.service.server.Config().ToMap()
	resp := &GetConfigResponse{config: make(map[string]string, len(props))}
	for k, v := range props {
		resp.config[k] = v
		bufSize += len(k) + len(v) + 8
	}
	bufResp := buffer.NewCapacity(4 + uint32(bufSize))
	protocol.PutResponse(resp, bufResp)
	return bufResp, nil
}

func saveServerConfig(proto *protocol.OperationVersion[*RequestPack], pack *RequestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
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
	return protocol.SuccessResponse(), nil
}

func loadServerConfig(proto *protocol.OperationVersion[*RequestPack], pack *RequestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
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
	return protocol.SuccessResponse(), nil
}

func (r *SetLogLevelRequest) ToBuffer(buf *buffer.ByteBuffer) {
	buf.PutString(r.appLogLevel.Name())
	buf.PutString(r.serverLogLevel.Name())
	buf.PutString(r.teLogLevel.Name())
}

func (r *SetLogLevelRequest) FromBuffer(buf *buffer.ByteBuffer) {
	r.appLogLevel = log.LogLevelFromName(buf.GetString())
	r.serverLogLevel = log.LogLevelFromName(buf.GetString())
	r.teLogLevel = log.LogLevelFromName(buf.GetString())
}

func setLogLevel(proto *protocol.OperationVersion[*RequestPack], pack *RequestPack) (*buffer.ByteBuffer, protocol.ErrorResponse) {
	log.Debug("admin_sessions_operations.setLogLevel()")
	if pack.handler.readOnly {
		return nil, NewError(NOT_ALLOWED_OPERATION, "Operation not allowed in read only session")
	}
	req := &SetLogLevelRequest{}
	req.FromBuffer(buffer.Wrap(pack.body))
	cfg := pack.handler.service.server.Config()
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
	return protocol.SuccessResponse(), nil
}
