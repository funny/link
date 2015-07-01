package link

import (
	"io"

	"github.com/funny/binary"
)

func SelfCodec() CodecType {
	return selfCodecType{}
}

type selfCodecType struct{}

func (_ selfCodecType) NewCodec(r io.Reader, w io.Writer) Codec {
	return selfCodec{
		binary.NewReader(r),
		binary.NewWriter(w),
	}
}

type SelfDecoder interface {
	SelfDecode(*binary.Reader) error
}

type SelfEncoder interface {
	SelfEncode(*binary.Writer) error
}

type selfCodec struct {
	Reader *binary.Reader
	Writer *binary.Writer
}

func (codec selfCodec) Decode(msg interface{}) error {
	if err := msg.(SelfDecoder).SelfDecode(codec.Reader); err != nil {
		return err
	}
	return codec.Reader.Error()
}

func (codec selfCodec) Encode(msg interface{}) error {
	if err := msg.(SelfEncoder).SelfEncode(codec.Writer); err != nil {
		return err
	}
	return codec.Writer.Flush()
}
