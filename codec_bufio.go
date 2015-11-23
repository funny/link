package link

import (
	"bufio"
	"io"
	"sync"
)

type bufioCodecType struct {
	base            CodecType
	readBufferSize  int
	writeBufferSize int
	encoderPool     sync.Pool
	decoderPool     sync.Pool
}

func Bufio(base CodecType) CodecType {
	return BufioSize(4096, 4096, base)
}

func BufioSize(readBufferSize, writeBufferSize int, base CodecType) CodecType {
	return &bufioCodecType{
		base:            base,
		readBufferSize:  readBufferSize,
		writeBufferSize: writeBufferSize,
	}
}

func (codecType *bufioCodecType) NewEncoder(w io.Writer) Encoder {
	if codecType.writeBufferSize == 0 {
		return codecType.base.NewEncoder(w)
	}
	if encoder, ok := codecType.encoderPool.Get().(*bufioEncoder); ok {
		encoder.writer.Reset(w)
		encoder.base = codecType.base.NewEncoder(encoder.writer)
		return encoder
	}
	bw := bufio.NewWriterSize(w, codecType.writeBufferSize)
	return &bufioEncoder{
		writer: bw,
		parent: codecType,
		base:   codecType.base.NewEncoder(bw),
	}
}

func (codecType *bufioCodecType) NewDecoder(r io.Reader) Decoder {
	if codecType.readBufferSize == 0 {
		return codecType.base.NewDecoder(r)
	}
	if decoder, ok := codecType.decoderPool.Get().(*bufioDecoder); ok {
		decoder.reader.Reset(r)
		decoder.base = codecType.base.NewDecoder(decoder.reader)
		return decoder
	}
	br := bufio.NewReaderSize(r, codecType.readBufferSize)
	return &bufioDecoder{
		reader: br,
		parent: codecType,
		base:   codecType.base.NewDecoder(br),
	}
}

type bufioEncoder struct {
	base   Encoder
	writer *bufio.Writer
	parent *bufioCodecType
}

type bufioDecoder struct {
	base   Decoder
	reader *bufio.Reader
	parent *bufioCodecType
}

func (encoder *bufioEncoder) Encode(msg interface{}) error {
	if err := encoder.base.Encode(msg); err != nil {
		return err
	}
	return encoder.writer.Flush()
}

func (decoder *bufioDecoder) Decode(msg interface{}) error {
	return decoder.base.Decode(msg)
}

func (encoder *bufioEncoder) Dispose() {
	if d, ok := encoder.base.(Disposeable); ok {
		d.Dispose()
	}
	encoder.base = nil
	encoder.parent.encoderPool.Put(encoder)
}

func (decoder *bufioDecoder) Dispose() {
	if d, ok := decoder.base.(Disposeable); ok {
		d.Dispose()
	}
	decoder.base = nil
	decoder.parent.decoderPool.Put(decoder)
}
