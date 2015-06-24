package link

import (
	"io"
	"net"
	"strings"
	"time"
)

func ParseAddr(address string) (net, addr string) {
	n := strings.Index(address, "://")
	return address[:n], address[n+3:]
}

func Echo(session *Session) {
	c := session.Conn().(IStreamConn)
	io.Copy(c.DownStream(), c.UpStream())
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

func Serve(address string, protocol ServerProtocol, codec CodecType) (*Server, error) {
	listener, err := Listen(address, protocol)
	if err != nil {
		return nil, err
	}
	return NewServer(listener, codec), nil
}

func Connect(address string, protocol ClientProtocol, codec CodecType) (*Session, error) {
	conn, err := Dial(address, protocol)
	if err != nil {
		return nil, err
	}
	return NewSession(conn, codec), nil
}

func ConnectTimeout(address string, timeout time.Duration, protocol ClientProtocol, codec CodecType) (*Session, error) {
	conn, err := DialTimeout(address, timeout, protocol)
	if err != nil {
		return nil, err
	}
	return NewSession(conn, codec), nil
}

func ListenPacket(address string, protocol PacketServerProtocol) (IPacketListener, error) {
	lnet, laddr := ParseAddr(address)
	listener, err := net.Listen(lnet, laddr)
	if err != nil {
		return nil, err
	}
	return protocol.NewPacketListener(listener)
}

func DialPacket(address string, protocol PacketClientProtocol) (IPacketConn, error) {
	lnet, laddr := ParseAddr(address)
	conn, err := net.Dial(lnet, laddr)
	if err != nil {
		return nil, err
	}
	return protocol.NewPacketClientConn(conn)
}

func DialPacketTimeout(address string, timeout time.Duration, protocol PacketClientProtocol) (IPacketConn, error) {
	lnet, laddr := ParseAddr(address)
	conn, err := net.DialTimeout(lnet, laddr, timeout)
	if err != nil {
		return nil, err
	}
	return protocol.NewPacketClientConn(conn)
}

func ServePacket(address string, protocol PacketServerProtocol, codec CodecType) (*Server, error) {
	listener, err := ListenPacket(address, protocol)
	if err != nil {
		return nil, err
	}
	return NewServer(listener, codec), nil
}

func ConnectPacket(address string, protocol PacketClientProtocol, codec CodecType) (*Session, error) {
	conn, err := DialPacket(address, protocol)
	if err != nil {
		return nil, err
	}
	return NewSession(conn, codec), nil
}

func ConnectPacketTimeout(address string, timeout time.Duration, protocol PacketClientProtocol, codec CodecType) (*Session, error) {
	conn, err := DialPacketTimeout(address, timeout, protocol)
	if err != nil {
		return nil, err
	}
	return NewSession(conn, codec), nil
}

func ListenStream(address string, protocol StreamServerProtocol) (IStreamListener, error) {
	lnet, laddr := ParseAddr(address)
	listener, err := net.Listen(lnet, laddr)
	if err != nil {
		return nil, err
	}
	return protocol.NewStreamListener(listener)
}

func DialStream(address string, protocol StreamClientProtocol) (IStreamConn, error) {
	lnet, laddr := ParseAddr(address)
	conn, err := net.Dial(lnet, laddr)
	if err != nil {
		return nil, err
	}
	return protocol.NewStreamClientConn(conn)
}

func DialStreamTimeout(address string, timeout time.Duration, protocol StreamClientProtocol) (IStreamConn, error) {
	lnet, laddr := ParseAddr(address)
	conn, err := net.DialTimeout(lnet, laddr, timeout)
	if err != nil {
		return nil, err
	}
	return protocol.NewStreamClientConn(conn)
}

func ServeStream(address string, protocol StreamServerProtocol, codec CodecType) (*Server, error) {
	listener, err := ListenStream(address, protocol)
	if err != nil {
		return nil, err
	}
	return NewServer(listener, codec), nil
}

func ConnectStream(address string, protocol StreamClientProtocol, codec CodecType) (*Session, error) {
	conn, err := DialStream(address, protocol)
	if err != nil {
		return nil, err
	}
	return NewSession(conn, codec), nil
}

func ConnectStreamTimeout(address string, timeout time.Duration, protocol StreamClientProtocol, codec CodecType) (*Session, error) {
	conn, err := DialStreamTimeout(address, timeout, protocol)
	if err != nil {
		return nil, err
	}
	return NewSession(conn, codec), nil
}
