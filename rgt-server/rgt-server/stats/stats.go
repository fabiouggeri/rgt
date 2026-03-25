package stats

import "sync/atomic"

type Stats struct {
	bytesReceived   atomic.Uint64
	bytesSent       atomic.Uint64
	packetsReceived atomic.Uint64
	packetsSent     atomic.Uint64
}

func New() *Stats {
	return &Stats{}
}

func (s *Stats) AddBytesReceived(value uint64) {
	s.bytesReceived.Add(value)
}

func (s *Stats) AddBytesSent(value uint64) {
	s.bytesSent.Add(value)
}

func (s *Stats) AddPacketsReceived(value uint64) {
	s.packetsReceived.Add(value)
}

func (s *Stats) AddPacketsSent(value uint64) {
	s.packetsSent.Add(value)
}

func (s *Stats) BytesReceived() uint64 {
	return s.bytesReceived.Load()
}

func (s *Stats) BytesSent() uint64 {
	return s.bytesSent.Load()
}

func (s *Stats) PacketsReceived() uint64 {
	return s.packetsReceived.Load()
}

func (s *Stats) PacketsSent() uint64 {
	return s.packetsSent.Load()
}

func (s *Stats) SetBytesReceived(value uint64) {
	s.bytesReceived.Store(value)
}

func (s *Stats) SetBytesSent(value uint64) {
	s.bytesSent.Store(value)
}

func (s *Stats) SetPacketsReceived(value uint64) {
	s.packetsReceived.Store(value)
}

func (s *Stats) SetPacketsSent(value uint64) {
	s.packetsSent.Store(value)
}
