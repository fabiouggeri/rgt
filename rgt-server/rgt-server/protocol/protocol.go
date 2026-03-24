package protocol

import (
	"fmt"
	"rgt-server/buffer"
)

type OperationCode uint8
type ResponseCode uint16

type Request interface {
	GetOperationCode() OperationCode
}

type Response interface {
	GetCode() ResponseCode
	GetMessage() string
}

type ErrorResponse interface {
	error
	GetResponseCode() ResponseCode
}

type Protocol[R Request, S Response] struct {
	bufferToRequest func(buf *buffer.ByteBuffer) R
	requestToBuffer func(request R, buf *buffer.ByteBuffer)

	bufferToResponse func(buf *buffer.ByteBuffer) S
	responseToBuffer func(response S, buf *buffer.ByteBuffer)
}

const (
	PACK_SIZE_FIELD_SIZE    int   = 4
	OP_FIELD_SIZE           int   = 1
	MAGIC_NUMBER_FIELD_SIZE int   = 4
	RESP_CODE_FIELD_SIZE    int   = 2
	HEADER_SIZE             int   = PACK_SIZE_FIELD_SIZE + OP_FIELD_SIZE
	FIRST_HEADER_SIZE       int   = HEADER_SIZE + MAGIC_NUMBER_FIELD_SIZE
	RESPONSE_HEADER_SIZE    int   = PACK_SIZE_FIELD_SIZE + RESP_CODE_FIELD_SIZE
	MAGIC_NUMBER            int32 = 0x5CDBA4EA
	DEFAULT_IO_BUFFER_SIZE  int   = 4096
)

func New[R Request, S Response](bufToReq func(buf *buffer.ByteBuffer) R, reqToBuf func(request R, buf *buffer.ByteBuffer),
	bufToResp func(buf *buffer.ByteBuffer) S, respToBuf func(response S, buf *buffer.ByteBuffer)) *Protocol[R, S] {
	return &Protocol[R, S]{
		bufferToRequest:  bufToReq,
		requestToBuffer:  reqToBuf,
		bufferToResponse: bufToResp,
		responseToBuffer: respToBuf}
}

func NewDefault() *Protocol[*BaseRequest, *BaseResponse] {
	return &Protocol[*BaseRequest, *BaseResponse]{
		bufferToRequest:  BufferToBaseRequest,
		requestToBuffer:  BaseRequestToBuffer,
		bufferToResponse: BufferToBaseResponse,
		responseToBuffer: BaseResponseToBuffer}
}

func NewResponse(code ResponseCode, message ...any) *BaseResponse {
	return &BaseResponse{Code: code, Message: fmt.Sprint(message...)}
}

func ResponseFromError(err ErrorResponse) *BaseResponse {
	return &BaseResponse{Code: err.GetResponseCode(), Message: err.Error()}
}

func NewBufferRequest(opCode OperationCode) *buffer.ByteBuffer {
	buf := buffer.NewCapacity(16)
	buf.PutUInt32(0)
	buf.PutUInt8(uint8(opCode))
	return buf
}

func FinalizeBufferRequest(buf *buffer.ByteBuffer) {
	pos := buf.Position()
	buf.Flip()
	buf.PutUInt32(uint32(pos - 4))
	buf.Rewind()
}

func (p *Protocol[R, S]) GetRequest(buf *buffer.ByteBuffer) R {
	return p.bufferToRequest(buf)
}

func (p *Protocol[R, S]) GetResponse(buf *buffer.ByteBuffer) S {
	return p.bufferToResponse(buf)
}

func (p *Protocol[R, S]) PutRequest(request R, buf *buffer.ByteBuffer) {
	buf.Clear()
	buf.PutUInt32(0)
	buf.PutInt8(int8(request.GetOperationCode()))
	p.requestToBuffer(request, buf)
	pos := buf.Position()
	buf.Flip()
	buf.PutUInt32(uint32(pos - 4))
	buf.Rewind()
}

func (p *Protocol[R, S]) PutRequestFirstOp(request R, buf *buffer.ByteBuffer) {
	buf.Clear()
	buf.PutInt32(MAGIC_NUMBER)
	buf.PutUInt32(0)
	buf.PutInt8(int8(request.GetOperationCode()))
	p.requestToBuffer(request, buf)
	pos := buf.Position()
	buf.Flip()
	buf.Skip(4)
	buf.PutUInt32(uint32(pos - 8))
	buf.Rewind()
}

func (p *Protocol[R, S]) PutResponse(response S, buf *buffer.ByteBuffer) {
	buf.Clear()
	buf.PutUInt32(0)
	buf.PutInt16(int16(response.GetCode()))
	p.responseToBuffer(response, buf)
	pos := buf.Position()
	buf.Flip()
	buf.PutUInt32(uint32(pos - 4))
	buf.Rewind()
}

func (p *Protocol[R, S]) PutTrmAppResponse(response S, buf *buffer.ByteBuffer) {
	buf.Clear()
	buf.PutUInt32(0)
	buf.PutInt16(int16(response.GetCode()))
	p.responseToBuffer(response, buf)
	pos := buf.Position()
	buf.Flip()
	buf.PutUInt32(uint32(pos - 4))
	buf.Rewind()
}
