package link

import (
	"encoding/binary"
	"github.com/funny/rush/libnet/netutil"
	"io"
	"math"
	"unicode/utf8"
)

type Buffer struct {
	*mem
	ReadPos int
	pool    *MemPool
}

func MakeBuffer(size, capacity int) *Buffer {
	return NewBuffer(make([]byte, size, capacity))
}

func NewBuffer(data []byte) *Buffer {
	return &Buffer{mem: &mem{Data: data}}
}

func NewPoolBuffer(size, capacity int, pool *MemPool) *Buffer {
	if pool == nil {
		return MakeBuffer(size, capacity)
	}
	return &Buffer{
		mem:  pool.Alloc(size, capacity),
		pool: pool,
	}
}

func (b *Buffer) Reset(size, capacity int) {
	b.ReadPos = 0
	if capacity > cap(b.Data) {
		b.grows(capacity - len(b.Data))
	}
	b.Data = b.Data[:size]
}

func (b *Buffer) free() {
	if b.pool != nil {
		b.pool.Free(b.mem)
	}
}

func (b *Buffer) grows(n int) (i int) {
	i = len(b.Data)

	newLen := len(b.Data) + n
	if cap(b.Data) >= newLen {
		b.Data = b.Data[:newLen]
		return
	}

	if b.pool != nil && newLen <= b.pool.max {
		mem := b.pool.Alloc(newLen, newLen)
		copy(mem.Data, b.Data)
		b.pool.Free(b.mem)
		b.mem = mem
	}

	data := make([]byte, newLen, cap(b.Data)/4+newLen)
	copy(data, b.Data)
	b.Data = data

	return
}

func (b *Buffer) Length() int {
	return len(b.Data)
}

// io.Writer
func (b *Buffer) Write(p []byte) (int, error) {
	b.WriteBytes(p)
	return len(p), nil
}

// io.ByteWriter
func (b *Buffer) WriteByte(c byte) error {
	b.WriteUint8(uint8(c))
	return nil
}

func (b *Buffer) WriteRune(r rune) {
	i := b.grows(utf8.UTFMax)
	n := utf8.UTFMax - utf8.EncodeRune(b.Data[i:], r)
	b.Data = b.Data[:len(b.Data)-n]
}

func (b *Buffer) WriteString(s string) {
	b.WriteBytes([]byte(s))
}

func (b *Buffer) WriteBytes(p []byte) {
	i := b.grows(len(p))
	copy(b.Data[i:], p)
}

func (b *Buffer) WriteVarint(v int64) {
	i := b.grows(binary.MaxVarintLen64)
	n := binary.MaxVarintLen64 - binary.PutVarint(b.Data[i:], v)
	b.Data = b.Data[:len(b.Data)-n]
}

func (b *Buffer) WriteUvarint(v uint64) {
	i := b.grows(binary.MaxVarintLen64)
	n := binary.MaxVarintLen64 - binary.PutUvarint(b.Data[i:], v)
	b.Data = b.Data[:len(b.Data)-n]
}

func (b *Buffer) WriteUint8(v uint8) {
	i := b.grows(1)
	b.Data[i] = byte(v)
}

func (b *Buffer) WriteUint16BE(v uint16) {
	i := b.grows(2)
	binary.BigEndian.PutUint16(b.Data[i:], v)
}

func (b *Buffer) WriteUint16LE(v uint16) {
	i := b.grows(2)
	binary.LittleEndian.PutUint16(b.Data[i:], v)
}

func (b *Buffer) WriteUint24BE(v uint32) {
	i := b.grows(3)
	netutil.PutUint24BE(b.Data[i:], v)
}

func (b *Buffer) WriteUint24LE(v uint32) {
	i := b.grows(3)
	netutil.PutUint24LE(b.Data[i:], v)
}

func (b *Buffer) WriteUint32BE(v uint32) {
	i := b.grows(4)
	binary.BigEndian.PutUint32(b.Data[i:], v)
}

func (b *Buffer) WriteUint32LE(v uint32) {
	i := b.grows(4)
	binary.LittleEndian.PutUint32(b.Data[i:], v)
}

func (b *Buffer) WriteUint40BE(v uint64) {
	i := b.grows(5)
	netutil.PutUint40BE(b.Data[i:], v)
}

func (b *Buffer) WriteUint40LE(v uint64) {
	i := b.grows(5)
	netutil.PutUint40LE(b.Data[i:], v)
}

func (b *Buffer) WriteUint48BE(v uint64) {
	i := b.grows(6)
	netutil.PutUint48BE(b.Data[i:], v)
}

func (b *Buffer) WriteUint48LE(v uint64) {
	i := b.grows(6)
	netutil.PutUint48LE(b.Data[i:], v)
}

func (b *Buffer) WriteUint56BE(v uint64) {
	i := b.grows(7)
	netutil.PutUint56BE(b.Data[i:], v)
}

func (b *Buffer) WriteUint56LE(v uint64) {
	i := b.grows(7)
	netutil.PutUint56LE(b.Data[i:], v)
}

func (b *Buffer) WriteUint64BE(v uint64) {
	i := b.grows(8)
	binary.BigEndian.PutUint64(b.Data[i:], v)
}

func (b *Buffer) WriteUint64LE(v uint64) {
	i := b.grows(8)
	binary.LittleEndian.PutUint64(b.Data[i:], v)
}

func (b *Buffer) WriteFloat32BE(v float32) {
	b.WriteUint32BE(math.Float32bits(v))
}

func (b *Buffer) WriteFloat32LE(v float32) {
	b.WriteUint32LE(math.Float32bits(v))
}

func (b *Buffer) WriteFloat64BE(v float64) {
	b.WriteUint64BE(math.Float64bits(v))
}

func (b *Buffer) WriteFloat64LE(v float64) {
	b.WriteUint64LE(math.Float64bits(v))
}

// io.Reader
func (b *Buffer) Read(p []byte) (int, error) {
	n, err := b.ReadAt(p, int64(b.ReadPos))
	b.ReadPos += n
	return n, err
}

// io.ByteReader
func (b *Buffer) ReadByte() (byte, error) {
	if b.ReadPos == len(b.Data) {
		return 0, io.EOF
	}
	return byte(b.ReadUint8()), nil
}

// io.ReaderAt
func (b *Buffer) ReadAt(p []byte, off int64) (int, error) {
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
func (b *Buffer) ReadRune() (rune, int, error) {
	r, n := utf8.DecodeRune(b.Data[b.ReadPos:])
	b.ReadPos += n
	return r, n, nil
}

func (b *Buffer) ReadString(n int) string {
	r := string(b.Data[b.ReadPos : b.ReadPos+n])
	b.ReadPos += n
	return r
}

func (b *Buffer) ReadBytes(n int) []byte {
	r := make([]byte, n)
	copy(r, b.Data[b.ReadPos:b.ReadPos+n])
	b.ReadPos += n
	return r
}

func (b *Buffer) ReadVarint() int64 {
	r, n := binary.Varint(b.Data[b.ReadPos:])
	b.ReadPos += n
	return r
}

func (b *Buffer) ReadUvarint() uint64 {
	r, n := binary.Uvarint(b.Data[b.ReadPos:])
	b.ReadPos += n
	return r
}

func (b *Buffer) ReadUint8() uint8 {
	r := b.Data[b.ReadPos]
	b.ReadPos += 1
	return r
}

func (b *Buffer) ReadUint16BE() uint16 {
	r := binary.BigEndian.Uint16(b.Data[b.ReadPos:])
	b.ReadPos += 2
	return r
}

func (b *Buffer) ReadUint16LE() uint16 {
	r := binary.LittleEndian.Uint16(b.Data[b.ReadPos:])
	b.ReadPos += 2
	return r
}

func (b *Buffer) ReadUint24BE() uint32 {
	r := netutil.GetUint24BE(b.Data[b.ReadPos:])
	b.ReadPos += 3
	return r
}

func (b *Buffer) ReadUint24LE() uint32 {
	r := netutil.GetUint24LE(b.Data[b.ReadPos:])
	b.ReadPos += 3
	return r
}

func (b *Buffer) ReadUint32BE() uint32 {
	r := binary.BigEndian.Uint32(b.Data[b.ReadPos:])
	b.ReadPos += 4
	return r
}

func (b *Buffer) ReadUint32LE() uint32 {
	r := binary.LittleEndian.Uint32(b.Data[b.ReadPos:])
	b.ReadPos += 4
	return r
}

func (b *Buffer) ReadUint40BE() uint64 {
	r := netutil.GetUint40BE(b.Data[b.ReadPos:])
	b.ReadPos += 5
	return r
}

func (b *Buffer) ReadUint40LE() uint64 {
	r := netutil.GetUint40LE(b.Data[b.ReadPos:])
	b.ReadPos += 5
	return r
}

func (b *Buffer) ReadUint48BE() uint64 {
	r := netutil.GetUint48BE(b.Data[b.ReadPos:])
	b.ReadPos += 6
	return r
}

func (b *Buffer) ReadUint48LE() uint64 {
	r := netutil.GetUint48LE(b.Data[b.ReadPos:])
	b.ReadPos += 6
	return r
}

func (b *Buffer) ReadUint56BE() uint64 {
	r := netutil.GetUint56BE(b.Data[b.ReadPos:])
	b.ReadPos += 7
	return r
}

func (b *Buffer) ReadUint56LE() uint64 {
	r := netutil.GetUint56LE(b.Data[b.ReadPos:])
	b.ReadPos += 7
	return r
}

func (b *Buffer) ReadUint64BE() uint64 {
	r := binary.BigEndian.Uint64(b.Data[b.ReadPos:])
	b.ReadPos += 8
	return r
}

func (b *Buffer) ReadUint64LE() uint64 {
	r := binary.LittleEndian.Uint64(b.Data[b.ReadPos:])
	b.ReadPos += 8
	return r
}

func (b *Buffer) ReadFloat32BE() float32 {
	return math.Float32frombits(b.ReadUint32BE())
}

func (b *Buffer) ReadFloat32LE() float32 {
	return math.Float32frombits(b.ReadUint32LE())
}

func (b *Buffer) ReadFloat64BE() float64 {
	return math.Float64frombits(b.ReadUint64BE())
}

func (b *Buffer) ReadFloat64LE() float64 {
	return math.Float64frombits(b.ReadUint64LE())
}
