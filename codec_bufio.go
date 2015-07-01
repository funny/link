package link

import (
	"bufio"
	"io"
)

const DEFAULT_BUFFER_SIZE = 4096

type BufioCodecType struct {
	CodecType       CodecType
	ReadBufferSize  int
	WriteBufferSize int
}

func Bufio(codecType CodecType) *BufioCodecType {
	return &BufioCodecType{
		codecType,
		DEFAULT_BUFFER_SIZE,
		DEFAULT_BUFFER_SIZE,
	}
}

func (codecType *BufioCodecType) NewCodec(r io.Reader, w io.Writer) Codec {
	br := bufio.NewReaderSize(r, codecType.ReadBufferSize)
	bw := bufio.NewWriterSize(w, codecType.WriteBufferSize)
	codec := &bufioCodec{
		Codec:  codecType.CodecType.NewCodec(br, bw),
		Writer: bw,
	}
	return codec
}

type bufioCodec struct {
	Codec  Codec
	Writer *bufio.Writer
}

func (codec *bufioCodec) Encode(msg interface{}) error {
	if err := codec.Codec.Encode(msg); err != nil {
		return err
	}
	return codec.Writer.Flush()
}

func (codec *bufioCodec) Decode(msg interface{}) error {
	return codec.Codec.Decode(msg)
}
