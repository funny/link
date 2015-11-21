package link

import (
	"io"
	"net"
	"sync"
	"time"

	"encoding/gob"
	"encoding/json"
	"encoding/xml"
)

type CodecType interface {
	NewEncoder(w io.Writer) Encoder
	NewDecoder(r io.Reader) Decoder
}

type Encoder interface {
	Encode(msg interface{}) error
}

type Decoder interface {
	Decode(msg interface{}) error
}

type Disposeable interface {
	Dispose()
}

func Serve(network, address string, codecType CodecType) (*Server, error) {
	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	return NewServer(listener, codecType), nil
}

func Connect(network, address string, codecType CodecType) (*Session, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return NewSession(conn, codecType), nil
}

func ConnectTimeout(network, address string, timeout time.Duration, codecType CodecType) (*Session, error) {
	conn, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return nil, err
	}
	return NewSession(conn, codecType), nil
}

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
