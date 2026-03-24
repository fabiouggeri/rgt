package buffer

import (
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"time"
)

const INITIAL_BUFFER_SIZE uint32 = 64

type ByteBuffer struct {
	buffer   []byte
	position int
}

func New() *ByteBuffer {
	return NewCapacity(INITIAL_BUFFER_SIZE)
}

func NewCapacity(capacity uint32) *ByteBuffer {
	return &ByteBuffer{
		position: 0,
		buffer:   make([]byte, 0, capacity),
	}
}

func NewLen(newLen uint32) *ByteBuffer {
	return &ByteBuffer{
		position: 0,
		buffer:   make([]byte, newLen),
	}
}

func Wrap(buf []byte) *ByteBuffer {
	return &ByteBuffer{position: 0,
		buffer: buf}
}

func (b *ByteBuffer) BackSlice() []byte {
	return b.buffer
}

func (b *ByteBuffer) RemainingSlice() []byte {
	return b.buffer[b.position:]
}

func (b *ByteBuffer) RemainingBuffer() *ByteBuffer {
	return Wrap(b.buffer[b.position:])
}

func (b *ByteBuffer) hasDataEnough(required int) bool {
	return required >= 0 && required <= len(b.buffer)-b.position
}

func (b *ByteBuffer) GetString() string {
	var strContent string
	strLen := b.GetInt32()
	if strLen > 0 {
		if !b.hasDataEnough(int(strLen)) {
			return ""
		}
		strContent = string(b.buffer[b.position : b.position+int(strLen)])
		b.position += int(strLen)
	} else {
		strContent = ""
	}
	return strContent
}

func (b *ByteBuffer) GetByte() byte {
	if !b.hasDataEnough(1) {
		return 0
	}
	result := b.buffer[b.position]
	b.position++
	return result
}

func (b *ByteBuffer) GetInt8() int8 {
	if !b.hasDataEnough(1) {
		return 0
	}
	result := int8(b.buffer[b.position])
	b.position++
	return result
}

func (b *ByteBuffer) GetUInt8() uint8 {
	if !b.hasDataEnough(1) {
		return 0
	}
	result := uint8(b.buffer[b.position])
	b.position++
	return result
}

func (b *ByteBuffer) GetInt16() int16 {
	var result int16
	if !b.hasDataEnough(2) {
		return 0
	}
	result = int16(b.buffer[b.position]) | int16(b.buffer[b.position+1])<<8
	b.position += 2
	return result
}

func (b *ByteBuffer) GetUInt16() uint16 {
	var result uint16
	if !b.hasDataEnough(2) {
		return 0
	}
	result = uint16(b.buffer[b.position]) | uint16(b.buffer[b.position+1])<<8
	b.position += 2
	return result
}

func (b *ByteBuffer) GetInt32() int32 {
	var result int32
	if !b.hasDataEnough(4) {
		return 0
	}
	result = int32(b.buffer[b.position]) | int32(b.buffer[b.position+1])<<8 | int32(b.buffer[b.position+2])<<16 | int32(b.buffer[b.position+3])<<24
	b.position += 4
	return result
}

func (b *ByteBuffer) GetUInt32() uint32 {
	var result uint32
	if !b.hasDataEnough(4) {
		return 0
	}
	result = uint32(b.buffer[b.position]) | uint32(b.buffer[b.position+1])<<8 | uint32(b.buffer[b.position+2])<<16 | uint32(b.buffer[b.position+3])<<24
	b.position += 4
	return result
}

func (b *ByteBuffer) GetInt64() int64 {
	var result int64
	if !b.hasDataEnough(8) {
		return 0
	}
	result = int64(b.buffer[b.position]) | int64(b.buffer[b.position+1])<<8 | int64(b.buffer[b.position+2])<<16 | int64(b.buffer[b.position+3])<<24 |
		int64(b.buffer[b.position+4])<<32 | int64(b.buffer[b.position+5])<<40 | int64(b.buffer[b.position+6])<<48 | int64(b.buffer[b.position+7])<<56
	b.position += 8
	return result
}

func (b *ByteBuffer) GetUInt64() uint64 {
	var result uint64
	if !b.hasDataEnough(8) {
		return 0
	}
	result = uint64(b.buffer[b.position]) | uint64(b.buffer[b.position+1])<<8 | uint64(b.buffer[b.position+2])<<16 | uint64(b.buffer[b.position+3])<<24 |
		uint64(b.buffer[b.position+4])<<32 | uint64(b.buffer[b.position+5])<<40 | uint64(b.buffer[b.position+6])<<48 | uint64(b.buffer[b.position+7])<<56
	b.position += 8
	return result
}

func (b *ByteBuffer) GetFloat32() float32 {
	if !b.hasDataEnough(4) {
		return 0.0
	}
	val := uint32(b.buffer[b.position]) | uint32(b.buffer[b.position+1])<<8 | uint32(b.buffer[b.position+2])<<16 | uint32(b.buffer[b.position+3])<<24
	result := math.Float32frombits(val)
	b.position += 4
	return result
}

func (b *ByteBuffer) GetFloat64() float64 {
	if !b.hasDataEnough(8) {
		return 0.0
	}
	value := uint64(b.buffer[b.position]) | uint64(b.buffer[b.position+1])<<8 | uint64(b.buffer[b.position+2])<<16 | uint64(b.buffer[b.position+3])<<24 |
		uint64(b.buffer[b.position+4])<<32 | uint64(b.buffer[b.position+5])<<40 | uint64(b.buffer[b.position+6])<<48 | uint64(b.buffer[b.position+7])<<56
	result := math.Float64frombits(value)
	b.position += 8
	return result
}

func (b *ByteBuffer) GetBool() bool {
	if !b.hasDataEnough(1) {
		return false
	}
	result := b.buffer[b.position] != 0
	b.position++
	return result
}

func (b *ByteBuffer) GetDate() time.Time {
	var date time.Time
	if !b.hasDataEnough(4) {
		return time.Time{}
	}
	year := int16(b.buffer[b.position]) | int16(b.buffer[b.position+1])<<8
	month := int8(b.buffer[b.position+2])
	day := int8(b.buffer[b.position+3])
	date = time.Date(int(year), time.Month(month), int(day), 0, 0, 0, 0, time.Local)
	b.position += 4
	return date
}

func (b *ByteBuffer) GetDateTime() time.Time {
	var date time.Time
	if !b.hasDataEnough(7) {
		return time.Time{}
	}
	year := int16(b.buffer[b.position]) | int16(b.buffer[b.position+1])<<8
	month := int8(b.buffer[b.position+2])
	day := int8(b.buffer[b.position+3])
	hour := int8(b.buffer[b.position+4])
	min := int8(b.buffer[b.position+5])
	sec := int8(b.buffer[b.position+6])
	date = time.Date(int(year), time.Month(month), int(day), int(hour), int(min), int(sec), 0, time.Local)
	b.position += 7
	return date
}

func (b *ByteBuffer) Rewind() {
	b.position = 0
}

func (b *ByteBuffer) Compact() error {
	if b.position < 0 || b.position > len(b.buffer) {
		return fmt.Errorf("position is out of limit. fun: Compact. pos: %d limit: %d", b.position, len(b.buffer))
	}
	remaining := len(b.buffer) - b.position
	copy(b.buffer, b.buffer[b.position:])
	b.buffer = b.buffer[:remaining]
	b.position = 0
	return nil
}

func (b *ByteBuffer) Clear() {
	b.position = 0
	b.buffer = b.buffer[:0]
}

func (b *ByteBuffer) HasRemaining() bool {
	return len(b.buffer) > b.position
}

func (b *ByteBuffer) Remaining() int {
	if len(b.buffer) > b.position {
		return len(b.buffer) - b.position
	} else {
		return 0
	}
}

func (b *ByteBuffer) Flip() {
	if b.position < len(b.buffer) {
		b.buffer = b.buffer[:b.position]
	}
	b.position = 0
}

func (b *ByteBuffer) Position() int {
	return b.position
}

func (b *ByteBuffer) SetPosition(newPos int) error {
	if newPos < 0 {
		return fmt.Errorf("new position (%d) can't be negative", newPos)
	} else if newPos > len(b.buffer) {
		return fmt.Errorf("new position (%d) exceed limit (%d)", newPos, len(b.buffer))
	}
	b.position = newPos
	return nil
}

func (b *ByteBuffer) Limit() int {
	return len(b.buffer)
}

func (b *ByteBuffer) SetLimit(newLimit int) error {
	if newLimit < 0 {
		return fmt.Errorf("new limit (%d) can't be negative", newLimit)
	} else if newLimit > len(b.buffer) {
		return fmt.Errorf("new limit (%d) exceed capacity (%d)", newLimit, len(b.buffer))
	}
	b.buffer = b.buffer[:newLimit]
	return nil
}

func (b *ByteBuffer) appendByte(data ...byte) {
	dataLen := len(data)
	if dataLen == 0 {
		return
	}
	newPosition := b.position + dataLen
	if newPosition <= len(b.buffer) {
		copy(b.buffer[b.position:], data)
	} else {
		b.buffer = append(b.buffer[:b.position], data...)
	}
	b.position = newPosition
}

func (b *ByteBuffer) SetByte(index int, value byte) {
	if index >= 0 && index < len(b.buffer) {
		b.buffer[index] = value
	}
}

func (b *ByteBuffer) PutByte(value byte) {
	b.appendByte(value)
}

func (b *ByteBuffer) PutString(value string) {
	if len(value) > 0 {
		array := []byte(value)
		b.PutInt32(int32(len(array)))
		b.appendByte(array...)
	} else {
		b.PutInt32(0)
	}
}

func (b *ByteBuffer) PutInt8(value int8) {
	b.PutByte(byte(value))
}

func (b *ByteBuffer) PutUInt8(value uint8) {
	b.PutByte(byte(value))
}

func (b *ByteBuffer) PutInt16(value int16) {
	b.appendByte(byte(value), byte(value>>8))
}

func (b *ByteBuffer) PutUInt16(value uint16) {
	b.appendByte(byte(value), byte(value>>8))
}

func (b *ByteBuffer) PutInt32(value int32) {
	b.appendByte(byte(value), byte(value>>8), byte(value>>16), byte(value>>24))
}

func (b *ByteBuffer) PutUInt32(value uint32) {
	b.appendByte(byte(value), byte(value>>8), byte(value>>16), byte(value>>24))
}

func (b *ByteBuffer) PutInt64(value int64) {
	b.appendByte(byte(value), byte(value>>8), byte(value>>16), byte(value>>24), byte(value>>32), byte(value>>40), byte(value>>48), byte(value>>56))
}

func (b *ByteBuffer) PutUInt64(value uint64) {
	b.appendByte(byte(value), byte(value>>8), byte(value>>16), byte(value>>24), byte(value>>32), byte(value>>40), byte(value>>48), byte(value>>56))
}

func (b *ByteBuffer) PutBool(value bool) {
	if value {
		b.appendByte(1)
	} else {
		b.appendByte(0)
	}
}

func (b *ByteBuffer) PutSlice(value []byte) {
	b.PutInt32(int32(len(value)))
	b.appendByte(value...)
}

func (b *ByteBuffer) Put(value []byte) {
	b.appendByte(value...)
}

func (b *ByteBuffer) PutFloat32(value float32) {
	ui32 := math.Float32bits(value)
	b.appendByte(byte(ui32), byte(ui32>>8), byte(ui32>>16), byte(ui32>>24))
}

func (b *ByteBuffer) PutFloat64(value float64) {
	ui64 := math.Float64bits(value)
	b.appendByte(byte(ui64), byte(ui64>>8), byte(ui64>>16), byte(ui64>>24), byte(ui64>>32), byte(ui64>>40), byte(ui64>>48), byte(ui64>>56))
}

func (b *ByteBuffer) PutDate(value time.Time) {
	b.PutInt16(int16(value.Year()))
	b.PutInt8(int8(value.Month()))
	b.PutInt8(int8(value.Day()))
}

func (b *ByteBuffer) PutDateTime(value time.Time) {
	b.PutInt16(int16(value.Year()))
	b.PutInt8(int8(value.Month()))
	b.PutInt8(int8(value.Day()))
	b.PutInt8(int8(value.Hour()))
	b.PutInt8(int8(value.Minute()))
	b.PutInt8(int8(value.Second()))
}

func (b *ByteBuffer) PutBuffer(other *ByteBuffer) {
	b.appendByte(other.buffer[other.position:]...)
}

func (b *ByteBuffer) GetSlice() []byte {
	sliceLen := b.GetInt32()
	if sliceLen > 0 && b.hasDataEnough(int(sliceLen)) {
		data := b.buffer[b.position : b.position+int(sliceLen)]
		b.position += int(sliceLen)
		return data
	}
	return []byte{}
}

func (b *ByteBuffer) GetBytes() []byte {
	if b.position < len(b.buffer) {
		res := b.buffer[b.position:]
		b.position = len(b.buffer)
		return res
	} else {
		return []byte{}
	}
}

func (b *ByteBuffer) GetBytesFrom(position int) []byte {
	if position >= 0 && position < len(b.buffer) {
		res := b.buffer[position:]
		b.position = len(b.buffer)
		return res
	} else {
		return []byte{}
	}
}

func (b *ByteBuffer) GetBufferFrom(position int) *ByteBuffer {
	if position >= 0 && position < len(b.buffer) {
		return Wrap(b.buffer[position:])
	} else {
		return &ByteBuffer{buffer: []byte{}, position: 0}
	}
}

func (b *ByteBuffer) Skip(toSkip int) {
	b.position += toSkip
	if b.position < 0 {
		b.position = 0
	} else if b.position > len(b.buffer) {
		b.position = len(b.buffer)
	}
}

func (b *ByteBuffer) GoString() string {
	if len(b.buffer) > 0 {
		var s strings.Builder
		h := hex.NewEncoder(&s)
		h.Write(b.buffer)
		return s.String()
	} else {
		return ""
	}
}

func (b *ByteBuffer) String() string {
	return b.GoString()
}
