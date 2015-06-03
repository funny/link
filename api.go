package link

import (
	"net"
	"sync/atomic"
	"time"
)

type SessionFetcher func(func(*Session))

type ServerProtocol interface {
	NewListener(listener net.Listener) Listener
	Broadcast(msg interface{}, fetcher SessionFetcher) error
}

type ClientProtocol interface {
	NewClientConn(conn net.Conn) (Conn, error)
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

var autoSessionId uint64

func Listen(network, address string, protocol ServerProtocol) (*Server, error) {
	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	return NewServer(protocol.NewListener(listener)), nil
}

func Dial(network, address string, protocol ClientProtocol) (*Session, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	id := atomic.AddUint64(&autoSessionId, 1)
	lconn, lerr := protocol.NewClientConn(conn)
	if lerr != nil {
		return nil, err
	}
	return NewSession(id, lconn), nil
}

func DialTimeout(network, address string, timeout time.Duration, protocol ClientProtocol) (*Session, error) {
	conn, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return nil, err
	}
	id := atomic.AddUint64(&autoSessionId, 1)
	lconn, lerr := protocol.NewClientConn(conn)
	if lerr != nil {
		return nil, err
	}
	return NewSession(id, lconn), nil
}

func DefaultBroadcast(msg interface{}, fetcher SessionFetcher) error {
	fetcher(func(session *Session) {
		session.AsyncSend(msg)
	})
	return nil
}
