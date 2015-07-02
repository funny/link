package link

import (
	"io"
	"net"
	"time"
)

type CodecType interface {
	NewCodec(r io.Reader, w io.Writer) Codec
}

type Codec interface {
	Decode(msg interface{}) error
	Encode(msg interface{}) error
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
