package codec

import (
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"io"

	"github.com/funny/link"
)

func Gob() link.CodecType {
	return &genCodecType{
		func(w io.Writer) link.Encoder { return gob.NewEncoder(w) },
		func(r io.Reader) link.Decoder { return gob.NewDecoder(r) },
	}
}

func Json() link.CodecType {
	return &genCodecType{
		func(w io.Writer) link.Encoder { return json.NewEncoder(w) },
		func(r io.Reader) link.Decoder { return json.NewDecoder(r) },
	}
}

func Xml() link.CodecType {
	return &genCodecType{
		func(w io.Writer) link.Encoder { return xml.NewEncoder(w) },
		func(r io.Reader) link.Decoder { return xml.NewDecoder(r) },
	}
}

func Mix(encodeType link.EncodeType, decodeType link.DecodeType) link.CodecType {
	return &genCodecType{
		encodeType.NewEncoder,
		decodeType.NewDecoder,
	}
}

type genCodecType struct {
	newEncoder func(io.Writer) link.Encoder
	newDecoder func(io.Reader) link.Decoder
}

func (codecType *genCodecType) NewEncoder(w io.Writer) link.Encoder {
	return codecType.newEncoder(w)
}

func (codecType *genCodecType) NewDecoder(r io.Reader) link.Decoder {
	return codecType.newDecoder(r)
}
