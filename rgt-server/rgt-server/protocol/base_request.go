package protocol

import "rgt-server/buffer"

type Request interface {
	GetOperationCode() OperationCode
}

type BaseRequest struct {
	OperationCode OperationCode
}

var _ RequestSerializerDeserializer = &BaseRequest{}

func baseRequestCreator() RequestSerializerDeserializer {
	return &BaseRequest{}
}

func (r *BaseRequest) GetOperationCode() OperationCode {
	return r.OperationCode
}

func (r *BaseRequest) FromBuffer(buf *buffer.ByteBuffer) {
	r.OperationCode = OperationCode(buf.GetUInt8())
}

func (r *BaseRequest) ToBuffer(buf *buffer.ByteBuffer) {
	buf.PutInt8(int8(r.OperationCode))
}
