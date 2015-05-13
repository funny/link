package fixhead

import (
	"encoding/binary"
	"github.com/funny/link"
	"github.com/funny/link/linkutil"
	"io"
)

var _ link.Protocol = &protocol{}
var _ link.Codec = &protocol{}

type protocol struct {
	n       int
	decoder func(*link.Buffer) int
	encoder func([]byte, int)
}

func (protocol *protocol) NewCodec() link.Codec {
	return protocol
}

func (codec *protocol) Handshake(rw io.ReadWriter, buf *link.Buffer) error {
	return nil
}

func (codec *protocol) Prepend(buf *link.Buffer, msg link.Message) {
	size := 1024
	if sizeable, ok := msg.(link.Sizeable); ok {
		size = sizeable.BufferSize()
	}
	buf.Reset(codec.n, codec.n+size)
}

func (codec *protocol) Write(w io.Writer, buf *link.Buffer) error {
	size := len(buf.Data) - codec.n
	codec.encoder(buf.Data, size)
	_, err := w.Write(buf.Data)
	return err
}

func (codec *protocol) Read(r io.Reader, buf *link.Buffer) error {
	buf.Reset(codec.n, codec.n)
	if _, err := io.ReadFull(r, buf.Data); err != nil {
		return err
	}
	size := codec.decoder(buf)
	buf.Reset(size, size)
	if _, err := io.ReadFull(r, buf.Data); err != nil {
		return err
	}
	return nil
}

var (
	Uint8 = &protocol{1,
		func(buf *link.Buffer) int {
			return int(buf.ReadUint8())
		},
		func(buf []byte, size int) {
			buf[0] = byte(size)
		},
	}
	Uint16BE = &protocol{2,
		func(buf *link.Buffer) int {
			return int(buf.ReadUint16BE())
		},
		func(buf []byte, size int) {
			binary.BigEndian.PutUint16(buf, uint16(size))
		},
	}
	Uint16LE = &protocol{2,
		func(buf *link.Buffer) int {
			return int(buf.ReadUint16LE())
		},
		func(buf []byte, size int) {
			binary.LittleEndian.PutUint16(buf, uint16(size))
		},
	}
	Uint24BE = &protocol{3,
		func(buf *link.Buffer) int {
			return int(buf.ReadUint24BE())
		},
		func(buf []byte, size int) {
			linkutil.PutUint24BE(buf, uint32(size))
		},
	}
	Uint24LE = &protocol{3,
		func(buf *link.Buffer) int {
			return int(buf.ReadUint24LE())
		},
		func(buf []byte, size int) {
			linkutil.PutUint24LE(buf, uint32(size))
		},
	}
	Uint32BE = &protocol{4,
		func(buf *link.Buffer) int {
			return int(buf.ReadUint32BE())
		},
		func(buf []byte, size int) {
			binary.BigEndian.PutUint32(buf, uint32(size))
		},
	}
	Uint32LE = &protocol{4,
		func(buf *link.Buffer) int {
			return int(buf.ReadUint32LE())
		},
		func(buf []byte, size int) {
			binary.LittleEndian.PutUint32(buf, uint32(size))
		},
	}
	Uint40BE = &protocol{5,
		func(buf *link.Buffer) int {
			return int(buf.ReadUint40BE())
		},
		func(buf []byte, size int) {
			linkutil.PutUint40BE(buf, uint64(size))
		},
	}
	Uint40LE = &protocol{5,
		func(buf *link.Buffer) int {
			return int(buf.ReadUint40LE())
		},
		func(buf []byte, size int) {
			linkutil.PutUint40LE(buf, uint64(size))
		},
	}
	Uint48BE = &protocol{6,
		func(buf *link.Buffer) int {
			return int(buf.ReadUint48BE())
		},
		func(buf []byte, size int) {
			linkutil.PutUint48BE(buf, uint64(size))
		},
	}
	Uint48LE = &protocol{6,
		func(buf *link.Buffer) int {
			return int(buf.ReadUint48LE())
		},
		func(buf []byte, size int) {
			linkutil.PutUint48LE(buf, uint64(size))
		},
	}
	Uint56BE = &protocol{7,
		func(buf *link.Buffer) int {
			return int(buf.ReadUint56BE())
		},
		func(buf []byte, size int) {
			linkutil.PutUint56BE(buf, uint64(size))
		},
	}
	Uint56LE = &protocol{7,
		func(buf *link.Buffer) int {
			return int(buf.ReadUint56LE())
		},
		func(buf []byte, size int) {
			linkutil.PutUint56LE(buf, uint64(size))
		},
	}
	Uint64BE = &protocol{8,
		func(buf *link.Buffer) int {
			return int(buf.ReadUint64BE())
		},
		func(buf []byte, size int) {
			binary.BigEndian.PutUint64(buf, uint64(size))
		},
	}
	Uint64LE = &protocol{8,
		func(buf *link.Buffer) int {
			return int(buf.ReadUint64LE())
		},
		func(buf []byte, size int) {
			binary.LittleEndian.PutUint64(buf, uint64(size))
		},
	}
)
