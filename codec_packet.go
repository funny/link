package link

import (
	"bufio"
	gobinary "encoding/binary"
	"github.com/funny/binary"
	"io"
	"sync"
)

var (
	BigEndian    = gobinary.BigEndian
	LittleEndian = gobinary.LittleEndian
)

func Packet(n, maxPacketSize, readBufferSize int, byteOrder gobinary.ByteOrder, base CodecType) CodecType {
	return &packetCodecType{
		n:              n,
		base:           base,
		maxPacketSize:  maxPacketSize,
		readBufferSize: readBufferSize,
		byteOrder:      byteOrder,
	}
}

func PacketPro(n, maxPacketSize, readBufferSize int, byteOrder gobinary.ByteOrder, bufferPool *binary.BufferPool, base CodecType) CodecType {
	return &packetCodecType{
		n:              n,
		base:           base,
		bufferPool:     bufferPool,
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
	bufferPool     *binary.BufferPool
	encoderPool    sync.Pool
	decoderPool    sync.Pool
	byteOrder      gobinary.ByteOrder
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
		if codecType.bufferPool != nil {
			codecType.bufferPool.Manage(&encoder.buffer)
		} else {
			encoder.buffer.Data = make([]byte, codecType.n+codecType.readBufferSize)
		}
		switch codecType.n {
		case 1:
			encoder.writeHead = codecType.writeHead1
		case 2:
			encoder.writeHead = codecType.writeHead2
		case 4:
			encoder.writeHead = codecType.writeHead4
		case 8:
			encoder.writeHead = codecType.writeHead8
		}
	}
	encoder.base = codecType.base.NewEncoder(&encoder.buffer)
	return encoder
}

func (codecType *packetCodecType) writeHead1(buf *binary.Buffer) {
	if n := len(buf.Data) - 1; n <= 254 && n <= codecType.maxPacketSize {
		buf.Data[0] = byte(n)
	} else {
		panic("too large packet size")
	}
}

func (codecType *packetCodecType) writeHead2(buf *binary.Buffer) {
	if n := len(buf.Data) - 2; n <= 65534 && n <= codecType.maxPacketSize {
		codecType.byteOrder.PutUint16(buf.Data, uint16(n))
	} else {
		panic("too large packet size")
	}
}

func (codecType *packetCodecType) writeHead4(buf *binary.Buffer) {
	if n := len(buf.Data) - 4; n <= codecType.maxPacketSize {
		codecType.byteOrder.PutUint32(buf.Data, uint32(n))
	} else {
		panic("too large packet size")
	}
}

func (codecType *packetCodecType) writeHead8(buf *binary.Buffer) {
	if n := len(buf.Data) - 8; n <= codecType.maxPacketSize {
		codecType.byteOrder.PutUint64(buf.Data, uint64(n))
	} else {
		panic("too large packet size")
	}
}

func (codecType *packetCodecType) NewDecoder(r io.Reader) Decoder {
	decoder, ok := codecType.decoderPool.Get().(*packetDecoder)
	if ok {
		decoder.reader.Reset(r)
	} else {
		decoder = &packetDecoder{
			n:      codecType.n,
			parent: codecType,
			reader: bufio.NewReaderSize(r, codecType.readBufferSize),
		}
		if codecType.bufferPool != nil {
			codecType.bufferPool.Manage(&decoder.buffer)
		} else {
			decoder.buffer.Data = make([]byte, codecType.n+codecType.readBufferSize)
		}
		switch codecType.n {
		case 1:
			decoder.readHead = codecType.readHead1
		case 2:
			decoder.readHead = codecType.readHead2
		case 4:
			decoder.readHead = codecType.readHead4
		case 8:
			decoder.readHead = codecType.readHead8
		}
	}
	decoder.base = codecType.base.NewDecoder(&decoder.buffer)
	return decoder
}

func (codecType *packetCodecType) readHead1(buf []byte) int {
	if n := int(buf[0]); n <= 254 && n <= codecType.maxPacketSize {
		return n
	} else {
		panic("too large packet size")
	}
}

func (codecType *packetCodecType) readHead2(buf []byte) int {
	if n := int(codecType.byteOrder.Uint16(buf)); n > 0 && n <= 65534 && n <= codecType.maxPacketSize {
		return n
	} else {
		panic("too large packet size")
	}
}

func (codecType *packetCodecType) readHead4(buf []byte) int {
	if n := int(codecType.byteOrder.Uint32(buf)); n > 0 && n <= codecType.maxPacketSize {
		return n
	} else {
		panic("too large packet size")
	}
}

func (codecType *packetCodecType) readHead8(buf []byte) int {
	if n := int(codecType.byteOrder.Uint64(buf)); n > 0 && n <= codecType.maxPacketSize {
		return n
	} else {
		panic("too large packet size")
	}
}

type packetEncoder struct {
	n         int
	base      Encoder
	buffer    binary.Buffer
	writer    io.Writer
	parent    *packetCodecType
	writeHead func(*binary.Buffer)
}

type packetDecoder struct {
	n        int
	base     Decoder
	buffer   binary.Buffer
	reader   *bufio.Reader
	parent   *packetCodecType
	readHead func([]byte) int
}

func (encoder *packetEncoder) Encode(msg interface{}) (err error) {
	encoder.buffer.Data = encoder.buffer.Data[:encoder.n]
	if err = encoder.base.Encode(msg); err != nil {
		return err
	}
	encoder.writeHead(&encoder.buffer)
	_, err = encoder.writer.Write(encoder.buffer.Data)
	return err
}

func (decoder *packetDecoder) Decode(msg interface{}) (err error) {
	decoder.buffer.Data = decoder.buffer.Data[:decoder.n]
	if _, err = io.ReadFull(decoder.reader, decoder.buffer.Data); err != nil {
		return err
	}
	n := decoder.readHead(decoder.buffer.Data)
	decoder.buffer.Renew(n)
	if _, err = io.ReadFull(decoder.reader, decoder.buffer.Data); err != nil {
		return err
	}
	decoder.buffer.ReadPos = 0
	return decoder.base.Decode(msg)
}

func (encoder *packetEncoder) Dispose() {
	if d, ok := encoder.base.(Disposeable); ok {
		d.Dispose()
	}
	encoder.parent.encoderPool.Put(encoder)
}

func (decoder *packetDecoder) Dispose() {
	if d, ok := decoder.base.(Disposeable); ok {
		d.Dispose()
	}
	decoder.parent.decoderPool.Put(decoder)
}
