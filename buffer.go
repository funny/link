package link

import (
	"encoding/binary"
	"unicode/utf8"
)

// Big endian message buffer factory.
type BufferFactoryBE struct {
}

// Create a big endian incoming message buffer.
func (_ BufferFactoryBE) NewInBuffer() InBuffer {
	return new(InBufferBE)
}

// Create a big endian outgoing message buffer.
func (_ BufferFactoryBE) NewOutBuffer() OutBuffer {
	return new(OutBufferBE)
}

// Little endian message buffer factory.
type BufferFactoryLE struct {
}

// Create a little endian incoming message buffer.
func (_ BufferFactoryLE) NewInBuffer() InBuffer {
	return new(InBufferLE)
}

// Create a little endian outgoing message buffer.
func (_ BufferFactoryLE) NewOutBuffer() OutBuffer {
	return new(OutBufferLE)
}

/*
Incoming
*/

// The base type of incoming message buffer.
type InBufferBase struct {
	b []byte
	i int
}

// Get internal buffer data.
func (m *InBufferBase) Get() []byte {
	return m.b
}

// Set internal buffer data.
func (m *InBufferBase) Set(b []byte) {
	m.b = b
}

func (m *InBufferBase) Prepare(size int) {
	if cap(m.b) >= size {
		m.b = m.b[0:size]
	} else {
		m.b = make([]byte, size)
	}
	m.i = 0
}

// Copy data.
func (m *InBufferBase) Copy() []byte {
	b := make([]byte, len(m.b))
	copy(b, m.b)
	return b
}

// Slice some bytes from buffer.
func (m *InBufferBase) ReadSlice(n int) []byte {
	r := m.b[m.i:n]
	m.i += n
	return r
}

// Copy some bytes from buffer.
func (m *InBufferBase) ReadBytes(n int) []byte {
	r := make([]byte, n)
	copy(r, m.ReadSlice(n))
	return r
}

// Read a string from buffer.
func (m *InBufferBase) ReadString(n int) string {
	return string(m.ReadSlice(n))
}

// Read a rune from buffer.
func (m *InBufferBase) ReadRune() rune {
	r, size := utf8.DecodeRune(m.b[m.i:])
	m.i += size
	return r
}

// Read a byte value from buffer.
func (m *InBufferBase) ReadByte() byte {
	r := m.b[m.i]
	m.i += 1
	return r
}

// Read a int8 value from buffer.
func (m *InBufferBase) ReadInt8() int8 {
	r := int8(m.b[m.i])
	m.i += 1
	return r
}

// Read a uint8 value from buffer.
func (m *InBufferBase) ReadUint8() uint8 {
	r := uint8(m.b[m.i])
	m.i += 1
	return r
}

/*
big endian
*/

// Big endian incoming message.
type InBufferBE struct {
	InBufferBase
}

// Read a big endian int16 value from buffer.
func (m *InBufferBE) ReadInt16() int16 {
	return int16(m.ReadUint16())
}

// Read a big endian uint16 value from buffer.
func (m *InBufferBE) ReadUint16() uint16 {
	r := binary.BigEndian.Uint16(m.b[m.i:])
	m.i += 2
	return r
}

// Read a big endian int32 value from buffer.
func (m *InBufferBE) ReadInt32() int32 {
	return int32(m.ReadUint32())
}

// Read a big endian uint32 value from buffer.
func (m *InBufferBE) ReadUint32() uint32 {
	r := binary.BigEndian.Uint32(m.b[m.i:])
	m.i += 4
	return r
}

// Read a big endian int64 value from buffer.
func (m *InBufferBE) ReadInt64() int64 {
	return int64(m.ReadUint64())
}

// Read a big endian uint64 value from buffer.
func (m *InBufferBE) ReadUint64() uint64 {
	r := binary.BigEndian.Uint64(m.b[m.i:])
	m.i += 8
	return r
}

/*
little endian
*/

// Little endian incoming message.
type InBufferLE struct {
	InBufferBase
}

// Read a little endian int16 value from buffer.
func (m *InBufferLE) ReadInt16() int16 {
	return int16(m.ReadUint16())
}

// Read a little endian uint16 value from buffer.
func (m *InBufferLE) ReadUint16() uint16 {
	r := binary.LittleEndian.Uint16(m.b[m.i:])
	m.i += 2
	return r
}

// Read a little endian int32 value from buffer.
func (m *InBufferLE) ReadInt32() int32 {
	return int32(m.ReadUint32())
}

// Read a little endian uint32 value from buffer.
func (m *InBufferLE) ReadUint32() uint32 {
	r := binary.LittleEndian.Uint32(m.b[m.i:])
	m.i += 4
	return r
}

// Read a little endian int64 value from buffer.
func (m *InBufferLE) ReadInt64() int64 {
	return int64(m.ReadUint64())
}

// Read a little endian uint64 value from buffer.
func (m *InBufferLE) ReadUint64() uint64 {
	r := binary.LittleEndian.Uint64(m.b[m.i:])
	m.i += 8
	return r
}

/*
Outgoing
*/

// The base type of outgoing message buffer.
type OutBufferBase struct {
	b []byte
}

// Convert to byte slice.
func (m *OutBufferBase) Get() []byte {
	return []byte(m.b)
}

// Get internal buffer.
func (m *OutBufferBase) Set(b []byte) {
	m.b = b
}

// Get message length.
func (m *OutBufferBase) Len() int {
	return len(m.b)
}

// Prepare buffer for next write.
func (m *OutBufferBase) Prepare(head, size int) {
	if cap(m.b) >= size {
		m.b = m.b[0:head]
	} else {
		m.b = make([]byte, head, size)
	}
}

// Copy data.
func (m *OutBufferBase) Copy() []byte {
	b := make([]byte, len(m.b))
	copy(b, m.b)
	return b
}

// Append a byte slice into buffer.
func (m *OutBufferBase) AppendBytes(d []byte) {
	m.b = append(m.b, d...)
}

// Append a string into buffer.
func (m *OutBufferBase) AppendString(s string) {
	m.b = append(m.b, s...)
}

// Append a rune into buffer.
func (m *OutBufferBase) AppendRune(r rune) {
	p := []byte{0, 0, 0, 0}
	n := utf8.EncodeRune(p, r)
	m.b = append(m.b, p[:n]...)
}

// Append a byte value into buffer.
func (m *OutBufferBase) AppendByte(v byte) {
	m.b = append(m.b, v)
}

// Append a int8 value into buffer.
func (m *OutBufferBase) AppendInt8(v int8) {
	m.b = append(m.b, byte(v))
}

// Append a uint8 value into buffer.
func (m *OutBufferBase) AppendUint8(v uint8) {
	m.b = append(m.b, byte(v))
}

/*
big endian
*/

// Big endian format outgoing message buffer.
type OutBufferBE struct {
	OutBufferBase
}

// Convert to byte slice.
func (m *OutBufferBE) Bytes() []byte {
	return []byte(m.b)
}

// Copy data.
func (m *OutBufferBE) Copy() []byte {
	b := make([]byte, len(m.b))
	copy(b, m.b)
	return b
}

// Append a byte slice into buffer.
func (m *OutBufferBE) AppendBytes(d []byte) {
	m.b = append(m.b, d...)
}

// Append a string into buffer.
func (m *OutBufferBE) AppendString(s string) {
	m.b = append(m.b, s...)
}

// Append a rune into buffer.
func (m *OutBufferBE) AppendRune(r rune) {
	p := []byte{0, 0, 0, 0}
	n := utf8.EncodeRune(p, r)
	m.b = append(m.b, p[:n]...)
}

// Append a byte value into buffer.
func (m *OutBufferBE) AppendByte(v byte) {
	m.b = append(m.b, v)
}

// Append a big endian int16 value into buffer.
func (m *OutBufferBE) AppendInt16(v int16) {
	m.AppendUint16(uint16(v))
}

// Append a big endian uint16 value into buffer.
func (m *OutBufferBE) AppendUint16(v uint16) {
	m.b = append(m.b, byte(v>>8), byte(v))
}

// Append a big endian int32 value into buffer.
func (m *OutBufferBE) AppendInt32(v int32) {
	m.AppendUint32(uint32(v))
}

// Append a big endian uint32 value into buffer.
func (m *OutBufferBE) AppendUint32(v uint32) {
	m.b = append(m.b, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

// Append a big endian int64 value into buffer.
func (m *OutBufferBE) AppendInt64(v int64) {
	m.AppendUint64(uint64(v))
}

// Append a big endian uint64 value into buffer.
func (m *OutBufferBE) AppendUint64(v uint64) {
	m.b = append(m.b,
		byte(v>>56),
		byte(v>>48),
		byte(v>>40),
		byte(v>>32),
		byte(v>>24),
		byte(v>>16),
		byte(v>>8),
		byte(v),
	)
}

/*
little endian
*/

// Little endian format outgoing message buffer.
type OutBufferLE struct {
	OutBufferBase
}

// Append a little endian int16 value into buffer.
func (m *OutBufferLE) AppendInt16(v int16) {
	m.AppendUint16(uint16(v))
}

// Append a little endian uint16 value into buffer.
func (m *OutBufferLE) AppendUint16(v uint16) {
	m.b = append(m.b, byte(v), byte(v>>8))
}

// Append a little endian int32 value into buffer.
func (m *OutBufferLE) AppendInt32(v int32) {
	m.AppendUint32(uint32(v))
}

// Append a little endian uint32 value into buffer.
func (m *OutBufferLE) AppendUint32(v uint32) {
	m.b = append(m.b, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}

// Append a little endian int64 value into buffer.
func (m *OutBufferLE) AppendInt64(v int64) {
	m.AppendUint64(uint64(v))
}

// Append a little endian uint64 value into buffer.
func (m *OutBufferLE) AppendUint64(v uint64) {
	m.b = append(m.b,
		byte(v),
		byte(v>>8),
		byte(v>>16),
		byte(v>>24),
		byte(v>>32),
		byte(v>>40),
		byte(v>>48),
		byte(v>>56),
	)
}
