package link

import (
	"bytes"
	"github.com/funny/unitest"
	"runtime"
	"testing"
)

func Test_Buffer(t *testing.T) {
	var buffer = NewOutBuffer()

	PrepareBuffer(buffer)

	VerifyBuffer(t, &InBuffer{Data: buffer.Data})
}

func PrepareBuffer(buffer *OutBuffer) {
	buffer.WriteUint8(123)
	buffer.WriteUint16LE(0xFFEE)
	buffer.WriteUint16BE(0xFFEE)
	buffer.WriteUint32LE(0xFFEEDDCC)
	buffer.WriteUint32BE(0xFFEEDDCC)
	buffer.WriteUint64LE(0xFFEEDDCCBBAA9988)
	buffer.WriteUint64BE(0xFFEEDDCCBBAA9988)
	buffer.WriteFloat32LE(88.01)
	buffer.WriteFloat64LE(99.02)
	buffer.WriteFloat32BE(88.01)
	buffer.WriteFloat64BE(99.02)
	buffer.WriteRune('好')
	buffer.WriteString("Hello1")
	buffer.WriteBytes([]byte("Hello2"))
	buffer.WriteBytes([]byte("Hello3"))
}

func VerifyBuffer(t *testing.T, buffer *InBuffer) {
	unitest.Pass(t, buffer.ReadUint8() == 123)
	unitest.Pass(t, buffer.ReadUint16LE() == 0xFFEE)
	unitest.Pass(t, buffer.ReadUint16BE() == 0xFFEE)
	unitest.Pass(t, buffer.ReadUint32LE() == 0xFFEEDDCC)
	unitest.Pass(t, buffer.ReadUint32BE() == 0xFFEEDDCC)
	unitest.Pass(t, buffer.ReadUint64LE() == 0xFFEEDDCCBBAA9988)
	unitest.Pass(t, buffer.ReadUint64BE() == 0xFFEEDDCCBBAA9988)
	unitest.Pass(t, buffer.ReadFloat32LE() == 88.01)
	unitest.Pass(t, buffer.ReadFloat64LE() == 99.02)
	unitest.Pass(t, buffer.ReadFloat32BE() == 88.01)
	unitest.Pass(t, buffer.ReadFloat64BE() == 99.02)
	unitest.Pass(t, buffer.ReadRune() == '好')
	unitest.Pass(t, buffer.ReadString(6) == "Hello1")
	unitest.Pass(t, bytes.Equal(buffer.ReadBytes(6), []byte("Hello2")))
	unitest.Pass(t, bytes.Equal(buffer.Slice(6), []byte("Hello3")))
}

func Benchmark_SetFinalizer1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var x = &InBuffer{}
		runtime.SetFinalizer(x, func(x *InBuffer) {
		})
	}
}

func Benchmark_SetFinalizer2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var x = &InBuffer{}
		runtime.SetFinalizer(x, nil)
	}
}

func Benchmark_MakeBytes_512(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = make([]byte, 512)
	}
}

func Benchmark_MakeBytes_1024(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = make([]byte, 1024)
	}
}

func Benchmark_MakeBytes_4096(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = make([]byte, 4096)
	}
}
