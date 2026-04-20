package stats

import "sync/atomic"

const cacheLineSize = 64

type ServerStats struct {
	bytesReceived   atomic.Uint64
	_pad0           [cacheLineSize - 8]byte
	bytesSent       atomic.Uint64
	_pad1           [cacheLineSize - 8]byte
	packetsReceived atomic.Uint64
	_pad2           [cacheLineSize - 8]byte
	packetsSent     atomic.Uint64
	_pad3           [cacheLineSize - 8]byte
}

func NewServerStats() *ServerStats {
	return &ServerStats{}
}

func (s *ServerStats) AddBytesReceived(value uint64) {
	s.bytesReceived.Add(value)
}

func (s *ServerStats) AddBytesSent(value uint64) {
	s.bytesSent.Add(value)
}

func (s *ServerStats) AddPacketsReceived(value uint64) {
	s.packetsReceived.Add(value)
}

func (s *ServerStats) AddPacketsSent(value uint64) {
	s.packetsSent.Add(value)
}

func (s *ServerStats) BytesReceived() uint64 {
	return s.bytesReceived.Load()
}

func (s *ServerStats) BytesSent() uint64 {
	return s.bytesSent.Load()
}

func (s *ServerStats) PacketsReceived() uint64 {
	return s.packetsReceived.Load()
}

func (s *ServerStats) PacketsSent() uint64 {
	return s.packetsSent.Load()
}

func (s *ServerStats) SetBytesReceived(value uint64) {
	s.bytesReceived.Store(value)
}

func (s *ServerStats) SetBytesSent(value uint64) {
	s.bytesSent.Store(value)
}

func (s *ServerStats) SetPacketsReceived(value uint64) {
	s.packetsReceived.Store(value)
}

func (s *ServerStats) SetPacketsSent(value uint64) {
	s.packetsSent.Store(value)
}
