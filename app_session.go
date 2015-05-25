package link

import (
	"container/list"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	SendToClosedError     = errors.New("Send to closed session")
	AsyncSendTimeoutError = errors.New("Async send timeout")
)

type SessionConfig struct {
	AutoFlush         bool
	AsyncSendTimeout  time.Duration
	AsyncSendChanSize int
}

type Session struct {
	id   uint64
	conn *Conn

	// About send and receive
	recvMutex        sync.Mutex
	sendMutex        sync.Mutex
	autoFlush        bool
	sendLoopFlag     int32
	asyncSendChan    chan asyncOutMessage
	asyncSendTimeout time.Duration

	// About session close
	closeChan       chan int
	closeFlag       int32
	closeEventMutex sync.Mutex
	closeCallbacks  *list.List

	// Put your session state here.
	State interface{}
}

func NewSession(id uint64, conn *Conn, config SessionConfig) *Session {
	return &Session{
		id:               id,
		conn:             conn,
		autoFlush:        config.AutoFlush,
		asyncSendChan:    make(chan asyncOutMessage, config.AsyncSendChanSize),
		asyncSendTimeout: config.AsyncSendTimeout,
		closeChan:        make(chan int),
		closeCallbacks:   list.New(),
	}
}

func (session *Session) Id() uint64 {
	return session.id
}

func (session *Session) Conn() *Conn {
	return session.conn
}

func (session *Session) IsClosed() bool {
	return atomic.LoadInt32(&session.closeFlag) != 0
}

func (session *Session) Close() {
	if atomic.CompareAndSwapInt32(&session.closeFlag, 0, 1) {
		session.conn.Close()
		close(session.closeChan)
		session.invokeCloseCallbacks()
	}
}

func (session *Session) Flush() {
	session.sendMutex.Lock()
	defer session.sendMutex.Unlock()

	session.conn.w.Flush()
}

func (session *Session) Send(msg OutMessage) error {
	session.sendMutex.Lock()
	defer session.sendMutex.Unlock()

	if err := msg.Send(session.conn.w); err != nil {
		session.Close()
		return err
	}

	if session.autoFlush {
		session.conn.w.Flush()
	}

	if session.conn.w.Error() != nil {
		session.Close()
		return session.conn.w.Error()
	}

	return nil
}

func (session *Session) Receive(message InMessage) error {
	session.recvMutex.Lock()
	defer session.recvMutex.Unlock()

	if err := message.Receive(session.conn.r); err != nil {
		session.Close()
		return err
	}

	if session.conn.r.Error() != nil {
		session.Close()
		return session.conn.r.Error()
	}

	return nil
}

type AsyncWork struct {
	C <-chan error
}

func (aw AsyncWork) Wait() error {
	return <-aw.C
}

type asyncOutMessage struct {
	C chan<- error
	M OutMessage
}

func (session *Session) initSendLoop() {
	if atomic.CompareAndSwapInt32(&session.sendLoopFlag, 0, 1) {
		go session.sendLoop()
	}
}

func (session *Session) sendLoop() {
	for {
		select {
		case msg := <-session.asyncSendChan:
			msg.C <- session.Send(msg.M)
		case <-session.closeChan:
			return
		}
	}
}

func (session *Session) AsyncSend(msg OutMessage) AsyncWork {
	c := make(chan error, 1)

	if session.IsClosed() {
		c <- SendToClosedError
		return AsyncWork{c}
	}

	session.initSendLoop()

	select {
	case session.asyncSendChan <- asyncOutMessage{c, msg}:
	default:
		if session.asyncSendTimeout == 0 {
			session.Close()
			c <- AsyncSendTimeoutError
		} else {
			go func() {
				select {
				case session.asyncSendChan <- asyncOutMessage{c, msg}:
				case <-session.closeChan:
					c <- SendToClosedError
				case <-time.After(session.asyncSendTimeout):
					session.Close()
					c <- AsyncSendTimeoutError
				}
			}()
		}
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
