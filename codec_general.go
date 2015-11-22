package link

import (
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"io"
)

func Gob() CodecType {
	return &genCodecType{
		func(w io.Writer) Encoder { return gob.NewEncoder(w) },
		func(r io.Reader) Decoder { return gob.NewDecoder(r) },
	}
}

func Json() CodecType {
	return &genCodecType{
		func(w io.Writer) Encoder { return json.NewEncoder(w) },
		func(r io.Reader) Decoder { return json.NewDecoder(r) },
	}
}

func Xml() CodecType {
	return &genCodecType{
		func(w io.Writer) Encoder { return xml.NewEncoder(w) },
		func(r io.Reader) Decoder { return xml.NewDecoder(r) },
	}
}

type genCodecType struct {
	newEncoder func(io.Writer) Encoder
	newDecoder func(io.Reader) Decoder
}

func (codecType *genCodecType) NewEncoder(w io.Writer) Encoder {
	return codecType.newEncoder(w)
}

func (codecType *genCodecType) NewDecoder(r io.Reader) Decoder {
	return codecType.newDecoder(r)
}
