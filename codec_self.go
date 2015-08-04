package link

import (
	"io"

	"github.com/funny/binary"
)

func SelfCodec() CodecType {
	return selfCodecType{}
}

type selfCodecType struct{}

func (_ selfCodecType) NewEncoder(w io.Writer) Encoder {
	return selfEncoder{
		binary.NewWriter(w),
	}
}

func (_ selfCodecType) NewDecoder(r io.Reader) Decoder {
	return selfDecoder{
		binary.NewReader(r),
	}
}

type SelfDecoder interface {
	SelfDecode(*binary.Reader) error
}

type SelfEncoder interface {
	SelfEncode(*binary.Writer) error
}

type selfEncoder struct {
	Writer *binary.Writer
}

func (encoder selfEncoder) Encode(msg interface{}) error {
	if err := msg.(SelfEncoder).SelfEncode(encoder.Writer); err != nil {
		return err
	}
	return encoder.Writer.Flush()
}

type selfDecoder struct {
	Reader *binary.Reader
}

func (deocder selfDecoder) Decode(msg interface{}) error {
	if err := msg.(SelfDecoder).SelfDecode(deocder.Reader); err != nil {
		return err
	}
	return deocder.Reader.Error()
}
