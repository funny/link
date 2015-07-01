package link

import (
	"io"

	"github.com/funny/binary"
)

func String(spliter binary.Spliter) CodecType {
	return stringCodecType{spliter}
}

type stringCodecType struct {
	Spliter binary.Spliter
}

func (codecType stringCodecType) NewCodec(r io.Reader, w io.Writer) Codec {
	return stringCodec{
		codecType.Spliter,
		binary.NewReader(r),
		binary.NewWriter(w),
	}
}

type stringCodec struct {
	Spliter binary.Spliter
	Reader  *binary.Reader
	Writer  *binary.Writer
}

func (codec stringCodec) Decode(msg interface{}) error {
	*(msg.(*string)) = string(codec.Reader.ReadPacket(codec.Spliter))
	return codec.Reader.Error()
}

func (codec stringCodec) Encode(msg interface{}) error {
	codec.Writer.WritePacket([]byte(msg.(string)), codec.Spliter)
	return codec.Writer.Flush()
}
