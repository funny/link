package link

import (
	"bufio"

	"github.com/funny/binary"
)

func SelfCodec() PSCodecType {
	return selfCodecType{}
}

type selfCodecType struct{}

func (_ selfCodecType) NewPacketCodec() PacketCodec {
	codec := &selfPacketCodec{}
	codec.r = binary.NewReader(&codec.rbuf)
	codec.w = binary.NewWriter(&codec.wbuf)
	return codec
}

func (_ selfCodecType) NewStreamCodec(r *bufio.Reader, w *bufio.Writer) StreamCodec {
	return selfStreamCodec{
		binary.NewReader(r),
		binary.NewWriter(w),
	}
}

type SelfDecoder interface {
	SelfDecode(*binary.Reader) error
}

type SelfEncoder interface {
	SelfEncode(*binary.Writer) error
}

type selfPacketCodec struct {
	rbuf binary.Buffer
	wbuf binary.Buffer
	r    *binary.Reader
	w    *binary.Writer
}

func (codec *selfPacketCodec) DecodePacket(msg interface{}, b []byte) error {
	codec.rbuf.Reset(b)
	return msg.(SelfDecoder).SelfDecode(codec.r)
}

func (codec *selfPacketCodec) EncodePacket(msg interface{}) ([]byte, error) {
	codec.wbuf.Reset(codec.wbuf.Data[0:0])
	if err := msg.(SelfEncoder).SelfEncode(codec.w); err != nil {
		return nil, err
	}
	return codec.wbuf.Bytes(), nil
}

type selfStreamCodec struct {
	r *binary.Reader
	w *binary.Writer
}

func (codec selfStreamCodec) DecodeStream(msg interface{}) error {
	if err := msg.(SelfDecoder).SelfDecode(codec.r); err != nil {
		return err
	}
	return codec.r.Error()
}

func (codec selfStreamCodec) EncodeStream(msg interface{}) error {
	if err := msg.(SelfEncoder).SelfEncode(codec.w); err != nil {
		return err
	}
	return codec.w.Flush()
}
