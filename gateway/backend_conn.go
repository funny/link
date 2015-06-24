package gateway

import (
	"io"
	"net"
	"sync/atomic"

	"github.com/funny/link"
)

var _ link.IPacketConn = &BackendConn{}

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
	link      *backendLink
	recvChan  chan []byte
	closeFlag int32

	readCount  uint32
	writeCount uint32
}

func newBackendConn(id, waitId uint64, addr clientAddr, link *backendLink) *BackendConn {
	return &BackendConn{
		id:       id,
		waitId:   waitId,
		addr:     addr,
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

func (c *BackendConn) ReadPacket() ([]byte, error) {
	data, ok := <-c.recvChan
	if !ok {
		return nil, io.EOF
	}
	atomic.AddUint32(&c.readCount, 1)
	return data, nil
}

func (c *BackendConn) WritePacket(msg []byte) error {
	atomic.AddUint32(&c.writeCount, 1)
	return c.link.session.Send(&gatewayMsg{
		Command: CMD_MSG, ClientId: c.id, Data: msg,
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
