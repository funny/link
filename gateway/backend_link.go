package gateway

import (
	"bytes"
	"sync"
	"sync/atomic"

	"github.com/funny/link"
)

var gloablClientId uint64

type backendLink struct {
	id        uint64
	listener  *BackendListener
	session   *link.Session
	connMutex sync.RWMutex
	conns     map[uint64]*BackendConn
	connWaits map[uint64]*BackendConn
	closeFlag int32
}

func newBackendLink(id uint64, listener *BackendListener, session *link.Session) *backendLink {
	this := &backendLink{
		id:        id,
		listener:  listener,
		session:   session,
		conns:     make(map[uint64]*BackendConn),
		connWaits: make(map[uint64]*BackendConn),
	}
	session.AddCloseCallback(this, func() {
		this.Close(false)
	})
	go this.loop()
	return this
}

func (this *backendLink) Close(byListener bool) {
	if atomic.CompareAndSwapInt32(&this.closeFlag, 0, 1) {
		this.session.RemoveCloseCallback(this)

		this.connMutex.Lock()
		defer this.connMutex.Unlock()

		for _, conn := range this.conns {
			conn.close(false)
		}

		if !byListener {
			this.listener.delLink(this.id)
		}
	}
}

func (this *backendLink) loop() {
	var msg gatewayMsg
	for {
		if err := this.session.Receive(&msg); err != nil {
			// TODO: log
			return
		}
		switch msg.Command {
		case CMD_MSG:
			this.dispathMsg(msg.ClientId, msg.Data)
		case CMD_BRD:
			this.listener.broadcast(msg.ClientIds, msg.Data)
		case CMD_DEL:
			this.delConn(msg.ClientId)
		case CMD_NEW_1:
			this.newConn(msg.ClientId, msg.Data)
		case CMD_NEW_3:
			this.acceptConn(msg.ClientId)
		}
	}
}

func (this *backendLink) dispathMsg(id uint64, data []byte) {
	this.connMutex.RLock()
	defer this.connMutex.RUnlock()

	if conn, exists := this.conns[id]; exists {
		select {
		case conn.recvChan <- data:
		default:
			go conn.close(true)
		}
	}
}

func (this *backendLink) newConn(waitId uint64, addr []byte) link.Conn {
	clientId := atomic.AddUint64(&gloablClientId, 1)

	this.connMutex.RLock()
	conn, exists := this.conns[clientId]
	this.connMutex.RUnlock()
	if exists {
		conn.close(true)
	}

	i := bytes.IndexByte(addr, ',')
	conn = newBackendConn(clientId, clientAddr{addr[:i], addr[i+1:]}, this)
	this.session.Send(&gatewayMsg{
		Command: CMD_NEW_2, ClientId: waitId, ClientIds: []uint64{clientId},
	})

	this.connMutex.Lock()
	defer this.connMutex.Unlock()
	this.connWaits[clientId] = conn

	return conn
}

func (this *backendLink) acceptConn(clientId uint64) {
	conn := func() *BackendConn {
		this.connMutex.Lock()
		defer this.connMutex.Unlock()

		conn, exists := this.connWaits[clientId]
		if !exists {
			return nil
		}

		delete(this.connWaits, clientId)
		this.conns[clientId] = conn
		return conn
	}()

	if conn == nil {
		return
	}

	this.listener.acceptChan <- conn
}

func (this *backendLink) delConn(id uint64) {
	this.connMutex.Lock()
	defer this.connMutex.Unlock()

	if conn, exists := this.conns[id]; exists {
		conn.close(false)
	}
}

func (this *backendLink) broadcast(ids []uint64, data []byte) {
	this.connMutex.RLock()
	defer this.connMutex.RUnlock()

	for _, id := range ids {
		if conn, exists := this.conns[id]; exists {
			select {
			case conn.recvChan <- data:
			default:
				go conn.close(true)
			}
		}
	}
}
