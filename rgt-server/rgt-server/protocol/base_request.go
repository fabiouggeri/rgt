package protocol

import "rgt-server/buffer"

type BaseRequest struct {
	OperationCode OperationCode
}

func (r *BaseRequest) GetOperationCode() OperationCode {
	return r.OperationCode
}

func BufferToBaseRequest(buf *buffer.ByteBuffer) *BaseRequest {
	return &BaseRequest{}
}

func BaseRequestToBuffer(req *BaseRequest, buf *buffer.ByteBuffer) {
	buf.PutInt8(int8(req.OperationCode))
}
