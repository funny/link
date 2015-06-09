package link

import (
	"net"
	"sync/atomic"
	"time"
)

type SessionFetcher func(func(*Session))

type ServerProtocol interface {
	NewListener(listener net.Listener) Listener
}

type ClientProtocol interface {
	NewClientConn(conn net.Conn) (Conn, error)
}

type BroadcastProtocol interface {
	Broadcast(msg interface{}, fetcher SessionFetcher) error
}

type Listener interface {
	Accept() (Conn, error)
	Handshake(Conn) error
	Addr() net.Addr
	Protocol() ServerProtocol
	Close() error
}

type Conn interface {
	Config() SessionConfig
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Receive(msg interface{}) error
	Send(msg interface{}) (err error)
	Close() error
}

var DefaultBroadcast = defaultBroadcast{}

type defaultBroadcast struct {
}

func (_ defaultBroadcast) Broadcast(msg interface{}, fetcher SessionFetcher) error {
	fetcher(func(session *Session) {
		session.AsyncSend(msg)
	})
	return nil
}

var autoSessionId uint64

func Listen(network, address string, protocol ServerProtocol) (Listener, error) {
	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	return protocol.NewListener(listener), nil
}

func Dial(network, address string, protocol ClientProtocol) (Conn, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return protocol.NewClientConn(conn)
}

func DialTimeout(network, address string, timeout time.Duration, protocol ClientProtocol) (Conn, error) {
	conn, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return nil, err
	}
	return protocol.NewClientConn(conn)
}

func Serve(network, address string, protocol ServerProtocol) (*Server, error) {
	listener, err := Listen(network, address, protocol)
	if err != nil {
		return nil, err
	}
	return NewServer(listener), nil
}

func Connect(network, address string, protocol ClientProtocol) (*Session, error) {
	conn, err := Dial(network, address, protocol)
	if err != nil {
		return nil, err
	}
	return NewSession(atomic.AddUint64(&autoSessionId, 1), conn), nil
}

func ConnectTimeout(network, address string, timeout time.Duration, protocol ClientProtocol) (*Session, error) {
	conn, err := DialTimeout(network, address, timeout, protocol)
	if err != nil {
		return nil, err
	}
	return NewSession(atomic.AddUint64(&autoSessionId, 1), conn), nil
}
