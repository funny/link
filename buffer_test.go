package link

import (
	"bytes"
	"github.com/funny/unitest"
	"testing"
)

func Test_Buffer(t *testing.T) {
	buffer := MakeBuffer(0, 0)
	PrepareBuffer(buffer)
	VerifyBuffer(t, buffer)
}

func Test_PoolBuffer(t *testing.T) {
	pool := NewMemPool(10, 1, 10)
	buffer := NewPoolBuffer(0, 0, pool)
	PrepareBuffer(buffer)
	VerifyBuffer(t, buffer)
}

func PrepareBuffer(buffer *Buffer) {
	buffer.WriteVarint(0x12345678AABBCCDD)
	buffer.WriteUvarint(0x12345678AABBCCDD)
	buffer.WriteUint8(0x12)
	buffer.WriteUint16LE(0x1234)
	buffer.WriteUint16BE(0x1234)
	buffer.WriteUint24LE(0x123456)
	buffer.WriteUint24BE(0x123456)
	buffer.WriteUint32LE(0x12345678)
	buffer.WriteUint32BE(0x12345678)
	buffer.WriteUint40LE(0x12345678AA)
	buffer.WriteUint40BE(0x12345678AA)
	buffer.WriteUint48LE(0x12345678AABB)
	buffer.WriteUint48BE(0x12345678AABB)
	buffer.WriteUint56LE(0x12345678AABBCC)
	buffer.WriteUint56BE(0x12345678AABBCC)
	buffer.WriteUint64LE(0x12345678AABBCCDD)
	buffer.WriteUint64BE(0x12345678AABBCCDD)
	buffer.WriteFloat32LE(88.01)
	buffer.WriteFloat64LE(99.02)
	buffer.WriteFloat32BE(88.01)
	buffer.WriteFloat64BE(99.02)
	buffer.WriteString("Hello1")
	buffer.WriteBytes([]byte("Hello2"))
	buffer.WriteBytes(bytes.Repeat([]byte("l"), 4096))

	buffer.WriteRune('好')
}

func VerifyBuffer(t *testing.T, buffer *Buffer) {
	unitest.Pass(t, buffer.ReadVarint() == 0x12345678AABBCCDD)
	unitest.Pass(t, buffer.ReadUvarint() == 0x12345678AABBCCDD)
	unitest.Pass(t, buffer.ReadUint8() == 0x12)
	unitest.Pass(t, buffer.ReadUint16LE() == 0x1234)
	unitest.Pass(t, buffer.ReadUint16BE() == 0x1234)
	unitest.Pass(t, buffer.ReadUint24LE() == 0x123456)
	unitest.Pass(t, buffer.ReadUint24BE() == 0x123456)
	unitest.Pass(t, buffer.ReadUint32LE() == 0x12345678)
	unitest.Pass(t, buffer.ReadUint32BE() == 0x12345678)
	unitest.Pass(t, buffer.ReadUint40LE() == 0x12345678AA)
	unitest.Pass(t, buffer.ReadUint40BE() == 0x12345678AA)
	unitest.Pass(t, buffer.ReadUint48LE() == 0x12345678AABB)
	unitest.Pass(t, buffer.ReadUint48BE() == 0x12345678AABB)
	unitest.Pass(t, buffer.ReadUint56LE() == 0x12345678AABBCC)
	unitest.Pass(t, buffer.ReadUint56BE() == 0x12345678AABBCC)
	unitest.Pass(t, buffer.ReadUint64LE() == 0x12345678AABBCCDD)
	unitest.Pass(t, buffer.ReadUint64BE() == 0x12345678AABBCCDD)
	unitest.Pass(t, buffer.ReadFloat32LE() == 88.01)
	unitest.Pass(t, buffer.ReadFloat64LE() == 99.02)
	unitest.Pass(t, buffer.ReadFloat32BE() == 88.01)
	unitest.Pass(t, buffer.ReadFloat64BE() == 99.02)
	unitest.Pass(t, buffer.ReadString(6) == "Hello1")
	unitest.Pass(t, bytes.Equal(buffer.ReadBytes(6), []byte("Hello2")))
	unitest.Pass(t, bytes.Equal(buffer.ReadBytes(4096), bytes.Repeat([]byte("l"), 4096)))

	r, _, _ := buffer.ReadRune()
	unitest.Pass(t, r == '好')
}
