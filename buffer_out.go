package link

import "unicode/utf8"

// Outgoing messsage buffer.
type OutBuffer interface {
	// Get internal buffer.
	Get() []byte

	// Set internal buffer.
	Set([]byte)

	// Get message length.
	Len() int

	// Prepare buffer for next write.
	Prepare(head, size int)

	// Copy data.
	Copy() []byte

	// Append a byte slice into buffer.
	AppendBytes(d []byte)

	// Append a string into buffer.
	AppendString(s string)

	// Append a rune into buffer.
	AppendRune(r rune)

	// Append a byte value into buffer.
	AppendByte(v byte)

	// Append a int8 value into buffer.
	AppendInt8(v int8)

	// Append a uint8 value into buffer.
	AppendUint8(v uint8)

	// Append a int16 value into buffer.
	AppendInt16(v int16)

	// Append a uint16 value into buffer.
	AppendUint16(v uint16)

	// Append a int32 value into buffer.
	AppendInt32(v int32)

	// Append a uint32 value into buffer.
	AppendUint32(v uint32)

	// Append a int64 value into buffer.
	AppendInt64(v int64)

	// Append a uint64 value into buffer.
	AppendUint64(v uint64)
}

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
