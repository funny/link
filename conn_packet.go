package link

import (
	"github.com/funny/binary"
	"net"
)

var (
	_ ServerProtocol = &PacketProtocol{}
	_ ClientProtocol = &PacketProtocol{}
	_ Listener       = &PacketListener{}
	_ Conn           = &PacketConn{}
)

var (
	Line     = binary.SplitByLine
	Null     = binary.SplitByNull
	Uvarint  = binary.SplitByUvarint
	Uint8    = binary.SplitByUint8
	Uint16BE = binary.SplitByUint16BE
	Uint16LE = binary.SplitByUint16LE
	Uint24BE = binary.SplitByUint24BE
	Uint24LE = binary.SplitByUint24LE
	Uint32BE = binary.SplitByUint32BE
	Uint32LE = binary.SplitByUint32LE
	Uint40BE = binary.SplitByUint40BE
	Uint40LE = binary.SplitByUint40LE
	Uint48BE = binary.SplitByUint48BE
	Uint48LE = binary.SplitByUint48LE
	Uint56BE = binary.SplitByUint56BE
	Uint56LE = binary.SplitByUint56LE
	Uint64BE = binary.SplitByUint64BE
	Uint64LE = binary.SplitByUint64LE
)

type PacketCodecType interface {
	NewPacketCodec() PacketCodec
}

type PacketCodec interface {
	DecodePacket(msg interface{}, b []byte) error
	EncodePacket(msg interface{}) ([]byte, error)
}

type PacketProtocol struct {
	Spliter          binary.Spliter
	CodecType        PacketCodecType
	ReadBufferSize   int
	WriteBufferSize  int
	ClientHandshaker func(*PacketConn) error
	ServerHandshaker func(*PacketConn) error
}

func Packet(spliter binary.Spliter, codecType PacketCodecType) *PacketProtocol {
	return &PacketProtocol{spliter, codecType, 8192, 8192, nil, nil}
}

func (protocol *PacketProtocol) NewListener(listener net.Listener) (Listener, error) {
	return NewPacketListener(listener, protocol), nil
}

func (protocol *PacketProtocol) NewClientConn(conn net.Conn) (Conn, error) {
	lconn := NewPacketConn(conn, protocol)
	if protocol.ClientHandshaker != nil {
		if err := protocol.ClientHandshaker(lconn); err != nil {
			lconn.Close()
			return nil, err
		}
	}
	return lconn, nil
}

type PacketListener struct {
	listener net.Listener
	protocol *PacketProtocol
}

func NewPacketListener(listener net.Listener, protocol *PacketProtocol) *PacketListener {
	return &PacketListener{listener, protocol}
}

func (l *PacketListener) Accept() (Conn, error) {
	conn, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}
	return NewPacketConn(conn, l.protocol), nil
}

func (l *PacketListener) Handshake(conn Conn) error {
	if l.protocol.ServerHandshaker != nil {
		return l.protocol.ServerHandshaker(conn.(*PacketConn))
	}
	return nil
}

func (l *PacketListener) Addr() net.Addr { return l.listener.Addr() }
func (l *PacketListener) Close() error   { return l.listener.Close() }

type PacketConn struct {
	Conn    net.Conn
	Codec   PacketCodec
	Spliter binary.Spliter
	Reader  *binary.Reader
	Writer  *binary.Writer
}

func NewPacketConn(conn net.Conn, protocol *PacketProtocol) *PacketConn {
	return &PacketConn{
		Conn:    conn,
		Codec:   protocol.CodecType.NewPacketCodec(),
		Spliter: protocol.Spliter,
		Reader:  binary.NewBufioReader(conn, protocol.ReadBufferSize),
		Writer:  binary.NewBufioWriter(conn, protocol.WriteBufferSize),
	}
}

func (conn *PacketConn) Send(msg interface{}) error {
	b, err := conn.Codec.EncodePacket(msg)
	if err != nil {
		return err
	}
	conn.Writer.WritePacket(b, conn.Spliter)
	return conn.Writer.Flush()
}

func (conn *PacketConn) Receive(msg interface{}) error {
	b := conn.Reader.ReadPacket(conn.Spliter)
	if conn.Reader.Error() != nil {
		return conn.Reader.Error()
	}
	return conn.Codec.DecodePacket(msg, b)
}

func (conn *PacketConn) Close() error         { return conn.Conn.Close() }
func (conn *PacketConn) LocalAddr() net.Addr  { return conn.Conn.LocalAddr() }
func (conn *PacketConn) RemoteAddr() net.Addr { return conn.Conn.RemoteAddr() }
