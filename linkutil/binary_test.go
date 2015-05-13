package linkutil

import (
	"encoding/binary"
	"github.com/funny/unitest"
	"testing"
)

func GetUintLE(b []byte, n uint) (r uint64) {
	for i := uint(0); i < n; i++ {
		r |= uint64(b[i]) << (8 * i)
	}
	return
}

func PutUintLE(b []byte, n uint, v uint64) {
	for i := uint(0); i < n; i++ {
		b[i] = byte(v >> (8 * i))
	}
}

func GetUintBE(b []byte, n uint) (r uint64) {
	for i := uint(0); i < n; i++ {
		r |= uint64(b[i]) << (8 * (n - 1 - i))
	}
	return
}

func PutUintBE(b []byte, n uint, v uint64) {
	for i := uint(0); i < n; i++ {
		b[i] = byte(v >> (8 * (n - 1 - i)))
	}
}

func Test_Binary(t *testing.T) {
	x := make([]byte, 8)

	PutUintLE(x, 2, uint64(0xAB))
	unitest.Pass(t, GetUintLE(x, 2) == 0xAB)
	unitest.Pass(t, binary.LittleEndian.Uint16(x) == 0xAB)

	PutUintBE(x, 2, uint64(0xAB))
	unitest.Pass(t, GetUintBE(x, 2) == 0xAB)
	unitest.Pass(t, binary.BigEndian.Uint16(x) == 0xAB)

	PutUintLE(x, 4, uint64(0xABCD))
	unitest.Pass(t, GetUintLE(x, 4) == 0xABCD)
	unitest.Pass(t, binary.LittleEndian.Uint32(x) == 0xABCD)

	PutUintBE(x, 4, uint64(0xABCD))
	unitest.Pass(t, GetUintBE(x, 4) == 0xABCD)
	unitest.Pass(t, binary.BigEndian.Uint32(x) == 0xABCD)

	PutUintLE(x, 8, uint64(0xABCDEFAB))
	unitest.Pass(t, GetUintLE(x, 8) == 0xABCDEFAB)
	unitest.Pass(t, binary.LittleEndian.Uint64(x) == 0xABCDEFAB)

	PutUintBE(x, 8, uint64(0xABCDEFAB))
	unitest.Pass(t, GetUintBE(x, 8) == 0xABCDEFAB)
	unitest.Pass(t, binary.BigEndian.Uint64(x) == 0xABCDEFAB)
}

func Benchmark_Binary_PutUintLE_2(b *testing.B) {
	x := make([]byte, 2)
	for i := 0; i < b.N; i++ {
		PutUintLE(x, 2, uint64(i))
	}
}

func Benchmark_Binary_PutUintLE_3(b *testing.B) {
	x := make([]byte, 3)
	for i := 0; i < b.N; i++ {
		PutUintLE(x, 3, uint64(i))
	}
}

func Benchmark_Binary_PutUintLE_4(b *testing.B) {
	x := make([]byte, 4)
	for i := 0; i < b.N; i++ {
		PutUintLE(x, 4, uint64(i))
	}
}

func Benchmark_Binary_PutUintLE_5(b *testing.B) {
	x := make([]byte, 5)
	for i := 0; i < b.N; i++ {
		PutUintLE(x, 5, uint64(i))
	}
}

func Benchmark_Binary_PutUintLE_6(b *testing.B) {
	x := make([]byte, 6)
	for i := 0; i < b.N; i++ {
		PutUintLE(x, 6, uint64(i))
	}
}

func Benchmark_Binary_PutUintLE_7(b *testing.B) {
	x := make([]byte, 7)
	for i := 0; i < b.N; i++ {
		PutUintLE(x, 7, uint64(i))
	}
}

func Benchmark_Binary_PutUintLE_8(b *testing.B) {
	x := make([]byte, 8)
	for i := 0; i < b.N; i++ {
		PutUintLE(x, 8, uint64(i))
	}
}

func Benchmark_Binary_PutUintBE_2(b *testing.B) {
	x := make([]byte, 2)
	for i := 0; i < b.N; i++ {
		PutUintBE(x, 2, uint64(i))
	}
}

func Benchmark_Binary_PutUintBE_3(b *testing.B) {
	x := make([]byte, 3)
	for i := 0; i < b.N; i++ {
		PutUintBE(x, 3, uint64(i))
	}
}

func Benchmark_Binary_PutUintBE_4(b *testing.B) {
	x := make([]byte, 4)
	for i := 0; i < b.N; i++ {
		PutUintBE(x, 4, uint64(i))
	}
}

func Benchmark_Binary_PutUintBE_5(b *testing.B) {
	x := make([]byte, 5)
	for i := 0; i < b.N; i++ {
		PutUintBE(x, 5, uint64(i))
	}
}

func Benchmark_Binary_PutUintBE_6(b *testing.B) {
	x := make([]byte, 6)
	for i := 0; i < b.N; i++ {
		PutUintBE(x, 6, uint64(i))
	}
}

func Benchmark_Binary_PutUintBE_7(b *testing.B) {
	x := make([]byte, 7)
	for i := 0; i < b.N; i++ {
		PutUintBE(x, 7, uint64(i))
	}
}

func Benchmark_Binary_PutUintBE_8(b *testing.B) {
	x := make([]byte, 8)
	for i := 0; i < b.N; i++ {
		PutUintBE(x, 8, uint64(i))
	}
}

func Benchmark_Binary_PutUint16BE(b *testing.B) {
	x := make([]byte, 2)
	for i := 0; i < b.N; i++ {
		binary.BigEndian.PutUint16(x, uint16(i))
	}
}

func Benchmark_Binary_PutUint24BE(b *testing.B) {
	x := make([]byte, 3)
	for i := 0; i < b.N; i++ {
		PutUint24BE(x, uint32(i))
	}
}

func Benchmark_Binary_PutUint32BE(b *testing.B) {
	x := make([]byte, 4)
	for i := 0; i < b.N; i++ {
		binary.BigEndian.PutUint32(x, uint32(i))
	}
}

func Benchmark_Binary_PutUint40BE(b *testing.B) {
	x := make([]byte, 5)
	for i := 0; i < b.N; i++ {
		PutUint40BE(x, uint64(i))
	}
}

func Benchmark_Binary_PutUint48BE(b *testing.B) {
	x := make([]byte, 6)
	for i := 0; i < b.N; i++ {
		PutUint48BE(x, uint64(i))
	}
}

func Benchmark_Binary_PutUint56BE(b *testing.B) {
	x := make([]byte, 7)
	for i := 0; i < b.N; i++ {
		PutUint56BE(x, uint64(i))
	}
}

func Benchmark_Binary_PutUint64BE(b *testing.B) {
	x := make([]byte, 8)
	for i := 0; i < b.N; i++ {
		binary.BigEndian.PutUint64(x, uint64(i))
	}
}

func Benchmark_Binary_PutUint16LE(b *testing.B) {
	x := make([]byte, 2)
	for i := 0; i < b.N; i++ {
		binary.LittleEndian.PutUint16(x, uint16(i))
	}
}

func Benchmark_Binary_PutUint24LE(b *testing.B) {
	x := make([]byte, 3)
	for i := 0; i < b.N; i++ {
		PutUint24LE(x, uint32(i))
	}
}

func Benchmark_Binary_PutUint32LE(b *testing.B) {
	x := make([]byte, 4)
	for i := 0; i < b.N; i++ {
		binary.LittleEndian.PutUint32(x, uint32(i))
	}
}

func Benchmark_Binary_PutUint40LE(b *testing.B) {
	x := make([]byte, 5)
	for i := 0; i < b.N; i++ {
		PutUint40LE(x, uint64(i))
	}
}

func Benchmark_Binary_PutUint48LE(b *testing.B) {
	x := make([]byte, 6)
	for i := 0; i < b.N; i++ {
		PutUint48LE(x, uint64(i))
	}
}

func Benchmark_Binary_PutUint56LE(b *testing.B) {
	x := make([]byte, 7)
	for i := 0; i < b.N; i++ {
		PutUint56LE(x, uint64(i))
	}
}

func Benchmark_Binary_PutUint64LE(b *testing.B) {
	x := make([]byte, 8)
	for i := 0; i < b.N; i++ {
		binary.LittleEndian.PutUint64(x, uint64(i))
	}
}
