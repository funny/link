package link

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"sync"
)

var (
	BigEndian    = binary.BigEndian
	LittleEndian = binary.LittleEndian
)

var ErrTooLargePacket = errors.New("too large packet")

type Sizer interface {
	Sizeof(msg interface{}) int
}

func Packet(n, maxPacketSize, readBufferSize int, byteOrder binary.ByteOrder, base CodecType) CodecType {
	return &packetCodecType{
		n:              n,
		base:           base,
		maxPacketSize:  maxPacketSize,
		readBufferSize: readBufferSize,
		byteOrder:      byteOrder,
	}
}

type packetCodecType struct {
	n              int
	maxPacketSize  int
	readBufferSize int
	base           CodecType
	encoderPool    sync.Pool
	decoderPool    sync.Pool
	byteOrder      binary.ByteOrder
}

func (codecType *packetCodecType) NewEncoder(w io.Writer) Encoder {
	encoder, ok := codecType.encoderPool.Get().(*packetEncoder)
	if ok {
		encoder.writer = w
	} else {
		encoder = &packetEncoder{
			n:      codecType.n,
			writer: w,
			parent: codecType,
		}
		encoder.buffer.max = codecType.maxPacketSize
		switch codecType.n {
		case 1:
			encoder.encodeHead = codecType.encodeHead1
			codecType.prepareBuffer1(&encoder.buffer.buf)
		case 2:
			encoder.encodeHead = codecType.encodeHead2
			codecType.prepareBuffer2(&encoder.buffer.buf)
		case 4:
			encoder.encodeHead = codecType.encodeHead4
			codecType.prepareBuffer4(&encoder.buffer.buf)
		case 8:
			encoder.encodeHead = codecType.encodeHead8
			codecType.prepareBuffer8(&encoder.buffer.buf)
		}
	}
	encoder.base = codecType.base.NewEncoder(&encoder.buffer)
	encoder.sizer, _ = encoder.base.(Sizer)
	return encoder
}

func (codecType *packetCodecType) prepareBuffer1(buf *bytes.Buffer) {
	buf.WriteByte(0)
}

func (codecType *packetCodecType) prepareBuffer2(buf *bytes.Buffer) {
	var b [2]byte
	buf.Write(b[:])
}

func (codecType *packetCodecType) prepareBuffer4(buf *bytes.Buffer) {
	var b [4]byte
	buf.Write(b[:])
}

func (codecType *packetCodecType) prepareBuffer8(buf *bytes.Buffer) {
	var b [8]byte
	buf.Write(b[:])
}

func (codecType *packetCodecType) encodeHead1(b []byte) {
	if n := len(b) - 1; n <= 254 && n <= codecType.maxPacketSize {
		b[0] = byte(n)
	} else {
		panic(ErrTooLargePacket)
	}
}

func (codecType *packetCodecType) encodeHead2(b []byte) {
	if n := len(b) - 2; n <= 65534 && n <= codecType.maxPacketSize {
		codecType.byteOrder.PutUint16(b, uint16(n))
	} else {
		panic(ErrTooLargePacket)
	}
}

func (codecType *packetCodecType) encodeHead4(b []byte) {
	if n := len(b) - 4; n <= codecType.maxPacketSize {
		codecType.byteOrder.PutUint32(b, uint32(n))
	} else {
		panic(ErrTooLargePacket)
	}
}

func (codecType *packetCodecType) encodeHead8(b []byte) {
	if n := len(b) - 8; n <= codecType.maxPacketSize {
		codecType.byteOrder.PutUint64(b, uint64(n))
	} else {
		panic(ErrTooLargePacket)
	}
}

func (codecType *packetCodecType) NewDecoder(r io.Reader) Decoder {
	decoder, ok := codecType.decoderPool.Get().(*packetDecoder)
	if ok {
		decoder.reader.R.(*bufio.Reader).Reset(r)
	} else {
		decoder = &packetDecoder{
			n:      codecType.n,
			parent: codecType,
		}
		decoder.reader.R = bufio.NewReaderSize(r, codecType.readBufferSize)
		switch codecType.n {
		case 1:
			decoder.decodeHead = codecType.decodeHead1
		case 2:
			decoder.decodeHead = codecType.decodeHead2
		case 4:
			decoder.decodeHead = codecType.decodeHead4
		case 8:
			decoder.decodeHead = codecType.decodeHead8
		}
	}
	decoder.base = codecType.base.NewDecoder(&decoder.reader)
	return decoder
}

func (codecType *packetCodecType) decodeHead1(b []byte) int {
	if n := int(b[0]); n <= 254 && n <= codecType.maxPacketSize {
		return n
	}
	panic(ErrTooLargePacket)
}

func (codecType *packetCodecType) decodeHead2(b []byte) int {
	if n := int(codecType.byteOrder.Uint16(b)); n > 0 && n <= 65534 && n <= codecType.maxPacketSize {
		return n
	}
	panic(ErrTooLargePacket)
}

func (codecType *packetCodecType) decodeHead4(b []byte) int {
	if n := int(codecType.byteOrder.Uint32(b)); n > 0 && n <= codecType.maxPacketSize {
		return n
	}
	panic(ErrTooLargePacket)
}

func (codecType *packetCodecType) decodeHead8(b []byte) int {
	if n := int(codecType.byteOrder.Uint64(b)); n > 0 && n <= codecType.maxPacketSize {
		return n
	}
	panic(ErrTooLargePacket)
}

type packetEncoder struct {
	n          int
	base       Encoder
	sizer      Sizer
	buffer     limitedBuffer
	writer     io.Writer
	parent     *packetCodecType
	encodeHead func([]byte)
}

type packetDecoder struct {
	n          int
	base       Decoder
	head       [8]byte
	reader     io.LimitedReader
	parent     *packetCodecType
	decodeHead func([]byte) int
}

func (encoder *packetEncoder) Encode(msg interface{}) (err error) {
	encoder.buffer.buf.Truncate(encoder.n)
	if encoder.sizer != nil {
		encoder.buffer.buf.Grow(encoder.sizer.Sizeof(msg))
	}
	encoder.buffer.n = 0
	if err = encoder.base.Encode(msg); err != nil {
		return
	}
	b := encoder.buffer.buf.Bytes()
	encoder.encodeHead(b)
	_, err = encoder.writer.Write(b)
	return
}

func (decoder *packetDecoder) Decode(msg interface{}) (err error) {
	head := decoder.head[:decoder.n]
	if _, err = io.ReadFull(decoder.reader.R, head); err != nil {
		return
	}
	n := decoder.decodeHead(head)
	decoder.reader.N = int64(n)
	if err = decoder.base.Decode(msg); err != nil {
		return
	}
	if decoder.reader.N != 0 {
		decoder.reader.R.(*bufio.Reader).Discard(int(decoder.reader.N))
	}
	return
}

func (encoder *packetEncoder) Dispose() {
	if d, ok := encoder.base.(Disposeable); ok {
		d.Dispose()
	}
	encoder.base = nil
	encoder.parent.encoderPool.Put(encoder)
}

func (decoder *packetDecoder) Dispose() {
	if d, ok := decoder.base.(Disposeable); ok {
		d.Dispose()
	}
	decoder.base = nil
	decoder.parent.decoderPool.Put(decoder)
}

type limitedBuffer struct {
	buf bytes.Buffer
	max int
	n   int
}

func (lb *limitedBuffer) Write(p []byte) (int, error) {
	lb.n += len(p)
	if lb.n > lb.max {
		panic(ErrTooLargePacket)
	}
	return lb.buf.Write(p)
}
