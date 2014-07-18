package packnet

import "encoding/binary"

// The buffer type. Can used for message encoding and decoding.
// When a writing operation needs more space, the buffer will auto grows.
// The reading cursor will auto move after each reading operation.
type Buffer struct {
	buff      []byte
	byteOrder binary.ByteOrder
	rpos      int
}

// Create an buffer instance.
func NewBuffer(buff []byte, byteOrder binary.ByteOrder) *Buffer {
	return &Buffer{
		buff:      buff,
		byteOrder: byteOrder,
	}
}

// Get current buffer capacity
func (b *Buffer) Cap() int {
	return cap(b.buff)
}

// Get current buffer length.
func (b *Buffer) Len() int {
	return len(b.buff)
}

// Get internal buffer. Had better do not modify it.
func (b *Buffer) GetBytes() []byte {
	return b.buff
}

// Replace buffer. Dirty operation, not recommend.
func (b *Buffer) SetBytes(buff []byte) {
	b.buff = buff
}

// Grows the buffer for writing operation.
// Return a slice of the new space.
func (b *Buffer) Grows(n int) []byte {
	oldLen := len(b.buff)
	newLen := oldLen + n
	if newLen <= cap(b.buff) {
		b.buff = b.buff[0:newLen]
	} else {
		buff := make([]byte, newLen, newLen+512)
		copy(buff, b.buff)
		b.buff = buff
	}
	return b.buff[oldLen:newLen]
}

// Write an uint8 value into buffer.
func (b *Buffer) WriteUint8(v uint8) {
	b.Grows(1)[0] = byte(v)
}

// Write a byte value into buffer.
func (b *Buffer) WriteByte(v byte) {
	b.Grows(1)[0] = v
}

// Write an int8 value into buffer.
func (b *Buffer) WriteInt8(v int8) {
	b.Grows(1)[0] = byte(v)
}

// Write an uint16 value into buffer.
func (b *Buffer) WriteUint16(v uint16) {
	b.byteOrder.PutUint16(b.Grows(2), v)
}

// Write an uint32 value into buffer.
func (b *Buffer) WriteUint32(v uint32) {
	b.byteOrder.PutUint32(b.Grows(4), v)
}

// Write an uint64 value into buffer.
func (b *Buffer) WriteUint64(v uint64) {
	b.byteOrder.PutUint64(b.Grows(8), v)
}

// Write an int16 value into buffer.
func (b *Buffer) WriteInt16(v int16) {
	b.byteOrder.PutUint16(b.Grows(2), uint16(v))
}

// Write an int32 value into buffer.
func (b *Buffer) WriteInt32(v int32) {
	b.byteOrder.PutUint32(b.Grows(4), uint32(v))
}

// Write an int64 value into buffer.
func (b *Buffer) WriteInt64(v int64) {
	b.byteOrder.PutUint64(b.Grows(8), uint64(v))
}

// Write a byte slice into buffer.
func (b *Buffer) WriteBytes(v []byte) {
	copy(b.Grows(len(v)), v)
}

// Write a string into buffer.
func (b *Buffer) WriteString(v string) {
	copy(b.Grows(len(v)), v)
}

// Get current read position.
func (b *Buffer) GetReadPosition() int {
	return b.rpos
}

// Get current read position.
// Dirty operation, not recommend.
func (b *Buffer) SetReadPosition(pos int) {
	b.rpos = pos
}

// Read an uint8 value from buffer.
func (b *Buffer) ReadUint8() uint8 {
	r := b.buff[b.rpos]
	b.rpos += 1
	return r
}

// Read a byte value from buffer.
func (b *Buffer) ReadByte() byte {
	r := b.buff[b.rpos]
	b.rpos += 1
	return byte(r)
}

// Read an int8 value from buffer.
func (b *Buffer) ReadInt8() int8 {
	r := b.buff[b.rpos]
	b.rpos += 1
	return int8(r)
}

// Read an uint16 value from buffer.
func (b *Buffer) ReadUint16() uint16 {
	r := b.byteOrder.Uint16(b.buff[b.rpos:])
	b.rpos += 2
	return r
}

// Read an uint32 value from buffer.
func (b *Buffer) ReadUint32() uint32 {
	r := b.byteOrder.Uint32(b.buff[b.rpos:])
	b.rpos += 4
	return r
}

// Read an uint64 value from buffer.
func (b *Buffer) ReadUint64() uint64 {
	r := b.byteOrder.Uint64(b.buff[b.rpos:])
	b.rpos += 8
	return r
}

// Read an int16 value from buffer.
func (b *Buffer) ReadInt16() int16 {
	r := b.byteOrder.Uint16(b.buff[b.rpos:])
	b.rpos += 2
	return int16(r)
}

// Read an int32 value from buffer.
func (b *Buffer) ReadInt32() int32 {
	r := b.byteOrder.Uint32(b.buff[b.rpos:])
	b.rpos += 4
	return int32(r)
}

// Read an int64 value from buffer.
func (b *Buffer) ReadInt64() int64 {
	r := b.byteOrder.Uint64(b.buff[b.rpos:])
	b.rpos += 8
	return int64(r)
}

// Read bytes.
func (b *Buffer) ReadBytes(length int) []byte {
	r := b.buff[b.rpos : b.rpos+length]
	b.rpos += length
	return r
}

// Read string.
func (b *Buffer) ReadString(length int) string {
	r := b.buff[b.rpos : b.rpos+length]
	b.rpos += length
	return string(r)
}
