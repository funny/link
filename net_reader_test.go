package link

import (
	"github.com/funny/unitest"
	"math/rand"
	"sync"
	"testing"
)

func NetTest(t *testing.T, n int, test func(r, w *Conn)) {
	listener, err := Listen("tcp", "0.0.0.0:0")
	unitest.NotError(t, err)
	defer func() { unitest.NotError(t, listener.Close()) }()

	var r *Conn
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, err := listener.Accept()
		unitest.NotError(t, err)
		r = conn
	}()

	addr := listener.l.Addr().String()
	w, err := Dial("tcp", addr)
	unitest.NotError(t, err)
	defer func() { unitest.NotError(t, w.Close()) }()

	wg.Wait()
	unitest.Pass(t, r != nil)
	defer func() { unitest.NotError(t, r.Close()) }()

	for i := 0; i < n; i++ {
		test(r, w)
	}
}

func Test_Uint8(t *testing.T) {
	NetTest(t, 10000, func(r, w *Conn) {
		v1 := uint8(rand.Intn(256))
		w.WriteUint8(v1)
		w.Flush()
		unitest.NotError(t, w.werr)

		v2 := r.ReadUint8()
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, v1 == v2)
	})
}

func Test_Uint16BE(t *testing.T) {
	NetTest(t, 10000, func(r, w *Conn) {
		v1 := uint16(rand.Intn(0xFFFF))
		w.WriteUint16BE(v1)
		w.Flush()
		unitest.NotError(t, w.werr)

		v2 := r.ReadUint16BE()
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, v1 == v2)
	})
}

func Test_Uint16LE(t *testing.T) {
	NetTest(t, 10000, func(r, w *Conn) {
		v1 := uint16(rand.Intn(0xFFFF))
		w.WriteUint16LE(v1)
		w.Flush()
		unitest.NotError(t, w.werr)

		v2 := r.ReadUint16LE()
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, v1 == v2)
	})
}

func Test_Uint24BE(t *testing.T) {
	NetTest(t, 10000, func(r, w *Conn) {
		v1 := uint32(rand.Intn(0xFFFFFF))
		w.WriteUint24BE(v1)
		w.Flush()
		unitest.NotError(t, w.werr)

		v2 := r.ReadUint24BE()
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, v1 == v2)
	})
}

func Test_Uint24LE(t *testing.T) {
	NetTest(t, 10000, func(r, w *Conn) {
		v1 := uint32(rand.Intn(0xFFFFFF))
		w.WriteUint24LE(v1)
		w.Flush()
		unitest.NotError(t, w.werr)

		v2 := r.ReadUint24LE()
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, v1 == v2)
	})
}

func Test_Uint32BE(t *testing.T) {
	NetTest(t, 10000, func(r, w *Conn) {
		v1 := uint32(rand.Intn(0xFFFFFFFF))
		w.WriteUint32BE(v1)
		w.Flush()
		unitest.NotError(t, w.werr)

		v2 := r.ReadUint32BE()
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, v1 == v2)
	})
}

func Test_Uint32LE(t *testing.T) {
	NetTest(t, 10000, func(r, w *Conn) {
		v1 := uint32(rand.Intn(0xFFFFFFFF))
		w.WriteUint32LE(v1)
		unitest.NotError(t, w.werr)
		w.Flush()
		unitest.NotError(t, w.werr)

		v2 := r.ReadUint32LE()
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, v1 == v2)
	})
}

func Test_Uint40BE(t *testing.T) {
	NetTest(t, 10000, func(r, w *Conn) {
		v1 := uint64(rand.Intn(0xFFFFFFFFFF))
		w.WriteUint64BE(v1)
		w.Flush()
		unitest.NotError(t, w.werr)

		v2 := r.ReadUint64BE()
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, v1 == v2)
	})
}

func Test_Uint40LE(t *testing.T) {
	NetTest(t, 10000, func(r, w *Conn) {
		v1 := uint64(rand.Intn(0xFFFFFFFFFF))
		w.WriteUint40LE(v1)
		w.Flush()
		unitest.NotError(t, w.werr)

		v2 := r.ReadUint40LE()
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, v1 == v2)
	})
}

func Test_Uint48BE(t *testing.T) {
	NetTest(t, 10000, func(r, w *Conn) {
		v1 := uint64(rand.Intn(0xFFFFFFFFFFFF))
		w.WriteUint48BE(v1)
		w.Flush()
		unitest.NotError(t, w.werr)

		v2 := r.ReadUint48BE()
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, v1 == v2)
	})
}

func Test_Uint48LE(t *testing.T) {
	NetTest(t, 10000, func(r, w *Conn) {
		v1 := uint64(rand.Intn(0xFFFFFFFFFFFF))
		w.WriteUint48LE(v1)
		w.Flush()
		unitest.NotError(t, w.werr)

		v2 := r.ReadUint48LE()
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, v1 == v2)
	})
}

func Test_Uint56BE(t *testing.T) {
	NetTest(t, 10000, func(r, w *Conn) {
		v1 := uint64(rand.Intn(0xFFFFFFFFFFFFFF))
		w.WriteUint56BE(v1)
		w.Flush()
		unitest.NotError(t, w.werr)

		v2 := r.ReadUint56BE()
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, v1 == v2)
	})
}

func Test_Uint56LE(t *testing.T) {
	NetTest(t, 10000, func(r, w *Conn) {
		v1 := uint64(rand.Intn(0xFFFFFFFFFFFFFF))
		w.WriteUint56LE(v1)
		w.Flush()
		unitest.NotError(t, w.werr)

		v2 := r.ReadUint56LE()
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, v1 == v2)
	})
}

func Test_Uint64BE(t *testing.T) {
	NetTest(t, 10000, func(r, w *Conn) {
		v1 := uint64(rand.Int63n(0x7FFFFFFFFFFFFFFF))
		w.WriteUint64BE(v1)
		w.Flush()
		unitest.NotError(t, w.werr)

		v2 := r.ReadUint64BE()
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, v1 == v2)
	})
}

func Test_Uint64LE(t *testing.T) {
	NetTest(t, 10000, func(r, w *Conn) {
		v1 := uint64(rand.Int63n(0x7FFFFFFFFFFFFFFF))
		w.WriteUint64LE(v1)
		w.Flush()
		unitest.NotError(t, w.werr)

		v2 := r.ReadUint64LE()
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, v1 == v2)
	})
}

func Test_Uvarint(t *testing.T) {
	NetTest(t, 10000, func(r, w *Conn) {
		v1 := uint64(rand.Int63n(0x7FFFFFFFFFFFFFFF))
		w.WriteUvarint(v1)
		w.Flush()
		unitest.NotError(t, w.werr)

		v2 := r.ReadUvarint()
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, v1 == v2)
	})
}

func Test_Varint(t *testing.T) {
	NetTest(t, 10000, func(r, w *Conn) {
		v1 := int64(rand.Int63n(0x7FFFFFFFFFFFFFFF))
		w.WriteVarint(v1)
		w.Flush()
		unitest.NotError(t, w.werr)

		v2 := r.ReadVarint()
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, v1 == v2)
	})
}

func Test_Float32BE(t *testing.T) {
	NetTest(t, 10000, func(r, w *Conn) {
		v1 := float32(rand.NormFloat64())
		w.WriteFloat32BE(v1)
		w.Flush()
		unitest.NotError(t, w.werr)

		v2 := r.ReadFloat32BE()
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, v1 == v2)
	})
}

func Test_Float32LE(t *testing.T) {
	NetTest(t, 10000, func(r, w *Conn) {
		v1 := float32(rand.NormFloat64())
		w.WriteFloat32LE(v1)
		w.Flush()
		unitest.NotError(t, w.werr)

		v2 := r.ReadFloat32LE()
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, v1 == v2)
	})
}

func Test_Float64BE(t *testing.T) {
	NetTest(t, 10000, func(r, w *Conn) {
		v1 := rand.NormFloat64()
		w.WriteFloat64BE(v1)
		w.Flush()
		unitest.NotError(t, w.werr)

		v2 := r.ReadFloat64BE()
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, v1 == v2)
	})
}

func Test_Float64LE(t *testing.T) {
	NetTest(t, 10000, func(r, w *Conn) {
		v1 := rand.NormFloat64()
		w.WriteFloat64LE(v1)
		w.Flush()
		unitest.NotError(t, w.werr)

		v2 := r.ReadFloat64LE()
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, v1 == v2)
	})
}
