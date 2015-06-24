package link

import (
	"bufio"
	"io"
)

func Bytes() PSCodecType {
	return bytesCodecType{}
}

func String() PacketCodecType {
	return stringCodecType{}
}

type bytesCodecType struct{}

func (_ bytesCodecType) NewPacketCodec() PacketCodec {
	return bytesPacketCodec{}
}

func (_ bytesCodecType) NewStreamCodec(r *bufio.Reader, w *bufio.Writer) StreamCodec {
	return bytesStreamCodec{r, w}
}

type bytesPacketCodec struct{}

func (codec bytesPacketCodec) DecodePacket(msg interface{}, b []byte) error {
	*(msg.(*[]byte)) = b
	return nil
}

func (codec bytesPacketCodec) EncodePacket(msg interface{}) ([]byte, error) {
	return msg.([]byte), nil
}

type bytesStreamCodec struct {
	r *bufio.Reader
	w *bufio.Writer
}

func (codec bytesStreamCodec) DecodeStream(msg interface{}) error {
	_, err := io.ReadFull(codec.r, msg.([]byte))
	return err
}

func (codec bytesStreamCodec) EncodeStream(msg interface{}) error {
	if _, err := codec.w.Write(msg.([]byte)); err != nil {
		return err
	}
	return codec.w.Flush()
}

type stringCodecType struct{}

func (_ stringCodecType) NewPacketCodec() PacketCodec {
	return stringPacketCodec{}
}

type stringPacketCodec struct{}

func (codec stringPacketCodec) DecodePacket(msg interface{}, b []byte) error {
	*(msg.(*string)) = string(b)
	return nil
}

func (codec stringPacketCodec) EncodePacket(msg interface{}) ([]byte, error) {
	return []byte(msg.(string)), nil
}
