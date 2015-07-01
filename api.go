package link

import (
	"net"
	"strings"
	"time"
)

type ServerProtocol interface {
	NewListener(net.Listener) (Listener, error)
}

type ClientProtocol interface {
	NewClientConn(net.Conn) (Conn, error)
}

type Listener interface {
	Addr() net.Addr
	Accept() (Conn, error)
	Handshake(Conn) error
	Close() error
}

type Conn interface {
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Send(interface{}) error
	Receive(interface{}) error
	Close() error
}

func ParseAddr(address string) (net, addr string) {
	n := strings.Index(address, "://")
	return address[:n], address[n+3:]
}

func Serve(address string, protocol ServerProtocol) (*Server, error) {
	listener, err := Listen(address, protocol)
	if err != nil {
		return nil, err
	}
	return NewServer(listener), nil
}

func Connect(address string, protocol ClientProtocol) (*Session, error) {
	conn, err := Dial(address, protocol)
	if err != nil {
		return nil, err
	}
	return NewSession(conn), nil
}

func ConnectTimeout(address string, timeout time.Duration, protocol ClientProtocol) (*Session, error) {
	conn, err := DialTimeout(address, timeout, protocol)
	if err != nil {
		return nil, err
	}
	return NewSession(conn), nil
}

func Listen(address string, protocol ServerProtocol) (Listener, error) {
	lnet, laddr := ParseAddr(address)
	listener, err := net.Listen(lnet, laddr)
	if err != nil {
		return nil, err
	}
	return protocol.NewListener(listener)
}

func Dial(address string, protocol ClientProtocol) (Conn, error) {
	lnet, laddr := ParseAddr(address)
	conn, err := net.Dial(lnet, laddr)
	if err != nil {
		return nil, err
	}
	return protocol.NewClientConn(conn)
}

func DialTimeout(address string, timeout time.Duration, protocol ClientProtocol) (Conn, error) {
	lnet, laddr := ParseAddr(address)
	conn, err := net.DialTimeout(lnet, laddr, timeout)
	if err != nil {
		return nil, err
	}
	return protocol.NewClientConn(conn)
}
