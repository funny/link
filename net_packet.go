package link

import (
	"io"
)

func (conn *Conn) ReadPacket(spliter Spliter) (b []byte) {
	if conn.rerr != nil {
		return nil
	}
	b, conn.rerr = spliter.Read(conn)
	return
}

func (conn *Conn) WritePacket(b []byte, spliter Spliter) {
	if conn.werr != nil {
		return
	}
	conn.werr = spliter.Write(conn, b)
}

type Spliter interface {
	Read(conn *Conn) ([]byte, error)
	Write(conn *Conn, b []byte) error
}

type Limiter interface {
	Limit(conn *Conn) *io.LimitedReader
}

type DelimSpliter struct {
	delim byte
}

func SplitByDelim(delim byte) DelimSpliter {
	return DelimSpliter{delim}
}

func (s DelimSpliter) Read(conn *Conn) ([]byte, error) {
	b, err := conn.r.ReadBytes(s.delim)
	if len(b) > 0 {
		b = b[:len(b)-1]
	}
	return b, err
}

func (s DelimSpliter) Write(conn *Conn, b []byte) error {
	if _, err := conn.w.Write(b); err != nil {
		return err
	}
	return conn.w.WriteByte(s.delim)
}

type HeadSpliter struct {
	ReadHead  func(conn *Conn) int
	WriteHead func(conn *Conn, l int)
}

func (s HeadSpliter) Read(conn *Conn) ([]byte, error) {
	n := s.ReadHead(conn)
	if conn.rerr != nil {
		return nil, conn.rerr
	}
	b := make([]byte, n)
	_, err := io.ReadFull(conn.r, b)
	return b, err
}

func (s HeadSpliter) Write(conn *Conn, b []byte) error {
	s.WriteHead(conn, len(b))
	if conn.werr != nil {
		return conn.werr
	}
	_, err := conn.w.Write(b)
	return err
}

func (s HeadSpliter) Limit(conn *Conn) *io.LimitedReader {
	n := s.ReadHead(conn)
	return &io.LimitedReader{conn.Reader(), int64(n)}
}

var (
	SplitByLine = SplitByDelim('\n')

	SplitByUvarint = HeadSpliter{
		ReadHead:  func(conn *Conn) int { return int(conn.ReadUvarint()) },
		WriteHead: func(conn *Conn, l int) { conn.WriteUvarint(uint64(l)) },
	}

	SplitByUint8 = HeadSpliter{
		ReadHead:  func(conn *Conn) int { return int(conn.ReadUint8()) },
		WriteHead: func(conn *Conn, l int) { conn.WriteUint8(uint8(l)) },
	}

	SplitByUint16BE = HeadSpliter{
		ReadHead:  func(conn *Conn) int { return int(conn.ReadUint16BE()) },
		WriteHead: func(conn *Conn, l int) { conn.WriteUint16BE(uint16(l)) },
	}
	SplitByUint16LE = HeadSpliter{
		ReadHead:  func(conn *Conn) int { return int(conn.ReadUint16LE()) },
		WriteHead: func(conn *Conn, l int) { conn.WriteUint16LE(uint16(l)) },
	}

	SplitByUint24BE = HeadSpliter{
		ReadHead:  func(conn *Conn) int { return int(conn.ReadUint24BE()) },
		WriteHead: func(conn *Conn, l int) { conn.WriteUint24BE(uint32(l)) },
	}
	SplitByUint24LE = HeadSpliter{
		ReadHead:  func(conn *Conn) int { return int(conn.ReadUint24LE()) },
		WriteHead: func(conn *Conn, l int) { conn.WriteUint24LE(uint32(l)) },
	}

	SplitByUint32BE = HeadSpliter{
		ReadHead:  func(conn *Conn) int { return int(conn.ReadUint32BE()) },
		WriteHead: func(conn *Conn, l int) { conn.WriteUint32BE(uint32(l)) },
	}
	SplitByUint32LE = HeadSpliter{
		ReadHead:  func(conn *Conn) int { return int(conn.ReadUint32LE()) },
		WriteHead: func(conn *Conn, l int) { conn.WriteUint32LE(uint32(l)) },
	}

	SplitByUint40BE = HeadSpliter{
		ReadHead:  func(conn *Conn) int { return int(conn.ReadUint40BE()) },
		WriteHead: func(conn *Conn, l int) { conn.WriteUint40BE(uint64(l)) },
	}
	SplitByUint40LE = HeadSpliter{
		ReadHead:  func(conn *Conn) int { return int(conn.ReadUint40LE()) },
		WriteHead: func(conn *Conn, l int) { conn.WriteUint40LE(uint64(l)) },
	}

	SplitByUint48BE = HeadSpliter{
		ReadHead:  func(conn *Conn) int { return int(conn.ReadUint48BE()) },
		WriteHead: func(conn *Conn, l int) { conn.WriteUint48BE(uint64(l)) },
	}
	SplitByUint48LE = HeadSpliter{
		ReadHead:  func(conn *Conn) int { return int(conn.ReadUint48LE()) },
		WriteHead: func(conn *Conn, l int) { conn.WriteUint48LE(uint64(l)) },
	}

	SplitByUint56BE = HeadSpliter{
		ReadHead:  func(conn *Conn) int { return int(conn.ReadUint56BE()) },
		WriteHead: func(conn *Conn, l int) { conn.WriteUint56BE(uint64(l)) },
	}
	SplitByUint56LE = HeadSpliter{
		ReadHead:  func(conn *Conn) int { return int(conn.ReadUint56LE()) },
		WriteHead: func(conn *Conn, l int) { conn.WriteUint56LE(uint64(l)) },
	}

	SplitByUint64BE = HeadSpliter{
		ReadHead:  func(conn *Conn) int { return int(conn.ReadUint64BE()) },
		WriteHead: func(conn *Conn, l int) { conn.WriteUint64BE(uint64(l)) },
	}
	SplitByUint64LE = HeadSpliter{
		ReadHead:  func(conn *Conn) int { return int(conn.ReadUint64LE()) },
		WriteHead: func(conn *Conn, l int) { conn.WriteUint64LE(uint64(l)) },
	}
)
