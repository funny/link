package link

import (
	"encoding/binary"
	"math"
	"unicode/utf8"
)

type Buffer struct {
	Data []byte
}

func (buffer *Buffer) Append(p ...byte) {
	buffer.Data = append(buffer.Data, p...)
}

func (buffer *Buffer) Write(p []byte) (int, error) {
	buffer.Data = append(buffer.Data, p...)
	return len(p), nil
}

// Buffer reader.
type BufferReader struct {
	*Buffer
	i int
}

func NewBufferReader(buffer *Buffer) *BufferReader {
	return &BufferReader{buffer, 0}
}

// Slice some bytes from buffer.
func (r *BufferReader) Slice(n int) []byte {
	x := r.Data[r.i : r.i+n]
	r.i += n
	return x
}

// Copy some bytes from buffer.
func (r *BufferReader) ReadBytes(n int) []byte {
	x := make([]byte, n)
	copy(x, r.Slice(n))
	return x
}

// Read a string from buffer.
func (r *BufferReader) ReadString(n int) string {
	return string(r.Slice(n))
}

// Read a rune from buffer.
func (r *BufferReader) ReadRune() rune {
	x, size := utf8.DecodeRune(r.Data[r.i:])
	r.i += size
	return x
}

// Read a uint8 value from buffer.
func (r *BufferReader) ReadUint8() uint8 {
	x := uint8(r.Data[r.i])
	r.i += 1
	return x
}

// Read a uint16 value from buffer using little endian byte order.
func (r *BufferReader) ReadUint16LE() uint16 {
	x := binary.LittleEndian.Uint16(r.Data[r.i:])
	r.i += 2
	return x
}

// Read a uint16 value from buffer using big endian byte order.
func (r *BufferReader) ReadUint16BE() uint16 {
	x := binary.BigEndian.Uint16(r.Data[r.i:])
	r.i += 2
	return x
}

// Read a uint32 value from buffer using little endian byte order.
func (r *BufferReader) ReadUint32LE() uint32 {
	x := binary.LittleEndian.Uint32(r.Data[r.i:])
	r.i += 4
	return x
}

// Read a uint32 value from buffer using big endian byte order.
func (r *BufferReader) ReadUint32BE() uint32 {
	x := binary.BigEndian.Uint32(r.Data[r.i:])
	r.i += 4
	return x
}

// Read a uint64 value from buffer using little endian byte order.
func (r *BufferReader) ReadUint64LE() uint64 {
	x := binary.LittleEndian.Uint64(r.Data[r.i:])
	r.i += 8
	return x
}

// Read a uint64 value from buffer using big endian byte order.
func (r *BufferReader) ReadUint64BE() uint64 {
	x := binary.BigEndian.Uint64(r.Data[r.i:])
	r.i += 8
	return x
}

// Read a float32 value from buffer using little endian byte order.
func (r *BufferReader) ReadFloat32LE() float32 {
	return math.Float32frombits(r.ReadUint32LE())
}

// Read a float32 value from buffer using big endian byte order.
func (r *BufferReader) ReadFloat32BE() float32 {
	return math.Float32frombits(r.ReadUint32BE())
}

// Read a float64 value from buffer using little endian byte order.
func (r *BufferReader) ReadFloat64LE() float64 {
	return math.Float64frombits(r.ReadUint64LE())
}

// Read a float64 value from buffer using big endian byte order.
func (r *BufferReader) ReadFloat64BE() float64 {
	return math.Float64frombits(r.ReadUint64BE())
}

// Buffer writer
type BufferWriter struct {
	*Buffer
}

func NewBufferWriter(buffer *Buffer) *BufferWriter {
	return &BufferWriter{buffer}
}

// Write a byte slice into buffer.
func (w *BufferWriter) WriteBytes(d []byte) {
	w.Append(d...)
}

// Write a string into buffer.
func (w *BufferWriter) WriteString(s string) {
	w.Append([]byte(s)...)
}

// Write a rune into buffer.
func (w *BufferWriter) WriteRune(r rune) {
	p := []byte{0, 0, 0, 0}
	n := utf8.EncodeRune(p, r)
	w.Append(p[:n]...)
}

// Write a uint8 value into buffer.
func (w *BufferWriter) WriteUint8(v uint8) {
	w.Append(byte(v))
}

// Write a uint16 value into buffer using little endian byte order.
func (w *BufferWriter) WriteUint16LE(v uint16) {
	w.Append(byte(v), byte(v>>8))
}

// Write a uint16 value into buffer using big endian byte order.
func (w *BufferWriter) WriteUint16BE(v uint16) {
	w.Append(byte(v>>8), byte(v))
}

// Write a uint32 value into buffer using little endian byte order.
func (w *BufferWriter) WriteUint32LE(v uint32) {
	w.Append(byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}

// Write a uint32 value into buffer using big endian byte order.
func (w *BufferWriter) WriteUint32BE(v uint32) {
	w.Append(byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

// Write a uint64 value into buffer using little endian byte order.
func (w *BufferWriter) WriteUint64LE(v uint64) {
	w.Append(
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
func (w *BufferWriter) WriteUint64BE(v uint64) {
	w.Append(
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
func (w *BufferWriter) WriteFloat32LE(v float32) {
	w.WriteUint32LE(math.Float32bits(v))
}

// Write a float32 value into buffer using big endian byte order.
func (w *BufferWriter) WriteFloat32BE(v float32) {
	w.WriteUint32BE(math.Float32bits(v))
}

// Write a float64 value into buffer using little endian byte order.
func (w *BufferWriter) WriteFloat64LE(v float64) {
	w.WriteUint64LE(math.Float64bits(v))
}

// Write a float64 value into buffer using big endian byte order.
func (w *BufferWriter) WriteFloat64BE(v float64) {
	w.WriteUint64BE(math.Float64bits(v))
}
