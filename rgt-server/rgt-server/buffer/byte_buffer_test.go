package buffer

import (
	"testing"
	"time"
)

// TestNew tests the default constructor
func TestNew(t *testing.T) {
	b := New()
	if b == nil {
		t.Fatal("New() returned nil")
	}
	if b.position != 0 {
		t.Errorf("Expected position 0, got %d", b.position)
	}
	if cap(b.buffer) != int(INITIAL_BUFFER_SIZE) {
		t.Errorf("Expected capacity %d, got %d", INITIAL_BUFFER_SIZE, cap(b.buffer))
	}
	if len(b.buffer) != 0 {
		t.Errorf("Expected length 0, got %d", len(b.buffer))
	}
}

// TestNewCapacity tests creating a buffer with specific capacity
func TestNewCapacity(t *testing.T) {
	tests := []uint32{10, 64, 128, 1024}
	for _, capacity := range tests {
		b := NewCapacity(capacity)
		if b == nil {
			t.Fatalf("NewCapacity(%d) returned nil", capacity)
		}
		if cap(b.buffer) != int(capacity) {
			t.Errorf("Expected capacity %d, got %d", capacity, cap(b.buffer))
		}
		if len(b.buffer) != 0 {
			t.Errorf("Expected length 0, got %d", len(b.buffer))
		}
	}
}

// TestWrap tests wrapping an existing byte slice
func TestWrap(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5}
	b := Wrap(data)
	if b == nil {
		t.Fatal("Wrap() returned nil")
	}
	if b.position != 0 {
		t.Errorf("Expected position 0, got %d", b.position)
	}
	if len(b.buffer) != len(data) {
		t.Errorf("Expected length %d, got %d", len(data), len(b.buffer))
	}
	for i, v := range data {
		if b.buffer[i] != v {
			t.Errorf("At index %d: expected %d, got %d", i, v, b.buffer[i])
		}
	}
}

// TestPutGetByte tests byte operations
func TestPutGetByte(t *testing.T) {
	b := New()
	testValues := []byte{0, 1, 127, 128, 255}

	for _, v := range testValues {
		b.PutByte(v)
	}

	b.Rewind()
	for i, expected := range testValues {
		got := b.GetByte()
		if got != expected {
			t.Errorf("At index %d: expected %d, got %d", i, expected, got)
		}
	}
}

// TestPutGetInt8 tests int8 operations
func TestPutGetInt8(t *testing.T) {
	b := New()
	testValues := []int8{-128, -1, 0, 1, 127}

	for _, v := range testValues {
		b.PutInt8(v)
	}

	b.Rewind()
	for i, expected := range testValues {
		got := b.GetInt8()
		if got != expected {
			t.Errorf("At index %d: expected %d, got %d", i, expected, got)
		}
	}
}

// TestPutGetUInt8 tests uint8 operations
func TestPutGetUInt8(t *testing.T) {
	b := New()
	testValues := []uint8{0, 1, 127, 128, 255}

	for _, v := range testValues {
		b.PutUInt8(v)
	}

	b.Rewind()
	for i, expected := range testValues {
		got := b.GetUInt8()
		if got != expected {
			t.Errorf("At index %d: expected %d, got %d", i, expected, got)
		}
	}
}

// TestPutGetInt16 tests int16 operations
func TestPutGetInt16(t *testing.T) {
	b := New()
	testValues := []int16{-32768, -1, 0, 1, 32767}

	for _, v := range testValues {
		b.PutInt16(v)
	}

	b.Rewind()
	for i, expected := range testValues {
		got := b.GetInt16()
		if got != expected {
			t.Errorf("At index %d: expected %d, got %d", i, expected, got)
		}
	}
}

// TestPutGetUInt16 tests uint16 operations
func TestPutGetUInt16(t *testing.T) {
	b := New()
	testValues := []uint16{0, 1, 32767, 32768, 65535}

	for _, v := range testValues {
		b.PutUInt16(v)
	}

	b.Rewind()
	for i, expected := range testValues {
		got := b.GetUInt16()
		if got != expected {
			t.Errorf("At index %d: expected %d, got %d", i, expected, got)
		}
	}
}

// TestPutGetInt32 tests int32 operations
func TestPutGetInt32(t *testing.T) {
	b := New()
	testValues := []int32{-2147483648, -1, 0, 1, 2147483647}

	for _, v := range testValues {
		b.PutInt32(v)
	}

	b.Rewind()
	for i, expected := range testValues {
		got := b.GetInt32()
		if got != expected {
			t.Errorf("At index %d: expected %d, got %d", i, expected, got)
		}
	}
}

// TestPutGetUInt32 tests uint32 operations
func TestPutGetUInt32(t *testing.T) {
	b := New()
	testValues := []uint32{0, 1, 2147483647, 2147483648, 4294967295}

	for _, v := range testValues {
		b.PutUInt32(v)
	}

	b.Rewind()
	for i, expected := range testValues {
		got := b.GetUInt32()
		if got != expected {
			t.Errorf("At index %d: expected %d, got %d", i, expected, got)
		}
	}
}

// TestPutGetInt64 tests int64 operations
func TestPutGetInt64(t *testing.T) {
	b := New()
	testValues := []int64{-9223372036854775808, -1, 0, 1, 9223372036854775807}

	for _, v := range testValues {
		b.PutInt64(v)
	}

	b.Rewind()
	for i, expected := range testValues {
		got := b.GetInt64()
		if got != expected {
			t.Errorf("At index %d: expected %d, got %d", i, expected, got)
		}
	}
}

// TestPutGetUInt64 tests uint64 operations
func TestPutGetUInt64(t *testing.T) {
	b := New()
	testValues := []uint64{0, 1, 9223372036854775807, 9223372036854775808, 18446744073709551615}

	for _, v := range testValues {
		b.PutUInt64(v)
	}

	b.Rewind()
	for i, expected := range testValues {
		got := b.GetUInt64()
		if got != expected {
			t.Errorf("At index %d: expected %d, got %d", i, expected, got)
		}
	}
}

// TestPutGetFloat32 tests float32 operations
func TestPutGetFloat32(t *testing.T) {
	b := New()
	testValues := []float32{-3.14, 0.0, 3.14, 1.234e10, -1.234e-10}

	for _, v := range testValues {
		b.PutFloat32(v)
	}

	b.Rewind()
	for i, expected := range testValues {
		got := b.GetFloat32()
		if got != expected {
			t.Errorf("At index %d: expected %f, got %f", i, expected, got)
		}
	}
}

// TestPutGetFloat64 tests float64 operations
func TestPutGetFloat64(t *testing.T) {
	b := New()
	testValues := []float64{-3.14159265359, 0.0, 3.14159265359, 1.234e100, -1.234e-100}

	for _, v := range testValues {
		b.PutFloat64(v)
	}

	b.Rewind()
	for i, expected := range testValues {
		got := b.GetFloat64()
		if got != expected {
			t.Errorf("At index %d: expected %f, got %f", i, expected, got)
		}
	}
}

// TestPutGetBool tests boolean operations
func TestPutGetBool(t *testing.T) {
	b := New()
	testValues := []bool{true, false, true, true, false}

	for _, v := range testValues {
		b.PutBool(v)
	}

	b.Rewind()
	for i, expected := range testValues {
		got := b.GetBool()
		if got != expected {
			t.Errorf("At index %d: expected %v, got %v", i, expected, got)
		}
	}
}

// TestPutGetString tests string operations
func TestPutGetString(t *testing.T) {
	b := New()
	testValues := []string{"", "Hello", "World", "Test String 123", "UTF-8: こんにちは"}

	for _, v := range testValues {
		b.PutString(v)
	}

	b.Rewind()
	for i, expected := range testValues {
		got := b.GetString()
		if got != expected {
			t.Errorf("At index %d: expected %q, got %q", i, expected, got)
		}
	}
}

// TestPutGetDate tests date operations
func TestPutGetDate(t *testing.T) {
	b := New()
	testDates := []time.Time{
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
		time.Date(2000, 12, 31, 0, 0, 0, 0, time.Local),
		time.Date(1999, 6, 15, 0, 0, 0, 0, time.Local),
	}

	for _, v := range testDates {
		b.PutDate(v)
	}

	b.Rewind()
	for i, expected := range testDates {
		got := b.GetDate()
		if got.Year() != expected.Year() || got.Month() != expected.Month() || got.Day() != expected.Day() {
			t.Errorf("At index %d: expected %v, got %v", i, expected, got)
		}
	}
}

// TestPutGetDateTime tests datetime operations
func TestPutGetDateTime(t *testing.T) {
	b := New()
	testDates := []time.Time{
		time.Date(2024, 1, 1, 12, 30, 45, 0, time.Local),
		time.Date(2000, 12, 31, 23, 59, 59, 0, time.Local),
		time.Date(1999, 6, 15, 0, 0, 0, 0, time.Local),
	}

	for _, v := range testDates {
		b.PutDateTime(v)
	}

	b.Rewind()
	for i, expected := range testDates {
		got := b.GetDateTime()
		if got.Year() != expected.Year() || got.Month() != expected.Month() ||
			got.Day() != expected.Day() || got.Hour() != expected.Hour() ||
			got.Minute() != expected.Minute() || got.Second() != expected.Second() {
			t.Errorf("At index %d: expected %v, got %v", i, expected, got)
		}
	}
}

// TestPutGetSlice tests slice operations
func TestPutGetSlice(t *testing.T) {
	b := New()
	testSlices := [][]byte{
		{},
		{1, 2, 3},
		{255, 0, 128},
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}

	for _, v := range testSlices {
		b.PutSlice(v)
	}

	b.Rewind()
	for i, expected := range testSlices {
		got := b.GetSlice()
		if len(got) != len(expected) {
			t.Errorf("At index %d: expected length %d, got %d", i, len(expected), len(got))
			continue
		}
		for j := range expected {
			if got[j] != expected[j] {
				t.Errorf("At index %d, byte %d: expected %d, got %d", i, j, expected[j], got[j])
			}
		}
	}
}

// TestPut tests Put operation
func TestPut(t *testing.T) {
	b := New()
	data := []byte{1, 2, 3, 4, 5}
	b.Put(data)

	if b.position != len(data) {
		t.Errorf("Expected position %d, got %d", len(data), b.position)
	}

	b.Rewind()
	result := b.BackSlice()[:len(data)]
	for i, v := range data {
		if result[i] != v {
			t.Errorf("At index %d: expected %d, got %d", i, v, result[i])
		}
	}
}

// TestPutBuffer tests PutBuffer operation
func TestPutBuffer(t *testing.T) {
	b1 := New()
	b1.PutInt32(12345)
	b1.PutString("test")
	b1.Flip() // Prepare b1 for reading by another buffer

	b2 := New()
	b2.PutBuffer(b1)

	b2.Rewind()
	got := b2.GetInt32()
	if got != 12345 {
		t.Errorf("Expected 12345, got %d", got)
	}
	gotStr := b2.GetString()
	if gotStr != "test" {
		t.Errorf("Expected 'test', got %q", gotStr)
	}
}

// TestRewind tests Rewind operation
func TestRewind(t *testing.T) {
	b := New()
	b.PutInt32(123)
	if b.position == 0 {
		t.Error("Position should be > 0 after Put")
	}
	b.Rewind()
	if b.position != 0 {
		t.Errorf("Expected position 0 after Rewind, got %d", b.position)
	}
}

// TestCompact tests Compact operation
func TestCompact(t *testing.T) {
	b := New()
	b.Put([]byte{1, 2, 3, 4, 5})
	b.position = 2

	err := b.Compact()
	if err != nil {
		t.Fatalf("Compact failed: %v", err)
	}

	if b.position != 0 {
		t.Errorf("Expected position 0 after Compact, got %d", b.position)
	}

	if len(b.buffer) != 3 {
		t.Errorf("Expected length 3 after Compact, got %d", len(b.buffer))
	}

	expected := []byte{3, 4, 5}
	for i, v := range expected {
		if b.buffer[i] != v {
			t.Errorf("At index %d: expected %d, got %d", i, v, b.buffer[i])
		}
	}
}

// TestCompactInvalidPosition tests Compact with invalid position
func TestCompactInvalidPosition(t *testing.T) {
	b := New()
	b.Put([]byte{1, 2, 3})
	b.position = 10

	err := b.Compact()
	if err == nil {
		t.Error("Expected error for invalid position, got nil")
	}
}

// TestClear tests Clear operation
func TestClear(t *testing.T) {
	b := New()
	originalCap := cap(b.buffer)
	b.Put([]byte{1, 2, 3, 4, 5})

	b.Clear()

	if b.position != 0 {
		t.Errorf("Expected position 0 after Clear, got %d", b.position)
	}
	if len(b.buffer) != 0 {
		t.Errorf("Expected length 0 after Clear, got %d", len(b.buffer))
	}
	if cap(b.buffer) != originalCap {
		t.Errorf("Capacity should be preserved after Clear, expected %d, got %d", originalCap, cap(b.buffer))
	}
}

// TestFlip tests Flip operation
func TestFlip(t *testing.T) {
	b := New()
	b.Put([]byte{1, 2, 3, 4, 5})

	b.Flip()

	if b.position != 0 {
		t.Errorf("Expected position 0 after Flip, got %d", b.position)
	}

	if b.Limit() != 5 {
		t.Errorf("Expected limit 5 after Flip, got %d", b.Limit())
	}
}

// TestHasRemaining tests HasRemaining operation
func TestHasRemaining(t *testing.T) {
	b := New()
	b.Put([]byte{1, 2, 3})

	b.position = 0
	if !b.HasRemaining() {
		t.Error("Expected HasRemaining to be true at position 0")
	}

	b.position = 3
	if b.HasRemaining() {
		t.Error("Expected HasRemaining to be false at end")
	}
}

// TestRemaining tests Remaining operation
func TestRemaining(t *testing.T) {
	b := New()
	b.Put([]byte{1, 2, 3, 4, 5})

	tests := []struct {
		position int
		expected int
	}{
		{0, 5},
		{2, 3},
		{5, 0},
		{10, 0}, // Beyond buffer
	}

	for _, tt := range tests {
		b.position = tt.position
		got := b.Remaining()
		if got != tt.expected {
			t.Errorf("At position %d: expected %d, got %d", tt.position, tt.expected, got)
		}
	}
}

// TestPosition tests Position getter
func TestPosition(t *testing.T) {
	b := New()
	if b.Position() != 0 {
		t.Errorf("Expected initial position 0, got %d", b.Position())
	}

	b.PutInt32(123)
	if b.Position() != 4 {
		t.Errorf("Expected position 4 after PutInt32, got %d", b.Position())
	}
}

// TestSetPosition tests SetPosition operation
func TestSetPosition(t *testing.T) {
	b := New()
	b.Put([]byte{1, 2, 3, 4, 5})

	// Valid position
	err := b.SetPosition(3)
	if err != nil {
		t.Errorf("SetPosition(3) failed: %v", err)
	}
	if b.position != 3 {
		t.Errorf("Expected position 3, got %d", b.position)
	}

	// Negative position
	err = b.SetPosition(-1)
	if err == nil {
		t.Error("Expected error for negative position")
	}

	// Position beyond buffer
	err = b.SetPosition(10)
	if err == nil {
		t.Error("Expected error for position beyond buffer")
	}
}

// TestLimit tests Limit getter
func TestLimit(t *testing.T) {
	b := New()
	b.Put([]byte{1, 2, 3})

	if b.Limit() != 3 {
		t.Errorf("Expected limit 3, got %d", b.Limit())
	}
}

// TestSetLimit tests SetLimit operation
func TestSetLimit(t *testing.T) {
	b := New()
	b.Put([]byte{1, 2, 3, 4, 5})

	// Valid limit
	err := b.SetLimit(3)
	if err != nil {
		t.Errorf("SetLimit(3) failed: %v", err)
	}
	if len(b.buffer) != 3 {
		t.Errorf("Expected length 3, got %d", len(b.buffer))
	}

	// Negative limit
	err = b.SetLimit(-1)
	if err == nil {
		t.Error("Expected error for negative limit")
	}
}

// TestSetByte tests SetByte operation
func TestSetByte(t *testing.T) {
	b := New()
	b.Put([]byte{1, 2, 3, 4, 5})

	b.SetByte(2, 99)
	if b.buffer[2] != 99 {
		t.Errorf("Expected buffer[2] = 99, got %d", b.buffer[2])
	}

	// Invalid index (should not panic)
	b.SetByte(-1, 100)
	b.SetByte(10, 100)
}

// TestGetBytes tests GetBytes operation
func TestGetBytes(t *testing.T) {
	b := New()
	b.Put([]byte{1, 2, 3, 4, 5})
	b.position = 2

	result := b.GetBytes()
	expected := []byte{3, 4, 5}

	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))
	}
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("At index %d: expected %d, got %d", i, v, result[i])
		}
	}

	if b.position != 5 {
		t.Errorf("Expected position 5 after GetBytes, got %d", b.position)
	}
}

// TestGetBytesFrom tests GetBytesFrom operation
func TestGetBytesFrom(t *testing.T) {
	b := New()
	b.Put([]byte{1, 2, 3, 4, 5})

	result := b.GetBytesFrom(2)
	expected := []byte{3, 4, 5}

	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))
	}
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("At index %d: expected %d, got %d", i, v, result[i])
		}
	}

	// Invalid position
	result = b.GetBytesFrom(-1)
	if len(result) != 0 {
		t.Errorf("Expected empty slice for negative position, got length %d", len(result))
	}

	result = b.GetBytesFrom(10)
	if len(result) != 0 {
		t.Errorf("Expected empty slice for out-of-bounds position, got length %d", len(result))
	}
}

// TestGetBufferFrom tests GetBufferFrom operation
func TestGetBufferFrom(t *testing.T) {
	b := New()
	b.Put([]byte{1, 2, 3, 4, 5})

	result := b.GetBufferFrom(2)
	if result == nil {
		t.Fatal("GetBufferFrom returned nil")
	}

	expected := []byte{3, 4, 5}
	if len(result.buffer) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result.buffer))
	}

	// Invalid position
	result = b.GetBufferFrom(-1)
	if result == nil {
		t.Error("Expected non-nil buffer for invalid position")
	}
	if len(result.buffer) != 0 {
		t.Errorf("Expected empty buffer for invalid position, got length %d", len(result.buffer))
	}
}

// TestSkip tests Skip operation
func TestSkip(t *testing.T) {
	b := New()
	b.Put([]byte{1, 2, 3, 4, 5})
	b.position = 2

	// Skip forward
	b.Skip(2)
	if b.position != 4 {
		t.Errorf("Expected position 4, got %d", b.position)
	}

	// Skip backward
	b.Skip(-2)
	if b.position != 2 {
		t.Errorf("Expected position 2, got %d", b.position)
	}

	// Skip beyond bounds (should clamp)
	b.Skip(100)
	if b.position != 5 {
		t.Errorf("Expected position clamped to 5, got %d", b.position)
	}

	b.Skip(-100)
	if b.position != 0 {
		t.Errorf("Expected position clamped to 0, got %d", b.position)
	}
}

// TestRemainingSlice tests RemainingSlice operation
func TestRemainingSlice(t *testing.T) {
	b := New()
	b.Put([]byte{1, 2, 3, 4, 5})
	b.position = 2

	result := b.RemainingSlice()
	expected := []byte{3, 4, 5}

	if len(result) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result))
	}
	for i, v := range expected {
		if result[i] != v {
			t.Errorf("At index %d: expected %d, got %d", i, v, result[i])
		}
	}
}

// TestRemainingBuffer tests RemainingBuffer operation
func TestRemainingBuffer(t *testing.T) {
	b := New()
	b.Put([]byte{1, 2, 3, 4, 5})
	b.position = 2

	result := b.RemainingBuffer()
	if result == nil {
		t.Fatal("RemainingBuffer returned nil")
	}

	expected := []byte{3, 4, 5}
	if len(result.buffer) != len(expected) {
		t.Errorf("Expected length %d, got %d", len(expected), len(result.buffer))
	}
}

// TestBackSlice tests BackSlice operation
func TestBackSlice(t *testing.T) {
	b := New()
	data := []byte{1, 2, 3, 4, 5}
	b.Put(data)

	result := b.BackSlice()
	if len(result) != len(data) {
		t.Errorf("Expected length %d, got %d", len(data), len(result))
	}
}

// TestGoString tests GoString operation
func TestGoString(t *testing.T) {
	b := New()
	b.Put([]byte{0xDE, 0xAD, 0xBE, 0xEF})

	str := b.GoString()
	if str == "" {
		t.Error("GoString returned empty string")
	}

	// Empty buffer
	b.Clear()
	str = b.GoString()
	if str != "" {
		t.Errorf("Expected empty string for empty buffer, got %q", str)
	}
}

// TestString tests String operation
func TestString(t *testing.T) {
	b := New()
	b.Put([]byte{0xDE, 0xAD, 0xBE, 0xEF})

	str := b.String()
	if str == "" {
		t.Error("String returned empty string")
	}
}

// TestAppendByteOverwrite tests that appendByte correctly overwrites
func TestAppendByteOverwrite(t *testing.T) {
	b := New()
	b.Put([]byte{1, 2, 3, 4, 5})

	// Move position back and overwrite
	b.position = 2
	b.Put([]byte{99, 99})

	expected := []byte{1, 2, 99, 99, 5}
	b.Rewind()
	for i, v := range expected {
		got := b.GetByte()
		if got != v {
			t.Errorf("At index %d: expected %d, got %d", i, v, got)
		}
	}
}
func TestAppendByteOverwriteGrowth(t *testing.T) {
	initial := []byte{1, 2, 3, 4, 5}
	b := New()
	b.Put(initial)

	if b.Limit() != len(initial) {
		t.Errorf("Expected limit %d, got %d", len(initial), b.Limit())
	}

	// Move position back and overwrite
	b.position = 2
	b.Put([]byte{99, 99, 99, 99, 99, 99})

	expected := []byte{1, 2, 99, 99, 99, 99, 99, 99}

	if b.Limit() != len(expected) {
		t.Errorf("Expected limit %d, got %d", len(expected), b.Limit())
	}

	b.Rewind()
	for i, v := range expected {
		got := b.GetByte()
		if got != v {
			t.Errorf("At index %d: expected %d, got %d", i, v, got)
		}
	}
}

// TestBufferGrowth tests that buffer grows correctly
func TestBufferGrowth(t *testing.T) {
	b := NewCapacity(4)

	// Add more data than initial capacity
	data := make([]byte, 100)
	for i := range data {
		data[i] = byte(i)
	}
	b.Put(data)

	if len(b.buffer) != 100 {
		t.Errorf("Expected buffer length 100, got %d", len(b.buffer))
	}

	b.Rewind()
	for i := 0; i < 100; i++ {
		got := b.GetByte()
		if got != byte(i) {
			t.Errorf("At index %d: expected %d, got %d", i, byte(i), got)
		}
	}
}

// TestGetWithInsufficientData tests Get operations with insufficient data
func TestGetWithInsufficientData(t *testing.T) {
	b := New()
	b.Put([]byte{1, 2}) // Only 2 bytes
	b.Rewind()          // Position is at 0 for reading

	// Try to read int32 (needs 4 bytes)
	result := b.GetInt32()
	if result != 0 {
		t.Errorf("Expected 0 for insufficient data, got %d", result)
	}

	// Position should not have changed when there's insufficient data
	if b.position != 0 {
		t.Errorf("Expected position unchanged at 0, got %d", b.position)
	}
}

// TestGetDateTimeWithInsufficientData tests GetDateTime with insufficient data
func TestGetDateTimeWithInsufficientData(t *testing.T) {
	b := New()
	b.Put([]byte{1, 2, 3}) // Only 3 bytes, needs 7

	result := b.GetDateTime()
	if !result.IsZero() {
		t.Errorf("Expected zero time for insufficient data, got %v", result)
	}
}

// TestGetStringWithInsufficientData tests GetString with insufficient data
func TestGetStringWithInsufficientData(t *testing.T) {
	b := New()
	b.PutInt32(100) // Claims string is 100 bytes long
	// But no actual string data follows

	b.Rewind()
	result := b.GetString()
	if result != "" {
		t.Errorf("Expected empty string for insufficient data, got %q", result)
	}
}
