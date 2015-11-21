package link

import (
	"bufio"
	"io"
	"sync"
)

type BufioCodecType struct {
	base            CodecType
	readBufferSize  int
	writeBufferSize int
	readerPool      sync.Pool
	writerPool      sync.Pool
}

func Bufio(base CodecType) *BufioCodecType {
	return BufioSize(base, 4096, 4096)
}

func BufioSize(base CodecType, readBufferSize, writeBufferSize int) *BufioCodecType {
	return &BufioCodecType{
		base:            base,
		readBufferSize:  readBufferSize,
		writeBufferSize: writeBufferSize,
	}
}

func (codecType *BufioCodecType) NewEncoder(w io.Writer) Encoder {
	if codecType.writeBufferSize == 0 {
		return codecType.base.NewEncoder(w)
	}
	bw, ok := codecType.writerPool.Get().(*bufio.Writer)
	if ok {
		bw.Reset(w)
	} else {
		bw = bufio.NewWriterSize(w, codecType.writeBufferSize)
	}
	return &bufioEncoder{
		writer: bw,
		pool:   &codecType.writerPool,
		base:   codecType.base.NewEncoder(bw),
	}
}

func (codecType *BufioCodecType) NewDecoder(r io.Reader) Decoder {
	if codecType.readBufferSize == 0 {
		return codecType.base.NewDecoder(r)
	}
	br, ok := codecType.readerPool.Get().(*bufio.Reader)
	if ok {
		br.Reset(r)
	} else {
		br = bufio.NewReaderSize(r, codecType.readBufferSize)
	}
	return &bufioDecoder{
		reader: br,
		pool:   &codecType.readerPool,
		base:   codecType.base.NewDecoder(br),
	}
}

type bufioEncoder struct {
	writer *bufio.Writer
	pool   *sync.Pool
	base   Encoder
}

func (encoder *bufioEncoder) Encode(msg interface{}) error {
	if err := encoder.base.Encode(msg); err != nil {
		return err
	}
	return encoder.writer.Flush()
}

func (encoder *bufioEncoder) Dispose() {
	if d, ok := encoder.base.(Disposeable); ok {
		d.Dispose()
	}
	encoder.pool.Put(encoder.writer)
}

type bufioDecoder struct {
	reader *bufio.Reader
	pool   *sync.Pool
	base   Decoder
}

func (decoder *bufioDecoder) Decode(msg interface{}) error {
	return decoder.base.Decode(msg)
}

func (decoder *bufioDecoder) Dispose() {
	if d, ok := decoder.base.(Disposeable); ok {
		d.Dispose()
	}
	decoder.pool.Put(decoder.reader)
}
