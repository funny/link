package link

import (
	"encoding/binary"
	"github.com/funny/link/linkutil"
	"io"
	"math"
	"unicode/utf8"
)

type Request struct {
	*mem
	ReadPos int
	pool    *MemPool
}

func MakeRequest(size, capacity int) *Request {
	return NewRequest(make([]byte, size, capacity))
}

func NewRequest(data []byte) *Request {
	return &Request{mem: &mem{Data: data}}
}

func NewPoolRequest(size, capacity int, pool *MemPool) *Request {
	if pool == nil {
		return MakeRequest(size, capacity)
	}
	return &Request{
		mem:  pool.Alloc(size, capacity),
		pool: pool,
	}
}

func (b *Request) free() {
	if b.pool != nil {
		b.pool.Free(b.mem)
	}
}

// io.Reader
func (b *Request) Read(p []byte) (int, error) {
	n, err := b.ReadAt(p, int64(b.ReadPos))
	b.ReadPos += n
	return n, err
}

// io.ByteReader
func (b *Request) ReadByte() (byte, error) {
	if b.ReadPos == len(b.Data) {
		return 0, io.EOF
	}
	return byte(b.ReadUint8()), nil
}

// io.ReaderAt
func (b *Request) ReadAt(p []byte, off int64) (int, error) {
	if int(off) >= len(b.Data) {
		return 0, io.EOF
	}
	n := len(p)
	if n+int(off) > len(b.Data) {
		n = len(b.Data) - int(off)
	}
	copy(p, b.Data[off:])
	return n, nil
}

// io.RuneReader
func (b *Request) ReadRune() (rune, int, error) {
	r, n := utf8.DecodeRune(b.Data[b.ReadPos:])
	b.ReadPos += n
	return r, n, nil
}

func (b *Request) ReadString(n int) string {
	r := string(b.Data[b.ReadPos : b.ReadPos+n])
	b.ReadPos += n
	return r
}

func (b *Request) ReadBytes(n int) []byte {
	r := make([]byte, n)
	copy(r, b.Data[b.ReadPos:b.ReadPos+n])
	b.ReadPos += n
	return r
}

func (b *Request) ReadVarint() int64 {
	r, n := binary.Varint(b.Data[b.ReadPos:])
	b.ReadPos += n
	return r
}

func (b *Request) ReadUvarint() uint64 {
	r, n := binary.Uvarint(b.Data[b.ReadPos:])
	b.ReadPos += n
	return r
}

func (b *Request) ReadUint8() uint8 {
	r := b.Data[b.ReadPos]
	b.ReadPos += 1
	return r
}

func (b *Request) ReadUint16BE() uint16 {
	r := binary.BigEndian.Uint16(b.Data[b.ReadPos:])
	b.ReadPos += 2
	return r
}

func (b *Request) ReadUint16LE() uint16 {
	r := binary.LittleEndian.Uint16(b.Data[b.ReadPos:])
	b.ReadPos += 2
	return r
}

func (b *Request) ReadUint24BE() uint32 {
	r := linkutil.GetUint24BE(b.Data[b.ReadPos:])
	b.ReadPos += 3
	return r
}

func (b *Request) ReadUint24LE() uint32 {
	r := linkutil.GetUint24LE(b.Data[b.ReadPos:])
	b.ReadPos += 3
	return r
}

func (b *Request) ReadUint32BE() uint32 {
	r := binary.BigEndian.Uint32(b.Data[b.ReadPos:])
	b.ReadPos += 4
	return r
}

func (b *Request) ReadUint32LE() uint32 {
	r := binary.LittleEndian.Uint32(b.Data[b.ReadPos:])
	b.ReadPos += 4
	return r
}

func (b *Request) ReadUint40BE() uint64 {
	r := linkutil.GetUint40BE(b.Data[b.ReadPos:])
	b.ReadPos += 5
	return r
}

func (b *Request) ReadUint40LE() uint64 {
	r := linkutil.GetUint40LE(b.Data[b.ReadPos:])
	b.ReadPos += 5
	return r
}

func (b *Request) ReadUint48BE() uint64 {
	r := linkutil.GetUint48BE(b.Data[b.ReadPos:])
	b.ReadPos += 6
	return r
}

func (b *Request) ReadUint48LE() uint64 {
	r := linkutil.GetUint48LE(b.Data[b.ReadPos:])
	b.ReadPos += 6
	return r
}

func (b *Request) ReadUint56BE() uint64 {
	r := linkutil.GetUint56BE(b.Data[b.ReadPos:])
	b.ReadPos += 7
	return r
}

func (b *Request) ReadUint56LE() uint64 {
	r := linkutil.GetUint56LE(b.Data[b.ReadPos:])
	b.ReadPos += 7
	return r
}

func (b *Request) ReadUint64BE() uint64 {
	r := binary.BigEndian.Uint64(b.Data[b.ReadPos:])
	b.ReadPos += 8
	return r
}

func (b *Request) ReadUint64LE() uint64 {
	r := binary.LittleEndian.Uint64(b.Data[b.ReadPos:])
	b.ReadPos += 8
	return r
}

func (b *Request) ReadFloat32BE() float32 {
	return math.Float32frombits(b.ReadUint32BE())
}

func (b *Request) ReadFloat32LE() float32 {
	return math.Float32frombits(b.ReadUint32LE())
}

func (b *Request) ReadFloat64BE() float64 {
	return math.Float64frombits(b.ReadUint64BE())
}

func (b *Request) ReadFloat64LE() float64 {
	return math.Float64frombits(b.ReadUint64LE())
}
