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

// In/Out message buffer base.
type BufferBase struct {
	b []byte
}

// Convert to byte slice.
func (m *BufferBase) Get() []byte {
	return []byte(m.b)
}

// Get message length.
func (m *BufferBase) Len() int {
	return len(m.b)
}

// Get buffer capacity.
func (m *BufferBase) Cap() int {
	return cap(m.b)
}

// Copy data.
func (m *BufferBase) Copy() []byte {
	b := make([]byte, len(m.b))
	copy(b, m.b)
	return b
}

/*
Incoming
*/

// The base type of incoming message buffer.
type InBufferBase struct {
	BufferBase
	i int
}

func (m *InBufferBase) Prepare(size int) {
	if cap(m.b) >= size {
		m.b = m.b[0:size]
	} else {
		m.b = make([]byte, size)
	}
	m.i = 0
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
	BufferBase
}

// Prepare buffer for next write.
func (m *OutBufferBase) Prepare(head, size int) {
	if cap(m.b) >= size {
		m.b = m.b[0:head]
	} else {
		m.b = make([]byte, head, size)
	}
}

// Write a byte slice into buffer.
func (m *OutBufferBase) WriteBytes(d []byte) {
	m.b = append(m.b, d...)
}

// Write a string into buffer.
func (m *OutBufferBase) WriteString(s string) {
	m.b = append(m.b, s...)
}

// Write a rune into buffer.
func (m *OutBufferBase) WriteRune(r rune) {
	p := []byte{0, 0, 0, 0}
	n := utf8.EncodeRune(p, r)
	m.b = append(m.b, p[:n]...)
}

// Write a byte value into buffer.
func (m *OutBufferBase) WriteByte(v byte) {
	m.b = append(m.b, v)
}

// Write a int8 value into buffer.
func (m *OutBufferBase) WriteInt8(v int8) {
	m.b = append(m.b, byte(v))
}

// Write a uint8 value into buffer.
func (m *OutBufferBase) WriteUint8(v uint8) {
	m.b = append(m.b, byte(v))
}

/*
big endian
*/

// Big endian format outgoing message buffer.
type OutBufferBE struct {
	OutBufferBase
}

// Write a byte slice into buffer.
func (m *OutBufferBE) WriteBytes(d []byte) {
	m.b = append(m.b, d...)
}

// Write a string into buffer.
func (m *OutBufferBE) WriteString(s string) {
	m.b = append(m.b, s...)
}

// Write a rune into buffer.
func (m *OutBufferBE) WriteRune(r rune) {
	p := []byte{0, 0, 0, 0}
	n := utf8.EncodeRune(p, r)
	m.b = append(m.b, p[:n]...)
}

// Write a byte value into buffer.
func (m *OutBufferBE) WriteByte(v byte) {
	m.b = append(m.b, v)
}

// Write a big endian int16 value into buffer.
func (m *OutBufferBE) WriteInt16(v int16) {
	m.WriteUint16(uint16(v))
}

// Write a big endian uint16 value into buffer.
func (m *OutBufferBE) WriteUint16(v uint16) {
	m.b = append(m.b, byte(v>>8), byte(v))
}

// Write a big endian int32 value into buffer.
func (m *OutBufferBE) WriteInt32(v int32) {
	m.WriteUint32(uint32(v))
}

// Write a big endian uint32 value into buffer.
func (m *OutBufferBE) WriteUint32(v uint32) {
	m.b = append(m.b, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

// Write a big endian int64 value into buffer.
func (m *OutBufferBE) WriteInt64(v int64) {
	m.WriteUint64(uint64(v))
}

// Write a big endian uint64 value into buffer.
func (m *OutBufferBE) WriteUint64(v uint64) {
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

// Write a little endian int16 value into buffer.
func (m *OutBufferLE) WriteInt16(v int16) {
	m.WriteUint16(uint16(v))
}

// Write a little endian uint16 value into buffer.
func (m *OutBufferLE) WriteUint16(v uint16) {
	m.b = append(m.b, byte(v), byte(v>>8))
}

// Write a little endian int32 value into buffer.
func (m *OutBufferLE) WriteInt32(v int32) {
	m.WriteUint32(uint32(v))
}

// Write a little endian uint32 value into buffer.
func (m *OutBufferLE) WriteUint32(v uint32) {
	m.b = append(m.b, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}

// Write a little endian int64 value into buffer.
func (m *OutBufferLE) WriteInt64(v int64) {
	m.WriteUint64(uint64(v))
}

// Write a little endian uint64 value into buffer.
func (m *OutBufferLE) WriteUint64(v uint64) {
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
