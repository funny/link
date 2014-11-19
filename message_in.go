package link

import (
	"encoding/binary"
	"unicode/utf8"
)

type InMessage []byte

func (m InMessage) Bytes() []byte {
	return []byte(m)
}

func (m InMessage) Copy() []byte {
	b := make([]byte, len(m))
	copy(b, m)
	return b
}

func (m *OutMessage) ReadSlice(n int) []byte {
	r := (*m)[0:n]
	*m = (*m)[n:]
	return r
}

func (m *OutMessage) ReadBytes(n int) []byte {
	r := make([]byte, n)
	copy(r, m.ReadSlice(n))
	return r
}

func (m *OutMessage) ReadString(n int) string {
	return string(m.ReadSlice(n))
}

func (m *OutMessage) ReadRune() rune {
	r, size := utf8.DecodeRune(*m)
	*m = (*m)[size:]
	return r
}

func (m *OutMessage) ReadByte() byte {
	r := (*m)[0]
	*m = (*m)[1:]
	return r
}

func (m *OutMessage) ReadInt8() int8 {
	r := int8((*m)[0])
	*m = (*m)[1:]
	return r
}

func (m *OutMessage) ReadUint8() uint8 {
	r := uint8((*m)[0])
	*m = (*m)[1:]
	return r
}

/*
little endian
*/

func (m *OutMessage) ReadInt16LE() int16 {
	return int16(m.ReadUint16LE())
}

func (m *OutMessage) ReadUint16LE() uint16 {
	r := binary.LittleEndian.Uint16(*m)
	*m = (*m)[2:]
	return r
}

func (m *OutMessage) ReadInt32LE() int32 {
	return int32(m.ReadUint32LE())
}

func (m *OutMessage) ReadUint32LE() uint32 {
	r := binary.LittleEndian.Uint32(*m)
	*m = (*m)[4:]
	return r
}

func (m *OutMessage) ReadInt64LE() int64 {
	return int64(m.ReadUint64LE())
}

func (m *OutMessage) ReadUint64LE() uint64 {
	r := binary.LittleEndian.Uint64(*m)
	*m = (*m)[8:]
	return r
}

/*
big endian
*/

func (m *OutMessage) ReadInt16BE() int16 {
	return int16(m.ReadUint16BE())
}

func (m *OutMessage) ReadUint16BE() uint16 {
	r := binary.BigEndian.Uint16(*m)
	*m = (*m)[2:]
	return r
}

func (m *OutMessage) ReadInt32BE() int32 {
	return int32(m.ReadUint32BE())
}

func (m *OutMessage) ReadUint32BE() uint32 {
	r := binary.BigEndian.Uint32(*m)
	*m = (*m)[4:]
	return r
}

func (m *OutMessage) ReadInt64BE() int64 {
	return int64(m.ReadUint64BE())
}

func (m *OutMessage) ReadUint64BE() uint64 {
	r := binary.BigEndian.Uint64(*m)
	*m = (*m)[8:]
	return r
}
