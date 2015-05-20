package link

import (
	"io"
)

var (
	DefaultConfig = Config{
		ConnConfig{
			ReadBufferSize:  2048,
			WriteBufferSize: 2048,
		},
		SessionConfig{
			AutoFlush:         true,
			AsyncSendTimeout:  0,
			AsyncSendChanSize: 1000,
		},
	}
)

type Reader interface {
	io.Reader
	io.ByteReader
	io.RuneReader

	ReadPacket(Spliter) []byte
	ReadBytes(n int) []byte
	ReadString(n int) string
	ReadUvarint() (v uint64)
	ReadVarint() (v int64)
	ReadUint8() (v uint8)
	ReadUint16BE() uint16
	ReadUint16LE() uint16
	ReadUint24BE() uint32
	ReadUint24LE() uint32
	ReadUint32BE() uint32
	ReadUint32LE() uint32
	ReadUint40BE() uint64
	ReadUint40LE() uint64
	ReadUint48BE() uint64
	ReadUint48LE() uint64
	ReadUint56BE() uint64
	ReadUint56LE() uint64
	ReadUint64BE() uint64
	ReadUint64LE() uint64
	ReadFloat32BE() float32
	ReadFloat32LE() float32
	ReadFloat64BE() float64
	ReadFloat64LE() float64
	Delim(byte) []byte

	ReaderError() error
}

type Writer interface {
	io.Writer
	io.ByteWriter

	WritePacket([]byte, Spliter)
	WriteBytes(b []byte)
	WriteString(s string)
	WriteUvarint(v uint64)
	WriteVarint(v int64)
	WriteUint8(v uint8)
	WriteUint16BE(v uint16)
	WriteUint16LE(v uint16)
	WriteUint24BE(v uint32)
	WriteUint24LE(v uint32)
	WriteUint32BE(v uint32)
	WriteUint32LE(v uint32)
	WriteUint40BE(v uint64)
	WriteUint40LE(v uint64)
	WriteUint48BE(v uint64)
	WriteUint48LE(v uint64)
	WriteUint56BE(v uint64)
	WriteUint56LE(v uint64)
	WriteUint64BE(v uint64)
	WriteUint64LE(v uint64)
	WriteFloat32BE(v float32)
	WriteFloat32LE(v float32)
	WriteFloat64BE(v float64)
	WriteFloat64LE(v float64)
	WriteRune(r rune)

	WriterError() error
}

type OutMessage interface {
	Send(Writer) error
}

type InMessage interface {
	Receive(Reader) error
}

type Spliter interface {
	Read(Reader) []byte
	Write(Writer, []byte)
}
