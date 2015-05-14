package fixhead

import (
	"encoding/binary"
	"github.com/funny/link"
	"github.com/funny/link/linkutil"
)

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
