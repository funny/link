package link

import (
	"container/list"
	"errors"
	"net"
	"sync"
	"sync/atomic"
)

var ErrClosed = errors.New("link.Session closed")

type Session struct {
	id              uint64
	conn            net.Conn
	encoder         Encoder
	decoder         Decoder
	closeChan       chan int
	closeFlag       int32
	closeEventMutex sync.Mutex
	closeCallbacks  *list.List
	State           interface{}
}

var globalSessionId uint64

func NewSession(conn net.Conn, codecType CodecType) *Session {
	session := &Session{
		id:             atomic.AddUint64(&globalSessionId, 1),
		conn:           conn,
		encoder:        codecType.NewEncoder(conn),
		decoder:        codecType.NewDecoder(conn),
		closeCallbacks: list.New(),
	}
	return session
}

func (session *Session) Id() uint64     { return session.id }
func (session *Session) Conn() net.Conn { return session.conn }
func (session *Session) IsClosed() bool { return atomic.LoadInt32(&session.closeFlag) == 1 }

func (session *Session) Close() {
	if atomic.CompareAndSwapInt32(&session.closeFlag, 0, 1) {
		session.invokeCloseCallbacks()
		if session.closeChan != nil {
			close(session.closeChan)
		}
		session.conn.Close()
		if d, ok := session.encoder.(Disposeable); ok {
			d.Dispose()
		}
		if d, ok := session.decoder.(Disposeable); ok {
			d.Dispose()
		}
	}
}

func (session *Session) Receive(msg interface{}) (err error) {
	if session.IsClosed() {
		return ErrClosed
	}
	err = session.decoder.Decode(msg)
	if err != nil {
		session.Close()
	}
	return
}

func (session *Session) Send(msg interface{}) (err error) {
	if session.IsClosed() {
		return ErrClosed
	}
	err = session.encoder.Encode(msg)
	if err != nil {
		session.Close()
	}
	return
}

type closeCallback struct {
	Handler interface{}
	Func    func()
}

func (session *Session) AddCloseCallback(handler interface{}, callback func()) {
	if session.IsClosed() {
		return
	}

	session.closeEventMutex.Lock()
	defer session.closeEventMutex.Unlock()

	session.closeCallbacks.PushBack(closeCallback{handler, callback})
}

func (session *Session) RemoveCloseCallback(handler interface{}) {
	if session.IsClosed() {
		return
	}

	session.closeEventMutex.Lock()
	defer session.closeEventMutex.Unlock()

	for i := session.closeCallbacks.Front(); i != nil; i = i.Next() {
		if i.Value.(closeCallback).Handler == handler {
			session.closeCallbacks.Remove(i)
			return
		}
	}
}

func (session *Session) invokeCloseCallbacks() {
	session.closeEventMutex.Lock()
	defer session.closeEventMutex.Unlock()

	for i := session.closeCallbacks.Front(); i != nil; i = i.Next() {
		callback := i.Value.(closeCallback)
		callback.Func()
	}
}
