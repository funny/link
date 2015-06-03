package link

import (
	"container/list"
	"errors"
	"sync"
	"sync/atomic"
)

var (
	ErrSendToClosed = errors.New("Send to closed session")
	ErrSendBlocking = errors.New("Async send blocking")
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
	sendChan     chan asyncOutMessage

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
		sendChan:       make(chan asyncOutMessage, conn.Config().SendChanSize),
		closeChan:      make(chan int),
		closeCallbacks: list.New(),
	}
}

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

func (session *Session) Id() uint64     { return session.id }
func (session *Session) Conn() Conn     { return session.conn }
func (session *Session) IsClosed() bool { return atomic.LoadInt32(&session.closeFlag) != 0 }

type AsyncWork struct {
	C <-chan error
}

func (aw AsyncWork) Wait() error {
	return <-aw.C
}

type asyncOutMessage struct {
	C chan<- error
	M interface{}
}

func (session *Session) initAsyncSendLoop() {
	if atomic.CompareAndSwapInt32(&session.sendLoopFlag, 0, 1) {
		go session.asyncSendLoop()
	}
}

func (session *Session) asyncSendLoop() {
	for {
		select {
		case msg := <-session.sendChan:
			msg.C <- session.Send(msg.M)
		case <-session.closeChan:
			return
		}
	}
}

func (session *Session) AsyncSend(msg interface{}) AsyncWork {
	c := make(chan error, 1)

	if session.IsClosed() {
		c <- ErrSendToClosed
		return AsyncWork{c}
	}

	session.initAsyncSendLoop()

	select {
	case session.sendChan <- asyncOutMessage{c, msg}:
	default:
		session.Close()
		c <- ErrSendBlocking
	}

	return AsyncWork{c}
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
