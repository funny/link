package link

import "unicode/utf8"

// OutMessage is bytes, bytes is OutMessage.
type OutMessage []byte

// Convert OutMessage to byte slice.
func (m OutMessage) Bytes() []byte {
	return []byte(m)
}

func (m OutMessage) Copy() []byte {
	b := make([]byte, len(m))
	copy(b, m)
	return b
}

func (m *OutMessage) AppendBytes(d []byte) {
	*m = append(*m, d...)
}

func (m *OutMessage) AppendString(s string) {
	*m = append(*m, s...)
}

func (m *OutMessage) AppendRune(r rune) {
	p := make([]byte, utf8.RuneLen(r))
	utf8.EncodeRune(p, r)
	*m = append(*m, p...)
}

// Append Byte value into packet.
func (m *OutMessage) AppendByte(v byte) {
	*m = append(*m, v)
}

// Append Int8 value into packet.
func (m *OutMessage) AppendInt8(v int8) {
	*m = append(*m, byte(v))
}

// Append Uint8 value into packet.
func (m *OutMessage) AppendUint8(v uint8) {
	*m = append(*m, byte(v))
}

/*
little endian
*/

// Append little endian Int16 value into packet.
func (m *OutMessage) AppendInt16LE(v int16) {
	m.AppendUint16LE(uint16(v))
}

// Append little endian Uint16 value into packet.
func (m *OutMessage) AppendUint16LE(v uint16) {
	*m = append(*m, byte(v), byte(v>>8))
}

// Append little endian Int32 value into packet.
func (m *OutMessage) AppendInt32LE(v int32) {
	m.AppendUint32LE(uint32(v))
}

// Append little endian Uint32 value into packet.
func (m *OutMessage) AppendUint32LE(v uint32) {
	*m = append(*m, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}

// Append little endian Int64 value into packet.
func (m *OutMessage) AppendInt64LE(v int64) {
	m.AppendUint64LE(uint64(v))
}

// Append little endian Uint64 value into packet.
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

// Append big endian Int16 value into packet.
func (m *OutMessage) AppendInt16BE(v int16) {
	m.AppendUint16BE(uint16(v))
}

// Append big endian Uint16 value into packet.
func (m *OutMessage) AppendUint16BE(v uint16) {
	*m = append(*m, byte(v>>8), byte(v))
}

// Append big endian Int32 value into packet.
func (m *OutMessage) AppendInt32BE(v int32) {
	m.AppendUint32BE(uint32(v))
}

// Append big endian Uint32 value into packet.
func (m *OutMessage) AppendUint32BE(v uint32) {
	*m = append(*m, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

// Append big endian Int64 value into packet.
func (m *OutMessage) AppendInt64BE(v int64) {
	m.AppendUint64BE(uint64(v))
}

// Append big endian Uint64 value into packet.
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
