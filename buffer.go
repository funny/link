package link

import (
	"encoding/binary"
	"math"
	"unicode/utf8"
)

// The base type of incoming message buffer.
type BufferReader struct {
	b []byte
	i int
}

func NewBufferReader(buffer []byte) *BufferReader {
	return &BufferReader{buffer, 0}
}

// Slice some bytes from buffer.
func (m *BufferReader) Slice(n int) []byte {
	r := m.b[m.i : m.i+n]
	m.i += n
	return r
}

// Copy some bytes from buffer.
func (m *BufferReader) ReadBytes(n int) []byte {
	r := make([]byte, n)
	copy(r, m.Slice(n))
	return r
}

// Read a string from buffer.
func (m *BufferReader) ReadString(n int) string {
	return string(m.Slice(n))
}

// Read a rune from buffer.
func (m *BufferReader) ReadRune() rune {
	r, size := utf8.DecodeRune(m.b[m.i:])
	m.i += size
	return r
}

// Read a uint8 value from buffer.
func (m *BufferReader) ReadUint8() uint8 {
	r := uint8(m.b[m.i])
	m.i += 1
	return r
}

// Read a uint16 value from buffer using little endian byte order.
func (m *BufferReader) ReadUint16LE() uint16 {
	r := binary.LittleEndian.Uint16(m.b[m.i:])
	m.i += 2
	return r
}

// Read a uint16 value from buffer using big endian byte order.
func (m *BufferReader) ReadUint16BE() uint16 {
	r := binary.BigEndian.Uint16(m.b[m.i:])
	m.i += 2
	return r
}

// Read a uint32 value from buffer using little endian byte order.
func (m *BufferReader) ReadUint32LE() uint32 {
	r := binary.LittleEndian.Uint32(m.b[m.i:])
	m.i += 4
	return r
}

// Read a uint32 value from buffer using big endian byte order.
func (m *BufferReader) ReadUint32BE() uint32 {
	r := binary.BigEndian.Uint32(m.b[m.i:])
	m.i += 4
	return r
}

// Read a uint64 value from buffer using little endian byte order.
func (m *BufferReader) ReadUint64LE() uint64 {
	r := binary.LittleEndian.Uint64(m.b[m.i:])
	m.i += 8
	return r
}

// Read a uint64 value from buffer using big endian byte order.
func (m *BufferReader) ReadUint64BE() uint64 {
	r := binary.BigEndian.Uint64(m.b[m.i:])
	m.i += 8
	return r
}

// Read a float32 value from buffer using little endian byte order.
func (m *BufferReader) ReadFloat32LE() float32 {
	return math.Float32frombits(m.ReadUint32LE())
}

// Read a float32 value from buffer using big endian byte order.
func (m *BufferReader) ReadFloat32BE() float32 {
	return math.Float32frombits(m.ReadUint32BE())
}

// Read a float64 value from buffer using little endian byte order.
func (m *BufferReader) ReadFloat64LE() float64 {
	return math.Float64frombits(m.ReadUint64LE())
}

// Read a float64 value from buffer using big endian byte order.
func (m *BufferReader) ReadFloat64BE() float64 {
	return math.Float64frombits(m.ReadUint64BE())
}

// The base type of outgoing message buffer.
type BufferWriter struct {
	b []byte
}

func NewBufferWriter(buffer []byte) *BufferWriter {
	return &BufferWriter{buffer}
}

func (m *BufferWriter) Bytes() []byte {
	return m.b
}

// Write a byte slice into buffer.
func (m *BufferWriter) WriteBytes(d []byte) {
	m.b = append(m.b, d...)
}

// Write a string into buffer.
func (m *BufferWriter) WriteString(s string) {
	m.WriteBytes([]byte(s))
}

// Write a rune into buffer.
func (m *BufferWriter) WriteRune(r rune) {
	p := []byte{0, 0, 0, 0}
	n := utf8.EncodeRune(p, r)
	m.b = append(m.b, p[:n]...)
}

// Write a uint8 value into buffer.
func (m *BufferWriter) WriteUint8(v uint8) {
	m.b = append(m.b, byte(v))
}

// Write a uint16 value into buffer using little endian byte order.
func (m *BufferWriter) WriteUint16LE(v uint16) {
	m.b = append(m.b, byte(v), byte(v>>8))
}

// Write a uint16 value into buffer using big endian byte order.
func (m *BufferWriter) WriteUint16BE(v uint16) {
	m.b = append(m.b, byte(v>>8), byte(v))
}

// Write a uint32 value into buffer using little endian byte order.
func (m *BufferWriter) WriteUint32LE(v uint32) {
	m.b = append(m.b, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}

// Write a uint32 value into buffer using big endian byte order.
func (m *BufferWriter) WriteUint32BE(v uint32) {
	m.b = append(m.b, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

// Write a uint64 value into buffer using little endian byte order.
func (m *BufferWriter) WriteUint64LE(v uint64) {
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
func (m *BufferWriter) WriteUint64BE(v uint64) {
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
func (m *BufferWriter) WriteFloat32LE(v float32) {
	m.WriteUint32LE(math.Float32bits(v))
}

// Write a float32 value into buffer using big endian byte order.
func (m *BufferWriter) WriteFloat32BE(v float32) {
	m.WriteUint32BE(math.Float32bits(v))
}

// Write a float64 value into buffer using little endian byte order.
func (m *BufferWriter) WriteFloat64LE(v float64) {
	m.WriteUint64LE(math.Float64bits(v))
}

// Write a float64 value into buffer using big endian byte order.
func (m *BufferWriter) WriteFloat64BE(v float64) {
	m.WriteUint64BE(math.Float64bits(v))
}
