package link

import (
	"bufio"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"github.com/funny/binary"
	"io"
)

func Gob() PSCodecType {
	return &genCodecType{
		func(r io.Reader) decoder { return gob.NewDecoder(r) },
		func(w io.Writer) encoder { return gob.NewEncoder(w) },
	}
}

func Json() PSCodecType {
	return &genCodecType{
		func(r io.Reader) decoder { return json.NewDecoder(r) },
		func(w io.Writer) encoder { return json.NewEncoder(w) },
	}
}

func Xml() PSCodecType {
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

func (streamType *genCodecType) NewPacketCodec() PacketCodec {
	codec := &genPacketCodec{}
	codec.Decoder = streamType.newDecoder(&codec.rbuf)
	codec.Encoder = streamType.newEncoder(&codec.wbuf)
	return codec
}

func (streamType *genCodecType) NewStreamCodec(up *bufio.Reader, down *bufio.Writer) StreamCodec {
	return &genStreamCodec{
		Writer:  down,
		Decoder: streamType.newDecoder(up),
		Encoder: streamType.newEncoder(down),
	}
}

type genPacketCodec struct {
	rbuf    binary.Buffer
	wbuf    binary.Buffer
	Decoder decoder
	Encoder encoder
}

func (codec *genPacketCodec) DecodePacket(msg interface{}, b []byte) error {
	codec.rbuf.Reset(b)
	return codec.Decoder.Decode(msg)
}

func (codec *genPacketCodec) EncodePacket(msg interface{}) ([]byte, error) {
	codec.wbuf.Reset(codec.wbuf.Data[0:0])
	if err := codec.Encoder.Encode(msg); err != nil {
		return nil, err
	}
	return codec.wbuf.Bytes(), nil
}

type genStreamCodec struct {
	Writer  *bufio.Writer
	Decoder decoder
	Encoder encoder
}

func (codec *genStreamCodec) DecodeStream(msg interface{}) error {
	return codec.Decoder.Decode(msg)
}

func (codec *genStreamCodec) EncodeStream(msg interface{}) error {
	if err := codec.Encoder.Encode(msg); err != nil {
		return err
	}
	return codec.Writer.Flush()
}
