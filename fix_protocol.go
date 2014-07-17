package packnet

import (
	"encoding/binary"
	"net"
	"time"
)

type FixProtocol struct {
	n  uint
	bo binary.ByteOrder
}

func NewFixProtocol(n uint, bo binary.ByteOrder) *FixProtocol {
	return &FixProtocol{
		n:  n,
		bo: bo,
	}
}

func (p FixProtocol) NewWriter() PacketWriter {
	return NewFixWriter(p.n, p.bo)
}

func (p FixProtocol) NewReader() PacketReader {
	return NewFixReader(p.n, p.bo)
}

type FixWriter struct {
	n       uint
	bo      binary.ByteOrder
	timeout time.Duration
	maxsize uint
}

func NewFixWriter(n uint, bo binary.ByteOrder) *FixWriter {
	return &FixWriter{
		n:  n,
		bo: bo,
	}
}

func (w *FixWriter) BeginPacket(size uint, buff []byte) []byte {
	packetLen := w.n + size
	if uint(cap(buff)) < packetLen {
		return make([]byte, w.n, packetLen)
	}
	return buff[0:w.n:packetLen]
}

func (w *FixWriter) EndPacket(packet []byte) []byte {
	size := uint(len(packet)) - w.n

	if w.maxsize > 0 && size > w.maxsize {
		panic("too large packet")
	}

	switch w.n {
	case 1:
		packet[0] = byte(size)
	case 2:
		w.bo.PutUint16(packet, uint16(size))
	case 4:
		w.bo.PutUint32(packet, uint32(size))
	case 8:
		w.bo.PutUint64(packet, uint64(size))
	default:
		panic("unsupported packet head size")
	}

	return packet
}

func (w *FixWriter) WritePacket(conn net.Conn, packet []byte) error {
	if w.timeout > 0 {
		conn.SetReadDeadline(time.Now().Add(w.timeout))
	} else {
		conn.SetReadDeadline(time.Time{})
	}

	if _, err := conn.Write(packet); err != nil {
		return err
	}

	return nil
}

func (w *FixWriter) SetTimeout(timeout time.Duration) {
	w.timeout = timeout
}

func (w *FixWriter) SetMaxSize(maxsize uint) {
	w.maxsize = maxsize
}

type FixReader struct {
	n       uint
	bo      binary.ByteOrder
	head    []byte
	timeout time.Duration
	maxsize uint
}

func NewFixReader(n uint, bo binary.ByteOrder) *FixReader {
	return &FixReader{
		n:    n,
		bo:   bo,
		head: make([]byte, n),
	}
}

func (r *FixReader) ReadPacket(conn net.Conn, b []byte) ([]byte, error) {
	if r.timeout > 0 {
		conn.SetReadDeadline(time.Now().Add(r.timeout))
	}

	if _, err := conn.Read(r.head); err != nil {
		return nil, err
	}

	size := uint(0)

	switch r.n {
	case 1:
		size = uint(r.head[0])
	case 2:
		size = uint(r.bo.Uint16(r.head))
	case 4:
		size = uint(r.bo.Uint32(r.head))
	case 8:
		size = uint(r.bo.Uint64(r.head))
	default:
		panic("unsupported packet head size")
	}

	if r.maxsize > 0 && size > r.maxsize {
		return nil, PacketTooLargeError
	}

	var data []byte

	if uint(cap(b)) >= size {
		data = b[0:size]
	} else {
		data = make([]byte, size)
	}

	if r.timeout > 0 {
		conn.SetReadDeadline(time.Now().Add(r.timeout))
	}

	_, err := conn.Read(data)

	return data, err
}

func (r *FixReader) SetTimeout(timeout time.Duration) {
	r.timeout = timeout
}

func (r *FixReader) SetMaxSize(maxsize uint) {
	r.maxsize = maxsize
}
