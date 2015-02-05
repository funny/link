package link

import (
	"encoding/binary"
	"io"
	"math"
	"sync/atomic"
	"unicode/utf8"
	"unsafe"
)

var (
	enableBufferPool = true
	globalPool       = newBufferPool()
)

// Turn On/Off buffer pool. Default is enable.
func BufferPoolEnable(enable bool) {
	enableBufferPool = enable
}

// Limit buffer pool memory usage. Default is 10M.
func BufferPoolLimit(size int) int {
	if size == 0 {
		return int(globalPool.sizeLimit)
	}
	old := globalPool.sizeLimit
	globalPool.sizeLimit = int64(size)
	return int(old)
}

// Get/Set initialization capacity for new buffer. Default is 4096.
func BufferInitSize(size int) int {
	if size == 0 {
		return globalPool.bufferInitSize
	}
	old := globalPool.bufferInitSize
	globalPool.bufferInitSize = size
	return old
}

// Limit buffer size in object pool.
// Large buffer will not return to object pool when it freed. Default is 102400.
func BufferSizeLimit(size int) int {
	if size == 0 {
		return globalPool.bufferSizeLimit
	}
	old := globalPool.bufferSizeLimit
	globalPool.bufferSizeLimit = size
	return old
}

// Buffer pool state.
type PoolState struct {
	InHitRate  float64 // Hit rate of InBuffer.
	InFreeRate float64 // InBuffer free rate.
	InDropRate float64 // Drop rate of large OutBuffer.

	OutHitRate  float64 // Hit rate of OutBuffer.
	OutFreeRate float64 // OutBuffer free rate.
	OutDropRate float64 // Drop rate of large OutBuffer.
}

// Get buffer pool state.
func BufferPoolState() PoolState {
	var (
		inGet  = float64(atomic.LoadUint64(&globalPool.inGet))
		inNew  = float64(atomic.LoadUint64(&globalPool.inNew))
		inFree = float64(atomic.LoadUint64(&globalPool.inFree))
		inDrop = float64(atomic.LoadUint64(&globalPool.inDrop))
	)
	var (
		outGet  = float64(atomic.LoadUint64(&globalPool.outGet))
		outNew  = float64(atomic.LoadUint64(&globalPool.outNew))
		outFree = float64(atomic.LoadUint64(&globalPool.outFree))
		outDrop = float64(atomic.LoadUint64(&globalPool.outDrop))
	)

	return PoolState{
		InHitRate:   (inGet - inNew) / inGet,
		InFreeRate:  inFree / inGet,
		InDropRate:  inDrop / inGet,
		OutHitRate:  (outGet - outNew) / outGet,
		OutFreeRate: outFree / outGet,
		OutDropRate: outDrop / outGet,
	}
}

type bufferPool struct {
	size int64

	// InBuffer
	in     unsafe.Pointer
	inGet  uint64
	inNew  uint64
	inFree uint64
	inDrop uint64

	// OutBuffer
	out     unsafe.Pointer
	outGet  uint64
	outNew  uint64
	outFree uint64
	outDrop uint64

	sizeLimit       int64
	bufferInitSize  int
	bufferSizeLimit int
}

func newBufferPool() *bufferPool {
	return &bufferPool{
		sizeLimit:       10240000,
		bufferInitSize:  1024,
		bufferSizeLimit: 102400,
	}
}

func (pool *bufferPool) GetInBuffer() (in *InBuffer) {
	var ptr unsafe.Pointer
	for {
		ptr = atomic.LoadPointer(&pool.in)
		if ptr == nil {
			break
		}
		if atomic.CompareAndSwapPointer(&pool.in, ptr, ((*InBuffer)(ptr)).next) {
			break
		}
	}

	atomic.AddUint64(&pool.inGet, 1)
	if ptr == nil {
		atomic.AddUint64(&pool.inNew, 1)
		in = &InBuffer{Data: make([]byte, 0, pool.bufferInitSize), pool: pool}
	} else {
		in = (*InBuffer)(ptr)
		atomic.AddInt64(&pool.size, -int64(cap(in.Data)))
	}

	in.isFreed = false
	return in
}

func (pool *bufferPool) GetOutBuffer() (out *OutBuffer) {
	var ptr unsafe.Pointer
	for {
		ptr = atomic.LoadPointer(&pool.out)
		if ptr == nil {
			break
		}
		if atomic.CompareAndSwapPointer(&pool.out, ptr, ((*OutBuffer)(ptr)).next) {
			break
		}
	}

	atomic.AddUint64(&pool.outGet, 1)
	if ptr == nil {
		atomic.AddUint64(&pool.outNew, 1)
		out = &OutBuffer{Data: make([]byte, 0, pool.bufferInitSize), pool: pool}
	} else {
		out = (*OutBuffer)(ptr)
		atomic.AddInt64(&pool.size, -int64(cap(out.Data)))
	}

	out.isFreed = false
	out.isBroadcast = false
	out.refCount = 0
	return out
}

func (pool *bufferPool) PutInBuffer(in *InBuffer) {
	atomic.AddUint64(&pool.inFree, 1)
	if cap(in.Data) >= pool.bufferSizeLimit || atomic.LoadInt64(&pool.size) >= pool.sizeLimit {
		atomic.AddUint64(&pool.inDrop, 1)
		return
	}

	in.isFreed = true

	for {
		in.next = atomic.LoadPointer(&pool.in)
		if atomic.CompareAndSwapPointer(&pool.in, in.next, unsafe.Pointer(in)) {
			atomic.AddInt64(&pool.size, int64(cap(in.Data)))
			break
		}
	}
}

func (pool *bufferPool) PutOutBuffer(out *OutBuffer) {
	atomic.AddUint64(&pool.outFree, 1)
	if cap(out.Data) >= pool.bufferSizeLimit || atomic.LoadInt64(&pool.size) >= pool.sizeLimit {
		atomic.AddUint64(&pool.outDrop, 1)
		return
	}

	out.isFreed = true

	for {
		out.next = atomic.LoadPointer(&pool.out)
		if atomic.CompareAndSwapPointer(&pool.out, out.next, unsafe.Pointer(out)) {
			atomic.AddInt64(&pool.size, int64(cap(out.Data)))
			break
		}
	}
}

// Incomming message buffer.
type InBuffer struct {
	Data    []byte // Buffer data.
	ReadPos int    // Read position.
	isFreed bool
	pool    *bufferPool
	next    unsafe.Pointer
}

func newInBuffer() *InBuffer {
	if enableBufferPool {
		return globalPool.GetInBuffer()
	}
	return &InBuffer{Data: make([]byte, 0, globalPool.bufferInitSize)}
}

func (in *InBuffer) reset() {
	in.Data = in.Data[0:0]
	in.ReadPos = 0
}

// Return the buffer to buffer pool.
func (in *InBuffer) free() {
	if enableBufferPool {
		if in.isFreed {
			panic("link.InBuffer: double free")
		}
		in.reset()
		in.pool.PutInBuffer(in)
	}
}

// Prepare buffer for next message.
// This method is for custom protocol only.
// Dont' use it in application logic.
func (in *InBuffer) Prepare(size int) {
	if cap(in.Data) < size {
		in.Data = make([]byte, size)
	} else {
		in.Data = in.Data[0:size]
	}
}

// Slice some bytes from buffer.
func (in *InBuffer) Slice(n int) []byte {
	r := in.Data[in.ReadPos : in.ReadPos+n]
	in.ReadPos += n
	return r
}

// Implement io.Reader interface
func (in *InBuffer) Read(b []byte) (int, error) {
	if in.ReadPos == len(in.Data) {
		return 0, io.EOF
	}
	n := len(b)
	if n+in.ReadPos > len(in.Data) {
		n = len(in.Data) - in.ReadPos
	}
	copy(b, in.Data[in.ReadPos:])
	in.ReadPos += n
	return n, nil
}

// Read some bytes from buffer.
func (in *InBuffer) ReadBytes(n int) []byte {
	x := make([]byte, n)
	copy(x, in.Slice(n))
	return x
}

// Read a string from buffer.
func (in *InBuffer) ReadString(n int) string {
	return string(in.Slice(n))
}

// Read a rune from buffer.
func (in *InBuffer) ReadRune() rune {
	x, size := utf8.DecodeRune(in.Data[in.ReadPos:])
	in.ReadPos += size
	return x
}

// Read a uint8 value from buffer.
func (in *InBuffer) ReadUint8() uint8 {
	return uint8(in.Slice(1)[0])
}

// Read a uint16 value from buffer using little endian byte order.
func (in *InBuffer) ReadUint16LE() uint16 {
	return binary.LittleEndian.Uint16(in.Slice(2))
}

// Read a uint16 value from buffer using big endian byte order.
func (in *InBuffer) ReadUint16BE() uint16 {
	return binary.BigEndian.Uint16(in.Slice(2))
}

// Read a uint32 value from buffer using little endian byte order.
func (in *InBuffer) ReadUint32LE() uint32 {
	return binary.LittleEndian.Uint32(in.Slice(4))
}

// Read a uint32 value from buffer using big endian byte order.
func (in *InBuffer) ReadUint32BE() uint32 {
	return binary.BigEndian.Uint32(in.Slice(4))
}

// Read a uint64 value from buffer using little endian byte order.
func (in *InBuffer) ReadUint64LE() uint64 {
	return binary.LittleEndian.Uint64(in.Slice(8))
}

// Read a uint64 value from buffer using big endian byte order.
func (in *InBuffer) ReadUint64BE() uint64 {
	return binary.BigEndian.Uint64(in.Slice(8))
}

// Read a float32 value from buffer using little endian byte order.
func (in *InBuffer) ReadFloat32LE() float32 {
	return math.Float32frombits(in.ReadUint32LE())
}

// Read a float32 value from buffer using big endian byte order.
func (in *InBuffer) ReadFloat32BE() float32 {
	return math.Float32frombits(in.ReadUint32BE())
}

// Read a float64 value from buffer using little endian byte order.
func (in *InBuffer) ReadFloat64LE() float64 {
	return math.Float64frombits(in.ReadUint64LE())
}

// Read a float64 value from buffer using big endian byte order.
func (in *InBuffer) ReadFloat64BE() float64 {
	return math.Float64frombits(in.ReadUint64BE())
}

// ReadVarint reads an encoded signed integer from buffer and returns it as an int64.
func (in *InBuffer) ReadVarint() int64 {
	v, n := binary.Varint(in.Data[in.ReadPos:])
	in.ReadPos += n
	return v
}

// ReadUvarint reads an encoded unsigned integer from buffer and returns it as a uint64.
func (in *InBuffer) ReadUvarint() uint64 {
	v, n := binary.Uvarint(in.Data[in.ReadPos:])
	in.ReadPos += n
	return v
}

// Outgoing message buffer.
type OutBuffer struct {
	Data        []byte // Buffer data.
	isFreed     bool
	isBroadcast bool
	refCount    int32
	pool        *bufferPool
	next        unsafe.Pointer
}

func newOutBuffer() *OutBuffer {
	if enableBufferPool {
		return globalPool.GetOutBuffer()
	}
	return &OutBuffer{Data: make([]byte, 0, globalPool.bufferInitSize)}
}

func (out *OutBuffer) broadcastUse() {
	if enableBufferPool {
		atomic.AddInt32(&out.refCount, 1)
	}
}

func (out *OutBuffer) broadcastFree() {
	if enableBufferPool {
		if out.isBroadcast && atomic.AddInt32(&out.refCount, -1) == 0 {
			out.free()
		}
	}
}

func (out *OutBuffer) reset() {
	out.Data = out.Data[0:0]
}

// Return the buffer to buffer pool.
func (out *OutBuffer) free() {
	if enableBufferPool {
		if out.isFreed {
			panic("link.OutBuffer: double free")
		}
		out.reset()
		out.pool.PutOutBuffer(out)
	}
}

// Prepare for next message.
// This method is for custom protocol only.
// Don't use it in application logic.
func (out *OutBuffer) Prepare(size int) {
	if cap(out.Data) < size {
		out.Data = make([]byte, 0, size)
	} else {
		out.Data = out.Data[0:0]
	}
}

// Append some bytes into buffer.
func (out *OutBuffer) Append(p ...byte) {
	out.Data = append(out.Data, p...)
}

// Implement io.Writer interface.
func (out *OutBuffer) Write(p []byte) (int, error) {
	out.Data = append(out.Data, p...)
	return len(p), nil
}

// Write a byte slice into buffer.
func (out *OutBuffer) WriteBytes(d []byte) {
	out.Append(d...)
}

// Write a string into buffer.
func (out *OutBuffer) WriteString(s string) {
	out.Append([]byte(s)...)
}

// Write a rune into buffer.
func (out *OutBuffer) WriteRune(r rune) {
	p := []byte{0, 0, 0, 0}
	n := utf8.EncodeRune(p, r)
	out.Append(p[:n]...)
}

// Write a uint8 value into buffer.
func (out *OutBuffer) WriteUint8(v uint8) {
	out.Append(byte(v))
}

// Write a uint16 value into buffer using little endian byte order.
func (out *OutBuffer) WriteUint16LE(v uint16) {
	out.Append(byte(v), byte(v>>8))
}

// Write a uint16 value into buffer using big endian byte order.
func (out *OutBuffer) WriteUint16BE(v uint16) {
	out.Append(byte(v>>8), byte(v))
}

// Write a uint32 value into buffer using little endian byte order.
func (out *OutBuffer) WriteUint32LE(v uint32) {
	out.Append(byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
}

// Write a uint32 value into buffer using big endian byte order.
func (out *OutBuffer) WriteUint32BE(v uint32) {
	out.Append(byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

// Write a uint64 value into buffer using little endian byte order.
func (out *OutBuffer) WriteUint64LE(v uint64) {
	out.Append(
		byte(v),
		byte(v>>8),
		byte(v>>16),
		byte(v>>24),
		byte(v>>32),
		byte(v>>40),
		byte(v>>48),
		byte(v>>56),
	)
}

// Write a uint64 value into buffer using big endian byte order.
func (out *OutBuffer) WriteUint64BE(v uint64) {
	out.Append(
		byte(v>>56),
		byte(v>>48),
		byte(v>>40),
		byte(v>>32),
		byte(v>>24),
		byte(v>>16),
		byte(v>>8),
		byte(v),
	)
}

// Write a float32 value into buffer using little endian byte order.
func (out *OutBuffer) WriteFloat32LE(v float32) {
	out.WriteUint32LE(math.Float32bits(v))
}

// Write a float32 value into buffer using big endian byte order.
func (out *OutBuffer) WriteFloat32BE(v float32) {
	out.WriteUint32BE(math.Float32bits(v))
}

// Write a float64 value into buffer using little endian byte order.
func (out *OutBuffer) WriteFloat64LE(v float64) {
	out.WriteUint64LE(math.Float64bits(v))
}

// Write a float64 value into buffer using big endian byte order.
func (out *OutBuffer) WriteFloat64BE(v float64) {
	out.WriteUint64BE(math.Float64bits(v))
}

// Write a uint64 value into buffer.
func (out *OutBuffer) WriteUvarint(v uint64) {
	for v >= 0x80 {
		out.Append(byte(v) | 0x80)
		v >>= 7
	}
	out.Append(byte(v))
}

// Write a int64 value into buffer.
func (out *OutBuffer) WriteVarint(v int64) {
	ux := uint64(v) << 1
	if v < 0 {
		ux = ^ux
	}
	out.WriteUvarint(ux)
}
