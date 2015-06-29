package gateway

import (
	"bytes"
	"github.com/funny/link"
	"sync"
	"sync/atomic"
)

var gloablClientId uint64

type backendLink struct {
	id        uint64
	listener  *BackendListener
	session   *link.Session
	connMutex sync.RWMutex
	conns     map[uint64]*BackendConn
	closeFlag int32
}

func newBackendLink(id uint64, listener *BackendListener, session *link.Session) *backendLink {
	session.EnableAsyncSend(10000)
	this := &backendLink{
		id:       id,
		listener: listener,
		session:  session,
		conns:    make(map[uint64]*BackendConn),
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
		case CMD_DEL:
			this.delConn(msg.ClientId)
		case CMD_NEW_1:
			this.newConn(msg.ClientId, msg.Data)
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
			conn.close(true)
		}
	}
}

func (this *backendLink) newConn(waitId uint64, addr []byte) {
	clientId := atomic.AddUint64(&gloablClientId, 1)

	this.connMutex.RLock()
	conn, exists := this.conns[clientId]
	this.connMutex.RUnlock()
	if exists {
		conn.close(true)
	}

	i := bytes.IndexByte(addr, ',')
	conn = newBackendConn(clientId, waitId, clientAddr{addr[:i], addr[i+1:]}, this.listener.codecType, this)

	this.connMutex.Lock()
	this.conns[clientId] = conn
	this.connMutex.Unlock()

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
