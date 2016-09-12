package link

import (
	"io"
	"net"
	"time"
)

type Context interface{}

type Protocol interface {
	NewCodec(rw io.ReadWriter) (Codec, Context, error)
}

type Codec interface {
	Receive() (interface{}, error)
	Send(interface{}) error
	Close() error
}

func Serve(network, address string, protocol Protocol, sendChanSize int) (*Server, error) {
	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	return NewServer(listener, protocol, sendChanSize), nil
}

func Connect(network, address string, protocol Protocol, sendChanSize int) (*Session, Context, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, nil, err
	}
	codec, ctx, err := protocol.NewCodec(conn)
	if err != nil {
		return nil, nil, err
	}
	return NewSession(codec, sendChanSize), ctx, nil
}

func ConnectTimeout(network, address string, timeout time.Duration, protocol Protocol, sendChanSize int) (*Session, Context, error) {
	conn, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return nil, nil, err
	}
	codec, ctx, err := protocol.NewCodec(conn)
	if err != nil {
		return nil, nil, err
	}
	return NewSession(codec, sendChanSize), ctx, nil
}
