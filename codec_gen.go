package link

import (
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"io"
)

func Gob() CodecType {
	return &genCodecType{
		func(r io.Reader) decoder { return gob.NewDecoder(r) },
		func(w io.Writer) encoder { return gob.NewEncoder(w) },
	}
}

func Json() CodecType {
	return &genCodecType{
		func(r io.Reader) decoder { return json.NewDecoder(r) },
		func(w io.Writer) encoder { return json.NewEncoder(w) },
	}
}

func Xml() CodecType {
	return &genCodecType{
		func(r io.Reader) decoder { return xml.NewDecoder(r) },
		func(w io.Writer) encoder { return xml.NewEncoder(w) },
	}
}

type encoder interface {
	Encode(interface{}) error
}

type decoder interface {
	Decode(interface{}) error
}

type genCodecType struct {
	newDecoder func(io.Reader) decoder
	newEncoder func(io.Writer) encoder
}

func (codecType *genCodecType) NewCodec(r io.Reader, w io.Writer) Codec {
	return &genCodec{
		Decoder: codecType.newDecoder(r),
		Encoder: codecType.newEncoder(w),
	}
}

type genCodec struct {
	Decoder decoder
	Encoder encoder
}

func (codec *genCodec) Decode(msg interface{}) error {
	return codec.Decoder.Decode(msg)
}

func (codec *genCodec) Encode(msg interface{}) error {
	return codec.Encoder.Encode(msg)
}
