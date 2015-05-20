package link

import (
	"io"
)

type DelimSpliter struct {
	delim byte
}

func SplitByDelim(delim byte) DelimSpliter {
	return DelimSpliter{delim}
}

func (s DelimSpliter) Read(r Reader) []byte {
	b := r.Delim(s.delim)
	if len(b) > 0 {
		b = b[:len(b)-1]
	}
	return b
}

func (s DelimSpliter) Write(w Writer, b []byte) {
	if _, err := w.Write(b); err != nil {
		return
	}
	w.WriteByte(s.delim)
}

type HeadSpliter struct {
	ReadHead  func(r Reader) int
	WriteHead func(w Writer, l int)
}

func (s HeadSpliter) Read(r Reader) []byte {
	n := s.ReadHead(r)
	if r.ReaderError() != nil {
		return nil
	}
	b := make([]byte, n)
	if _, err := io.ReadFull(r, b); err != nil {
		return nil
	}
	return b
}

func (s HeadSpliter) Write(w Writer, b []byte) {
	s.WriteHead(w, len(b))
	if w.WriterError() != nil {
		return
	}
	w.Write(b)
}

func (s HeadSpliter) Limit(r Reader) *io.LimitedReader {
	n := s.ReadHead(r)
	return &io.LimitedReader{r, int64(n)}
}

var (
	SplitByLine = SplitByDelim('\n')

	SplitByUvarint = HeadSpliter{
		ReadHead:  func(r Reader) int { return int(r.ReadUvarint()) },
		WriteHead: func(w Writer, l int) { w.WriteUvarint(uint64(l)) },
	}

	SplitByUint8 = HeadSpliter{
		ReadHead:  func(r Reader) int { return int(r.ReadUint8()) },
		WriteHead: func(w Writer, l int) { w.WriteUint8(uint8(l)) },
	}

	SplitByUint16BE = HeadSpliter{
		ReadHead:  func(r Reader) int { return int(r.ReadUint16BE()) },
		WriteHead: func(w Writer, l int) { w.WriteUint16BE(uint16(l)) },
	}
	SplitByUint16LE = HeadSpliter{
		ReadHead:  func(r Reader) int { return int(r.ReadUint16LE()) },
		WriteHead: func(w Writer, l int) { w.WriteUint16LE(uint16(l)) },
	}

	SplitByUint24BE = HeadSpliter{
		ReadHead:  func(r Reader) int { return int(r.ReadUint24BE()) },
		WriteHead: func(w Writer, l int) { w.WriteUint24BE(uint32(l)) },
	}
	SplitByUint24LE = HeadSpliter{
		ReadHead:  func(r Reader) int { return int(r.ReadUint24LE()) },
		WriteHead: func(w Writer, l int) { w.WriteUint24LE(uint32(l)) },
	}

	SplitByUint32BE = HeadSpliter{
		ReadHead:  func(r Reader) int { return int(r.ReadUint32BE()) },
		WriteHead: func(w Writer, l int) { w.WriteUint32BE(uint32(l)) },
	}
	SplitByUint32LE = HeadSpliter{
		ReadHead:  func(r Reader) int { return int(r.ReadUint32LE()) },
		WriteHead: func(w Writer, l int) { w.WriteUint32LE(uint32(l)) },
	}

	SplitByUint40BE = HeadSpliter{
		ReadHead:  func(r Reader) int { return int(r.ReadUint40BE()) },
		WriteHead: func(w Writer, l int) { w.WriteUint40BE(uint64(l)) },
	}
	SplitByUint40LE = HeadSpliter{
		ReadHead:  func(r Reader) int { return int(r.ReadUint40LE()) },
		WriteHead: func(w Writer, l int) { w.WriteUint40LE(uint64(l)) },
	}

	SplitByUint48BE = HeadSpliter{
		ReadHead:  func(r Reader) int { return int(r.ReadUint48BE()) },
		WriteHead: func(w Writer, l int) { w.WriteUint48BE(uint64(l)) },
	}
	SplitByUint48LE = HeadSpliter{
		ReadHead:  func(r Reader) int { return int(r.ReadUint48LE()) },
		WriteHead: func(w Writer, l int) { w.WriteUint48LE(uint64(l)) },
	}

	SplitByUint56BE = HeadSpliter{
		ReadHead:  func(r Reader) int { return int(r.ReadUint56BE()) },
		WriteHead: func(w Writer, l int) { w.WriteUint56BE(uint64(l)) },
	}
	SplitByUint56LE = HeadSpliter{
		ReadHead:  func(r Reader) int { return int(r.ReadUint56LE()) },
		WriteHead: func(w Writer, l int) { w.WriteUint56LE(uint64(l)) },
	}

	SplitByUint64BE = HeadSpliter{
		ReadHead:  func(r Reader) int { return int(r.ReadUint64BE()) },
		WriteHead: func(w Writer, l int) { w.WriteUint64BE(uint64(l)) },
	}
	SplitByUint64LE = HeadSpliter{
		ReadHead:  func(r Reader) int { return int(r.ReadUint64LE()) },
		WriteHead: func(w Writer, l int) { w.WriteUint64LE(uint64(l)) },
	}
)
