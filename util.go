package link

import (
	"net"
	"sync/atomic"
	"time"
)

var (
	gloablSessionId uint64
	DefaultConfig   = Config{
		ConnConfig{
			ReadBufferSize:  2048,
			WriteBufferSize: 2048,
		},
		SessionConfig{
			AutoFlush:         true,
			AsyncSendTimeout:  0,
			AsyncSendChanSize: 1000,
		},
	}
)

func Listen(network, address string) (*Listener, error) {
	l, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	return NewListener(l, DefaultConfig.ConnConfig), nil
}

func Serve(network, address string) (*Server, error) {
	listener, err := Listen(network, address)
	if err != nil {
		return nil, err
	}
	return NewServer(listener, DefaultConfig), nil
}

func Dial(network, address string) (*Conn, error) {
	c, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return NewConn(c, DefaultConfig.ConnConfig), nil
}

func Connet(network, address string) (*Session, error) {
	conn, err := Dial(network, address)
	if err != nil {
		return nil, err
	}
	id := atomic.AddUint64(&gloablSessionId, 1)
	return NewSession(id, conn, DefaultConfig.SessionConfig), nil
}

func DialTimeout(network, address string, timeout time.Duration) (*Conn, error) {
	c, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return nil, err
	}
	return NewConn(c, DefaultConfig.ConnConfig), nil
}

func ConnectTimeout(network, address string, timeout time.Duration) (*Session, error) {
	conn, err := DialTimeout(network, address, timeout)
	if err != nil {
		return nil, err
	}
	id := atomic.AddUint64(&gloablSessionId, 1)
	return NewSession(id, conn, DefaultConfig.SessionConfig), nil
}
