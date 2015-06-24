package link

import (
	"bufio"
	"io"
)

func Raw() PSCodecType {
	return rawCodecType{}
}

type rawCodecType struct{}

func (_ rawCodecType) NewPacketCodec() PacketCodec {
	return rawPacketCodec{}
}

func (_ rawCodecType) NewStreamCodec(r *bufio.Reader, w *bufio.Writer) StreamCodec {
	return rawStreamCodec{r, w}
}

type rawPacketCodec struct{}

func (codec rawPacketCodec) DecodePacket(msg interface{}, b []byte) error {
	*(msg.(*[]byte)) = b
	return nil
}

func (codec rawPacketCodec) EncodePacket(msg interface{}) ([]byte, error) {
	return msg.([]byte), nil
}

type rawStreamCodec struct {
	r *bufio.Reader
	w *bufio.Writer
}

func (codec rawStreamCodec) DecodeStream(msg interface{}) error {
	_, err := io.ReadFull(codec.r, msg.([]byte))
	return err
}

func (codec rawStreamCodec) EncodeStream(msg interface{}) error {
	if _, err := codec.w.Write(msg.([]byte)); err != nil {
		return err
	}
	return codec.w.Flush()
}
