package link

import (
	"io"
	"net"
	"time"
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
