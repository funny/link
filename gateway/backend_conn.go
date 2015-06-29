package gateway

import (
	"io"
	"net"
	"sync/atomic"

	"github.com/funny/link"
)

var _ link.Conn = &BackendConn{}

type clientAddr struct {
	network []byte
	data    []byte
}

func (addr clientAddr) Network() string {
	return string(addr.network)
}

func (addr clientAddr) String() string {
	return string(addr.data)
}

type BackendConn struct {
	id        uint64
	waitId    uint64
	addr      clientAddr
	codec     link.PacketCodec
	link      *backendLink
	recvChan  chan []byte
	closeFlag int32
}

func newBackendConn(id, waitId uint64, addr clientAddr, codecType link.PacketCodecType, link *backendLink) *BackendConn {
	return &BackendConn{
		id:       id,
		waitId:   waitId,
		addr:     addr,
		codec:    codecType.NewPacketCodec(),
		link:     link,
		recvChan: make(chan []byte, 1024),
	}
}

func (c *BackendConn) LocalAddr() net.Addr {
	return c.link.session.Conn().LocalAddr()
}

func (c *BackendConn) RemoteAddr() net.Addr {
	return c.addr
}

func (c *BackendConn) Receive(msg interface{}) error {
	b, ok := <-c.recvChan
	if !ok {
		return io.EOF
	}
	return c.codec.DecodePacket(msg, b)
}

func (c *BackendConn) Send(msg interface{}) error {
	b, err := c.codec.EncodePacket(msg)
	if err != nil {
		return err
	}
	return c.link.session.Send(&gatewayMsg{
		Command: CMD_MSG, ClientId: c.id, Data: b,
	})
}

func (c *BackendConn) Close() error {
	c.close(true)
	return nil
}

func (c *BackendConn) close(feedback bool) bool {
	if atomic.CompareAndSwapInt32(&c.closeFlag, 0, 1) {
		go func() {
			if feedback {
				c.link.delConn(c.id)
				c.link.session.Send(&gatewayMsg{
					Command: CMD_DEL, ClientId: c.id,
				})
			}
			close(c.recvChan)
		}()
		return true
	}
	return false
}
