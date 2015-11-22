package link

import (
	"io"
	"sync"
)

func ThreadSafe(base CodecType) CodecType {
	return safeCodecType{
		base: base,
	}
}

type safeCodecType struct {
	base CodecType
}

type safeDecoder struct {
	sync.Mutex
	base Decoder
}

type safeEncoder struct {
	sync.Mutex
	base Encoder
}

func (codecType safeCodecType) NewEncoder(w io.Writer) Encoder {
	return &safeEncoder{
		base: codecType.base.NewEncoder(w),
	}
}

func (codecType safeCodecType) NewDecoder(r io.Reader) Decoder {
	return &safeDecoder{
		base: codecType.base.NewDecoder(r),
	}
}

func (encoder *safeEncoder) Encode(msg interface{}) error {
	encoder.Lock()
	defer encoder.Unlock()
	return encoder.base.Encode(msg)
}

func (decoder *safeDecoder) Decode(msg interface{}) error {
	decoder.Lock()
	defer decoder.Unlock()
	return decoder.base.Decode(msg)
}

func (encoder *safeEncoder) Dispose() {
	if d, ok := encoder.base.(Disposeable); ok {
		d.Dispose()
	}
}

func (decoder *safeDecoder) Dispose() {
	if d, ok := decoder.base.(Disposeable); ok {
		d.Dispose()
	}
}
