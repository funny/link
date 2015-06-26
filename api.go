package link

import (
	"bufio"
	"net"
)

// Server side protocol

type ServerProtocol interface {
	NewListener(net.Listener) (Listener, error)
}

type PacketServerProtocol interface {
	ServerProtocol
	NewPacketListener(net.Listener) (IPacketListener, error)
}

type StreamServerProtocol interface {
	ServerProtocol
	NewStreamListener(net.Listener) (IStreamListener, error)
}

// Client side protocol

type ClientProtocol interface {
	NewClientConn(net.Conn) (Conn, error)
}

type PacketClientProtocol interface {
	ClientProtocol
	NewPacketClientConn(net.Conn) (IPacketConn, error)
}

type StreamClientProtocol interface {
	ClientProtocol
	NewStreamClientConn(net.Conn) (IStreamConn, error)
}

// Listener

type Listener interface {
	Addr() net.Addr
	Accept() (Conn, error)
	Handshake(Conn) error
	Close() error
}

type IPacketListener interface {
	Listener
	AcceptPacket() (IPacketConn, error)
}

type IStreamListener interface {
	Listener
	AcceptStream() (IStreamConn, error)
}

// Connection

type Conn interface {
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Close() error
}

type IPacketConn interface {
	Conn
	ReadPacket() ([]byte, error)
	WritePacket([]byte) error
}

type IStreamConn interface {
	Conn
	UpStream() *bufio.Reader
	DownStream() *bufio.Writer
}

// Codec

type CodecType interface{}

type PSCodecType interface {
	PacketCodecType
	StreamCodecType
}

type PacketCodecType interface {
	NewPacketCodec() PacketCodec
}

type StreamCodecType interface {
	NewStreamCodec(*bufio.Reader, *bufio.Writer) StreamCodec
}

type PacketCodec interface {
	DecodePacket(interface{}, []byte) error
	EncodePacket(interface{}) ([]byte, error)
}

type StreamCodec interface {
	DecodeStream(interface{}) error
	EncodeStream(interface{}) error
}
