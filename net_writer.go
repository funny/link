package link

import (
	"github.com/funny/binary"
)

func (conn *Conn) Flush() {
	if conn.werr != nil {
		return
	}
	conn.w.Flush()
}

func (conn *Conn) WriteBytes(b []byte) {
	if conn.werr != nil {
		return
	}
	_, conn.werr = conn.w.Write(b)
}

func (conn *Conn) WriteString(s string) {
	conn.WriteBytes([]byte(s))
}

func (conn *Conn) WriteUvarint(v uint64) {
	if conn.werr != nil {
		return
	}
	b := conn.b[:]
	n := binary.PutUvarint(b, v)
	b = b[:n]
	_, conn.werr = conn.w.Write(b)
}

func (conn *Conn) WriteVarint(v int64) {
	if conn.werr != nil {
		return
	}
	b := conn.b[:]
	n := binary.PutVarint(b, v)
	b = b[:n]
	_, conn.werr = conn.w.Write(b)
}

func (conn *Conn) WriteUint8(v uint8) {
	if conn.werr != nil {
		return
	}
	conn.werr = conn.w.WriteByte(byte(v))
}

func (conn *Conn) WriteUint16BE(v uint16) {
	if conn.werr != nil {
		return
	}
	b := conn.b[:2]
	binary.PutUint16BE(b, v)
	_, conn.werr = conn.w.Write(b)
}

func (conn *Conn) WriteUint16LE(v uint16) {
	if conn.werr != nil {
		return
	}
	b := conn.b[:2]
	binary.PutUint16LE(b, v)
	_, conn.werr = conn.w.Write(b)
}

func (conn *Conn) WriteUint24BE(v uint32) {
	if conn.werr != nil {
		return
	}
	b := conn.b[:3]
	binary.PutUint24BE(b, v)
	_, conn.werr = conn.w.Write(b)
}

func (conn *Conn) WriteUint24LE(v uint32) {
	if conn.werr != nil {
		return
	}
	b := conn.b[:3]
	binary.PutUint24LE(b, v)
	_, conn.werr = conn.w.Write(b)
}

func (conn *Conn) WriteUint32BE(v uint32) {
	if conn.werr != nil {
		return
	}
	b := conn.b[:4]
	binary.PutUint32BE(b, v)
	_, conn.werr = conn.w.Write(b)
}

func (conn *Conn) WriteUint32LE(v uint32) {
	if conn.werr != nil {
		return
	}
	b := conn.b[:4]
	binary.PutUint32LE(b, v)
	_, conn.werr = conn.w.Write(b)
}

func (conn *Conn) WriteUint40BE(v uint64) {
	if conn.werr != nil {
		return
	}
	b := conn.b[:5]
	binary.PutUint40BE(b, v)
	_, conn.werr = conn.w.Write(b)
}

func (conn *Conn) WriteUint40LE(v uint64) {
	if conn.werr != nil {
		return
	}
	b := conn.b[:5]
	binary.PutUint40LE(b, v)
	_, conn.werr = conn.w.Write(b)
}

func (conn *Conn) WriteUint48BE(v uint64) {
	if conn.werr != nil {
		return
	}
	b := conn.b[:6]
	binary.PutUint48BE(b, v)
	_, conn.werr = conn.w.Write(b)
}

func (conn *Conn) WriteUint48LE(v uint64) {
	if conn.werr != nil {
		return
	}
	b := conn.b[:6]
	binary.PutUint48LE(b, v)
	_, conn.werr = conn.w.Write(b)
}

func (conn *Conn) WriteUint56BE(v uint64) {
	if conn.werr != nil {
		return
	}
	b := conn.b[:7]
	binary.PutUint56BE(b, v)
	_, conn.werr = conn.w.Write(b)
}

func (conn *Conn) WriteUint56LE(v uint64) {
	if conn.werr != nil {
		return
	}
	b := conn.b[:7]
	binary.PutUint56LE(b, v)
	_, conn.werr = conn.w.Write(b)
}

func (conn *Conn) WriteUint64BE(v uint64) {
	if conn.werr != nil {
		return
	}
	b := conn.b[:8]
	binary.PutUint64BE(b, v)
	_, conn.werr = conn.w.Write(b)
}

func (conn *Conn) WriteUint64LE(v uint64) {
	if conn.werr != nil {
		return
	}
	b := conn.b[:8]
	binary.PutUint64LE(b, v)
	_, conn.werr = conn.w.Write(b)
}

func (conn *Conn) WriteFloat32BE(v float32) {
	if conn.werr != nil {
		return
	}
	b := conn.b[:4]
	binary.PutFloat32BE(b, v)
	_, conn.werr = conn.w.Write(b)
}

func (conn *Conn) WriteFloat32LE(v float32) {
	if conn.werr != nil {
		return
	}
	b := conn.b[:4]
	binary.PutFloat32LE(b, v)
	_, conn.werr = conn.w.Write(b)
}

func (conn *Conn) WriteFloat64BE(v float64) {
	if conn.werr != nil {
		return
	}
	b := conn.b[:8]
	binary.PutFloat64BE(b, v)
	_, conn.werr = conn.w.Write(b)
}

func (conn *Conn) WriteFloat64LE(v float64) {
	if conn.werr != nil {
		return
	}
	b := conn.b[:8]
	binary.PutFloat64LE(b, v)
	_, conn.werr = conn.w.Write(b)
}
