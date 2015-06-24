package link

import (
	"bufio"
	"net"
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
	Close() error
}

type PacketServerProtocol interface {
	ServerProtocol
	NewPacketListener(net.Listener) (IPacketListener, error)
}

type PacketClientProtocol interface {
	ClientProtocol
	NewPacketClientConn(net.Conn) (IPacketConn, error)
}

type IPacketListener interface {
	Listener
	AcceptPacket() (IPacketConn, error)
}

type IPacketConn interface {
	Conn
	ReadPacket() ([]byte, error)
	WritePacket([]byte) error
}

type StreamServerProtocol interface {
	ServerProtocol
	NewStreamListener(net.Listener) (IStreamListener, error)
}

type StreamClientProtocol interface {
	ClientProtocol
	NewStreamClientConn(net.Conn) (IStreamConn, error)
}

type IStreamListener interface {
	Listener
	AcceptStream() (IStreamConn, error)
}

type IStreamConn interface {
	Conn
	UpStream() *bufio.Reader
	DownStream() *bufio.Writer
}

type CodecType interface{}

type PSCodecType interface {
	PacketCodecType
	StreamCodecType
}

type PacketCodecType interface {
	NewPacketCodec() PacketCodec
}

type PacketCodec interface {
	DecodePacket(interface{}, []byte) error
	EncodePacket(interface{}) ([]byte, error)
}

type StreamCodecType interface {
	NewStreamCodec(*bufio.Reader, *bufio.Writer) StreamCodec
}

type StreamCodec interface {
	DecodeStream(interface{}) error
	EncodeStream(interface{}) error
}
