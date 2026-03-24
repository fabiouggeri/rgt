package protocol

import "rgt-server/buffer"

type BaseResponse struct {
	Message string
	Code    ResponseCode
}

func (r *BaseResponse) GetCode() ResponseCode {
	return r.Code
}

func (r *BaseResponse) GetMessage() string {
	return r.Message
}

func BaseResponseToBuffer(resp *BaseResponse, buf *buffer.ByteBuffer) {
	buf.PutString(resp.Message)
}

func BufferToBaseResponse(buf *buffer.ByteBuffer) *BaseResponse {
	return &BaseResponse{
		Code:    ResponseCode(buf.GetInt16()),
		Message: buf.GetString()}
}
