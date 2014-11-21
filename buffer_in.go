package link

import (
	"encoding/binary"
	"unicode/utf8"
)

// Incoming message buffer.
type InBuffer interface {
	// Get internal buffer data.
	Get() []byte

	// Set internal buffer data.
	Set([]byte)

	// Prepare buffer for next read.
	Prepare(size int)

	// Copy data.
	Copy() []byte

	// Slice some bytes from buffer.
	ReadSlice(n int) []byte

	// Copy some bytes from buffer.
	ReadBytes(n int) []byte

	// Read a string from buffer.
	ReadString(n int) string

	// Read a rune from buffer.
	ReadRune() rune

	// Read a byte value from buffer.
	ReadByte() byte

	// Read a int8 value from buffer.
	ReadInt8() int8

	// Read a uint8 value from buffer.
	ReadUint8() uint8

	// Read a int16 value from buffer.
	ReadInt16() int16

	// Read a uint16 value from buffer.
	ReadUint16() uint16

	// Read a int32 value from buffer.
	ReadInt32() int32

	// Read a uint32 value from buffer.
	ReadUint32() uint32

	// Read a int64 value from buffer.
	ReadInt64() int64

	// Read a uint64 value from buffer.
	ReadUint64() uint64
}

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
