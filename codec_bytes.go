package link

import (
	"io"

	"github.com/funny/binary"
)

func Bytes(spliter binary.Spliter) CodecType {
	return bytesCodecType{spliter}
}

type bytesCodecType struct {
	Spliter binary.Spliter
}

func (codecType bytesCodecType) NewEncoder(w io.Writer) Encoder {
	return bytesEncoder{
		codecType.Spliter,
		binary.NewWriter(w),
	}
}

func (codecType bytesCodecType) NewDecoder(r io.Reader) Decoder {
	return bytesDecoder{
		codecType.Spliter,
		binary.NewReader(r),
	}
}

type bytesEncoder struct {
	Spliter binary.Spliter
	Writer  *binary.Writer
}

func (encoder bytesEncoder) Encode(msg interface{}) error {
	encoder.Writer.WritePacket(msg.([]byte), encoder.Spliter)
	return encoder.Writer.Flush()
}

type bytesDecoder struct {
	Spliter binary.Spliter
	Reader  *binary.Reader
}

func (decoder bytesDecoder) Decode(msg interface{}) error {
	*(msg.(*[]byte)) = decoder.Reader.ReadPacket(decoder.Spliter)
	return decoder.Reader.Error()
}
