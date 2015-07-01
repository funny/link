package link

import (
	"bufio"
	"net"

	"github.com/funny/binary"
)

var (
	_ ServerProtocol  = &PacketProtocol{}
	_ ClientProtocol  = &PacketProtocol{}
	_ StreamCodecType = &PacketProtocol{}
)

var (
	Line     = binary.SplitByLine
	Null     = binary.SplitByNull
	Uvarint  = binary.SplitByUvarint
	Uint8    = binary.SplitByUint8
	Uint16BE = binary.SplitByUint16BE
	Uint16LE = binary.SplitByUint16LE
	Uint24BE = binary.SplitByUint24BE
	Uint24LE = binary.SplitByUint24LE
	Uint32BE = binary.SplitByUint32BE
	Uint32LE = binary.SplitByUint32LE
	Uint40BE = binary.SplitByUint40BE
	Uint40LE = binary.SplitByUint40LE
	Uint48BE = binary.SplitByUint48BE
	Uint48LE = binary.SplitByUint48LE
	Uint56BE = binary.SplitByUint56BE
	Uint56LE = binary.SplitByUint56LE
	Uint64BE = binary.SplitByUint64BE
	Uint64LE = binary.SplitByUint64LE
)

type PacketCodecType interface {
	NewPacketCodec() PacketCodec
}

type PacketCodec interface {
	DecodePacket(msg interface{}, b []byte) error
	EncodePacket(msg interface{}) ([]byte, error)
}

type PacketProtocol struct {
	packetCodecType
	StreamProtocol
}

func Packet(spliter binary.Spliter, codecType PacketCodecType) *PacketProtocol {
	protocol := &PacketProtocol{
		packetCodecType{spliter, codecType},
		StreamProtocol{nil, 8192, 8192, nil, nil},
	}
	protocol.StreamProtocol.CodecType = &protocol.packetCodecType
	return protocol
}

func (protocol *PacketProtocol) NewListener(listener net.Listener) (Listener, error) {
	return protocol.StreamProtocol.NewListener(listener)
}

func (protocol *PacketProtocol) NewClientConn(conn net.Conn) (Conn, error) {
	return protocol.StreamProtocol.NewClientConn(conn)
}

type packetCodecType struct {
	Spliter   binary.Spliter
	CodecType PacketCodecType
}

func (codecType *packetCodecType) NewStreamCodec(r *bufio.Reader, w *bufio.Writer) StreamCodec {
	return &packetStreamCodec{
		Codec:   codecType.CodecType.NewPacketCodec(),
		Spliter: codecType.Spliter,
		Reader:  binary.NewReader(r),
		Writer:  binary.NewWriter(w),
	}
}

func (codecType *packetCodecType) NewPacketCodec() PacketCodec {
	codec := &packetPacketCodec{
		Codec:   codecType.CodecType.NewPacketCodec(),
		Spliter: codecType.Spliter,
	}
	codec.Reader = binary.NewReader(&codec.ReadBuffer)
	codec.Writer = binary.NewWriter(&codec.WriteBuffer)
	return codec
}

type packetStreamCodec struct {
	Codec   PacketCodec
	Spliter binary.Spliter
	Reader  *binary.Reader
	Writer  *binary.Writer
}

func (codec *packetStreamCodec) EncodeStream(msg interface{}) error {
	b, err := codec.Codec.EncodePacket(msg)
	if err != nil {
		return err
	}
	codec.Writer.WritePacket(b, codec.Spliter)
	return codec.Writer.Flush()
}

func (codec *packetStreamCodec) DecodeStream(msg interface{}) error {
	b := codec.Reader.ReadPacket(codec.Spliter)
	if codec.Reader.Error() != nil {
		return codec.Reader.Error()
	}
	return codec.Codec.DecodePacket(msg, b)
}

type packetPacketCodec struct {
	Codec       PacketCodec
	Spliter     binary.Spliter
	ReadBuffer  binary.Buffer
	WriteBuffer binary.Buffer
	Reader      *binary.Reader
	Writer      *binary.Writer
}

func (codec *packetPacketCodec) EncodePacket(msg interface{}) ([]byte, error) {
	b, err := codec.Codec.EncodePacket(msg)
	if err != nil {
		return nil, err
	}
	codec.WriteBuffer.Reset(codec.WriteBuffer.Data[0:0])
	codec.Writer.WritePacket(b, codec.Spliter)
	return codec.WriteBuffer.Data, nil
}

func (codec *packetPacketCodec) DecodePacket(msg interface{}, b []byte) error {
	codec.ReadBuffer.Reset(b)
	b = codec.Reader.ReadPacket(codec.Spliter)
	if codec.Reader.Error() != nil {
		return codec.Reader.Error()
	}
	return codec.Codec.DecodePacket(msg, b)
}
