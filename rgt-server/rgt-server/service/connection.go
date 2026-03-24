package service

import (
	"rgt-server/buffer"
	"rgt-server/protocol"
	"time"
)

type ConnectionStatus uint8
type ConnectionType uint8

const (
	CREATED   ConnectionStatus = 0
	WORKING   ConnectionStatus = 1
	KILLING   ConnectionStatus = 2
	FINISHING ConnectionStatus = 3
	FINISHED  ConnectionStatus = 4
)

const (
	UNKNOWN     ConnectionType = 0
	TERMINAL    ConnectionType = 1
	APPLICATION ConnectionType = 2
	LAUNCHER    ConnectionType = 3
)

type ConnectionHandler interface {
	Id() uint64
	Handle()
	GetRemoteAddr() string
	Send(buf *buffer.ByteBuffer) error
	Close() error
	Connected() bool
}

type TerminalConnectionHandler interface {
	ConnectionHandler
	SendLogout(message string)
	GetEndpoint() TerminalConnectionHandler
	SetEndpoint(endpoint TerminalConnectionHandler)
	GetLastDataReadTime() time.Time
	GetLastAppOperationTime() time.Time
	RegisterAdminClient(conn ConnectionHandler) AdminClient
	UnregisterAdminClient(conn ConnectionHandler)
}

type AdminClient interface {
	SendRequest(requestCode protocol.OperationCode, data []byte) (*buffer.ByteBuffer, protocol.ErrorResponse)
}

func (c ConnectionType) GoString() string {
	switch c {
	case APPLICATION:
		return "APP"
	case TERMINAL:
		return "TE"
	case LAUNCHER:
		return "LAUNCHER"
	default:
		return "UNKNOWN"
	}
}

func (c ConnectionType) String() string {
	return c.GoString()
}
