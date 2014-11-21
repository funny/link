package link

import "unicode/utf8"

// Outgoing message.
type OutMessage []byte

// Convert to byte slice.
func (m OutMessage) Bytes() []byte {
	return []byte(m)
}

// Copy data.
func (m OutMessage) Copy() []byte {
	b := make([]byte, len(m))
	copy(b, m)
	return b
}

// Append a byte slice into buffer.
func (m *OutMessage) AppendBytes(d []byte) {
	*m = append(*m, d...)
}

// Append a string into buffer.
func (m *OutMessage) AppendString(s string) {
	*m = append(*m, s...)
}

// Append a rune into buffer.
func (m *OutMessage) AppendRune(r rune) {
	p := []byte{0, 0, 0, 0}
	n := utf8.EncodeRune(p, r)
	*m = append(*m, p[:n]...)
}

// Append a byte value into buffer.
func (m *OutMessage) AppendByte(v byte) {
	*m = append(*m, v)
}

// Append a int8 value into buffer.
func (m *OutMessage) AppendInt8(v int8) {
	*m = append(*m, byte(v))
}

// Append a uint8 value into buffer.
func (m *OutMessage) AppendUint8(v uint8) {
	*m = append(*m, byte(v))
}

/*
little endian
*/

// Append a little endian int16 value into buffer.
func (m *OutMessage) AppendInt16LE(v int16) {
	m.AppendUint16LE(uint16(v))
}

// Append a little endian uint16 value into buffer.
func (m *OutMessage) AppendUint16LE(v uint16) {
	*m = append(*m, byte(v), byte(v>>8))
}

// Append a little endian int32 value into buffer.
func (m *OutMessage) AppendInt32LE(v int32) {
	m.AppendUint32LE(uint32(v))
}

// Append a little endian uint32 value into buffer.
func (m *OutMessage) AppendUint32LE(v uint32) {
	*m = append(*m, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}

// Append a little endian int64 value into buffer.
func (m *OutMessage) AppendInt64LE(v int64) {
	m.AppendUint64LE(uint64(v))
}

// Append a little endian uint64 value into buffer.
func (m *OutMessage) AppendUint64LE(v uint64) {
	*m = append(*m,
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

/*
big endian
*/

// Append a big endian int16 value into buffer.
func (m *OutMessage) AppendInt16BE(v int16) {
	m.AppendUint16BE(uint16(v))
}

// Append a big endian uint16 value into buffer.
func (m *OutMessage) AppendUint16BE(v uint16) {
	*m = append(*m, byte(v>>8), byte(v))
}

// Append a big endian int32 value into buffer.
func (m *OutMessage) AppendInt32BE(v int32) {
	m.AppendUint32BE(uint32(v))
}

// Append a big endian uint32 value into buffer.
func (m *OutMessage) AppendUint32BE(v uint32) {
	*m = append(*m, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

// Append a big endian int64 value into buffer.
func (m *OutMessage) AppendInt64BE(v int64) {
	m.AppendUint64BE(uint64(v))
}

// Append a big endian uint64 value into buffer.
func (m *OutMessage) AppendUint64BE(v uint64) {
	*m = append(*m,
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
