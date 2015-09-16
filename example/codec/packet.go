package codec

import (
	"io"

	"github.com/funny/binary"
	"github.com/funny/link"
)

func Packet(spliter binary.Spliter, codecType link.CodecType) link.CodecType {
	return packetCodecType{spliter, codecType}
}

type packetCodecType struct {
	Spliter binary.Spliter
	Base    link.CodecType
}

func (codecType packetCodecType) NewEncoder(w io.Writer) link.Encoder {
	pw := binary.NewPacketWriter(codecType.Spliter, w)
	return &packetEncoder{
		Writer: pw,
		Base:   codecType.Base.NewEncoder(pw),
	}
}

func (codecType packetCodecType) NewDecoder(r io.Reader) link.Decoder {
	return &packetDecoder{
		Base: codecType.Base.NewDecoder(
			binary.NewPacketReader(codecType.Spliter, r),
		),
	}
}

type packetEncoder struct {
	Base   link.Encoder
	Writer *binary.PacketWriter
}

func (encoder *packetEncoder) Encode(msg interface{}) error {
	if err := encoder.Base.Encode(msg); err != nil {
		return err
	}
	return encoder.Writer.Flush()
}

type packetDecoder struct {
	Base link.Decoder
}

func (deocder *packetDecoder) Decode(msg interface{}) error {
	return deocder.Base.Decode(msg)
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
