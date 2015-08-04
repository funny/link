package link

import (
	"bufio"
	"io"
)

const DEFAULT_BUFFER_SIZE = 4096

type BufioCodecType struct {
	Base            CodecType
	ReadBufferSize  int
	WriteBufferSize int
}

func Bufio(base CodecType) *BufioCodecType {
	return &BufioCodecType{
		base,
		DEFAULT_BUFFER_SIZE,
		DEFAULT_BUFFER_SIZE,
	}
}

func (codecType *BufioCodecType) NewEncoder(w io.Writer) Encoder {
	bw := bufio.NewWriterSize(w, codecType.WriteBufferSize)
	codec := &bufioEncoder{
		Writer: bw,
		Base:   codecType.Base.NewEncoder(bw),
	}
	return codec
}

func (codecType *BufioCodecType) NewDecoder(r io.Reader) Decoder {
	return &bufioDecoder{
		Base: codecType.Base.NewDecoder(
			bufio.NewReaderSize(r, codecType.ReadBufferSize),
		),
	}
}

type bufioEncoder struct {
	Base   Encoder
	Writer *bufio.Writer
}

func (encoder *bufioEncoder) Encode(msg interface{}) error {
	if err := encoder.Base.Encode(msg); err != nil {
		return err
	}
	return encoder.Writer.Flush()
}

type bufioDecoder struct {
	Base Decoder
}

func (decoder *bufioDecoder) Decode(msg interface{}) error {
	return decoder.Base.Decode(msg)
}
