package link

import (
	"io"
	"net"
	"strings"
	"time"
)

type CodecType interface {
	NewCodec(r io.Reader, w io.Writer) Codec
}

type Codec interface {
	Decode(msg interface{}) error
	Encode(msg interface{}) error
}

func ParseAddr(address string) (net, addr string) {
	n := strings.Index(address, "://")
	return address[:n], address[n+3:]
}

func Serve(address string, codecType CodecType) (*Server, error) {
	lnet, laddr := ParseAddr(address)
	listener, err := net.Listen(lnet, laddr)
	if err != nil {
		return nil, err
	}
	return NewServer(listener, codecType), nil
}

func Connect(address string, codecType CodecType) (*Session, error) {
	lnet, laddr := ParseAddr(address)
	conn, err := net.Dial(lnet, laddr)
	if err != nil {
		return nil, err
	}
	return NewSession(conn, codecType), nil
}

func ConnectTimeout(address string, timeout time.Duration, codecType CodecType) (*Session, error) {
	lnet, laddr := ParseAddr(address)
	conn, err := net.DialTimeout(lnet, laddr, timeout)
	if err != nil {
		return nil, err
	}
	return NewSession(conn, codecType), nil
}
