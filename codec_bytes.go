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

func (codecType bytesCodecType) NewCodec(r io.Reader, w io.Writer) Codec {
	return bytesCodec{
		codecType.Spliter,
		binary.NewReader(r),
		binary.NewWriter(w),
	}
}

type bytesCodec struct {
	Spliter binary.Spliter
	Reader  *binary.Reader
	Writer  *binary.Writer
}

func (codec bytesCodec) Decode(msg interface{}) error {
	*(msg.(*[]byte)) = codec.Reader.ReadPacket(codec.Spliter)
	return codec.Reader.Error()
}

func (codec bytesCodec) Encode(msg interface{}) error {
	codec.Writer.WritePacket(msg.([]byte), codec.Spliter)
	return codec.Writer.Flush()
}
