package link

import (
	"encoding/binary"
	"unicode/utf8"
)

// Incoming message.
type InMessage []byte

// Convert to byte slice.
func (m InMessage) Bytes() []byte {
	return []byte(m)
}

// Copy data.
func (m InMessage) Copy() []byte {
	b := make([]byte, len(m))
	copy(b, m)
	return b
}

// Slice some bytes from buffer.
func (m *InMessage) ReadSlice(n int) []byte {
	r := (*m)[0:n]
	*m = (*m)[n:]
	return r
}

// Copy some bytes from buffer.
func (m *InMessage) ReadBytes(n int) []byte {
	r := make([]byte, n)
	copy(r, m.ReadSlice(n))
	return r
}

// Read a string from buffer.
func (m *InMessage) ReadString(n int) string {
	return string(m.ReadSlice(n))
}

// Read a rune from buffer.
func (m *InMessage) ReadRune() rune {
	r, size := utf8.DecodeRune(*m)
	*m = (*m)[size:]
	return r
}

// Read a byte value from buffer.
func (m *InMessage) ReadByte() byte {
	r := (*m)[0]
	*m = (*m)[1:]
	return r
}

// Read a int8 value from buffer.
func (m *InMessage) ReadInt8() int8 {
	r := int8((*m)[0])
	*m = (*m)[1:]
	return r
}

// Read a uint8 value from buffer.
func (m *InMessage) ReadUint8() uint8 {
	r := uint8((*m)[0])
	*m = (*m)[1:]
	return r
}

/*
little endian
*/

// Read a little endian int16 value from buffer.
func (m *InMessage) ReadInt16LE() int16 {
	return int16(m.ReadUint16LE())
}

// Read a little endian uint16 value from buffer.
func (m *InMessage) ReadUint16LE() uint16 {
	r := binary.LittleEndian.Uint16(*m)
	*m = (*m)[2:]
	return r
}

// Read a little endian int32 value from buffer.
func (m *InMessage) ReadInt32LE() int32 {
	return int32(m.ReadUint32LE())
}

// Read a little endian uint32 value from buffer.
func (m *InMessage) ReadUint32LE() uint32 {
	r := binary.LittleEndian.Uint32(*m)
	*m = (*m)[4:]
	return r
}

// Read a little endian int64 value from buffer.
func (m *InMessage) ReadInt64LE() int64 {
	return int64(m.ReadUint64LE())
}

// Read a little endian uint64 value from buffer.
func (m *InMessage) ReadUint64LE() uint64 {
	r := binary.LittleEndian.Uint64(*m)
	*m = (*m)[8:]
	return r
}

/*
big endian
*/

// Read a big endian int16 value from buffer.
func (m *InMessage) ReadInt16BE() int16 {
	return int16(m.ReadUint16BE())
}

// Read a big endian uint16 value from buffer.
func (m *InMessage) ReadUint16BE() uint16 {
	r := binary.BigEndian.Uint16(*m)
	*m = (*m)[2:]
	return r
}

// Read a big endian int32 value from buffer.
func (m *InMessage) ReadInt32BE() int32 {
	return int32(m.ReadUint32BE())
}

// Read a big endian uint32 value from buffer.
func (m *InMessage) ReadUint32BE() uint32 {
	r := binary.BigEndian.Uint32(*m)
	*m = (*m)[4:]
	return r
}

// Read a big endian int64 value from buffer.
func (m *InMessage) ReadInt64BE() int64 {
	return int64(m.ReadUint64BE())
}

// Read a big endian uint64 value from buffer.
func (m *InMessage) ReadUint64BE() uint64 {
	r := binary.BigEndian.Uint64(*m)
	*m = (*m)[8:]
	return r
}
