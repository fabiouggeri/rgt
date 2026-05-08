package protocol

import (
	"fmt"
	"rgt-server/buffer"
)

type OperationCode uint8

type operationHandle[T any] func(op *OperationVersion[T], pack T) (*buffer.ByteBuffer, ErrorResponse)

type SerializerDeserializer interface {
	FromBuffer(buf *buffer.ByteBuffer)
	ToBuffer(buf *buffer.ByteBuffer)
}

type ResponseSerializerDeserializer interface {
	Response
	SerializerDeserializer
}

type RequestSerializerDeserializer interface {
	Request
	SerializerDeserializer
}

type OperationVersion[T any] struct {
	version  int16
	executor operationHandle[T]
}

type Operation[T any] struct {
	versions    []OperationVersion[T]
	description string
	code        OperationCode
}

const (
	SUCCESS                   ResponseCode  = 0x00
	NOT_IMPLEMENTED_ERROR     ResponseCode  = 0x7CFF // 31999
	INVALID_OPERATION_VERSION               = -1
	UNKNOWN_OPERATION         OperationCode = 0xFF
)

func (p *Operation[T]) FindVersion(version int16) (*OperationVersion[T], error) {
	var opVersion *OperationVersion[T]
	if len(p.versions) == 0 || version < 0 {
		return nil, fmt.Errorf("Protocol version (%v) not found", version)
	}
	if int(version) >= len(p.versions) {
		opVersion = &p.versions[len(p.versions)-1]
	} else {
		opVersion = &p.versions[version]
	}
	if opVersion.version == INVALID_OPERATION_VERSION {
		return nil, fmt.Errorf("Protocol version (%v) not found", version)
	}
	return opVersion, nil
}

func (p *Operation[T]) Version(version int) *OperationVersion[T] {
	if version < len(p.versions) {
		return &p.versions[version]
	}
	opVersion := p.appendVersions(version)
	opVersion.version = int16(version)
	return opVersion
}

func (p *Operation[T]) appendVersions(version int) *OperationVersion[T] {
	var previousOpVersion OperationVersion[T]
	if len(p.versions) > 0 {
		previousOpVersion = p.versions[len(p.versions)-1]
	} else {
		previousOpVersion = OperationVersion[T]{
			version:  INVALID_OPERATION_VERSION,
			executor: defaultExecutor[T],
		}
	}
	for i := len(p.versions); i <= version; i++ {
		p.versions = append(p.versions, previousOpVersion)
	}
	return &p.versions[version]
}

func defaultExecutor[T any](op *OperationVersion[T], pack T) (*buffer.ByteBuffer, ErrorResponse) {
	return nil, &ProtocolError{
		errorCode: NOT_IMPLEMENTED_ERROR,
		message:   "Operation not implemented",
	}
}

func (p *OperationVersion[T]) Executor(executor operationHandle[T]) *OperationVersion[T] {
	p.executor = executor
	return p
}

func (p *OperationVersion[T]) Execute(pack T) (*buffer.ByteBuffer, ErrorResponse) {
	return p.executor(p, pack)
}
