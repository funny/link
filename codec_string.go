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

func (codecType stringCodecType) NewEncoder(w io.Writer) Encoder {
	return stringEncoder{
		codecType.Spliter,
		binary.NewWriter(w),
	}
}

func (codecType stringCodecType) NewDecoder(r io.Reader) Decoder {
	return stringDecoder{
		codecType.Spliter,
		binary.NewReader(r),
	}
}

type stringEncoder struct {
	Spliter binary.Spliter
	Writer  *binary.Writer
}

func (encoder stringEncoder) Encode(msg interface{}) error {
	encoder.Writer.WritePacket([]byte(msg.(string)), encoder.Spliter)
	return encoder.Writer.Flush()
}

type stringDecoder struct {
	Spliter binary.Spliter
	Reader  *binary.Reader
}

func (decoder stringDecoder) Decode(msg interface{}) error {
	*(msg.(*string)) = string(decoder.Reader.ReadPacket(decoder.Spliter))
	return decoder.Reader.Error()
}
