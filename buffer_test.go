package link

import (
	"bytes"
	"github.com/funny/unitest"
	"testing"
)

func Test_Buffer(t *testing.T) {
	outBuffer := new(SimpleOutBuffer)
	PrepareBuffer(outBuffer)

	inBuffer := new(SimpleInBuffer)
	inBuffer.b = outBuffer.b
	VerifyBuffer(t, inBuffer)
}

func PrepareBuffer(buffer OutBuffer) {
	buffer.WriteUint8(123)
	buffer.WriteUint16LE(0xFFEE)
	buffer.WriteUint16BE(0xFFEE)
	buffer.WriteUint32LE(0xFFEEDDCC)
	buffer.WriteUint32BE(0xFFEEDDCC)
	buffer.WriteUint64LE(0xFFEEDDCCBBAA9988)
	buffer.WriteUint64BE(0xFFEEDDCCBBAA9988)
	buffer.WriteFloat32(88.01)
	buffer.WriteFloat64(99.02)
	buffer.WriteRune('好')
	buffer.WriteString("Hello1")
	buffer.WriteBytes([]byte("Hello2"))
	buffer.WriteBytes([]byte("Hello3"))
}

func VerifyBuffer(t *testing.T, buffer InBuffer) {
	unitest.Pass(t, buffer.ReadUint8() == 123)
	unitest.Pass(t, buffer.ReadUint16LE() == 0xFFEE)
	unitest.Pass(t, buffer.ReadUint16BE() == 0xFFEE)
	unitest.Pass(t, buffer.ReadUint32LE() == 0xFFEEDDCC)
	unitest.Pass(t, buffer.ReadUint32BE() == 0xFFEEDDCC)
	unitest.Pass(t, buffer.ReadUint64LE() == 0xFFEEDDCCBBAA9988)
	unitest.Pass(t, buffer.ReadUint64BE() == 0xFFEEDDCCBBAA9988)
	unitest.Pass(t, buffer.ReadFloat32() == 88.01)
	unitest.Pass(t, buffer.ReadFloat64() == 99.02)
	unitest.Pass(t, buffer.ReadRune() == '好')
	unitest.Pass(t, buffer.ReadString(6) == "Hello1")
	unitest.Pass(t, bytes.Equal(buffer.ReadBytes(6), []byte("Hello2")))
	unitest.Pass(t, bytes.Equal(buffer.ReadSlice(6), []byte("Hello3")))
}
