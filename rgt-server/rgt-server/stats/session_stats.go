package stats

import "sync/atomic"

type SessionStats struct {
	bytesReceived   atomic.Uint64
	bytesSent       atomic.Uint64
	packetsReceived atomic.Uint64
	packetsSent     atomic.Uint64
}

func NewSessionStats() *SessionStats {
	return &SessionStats{}
}

func (s *SessionStats) AddBytesReceived(value uint64) {
	s.bytesReceived.Add(value)
}

func (s *SessionStats) AddBytesSent(value uint64) {
	s.bytesSent.Add(value)
}

func (s *SessionStats) AddPacketsReceived(value uint64) {
	s.packetsReceived.Add(value)
}

func (s *SessionStats) AddPacketsSent(value uint64) {
	s.packetsSent.Add(value)
}

func (s *SessionStats) BytesReceived() uint64 {
	return s.bytesReceived.Load()
}

func (s *SessionStats) BytesSent() uint64 {
	return s.bytesSent.Load()
}

func (s *SessionStats) PacketsReceived() uint64 {
	return s.packetsReceived.Load()
}

func (s *SessionStats) PacketsSent() uint64 {
	return s.packetsSent.Load()
}

func (s *SessionStats) SetBytesReceived(value uint64) {
	s.bytesReceived.Store(value)
}

func (s *SessionStats) SetBytesSent(value uint64) {
	s.bytesSent.Store(value)
}

func (s *SessionStats) SetPacketsReceived(value uint64) {
	s.packetsReceived.Store(value)
}

func (s *SessionStats) SetPacketsSent(value uint64) {
	s.packetsSent.Store(value)
}
