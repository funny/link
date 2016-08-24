package link

import (
	"testing"
)

func Benchmark_BytesToInterface(b *testing.B) {
	var a = []byte{}
	var x interface{}
	for i := 0; i < b.N; i++ {
		x = a
	}
	_ = x
}

func Benchmark_InterfaceToBytes(b *testing.B) {
	var x interface{} = []byte{}
	var a []byte
	for i := 0; i < b.N; i++ {
		a = x.([]byte)
	}
	_ = a
}

func Benchmark_PointerToInterface(b *testing.B) {
	var a = struct{}{}
	var x interface{}
	for i := 0; i < b.N; i++ {
		x = &a
	}
	_ = x
}

func Benchmark_InterfaceToPointer(b *testing.B) {
	var x interface{} = new(struct{})
	var a *struct{}
	for i := 0; i < b.N; i++ {
		a = x.(*struct{})
	}
	_ = a
}
