package protocol

import (
	"fmt"
	"rgt-server/buffer"
)

type Protocol[T any] struct {
	operations map[OperationCode]*Operation[T]
	responses  map[ResponseCode]string
}

type ProtocolError struct {
	message   string
	errorCode ResponseCode
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

func New[T any](responses map[ResponseCode]string) *Protocol[T] {
	return &Protocol[T]{
		operations: make(map[OperationCode]*Operation[T]),
		responses:  responses,
	}
}

func (p *Protocol[T]) Operation(op OperationCode, description string) *Operation[T] {
	if operation, found := p.operations[op]; found {
		return operation
	}
	operation := &Operation[T]{
		code:        op,
		description: description,
		versions:    make([]OperationVersion[T], 0, 8),
	}
	p.operations[op] = operation
	return operation
}

func (p *Protocol[T]) FindOperation(op OperationCode, version int16) (*OperationVersion[T], error) {
	operation, found := p.operations[op]
	if !found {
		return nil, fmt.Errorf("Protocol not found for operation %v", op)
	}
	return operation.FindVersion(version)
}

func (p *Protocol[T]) NewError(respCode ResponseCode, message ...any) *ProtocolError {
	if len(message) > 0 {
		msg := fmt.Sprint(message...)
		if msg != "" {
			return &ProtocolError{
				errorCode: respCode,
				message:   msg,
			}
		}
	}
	return &ProtocolError{
		errorCode: respCode,
		message:   p.GetResponseCodeDescription(respCode),
	}
}

func (p *Protocol[T]) GetResponseCodeDescription(respCode ResponseCode) string {
	if msg, found := p.responses[respCode]; found {
		return msg
	}
	return fmt.Sprint("Unknown error: ", respCode)
}

func (e *ProtocolError) GetResponseCode() ResponseCode {
	return e.errorCode
}

func (e *ProtocolError) Error() string {
	return e.message
}

// func RequestFromBuffer(req RequestSerializerDeserializer, buffer *buffer.ByteBuffer) {
// 	req.FromBuffer(buffer)
// }

// func RequestToBuffer(req RequestSerializerDeserializer, buffer *buffer.ByteBuffer) {
// 	req.ToBuffer(buffer)
// }

// func ResponseFromBuffer(req ResponseSerializerDeserializer, buffer *buffer.ByteBuffer) {
// 	req.FromBuffer(buffer)
// }

// func ResponseToBuffer(req ResponseSerializerDeserializer, buffer *buffer.ByteBuffer) {
// 	req.ToBuffer(buffer)
// }

func PutRequest(request RequestSerializerDeserializer, buf *buffer.ByteBuffer) {
	buf.Clear()
	buf.PutUInt32(0)
	buf.PutInt8(int8(request.GetOperationCode()))
	request.ToBuffer(buf)
	pos := buf.Position()
	buf.Flip()
	buf.PutUInt32(uint32(pos - 4))
	buf.Rewind()
}

func PutRequestFirstOp(request RequestSerializerDeserializer, buf *buffer.ByteBuffer) {
	buf.Clear()
	buf.PutInt32(MAGIC_NUMBER)
	buf.PutUInt32(0)
	buf.PutInt8(int8(request.GetOperationCode()))
	request.ToBuffer(buf)
	pos := buf.Position()
	buf.Flip()
	buf.Skip(4)
	buf.PutUInt32(uint32(pos - 8))
	buf.Rewind()
}

func PutResponse(response ResponseSerializerDeserializer, buf *buffer.ByteBuffer) {
	buf.Clear()
	buf.PutUInt32(0)
	buf.PutInt16(int16(response.GetCode()))
	response.ToBuffer(buf)
	pos := buf.Position()
	buf.Flip()
	buf.PutUInt32(uint32(pos - 4))
	buf.Rewind()
}
