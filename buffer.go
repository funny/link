package link

import (
	"encoding/binary"
	"io"
	"math"
	"unicode/utf8"
)

// Big endian message buffer factory.
type SimpleBufferFactory struct {
}

// Create a big endian incoming message buffer.
func (_ SimpleBufferFactory) NewInBuffer() InBuffer {
	return new(SimpleInBuffer)
}

// Create a big endian outgoing message buffer.
func (_ SimpleBufferFactory) NewOutBuffer() OutBuffer {
	return new(SimpleOutBuffer)
}

// In/Out message buffer base.
type SimpleBufferBase struct {
	b []byte
}

// Get internal buffer.
func (m *SimpleBufferBase) Get() []byte {
	return m.b
}

// Get buffer length.
func (m *SimpleBufferBase) Len() int {
	return len(m.b)
}

// Get buffer capacity.
func (m *SimpleBufferBase) Cap() int {
	return cap(m.b)
}

// The base type of incoming message buffer.
type SimpleInBuffer struct {
	SimpleBufferBase
	i int
}

// Implement io.Reader interface.
func (m *SimpleInBuffer) Read(p []byte) (int, error) {
	if m.i == len(m.b) {
		return 0, io.EOF
	}
	n := copy(p, m.b[m.i:])
	m.i += n
	return n, nil
}

// Prepare buffer for next read.
func (m *SimpleInBuffer) Prepare(size int) {
	if cap(m.b) >= size {
		m.b = m.b[0:size]
	} else {
		m.b = make([]byte, size)
	}
	m.i = 0
}

// Slice some bytes from buffer.
func (m *SimpleInBuffer) Slice(n int) []byte {
	r := m.b[m.i : m.i+n]
	m.i += n
	return r
}

// Copy some bytes from buffer.
func (m *SimpleInBuffer) ReadBytes(n int) []byte {
	r := make([]byte, n)
	copy(r, m.Slice(n))
	return r
}

// Read a string from buffer.
func (m *SimpleInBuffer) ReadString(n int) string {
	return string(m.Slice(n))
}

// Read a rune from buffer.
func (m *SimpleInBuffer) ReadRune() rune {
	r, size := utf8.DecodeRune(m.b[m.i:])
	m.i += size
	return r
}

// Read a uint8 value from buffer.
func (m *SimpleInBuffer) ReadUint8() uint8 {
	r := uint8(m.b[m.i])
	m.i += 1
	return r
}

// Read a uint16 value from buffer using little endian byte order.
func (m *SimpleInBuffer) ReadUint16LE() uint16 {
	r := binary.LittleEndian.Uint16(m.b[m.i:])
	m.i += 2
	return r
}

// Read a uint16 value from buffer using big endian byte order.
func (m *SimpleInBuffer) ReadUint16BE() uint16 {
	r := binary.BigEndian.Uint16(m.b[m.i:])
	m.i += 2
	return r
}

// Read a uint32 value from buffer using little endian byte order.
func (m *SimpleInBuffer) ReadUint32LE() uint32 {
	r := binary.LittleEndian.Uint32(m.b[m.i:])
	m.i += 4
	return r
}

// Read a uint32 value from buffer using big endian byte order.
func (m *SimpleInBuffer) ReadUint32BE() uint32 {
	r := binary.BigEndian.Uint32(m.b[m.i:])
	m.i += 4
	return r
}

// Read a uint64 value from buffer using little endian byte order.
func (m *SimpleInBuffer) ReadUint64LE() uint64 {
	r := binary.LittleEndian.Uint64(m.b[m.i:])
	m.i += 8
	return r
}

// Read a uint64 value from buffer using big endian byte order.
func (m *SimpleInBuffer) ReadUint64BE() uint64 {
	r := binary.BigEndian.Uint64(m.b[m.i:])
	m.i += 8
	return r
}

// Read a float32 value from buffer using little endian byte order.
func (m *SimpleInBuffer) ReadFloat32LE() float32 {
	return math.Float32frombits(m.ReadUint32LE())
}

// Read a float32 value from buffer using big endian byte order.
func (m *SimpleInBuffer) ReadFloat32BE() float32 {
	return math.Float32frombits(m.ReadUint32BE())
}

// Read a float64 value from buffer using little endian byte order.
func (m *SimpleInBuffer) ReadFloat64LE() float64 {
	return math.Float64frombits(m.ReadUint64LE())
}

// Read a float64 value from buffer using big endian byte order.
func (m *SimpleInBuffer) ReadFloat64BE() float64 {
	return math.Float64frombits(m.ReadUint64BE())
}

// The base type of outgoing message buffer.
type SimpleOutBuffer struct {
	SimpleBufferBase
}

// Implement io.Writer interface.
func (m *SimpleOutBuffer) Write(d []byte) (int, error) {
	m.b = append(m.b, d...)
	return len(d), nil
}

// Prepare buffer for next write.
func (m *SimpleOutBuffer) Prepare(size int) {
	if cap(m.b) >= size {
		m.b = m.b[0:0]
	} else {
		m.b = make([]byte, 0, size)
	}
}

// Write a byte slice into buffer.
func (m *SimpleOutBuffer) WriteBytes(d []byte) {
	m.b = append(m.b, d...)
}

// Write a string into buffer.
func (m *SimpleOutBuffer) WriteString(s string) {
	m.b = append(m.b, s...)
}

// Write a rune into buffer.
func (m *SimpleOutBuffer) WriteRune(r rune) {
	p := []byte{0, 0, 0, 0}
	n := utf8.EncodeRune(p, r)
	m.b = append(m.b, p[:n]...)
}

// Write a uint8 value into buffer.
func (m *SimpleOutBuffer) WriteUint8(v uint8) {
	m.b = append(m.b, byte(v))
}

// Write a uint16 value into buffer using little endian byte order.
func (m *SimpleOutBuffer) WriteUint16LE(v uint16) {
	m.b = append(m.b, byte(v), byte(v>>8))
}

// Write a uint16 value into buffer using big endian byte order.
func (m *SimpleOutBuffer) WriteUint16BE(v uint16) {
	m.b = append(m.b, byte(v>>8), byte(v))
}

// Write a uint32 value into buffer using little endian byte order.
func (m *SimpleOutBuffer) WriteUint32LE(v uint32) {
	m.b = append(m.b, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}

// Write a uint32 value into buffer using big endian byte order.
func (m *SimpleOutBuffer) WriteUint32BE(v uint32) {
	m.b = append(m.b, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

// Write a uint64 value into buffer using little endian byte order.
func (m *SimpleOutBuffer) WriteUint64LE(v uint64) {
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

// Write a uint64 value into buffer using big endian byte order.
func (m *SimpleOutBuffer) WriteUint64BE(v uint64) {
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

// Write a float32 value into buffer using little endian byte order.
func (m *SimpleOutBuffer) WriteFloat32LE(v float32) {
	m.WriteUint32LE(math.Float32bits(v))
}

// Write a float32 value into buffer using big endian byte order.
func (m *SimpleOutBuffer) WriteFloat32BE(v float32) {
	m.WriteUint32BE(math.Float32bits(v))
}

// Write a float64 value into buffer using little endian byte order.
func (m *SimpleOutBuffer) WriteFloat64LE(v float64) {
	m.WriteUint64LE(math.Float64bits(v))
}

// Write a float64 value into buffer using big endian byte order.
func (m *SimpleOutBuffer) WriteFloat64BE(v float64) {
	m.WriteUint64BE(math.Float64bits(v))
}
