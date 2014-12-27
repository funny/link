package link

import (
	"encoding/binary"
	"io"
	"math"
	"sync/atomic"
	"unicode/utf8"
)

// Incomming message buffer.
type InBuffer struct {
	Data    []byte // Buffer data.
	ReadPos int    // Read position.
	isFreed bool
}

// Create a new incomming message buffer.
func NewInBuffer() *InBuffer {
	return &InBuffer{Data: make([]byte, 0, 1024)}
}

func (in *InBuffer) reset() {
	in.Data = nil
	in.ReadPos = 0
	in.isFreed = false
}

// Return the buffer to buffer pool.
func (in *InBuffer) Free() {
	if in.isFreed {
		panic("link.InBuffer: double free")
	}
	// TODO
}

// Prepare buffer for next message.
// This method is for custom protocol only.
// Dont' use it in application logic.
func (in *InBuffer) Prepare(size int) {
	if cap(in.Data) < size {
		in.Data = make([]byte, size)
	} else {
		in.Data = in.Data[0:size]
	}
}

// Slice some bytes from buffer.
func (in *InBuffer) Slice(n int) []byte {
	r := in.Data[in.ReadPos : in.ReadPos+n]
	in.ReadPos += n
	return r
}

// Implement io.Reader interface
func (in *InBuffer) Read(b []byte) (int, error) {
	if in.ReadPos == len(in.Data) {
		return 0, io.EOF
	}
	n := len(b)
	if n+in.ReadPos > len(in.Data) {
		n = len(in.Data) - in.ReadPos
	}
	copy(b, in.Data[in.ReadPos:])
	in.ReadPos += n
	return n, nil
}

// Read some bytes from buffer.
func (in *InBuffer) ReadBytes(n int) []byte {
	x := make([]byte, n)
	copy(x, in.Slice(n))
	return x
}

// Read a string from buffer.
func (in *InBuffer) ReadString(n int) string {
	return string(in.Slice(n))
}

// Read a rune from buffer.
func (in *InBuffer) ReadRune() rune {
	x, size := utf8.DecodeRune(in.Data[in.ReadPos:])
	in.ReadPos += size
	return x
}

// Read a uint8 value from buffer.
func (in *InBuffer) ReadUint8() uint8 {
	return uint8(in.Slice(1)[0])
}

// Read a uint16 value from buffer using little endian byte order.
func (in *InBuffer) ReadUint16LE() uint16 {
	return binary.LittleEndian.Uint16(in.Slice(2))
}

// Read a uint16 value from buffer using big endian byte order.
func (in *InBuffer) ReadUint16BE() uint16 {
	return binary.BigEndian.Uint16(in.Slice(2))
}

// Read a uint32 value from buffer using little endian byte order.
func (in *InBuffer) ReadUint32LE() uint32 {
	return binary.LittleEndian.Uint32(in.Slice(4))
}

// Read a uint32 value from buffer using big endian byte order.
func (in *InBuffer) ReadUint32BE() uint32 {
	return binary.BigEndian.Uint32(in.Slice(4))
}

// Read a uint64 value from buffer using little endian byte order.
func (in *InBuffer) ReadUint64LE() uint64 {
	return binary.LittleEndian.Uint64(in.Slice(8))
}

// Read a uint64 value from buffer using big endian byte order.
func (in *InBuffer) ReadUint64BE() uint64 {
	return binary.BigEndian.Uint64(in.Slice(8))
}

// Read a float32 value from buffer using little endian byte order.
func (in *InBuffer) ReadFloat32LE() float32 {
	return math.Float32frombits(in.ReadUint32LE())
}

// Read a float32 value from buffer using big endian byte order.
func (in *InBuffer) ReadFloat32BE() float32 {
	return math.Float32frombits(in.ReadUint32BE())
}

// Read a float64 value from buffer using little endian byte order.
func (in *InBuffer) ReadFloat64LE() float64 {
	return math.Float64frombits(in.ReadUint64LE())
}

// Read a float64 value from buffer using big endian byte order.
func (in *InBuffer) ReadFloat64BE() float64 {
	return math.Float64frombits(in.ReadUint64BE())
}

// Outgoing message buffer.
type OutBuffer struct {
	Data        []byte // Buffer data.
	isFreed     bool
	isBroadcast bool
	refCount    int32
}

// Create a new outgoing message buffer.
func NewOutBuffer() *OutBuffer {
	return &OutBuffer{Data: make([]byte, 0, 1024)}
}

func (out *OutBuffer) reset() {
	out.Data = nil
	out.isFreed = false
	out.isBroadcast = false
	out.refCount = 0
}

func (out *OutBuffer) broadcastUse() {
	atomic.AddInt32(&out.refCount, 1)
}

func (out *OutBuffer) broadcastFree() {
	if out.isBroadcast && atomic.AddInt32(&out.refCount, -1) == 0 {
		out.Free()
	}
}

// Return the buffer to buffer pool.
func (out *OutBuffer) Free() {
	if out.isFreed {
		panic("link.OutBuffer: double free")
	}
	// TODO
}

// Prepare for next message.
// This method is for custom protocol only.
// Don't use it in application logic.
func (out *OutBuffer) Prepare(size int) {
	if cap(out.Data) < size {
		out.Data = make([]byte, 0, size)
	} else {
		out.Data = out.Data[0:0]
	}
}

// Append some bytes into buffer.
func (out *OutBuffer) Append(p ...byte) {
	out.Data = append(out.Data, p...)
}

// Implement io.Writer interface.
func (out *OutBuffer) Write(p []byte) (int, error) {
	out.Data = append(out.Data, p...)
	return len(p), nil
}

// Write a byte slice into buffer.
func (out *OutBuffer) WriteBytes(d []byte) {
	out.Append(d...)
}

// Write a string into buffer.
func (out *OutBuffer) WriteString(s string) {
	out.Append([]byte(s)...)
}

// Write a rune into buffer.
func (out *OutBuffer) WriteRune(r rune) {
	p := []byte{0, 0, 0, 0}
	n := utf8.EncodeRune(p, r)
	out.Append(p[:n]...)
}

// Write a uint8 value into buffer.
func (out *OutBuffer) WriteUint8(v uint8) {
	out.Append(byte(v))
}

// Write a uint16 value into buffer using little endian byte order.
func (out *OutBuffer) WriteUint16LE(v uint16) {
	out.Append(byte(v), byte(v>>8))
}

// Write a uint16 value into buffer using big endian byte order.
func (out *OutBuffer) WriteUint16BE(v uint16) {
	out.Append(byte(v>>8), byte(v))
}

// Write a uint32 value into buffer using little endian byte order.
func (out *OutBuffer) WriteUint32LE(v uint32) {
	out.Append(byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}

// Write a uint32 value into buffer using big endian byte order.
func (out *OutBuffer) WriteUint32BE(v uint32) {
	out.Append(byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

// Write a uint64 value into buffer using little endian byte order.
func (out *OutBuffer) WriteUint64LE(v uint64) {
	out.Append(
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
func (out *OutBuffer) WriteUint64BE(v uint64) {
	out.Append(
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
func (out *OutBuffer) WriteFloat32LE(v float32) {
	out.WriteUint32LE(math.Float32bits(v))
}

// Write a float32 value into buffer using big endian byte order.
func (out *OutBuffer) WriteFloat32BE(v float32) {
	out.WriteUint32BE(math.Float32bits(v))
}

// Write a float64 value into buffer using little endian byte order.
func (out *OutBuffer) WriteFloat64LE(v float64) {
	out.WriteUint64LE(math.Float64bits(v))
}

// Write a float64 value into buffer using big endian byte order.
func (out *OutBuffer) WriteFloat64BE(v float64) {
	out.WriteUint64BE(math.Float64bits(v))
}
