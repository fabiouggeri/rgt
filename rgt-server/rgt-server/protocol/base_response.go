package protocol

import "rgt-server/buffer"

type ResponseCode uint16

type Response interface {
	GetCode() ResponseCode
	GetMessage() string
}

type ErrorResponse interface {
	error
	GetResponseCode() ResponseCode
}

type BaseResponse struct {
	Message string
	Code    ResponseCode
}

var _ ResponseSerializerDeserializer = &BaseResponse{}

func baseResponseCreator() ResponseSerializerDeserializer {
	return &BaseResponse{}
}

func SuccessResponse() *buffer.ByteBuffer {
	resp := &BaseResponse{
		Code: SUCCESS,
	}
	respBuf := buffer.NewCapacity(8)
	resp.ToBuffer(respBuf)
	return respBuf
}

func (r *BaseResponse) GetCode() ResponseCode {
	return r.Code
}

func (r *BaseResponse) GetMessage() string {
	return r.Message
}

func (r *BaseResponse) FromBuffer(buf *buffer.ByteBuffer) {
	r.Code = ResponseCode(buf.GetInt16())
	r.Message = buf.GetString()
}

func (r *BaseResponse) ToBuffer(buf *buffer.ByteBuffer) {
	buf.PutString(r.Message)
}
