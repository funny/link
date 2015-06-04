package link

import (
	"container/list"
	"errors"
	"sync"
	"sync/atomic"
)

var (
	ErrClosed   = errors.New("Session closed")
	ErrBlocking = errors.New("Operation blocking")
)

var DefaultConfig = SessionConfig{
	SendChanSize: 1000,
}

type SessionConfig struct {
	SendChanSize int
}

type Session struct {
	id   uint64
	conn Conn

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

func NewSession(id uint64, conn Conn) *Session {
	return &Session{
		id:             id,
		conn:           conn,
		sendChan:       make(chan interface{}, conn.Config().SendChanSize),
		closeChan:      make(chan int),
		closeCallbacks: list.New(),
	}
}

func (session *Session) Id() uint64     { return session.id }
func (session *Session) Conn() Conn     { return session.conn }
func (session *Session) IsClosed() bool { return atomic.LoadInt32(&session.closeFlag) != 0 }

func (session *Session) Close() {
	if atomic.CompareAndSwapInt32(&session.closeFlag, 0, 1) {
		session.invokeCloseCallbacks()
		close(session.closeChan)
		session.conn.Close()
	}
}

func (session *Session) Receive(msg interface{}) error {
	session.recvMutex.Lock()
	defer session.recvMutex.Unlock()

	if err := session.conn.Receive(msg); err != nil {
		session.Close()
		return err
	}
	return nil
}

func (session *Session) Send(msg interface{}) error {
	session.sendMutex.Lock()
	defer session.sendMutex.Unlock()

	if err := session.conn.Send(msg); err != nil {
		session.Close()
		return err
	}
	return nil
}

func (session *Session) AsyncSend(msg interface{}) error {
	if session.IsClosed() {
		return ErrClosed
	}

	if atomic.CompareAndSwapInt32(&session.sendLoopFlag, 0, 1) {
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
