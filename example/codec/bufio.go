package codec

import (
	"bufio"
	"io"

	"github.com/funny/link"
)

const DEFAULT_BUFFER_SIZE = 4096

type BufioCodecType struct {
	Base            link.CodecType
	ReadBufferSize  int
	WriteBufferSize int
}

func Bufio(base link.CodecType) *BufioCodecType {
	return &BufioCodecType{
		base,
		DEFAULT_BUFFER_SIZE,
		DEFAULT_BUFFER_SIZE,
	}
}

func (codecType *BufioCodecType) NewEncoder(w io.Writer) link.Encoder {
	bw := bufio.NewWriterSize(w, codecType.WriteBufferSize)
	codec := &bufioEncoder{
		Writer: bw,
		Base:   codecType.Base.NewEncoder(bw),
	}
	return codec
}

func (codecType *BufioCodecType) NewDecoder(r io.Reader) link.Decoder {
	return &bufioDecoder{
		Base: codecType.Base.NewDecoder(
			bufio.NewReaderSize(r, codecType.ReadBufferSize),
		),
	}
}

type bufioEncoder struct {
	Base   link.Encoder
	Writer *bufio.Writer
}

func (encoder *bufioEncoder) Encode(msg interface{}) error {
	if err := encoder.Base.Encode(msg); err != nil {
		return err
	}
	return encoder.Writer.Flush()
}

type bufioDecoder struct {
	Base link.Decoder
}

func (decoder *bufioDecoder) Decode(msg interface{}) error {
	return decoder.Base.Decode(msg)
}
