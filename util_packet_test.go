package link

import (
	"bytes"
	"encoding/base64"
	"github.com/funny/unitest"
	"math/rand"
	"testing"
)

func RandBytes(n int) []byte {
	n = rand.Intn(n)
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		b[i] = byte(rand.Intn(255))
	}
	return b
}

func Test_Delim_Spliter(t *testing.T) {
	ConnTest(t, 1000, func(r, w *Conn) {
		b1 := RandBytes(1024)
		b2 := make([]byte, base64.StdEncoding.EncodedLen(len(b1)))
		base64.StdEncoding.Encode(b2, b1)

		w.WritePacket(b2, SplitByLine)
		w.Flush()
		unitest.NotError(t, w.werr)

		b3 := r.ReadPacket(SplitByLine)
		unitest.NotError(t, r.rerr)

		b4 := make([]byte, base64.StdEncoding.DecodedLen(len(b3)))
		n, err := base64.StdEncoding.Decode(b4, b3)
		unitest.NotError(t, err)
		unitest.Pass(t, bytes.Equal(b1, b4[:n]))
	})
}

func Test_Uvarint_Spliter(t *testing.T) {
	ConnTest(t, 10000, func(r, w *Conn) {
		b1 := RandBytes(1024)
		w.WritePacket(b1, SplitByUvarint)
		w.Flush()
		unitest.NotError(t, w.werr)

		b2 := r.ReadPacket(SplitByUvarint)
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, bytes.Equal(b1, b2))
	})
}

func Test_Uint8_Spliter(t *testing.T) {
	ConnTest(t, 10000, func(r, w *Conn) {
		b1 := RandBytes(256)
		w.WritePacket(b1, SplitByUint8)
		w.Flush()
		unitest.NotError(t, w.werr)

		b2 := r.ReadPacket(SplitByUint8)
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, bytes.Equal(b1, b2))
	})
}

func Test_Uint16BE_Spliter(t *testing.T) {
	ConnTest(t, 10000, func(r, w *Conn) {
		b1 := RandBytes(1024)
		w.WritePacket(b1, SplitByUint16BE)
		w.Flush()
		unitest.NotError(t, w.werr)

		b2 := r.ReadPacket(SplitByUint16BE)
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, bytes.Equal(b1, b2))
	})
}

func Test_Uint16LE_Spliter(t *testing.T) {
	ConnTest(t, 10000, func(r, w *Conn) {
		b1 := RandBytes(1024)
		w.WritePacket(b1, SplitByUint16LE)
		w.Flush()
		unitest.NotError(t, w.werr)

		b2 := r.ReadPacket(SplitByUint16LE)
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, bytes.Equal(b1, b2))
	})
}

func Test_Uint24BE_Spliter(t *testing.T) {
	ConnTest(t, 10000, func(r, w *Conn) {
		b1 := RandBytes(1024)
		w.WritePacket(b1, SplitByUint24BE)
		w.Flush()
		unitest.NotError(t, w.werr)

		b2 := r.ReadPacket(SplitByUint24BE)
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, bytes.Equal(b1, b2))
	})
}

func Test_Uint24LE_Spliter(t *testing.T) {
	ConnTest(t, 10000, func(r, w *Conn) {
		b1 := RandBytes(1024)
		w.WritePacket(b1, SplitByUint24LE)
		w.Flush()
		unitest.NotError(t, w.werr)

		b2 := r.ReadPacket(SplitByUint24LE)
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, bytes.Equal(b1, b2))
	})
}

func Test_Uint32BE_Spliter(t *testing.T) {
	ConnTest(t, 10000, func(r, w *Conn) {
		b1 := RandBytes(1024)
		w.WritePacket(b1, SplitByUint32BE)
		w.Flush()
		unitest.NotError(t, w.werr)

		b2 := r.ReadPacket(SplitByUint32BE)
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, bytes.Equal(b1, b2))
	})
}

func Test_Uint32LE_Spliter(t *testing.T) {
	ConnTest(t, 10000, func(r, w *Conn) {
		b1 := RandBytes(1024)
		w.WritePacket(b1, SplitByUint32LE)
		unitest.NotError(t, w.werr)
		w.Flush()
		unitest.NotError(t, w.werr)

		b2 := r.ReadPacket(SplitByUint32LE)
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, bytes.Equal(b1, b2))
	})
}

func Test_Uint40BE_Spliter(t *testing.T) {
	ConnTest(t, 10000, func(r, w *Conn) {
		b1 := RandBytes(1024)
		w.WritePacket(b1, SplitByUint40BE)
		w.Flush()
		unitest.NotError(t, w.werr)

		b2 := r.ReadPacket(SplitByUint40BE)
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, bytes.Equal(b1, b2))
	})
}

func Test_Uint40LE_Spliter(t *testing.T) {
	ConnTest(t, 10000, func(r, w *Conn) {
		b1 := RandBytes(1024)
		w.WritePacket(b1, SplitByUint40LE)
		w.Flush()
		unitest.NotError(t, w.werr)

		b2 := r.ReadPacket(SplitByUint40LE)
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, bytes.Equal(b1, b2))
	})
}

func Test_Uint48BE_Spliter(t *testing.T) {
	ConnTest(t, 10000, func(r, w *Conn) {
		b1 := RandBytes(1024)
		w.WritePacket(b1, SplitByUint48BE)
		w.Flush()
		unitest.NotError(t, w.werr)

		b2 := r.ReadPacket(SplitByUint48BE)
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, bytes.Equal(b1, b2))
	})
}

func Test_Uint48LE_Spliter(t *testing.T) {
	ConnTest(t, 10000, func(r, w *Conn) {
		b1 := RandBytes(1024)
		w.WritePacket(b1, SplitByUint48LE)
		w.Flush()
		unitest.NotError(t, w.werr)

		b2 := r.ReadPacket(SplitByUint48LE)
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, bytes.Equal(b1, b2))
	})
}

func Test_Uint56BE_Spliter(t *testing.T) {
	ConnTest(t, 10000, func(r, w *Conn) {
		b1 := RandBytes(1024)
		w.WritePacket(b1, SplitByUint56BE)
		w.Flush()
		unitest.NotError(t, w.werr)

		b2 := r.ReadPacket(SplitByUint56BE)
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, bytes.Equal(b1, b2))
	})
}

func Test_Uint56LE_Spliter(t *testing.T) {
	ConnTest(t, 10000, func(r, w *Conn) {
		b1 := RandBytes(1024)
		w.WritePacket(b1, SplitByUint56LE)
		unitest.NotError(t, w.werr)
		w.Flush()
		unitest.NotError(t, w.werr)

		b2 := r.ReadPacket(SplitByUint56LE)
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, bytes.Equal(b1, b2))
	})
}

func Test_Uint64BE_Spliter(t *testing.T) {
	ConnTest(t, 10000, func(r, w *Conn) {
		b1 := RandBytes(1024)
		w.WritePacket(b1, SplitByUint64BE)
		w.Flush()
		unitest.NotError(t, w.werr)

		b2 := r.ReadPacket(SplitByUint64BE)
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, bytes.Equal(b1, b2))
	})
}

func Test_Uint64LE_Spliter(t *testing.T) {
	ConnTest(t, 10000, func(r, w *Conn) {
		b1 := RandBytes(1024)
		w.WritePacket(b1, SplitByUint64LE)
		w.Flush()
		unitest.NotError(t, w.werr)

		b2 := r.ReadPacket(SplitByUint64LE)
		unitest.NotError(t, r.rerr)
		unitest.Pass(t, bytes.Equal(b1, b2))
	})
}
