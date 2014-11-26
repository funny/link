package link

import (
	"bytes"
	"github.com/funny/ceshi"
	"testing"
)

func init() {
	ceshi.StartMonitor(nil)
}

func Test_BigEndianBuffer(t *testing.T) {
	outBuffer := new(OutBufferBE)
	PrepareBuffer(outBuffer)

	inBuffer := new(InBufferBE)
	inBuffer.b = outBuffer.b
	VerifyBuffer(t, inBuffer)
}

func Test_LittleEndianBuffer(t *testing.T) {
	outBuffer := new(OutBufferLE)
	PrepareBuffer(outBuffer)

	inBuffer := new(InBufferLE)
	inBuffer.b = outBuffer.b
	VerifyBuffer(t, inBuffer)
}

func PrepareBuffer(buffer OutBuffer) {
	buffer.WriteByte(99)
	buffer.WriteInt8(-2)
	buffer.WriteUint8(1)
	buffer.WriteInt16(0x7FEE)
	buffer.WriteUint16(0xFFEE)
	buffer.WriteInt32(0x7FEEDDCC)
	buffer.WriteUint32(0xFFEEDDCC)
	buffer.WriteInt64(0x7FEEDDCCBBAA9988)
	buffer.WriteUint64(0xFFEEDDCCBBAA9988)
	buffer.WriteRune('好')
	buffer.WriteString("Hello1")
	buffer.WriteBytes([]byte("Hello2"))
	buffer.WriteBytes([]byte("Hello3"))
}

func VerifyBuffer(t *testing.T, buffer InBuffer) {
	ceshi.Pass(t, buffer.ReadByte() == 99)
	ceshi.Pass(t, buffer.ReadInt8() == -2)
	ceshi.Pass(t, buffer.ReadUint8() == 1)
	ceshi.Pass(t, buffer.ReadInt16() == 0x7FEE)
	ceshi.Pass(t, buffer.ReadUint16() == 0xFFEE)
	ceshi.Pass(t, buffer.ReadInt32() == 0x7FEEDDCC)
	ceshi.Pass(t, buffer.ReadUint32() == 0xFFEEDDCC)
	ceshi.Pass(t, buffer.ReadInt64() == 0x7FEEDDCCBBAA9988)
	ceshi.Pass(t, buffer.ReadUint64() == 0xFFEEDDCCBBAA9988)
	ceshi.Pass(t, buffer.ReadRune() == '好')
	ceshi.Pass(t, buffer.ReadString(6) == "Hello1")
	ceshi.Pass(t, bytes.Equal(buffer.ReadBytes(6), []byte("Hello2")))
	ceshi.Pass(t, bytes.Equal(buffer.ReadSlice(6), []byte("Hello3")))
}
