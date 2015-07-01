package link

import (
	"io"

	"github.com/funny/binary"
)

func Packet(spliter binary.Spliter, codecType CodecType) CodecType {
	return packetCodecType{spliter, codecType}
}

type packetCodecType struct {
	Spliter   binary.Spliter
	CodecType CodecType
}

func (codecType packetCodecType) NewCodec(r io.Reader, w io.Writer) Codec {
	pr := binary.NewPacketReader(codecType.Spliter, r)
	pw := binary.NewPacketWriter(codecType.Spliter, w)
	return &packetCodec{
		Codec:  codecType.CodecType.NewCodec(pr, pw),
		Writer: pw,
	}
}

type packetCodec struct {
	Codec  Codec
	Writer *binary.PacketWriter
}

func (codec *packetCodec) Encode(msg interface{}) error {
	if err := codec.Codec.Encode(msg); err != nil {
		return err
	}
	return codec.Writer.Flush()
}

func (codec *packetCodec) Decode(msg interface{}) error {
	return codec.Codec.Decode(msg)
}

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
