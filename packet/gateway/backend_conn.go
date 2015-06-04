package gateway

import (
	"io"
	"net"
	"sync/atomic"

	"github.com/funny/link"
	"github.com/funny/link/packet"
)

type RAW []byte

func (r *RAW) Unmarshal(b []byte) error {
	*r = b
	return nil
}

func (r RAW) Marshal() ([]byte, error) {
	return r, nil
}

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
	addr      clientAddr
	link      *backendLink
	recvChan  chan []byte
	closeFlag int32
}

func newBackendConn(id uint64, addr clientAddr, link *backendLink) *BackendConn {
	return &BackendConn{
		id:       id,
		addr:     addr,
		link:     link,
		recvChan: make(chan []byte, 1024),
	}
}

func (c *BackendConn) Config() link.SessionConfig {
	return link.SessionConfig{
		1024,
	}
}

func (c *BackendConn) LocalAddr() net.Addr {
	return c.link.session.Conn().LocalAddr()
}

func (c *BackendConn) RemoteAddr() net.Addr {
	return c.addr
}

func (c *BackendConn) Receive(msg interface{}) error {
	data, ok := <-c.recvChan
	if !ok {
		return io.EOF
	}
	return msg.(packet.InMessage).Unmarshal(data)
}

func (c *BackendConn) Send(msg interface{}) error {
	data, err := msg.(packet.OutMessage).Marshal()
	if err != nil {
		return err
	}
	return c.link.session.Send(&gatewayMsg{
		Command: CMD_MSG, ClientId: c.id, Data: data,
	})
}

func (c *BackendConn) Close() error {
	return c.close(true)
}

func (c *BackendConn) close(feedback bool) error {
	if atomic.CompareAndSwapInt32(&c.closeFlag, 0, 1) {
		if feedback {
			c.link.delConn(c.id)
			c.link.session.Send(&gatewayMsg{
				Command: CMD_DEL, ClientId: c.id,
			})
		}
		close(c.recvChan)
	}
	return nil
}
