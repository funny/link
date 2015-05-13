package link

import (
	"github.com/funny/unitest"
	"math/rand"
	"runtime"
	"testing"
	"time"
)

func Test_MemPool(t *testing.T) {
	MemPoolTest(t, 1, 1, 10)
	MemPoolTest(t, 1, 2, 16)
}

func MemPoolTest(t *testing.T, total, min, max int) {
	pool := NewMemPool(total, min, max)

	for i := min; i <= max; i++ {
		min1 := (i-1)*1024 + 1
		max1 := i * 1024

		b1 := pool.Alloc(min1, min1)
		b2 := pool.Alloc(max1, max1)

		unitest.Pass(t, cap(b1.Data) == max1)
		unitest.Pass(t, cap(b2.Data) == max1)

		pool.Free(b1)
		pool.Free(b2)
	}

	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 100000; i++ {
		size := (rand.Intn(max-min+1) + min) * 1024
		b := pool.Alloc(size, size)
		pool.Free(b)
	}

	for _, class := range pool.classes {
		unitest.Pass(t, class.length == class.maxlen)
	}
}

func Benchmark_MemPool(b *testing.B) {
	pool := NewMemPool(20, 2, 32)
	size := (rand.Intn(30) + 2) * 1024
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		x := pool.Alloc(0, size)
		pool.Free(x)
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

func Benchmark_MakeBytes_8192(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = make([]byte, 8192)
	}
}

func Benchmark_BufferedChan_Put(b *testing.B) {
	c := make(chan int, b.N)
	for i := 0; i < b.N; i++ {
		c <- 1
	}
}

func Benchmark_BufferedChan_Get(b *testing.B) {
	c := make(chan int, b.N)
	for i := 0; i < b.N; i++ {
		c <- 1
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		<-c
	}
}

func Benchmark_BufferedChan_SelectGet(b *testing.B) {
	c := make(chan int, b.N)
	for i := 0; i < b.N; i++ {
		c <- 1
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		select {
		case <-c:
		default:
		}
	}
}

func Benchmark_SetFinalizer1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var x = &Buffer{}
		runtime.SetFinalizer(x, func(x *Buffer) {
		})
	}
}

func Benchmark_SetFinalizer2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var x = &Buffer{}
		runtime.SetFinalizer(x, nil)
	}
}
