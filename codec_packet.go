package link

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"sync"
)

type ByteOrder binary.ByteOrder

var (
	BigEndian    = binary.BigEndian
	LittleEndian = binary.LittleEndian
)

var (
	ErrUnsupportedPacketType = errors.New("unsupported packet type")
	ErrTooLargePacket        = errors.New("too large packet")
)

func Packet(n, maxPacketSize, readBufferSize int, byteOrder ByteOrder, base CodecType) CodecType {
	if n != 1 && n != 2 && n != 4 && n != 8 {
		panic(ErrUnsupportedPacketType)
	}
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
			writer: w,
			parent: codecType,
		}
		encoder.buffer.data = make([]byte, 1024)
		encoder.buffer.n = codecType.n
		encoder.buffer.max = codecType.maxPacketSize
		switch codecType.n {
		case 1:
			encoder.encodeHead = codecType.encodeHead1
		case 2:
			encoder.encodeHead = codecType.encodeHead2
		case 4:
			encoder.encodeHead = codecType.encodeHead4
		case 8:
			encoder.encodeHead = codecType.encodeHead8
		}
	}
	encoder.base = codecType.base.NewEncoder(&encoder.buffer)
	return encoder
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
	base       Encoder
	buffer     PacketBuffer
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
	encoder.buffer.reset()
	if err = encoder.base.Encode(msg); err != nil {
		return
	}
	b := encoder.buffer.bytes()
	encoder.encodeHead(b)
	_, err = encoder.writer.Write(b)
	return
}

func (decoder *packetDecoder) Decode(msg interface{}) (err error) {
	head := decoder.head[:decoder.n]
	if _, err = io.ReadFull(decoder.reader.R, head); err != nil {
		return
	}
	decoder.reader.N = int64(decoder.decodeHead(head))
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

type PacketBuffer struct {
	data []byte
	n    int
	max  int
	wpos int
}

func (pb *PacketBuffer) bytes() []byte {
	return pb.data[:pb.wpos]
}

func (pb *PacketBuffer) reset() {
	pb.wpos = pb.n
}

func (pb *PacketBuffer) gorws(n int) {
	if newLen := pb.wpos + n; newLen > len(pb.data) {
		newData := make([]byte, newLen, newLen+512)
		copy(newData, pb.data)
		pb.data = newData
	}
}

func (pb *PacketBuffer) Next(n int) (b []byte) {
	pb.gorws(n)
	n += pb.wpos
	if n > pb.max {
		panic(ErrTooLargePacket)
	}
	b = pb.data[pb.wpos:n]
	pb.wpos = n
	return
}

func (pb *PacketBuffer) Write(b []byte) (int, error) {
	n := len(b)
	copy(pb.Next(n), b)
	return n, nil
}
