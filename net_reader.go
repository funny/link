package link

import (
	"github.com/funny/binary"
	"io"
	"math"
)

func (conn *Conn) ReadFull(b []byte) (n int, err error) {
	if conn.rerr != nil {
		return 0, conn.rerr
	}
	n, err = io.ReadFull(conn.r, b)
	conn.rerr = err
	return
}

func (conn *Conn) ReadBytes(n int) (b []byte) {
	b = make([]byte, n)
	nn, _ := conn.ReadFull(b)
	return b[:nn]
}

func (conn *Conn) ReadString(n int) string {
	return string(conn.ReadBytes(n))
}

func (conn *Conn) ReadUvarint() (v uint64) {
	if conn.rerr != nil {
		return
	}
	v, conn.rerr = binary.ReadUvarint(conn.r)
	return
}

func (conn *Conn) ReadVarint() (v int64) {
	if conn.rerr != nil {
		return
	}
	v, conn.rerr = binary.ReadVarint(conn.r)
	return
}

func (conn *Conn) ReadUint8() (v uint8) {
	if conn.rerr != nil {
		return
	}
	v, conn.rerr = conn.r.ReadByte()
	return
}

func (conn *Conn) seek(n int) []byte {
	if conn.rerr != nil {
		return nil
	}
	b := conn.b[:n]
	_, conn.rerr = io.ReadFull(conn.r, b)
	return b
}

func (conn *Conn) ReadUint16BE() uint16 {
	b := conn.seek(2)
	if conn.rerr != nil {
		return 0
	}
	return binary.GetUint16BE(b)
}

func (conn *Conn) ReadUint16LE() uint16 {
	b := conn.seek(2)
	if conn.rerr != nil {
		return 0
	}
	return binary.GetUint16LE(b)
}

func (conn *Conn) ReadUint24BE() uint32 {
	b := conn.seek(3)
	if conn.rerr != nil {
		return 0
	}
	return binary.GetUint24BE(b)
}

func (conn *Conn) ReadUint24LE() uint32 {
	b := conn.seek(3)
	if conn.rerr != nil {
		return 0
	}
	return binary.GetUint24LE(b)
}

func (conn *Conn) ReadUint32BE() uint32 {
	b := conn.seek(4)
	if conn.rerr != nil {
		return 0
	}
	return binary.GetUint32BE(b)
}

func (conn *Conn) ReadUint32LE() uint32 {
	b := conn.seek(4)
	if conn.rerr != nil {
		return 0
	}
	return binary.GetUint32LE(b)
}

func (conn *Conn) ReadUint40BE() uint64 {
	b := conn.seek(5)
	if conn.rerr != nil {
		return 0
	}
	return binary.GetUint40BE(b)
}

func (conn *Conn) ReadUint40LE() uint64 {
	b := conn.seek(5)
	if conn.rerr != nil {
		return 0
	}
	return binary.GetUint40LE(b)
}

func (conn *Conn) ReadUint48BE() uint64 {
	b := conn.seek(6)
	if conn.rerr != nil {
		return 0
	}
	return binary.GetUint48BE(b)
}

func (conn *Conn) ReadUint48LE() uint64 {
	b := conn.seek(6)
	if conn.rerr != nil {
		return 0
	}
	return binary.GetUint48LE(b)
}

func (conn *Conn) ReadUint56BE() uint64 {
	b := conn.seek(7)
	if conn.rerr != nil {
		return 0
	}
	return binary.GetUint56BE(b)
}

func (conn *Conn) ReadUint56LE() uint64 {
	b := conn.seek(7)
	if conn.rerr != nil {
		return 0
	}
	return binary.GetUint56LE(b)
}

func (conn *Conn) ReadUint64BE() uint64 {
	b := conn.seek(8)
	if conn.rerr != nil {
		return 0
	}
	return binary.GetUint64BE(b)
}

func (conn *Conn) ReadUint64LE() uint64 {
	b := conn.seek(8)
	if conn.rerr != nil {
		return 0
	}
	return binary.GetUint64LE(b)
}

func (conn *Conn) ReadFloat32BE() float32 {
	b := conn.seek(4)
	if conn.rerr != nil {
		return float32(math.NaN())
	}
	return binary.GetFloat32BE(b)
}

func (conn *Conn) ReadFloat32LE() float32 {
	b := conn.seek(4)
	if conn.rerr != nil {
		return float32(math.NaN())
	}
	return binary.GetFloat32LE(b)
}

func (conn *Conn) ReadFloat64BE() float64 {
	b := conn.seek(8)
	if conn.rerr != nil {
		return math.NaN()
	}
	return binary.GetFloat64BE(b)
}

func (conn *Conn) ReadFloat64LE() float64 {
	b := conn.seek(8)
	if conn.rerr != nil {
		return math.NaN()
	}
	return binary.GetFloat64LE(b)
}
