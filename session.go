package link

import (
	"container/list"
	"errors"
	"net"
	"sync"
	"sync/atomic"
)

var (
	ErrClosed   = errors.New("Session closed")
	ErrBlocking = errors.New("Operation blocking")
)

type Session struct {
	id    uint64
	conn  net.Conn
	codec Codec

	// About send and receive
	recvMutex    sync.Mutex
	sendMutex    sync.Mutex
	sendLoopFlag int32
	sendChan     chan interface{}

	// About session close
	closeChan       chan int
	closeFlag       int32
	closeEventMutex sync.Mutex
	closeCallbacks  *list.List

	// Session state
	State interface{}
}

var globalSessionId uint64

func NewSession(conn net.Conn, codecType CodecType) *Session {
	session := &Session{
		id:             atomic.AddUint64(&globalSessionId, 1),
		conn:           conn,
		codec:          codecType.NewCodec(conn, conn),
		closeChan:      make(chan int),
		closeCallbacks: list.New(),
	}
	return session
}

func (session *Session) Id() uint64     { return session.id }
func (session *Session) Conn() net.Conn { return session.conn }
func (session *Session) IsClosed() bool { return atomic.LoadInt32(&session.closeFlag) != 0 }

func (session *Session) Close() {
	if atomic.CompareAndSwapInt32(&session.closeFlag, 0, 1) {
		session.invokeCloseCallbacks()
		close(session.closeChan)
		session.conn.Close()
	}
}

func (session *Session) Receive(msg interface{}) (err error) {
	session.recvMutex.Lock()
	defer session.recvMutex.Unlock()

	err = session.codec.Decode(msg)
	if err != nil {
		session.Close()
	}
	return
}

func (session *Session) Send(msg interface{}) (err error) {
	session.sendMutex.Lock()
	defer session.sendMutex.Unlock()

	err = session.codec.Encode(msg)
	if err != nil {
		session.Close()
	}
	return
}

func (session *Session) EnableAsyncSend(sendChanSize int) {
	if atomic.CompareAndSwapInt32(&session.sendLoopFlag, 0, 1) {
		session.sendChan = make(chan interface{}, sendChanSize)
		go func() {
			for {
				select {
				case msg := <-session.sendChan:
					if err := session.Send(msg); err != nil {
						return
					}
				case <-session.closeChan:
					return
				}
			}
		}()
	}
}

func (session *Session) AsyncSend(msg interface{}) error {
	if session.IsClosed() {
		return ErrClosed
	}

	if session.sendLoopFlag != 1 {
		panic("AsyncSend not enable")
	}

	select {
	case session.sendChan <- msg:
	default:
		session.Close()
		return ErrBlocking
	}
	return nil
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
