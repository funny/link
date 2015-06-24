package gateway

import (
	"github.com/funny/binary"
	"github.com/funny/link"
)

const (
	CMD_NEW_1 = 1
	CMD_NEW_2 = 2
	CMD_DEL   = 3
	CMD_MSG   = 4
	CMD_BRD   = 5
	CMD_PING  = 6
	CMD_PONG  = 7
)

var _ link.SelfEncode = &gatewayMsg{}
var _ link.SelfDecode = &gatewayMsg{}

type gatewayMsg struct {
	Command   uint8
	ClientId  uint64
	ClientIds []uint64
	Data      []byte
}

func (msg *gatewayMsg) BinaryDecode(r *binary.Reader) error {
	msg.Command = r.ReadUint8()
	switch msg.Command {
	case CMD_NEW_1:
		msg.ClientId = r.ReadUint64BE()
		msg.Data = r.ReadPacket(binary.SplitByUint8)
	case CMD_NEW_2:
		msg.ClientId = r.ReadUint64BE()
		msg.ClientIds = []uint64{r.ReadUint64BE()}
	case CMD_DEL:
		msg.ClientId = r.ReadUint64BE()
	case CMD_MSG:
		msg.ClientId = r.ReadUint64BE()
		msg.Data = r.ReadPacket(binary.SplitByUvarint)
	case CMD_BRD:
		num := int(r.ReadUvarint())
		msg.ClientIds = make([]uint64, num)
		for i := 0; i < num; i++ {
			msg.ClientIds[i] = r.ReadUvarint()
		}
		msg.Data = r.ReadPacket(binary.SplitByUvarint)
	}
	return nil
}

func (msg *gatewayMsg) BinaryEncode(w *binary.Writer) error {
	w.WriteUint8(msg.Command)
	switch msg.Command {
	case CMD_NEW_1:
		w.WriteUint64BE(msg.ClientId)
		w.WritePacket(msg.Data, binary.SplitByUint8)
	case CMD_NEW_2:
		w.WriteUint64BE(msg.ClientId)
		w.WriteUint64BE(msg.ClientIds[0])
	case CMD_DEL:
		w.WriteUint64BE(msg.ClientId)
	case CMD_MSG:
		w.WriteUint64BE(msg.ClientId)
		w.WritePacket(msg.Data, binary.SplitByUvarint)
	case CMD_BRD:
		w.WriteUvarint(uint64(len(msg.ClientIds)))
		for i := 0; i < len(msg.ClientIds); i++ {
			w.WriteUvarint(msg.ClientIds[i])
		}
		w.WritePacket(msg.Data, binary.SplitByUvarint)
	}
	return nil
}
