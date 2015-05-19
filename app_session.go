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

type OutMessage interface {
	Send(conn *Conn) error
}

type InMessage interface {
	Receive(conn *Conn) error
}

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
	session := &Session{
		id:               id,
		conn:             conn,
		autoFlush:        config.AutoFlush,
		asyncSendChan:    make(chan asyncOutMessage, config.AsyncSendChanSize),
		asyncSendTimeout: config.AsyncSendTimeout,
		closeChan:        make(chan int),
		closeCallbacks:   list.New(),
	}

	go session.sendLoop()

	return session
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

	session.conn.Flush()
}

func (session *Session) Send(msg OutMessage) error {
	session.sendMutex.Lock()
	defer session.sendMutex.Unlock()

	if err := msg.Send(session.conn); err != nil {
		session.Close()
		return err
	}

	if session.autoFlush {
		session.conn.Flush()
	}

	if session.conn.WError() != nil {
		session.Close()
		return session.conn.WError()
	}

	return nil
}

func (session *Session) Receive(message InMessage) error {
	session.recvMutex.Lock()
	defer session.recvMutex.Unlock()

	if err := message.Receive(session.conn); err != nil {
		session.Close()
		return err
	}

	if session.conn.RError() != nil {
		session.Close()
		return session.conn.RError()
	}
	return nil
}

type AsyncWork struct {
	c <-chan error
}

func (aw AsyncWork) Wait() error {
	return <-aw.c
}

type asyncOutMessage struct {
	C chan<- error
	M OutMessage
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
