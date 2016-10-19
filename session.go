package link

import (
	"container/list"
	"errors"
	"sync"
	"sync/atomic"
)

var SessionClosedError = errors.New("Session Closed")
var SessionBlockedError = errors.New("Session Blocked")

var globalSessionId uint64

type Session struct {
	id        uint64
	codec     Codec
	manager   *Manager
	sendChan  chan interface{}
	sendMutex sync.RWMutex

	closeFlag      int32
	closeChan      chan int
	closeMutex     sync.Mutex
	closeCallbacks *list.List

	State interface{}
}

func NewSession(codec Codec, sendChanSize int) *Session {
	return newSession(nil, codec, sendChanSize)
}

func newSession(manager *Manager, codec Codec, sendChanSize int) *Session {
	session := &Session{
		codec:     codec,
		manager:   manager,
		closeChan: make(chan int),
		id:        atomic.AddUint64(&globalSessionId, 1),
	}
	if sendChanSize > 0 {
		session.sendChan = make(chan interface{}, sendChanSize)
		go session.sendLoop()
	}
	return session
}

func (session *Session) ID() uint64 {
	return session.id
}

func (session *Session) IsClosed() bool {
	return atomic.LoadInt32(&session.closeFlag) == 1
}

func (session *Session) Close() error {
	if atomic.CompareAndSwapInt32(&session.closeFlag, 0, 1) {
		close(session.closeChan)

		if session.sendChan != nil {
			session.sendMutex.Lock()
			close(session.sendChan)
			if clear, ok := session.codec.(ClearSendChan); ok {
				clear.ClearSendChan(session.sendChan)
			}
			session.sendMutex.Unlock()
		}

		err := session.codec.Close()
		if session.manager != nil {
			session.manager.delSession(session)
		}
		session.invokeCloseCallbacks()
		return err
	}
	return SessionClosedError
}

func (session *Session) Codec() Codec {
	return session.codec
}

func (session *Session) Receive() (interface{}, error) {
	msg, err := session.codec.Receive()
	if err != nil {
		session.Close()
	}
	return msg, err
}

func (session *Session) sendLoop() {
	defer session.Close()
	for {
		select {
		case msg, ok := <-session.sendChan:
			if !ok || session.codec.Send(msg) != nil {
				return
			}
		case <-session.closeChan:
			return
		}
	}
}

func (session *Session) Send(msg interface{}) error {
	if session.IsClosed() {
		return SessionClosedError
	}

	if session.sendChan == nil {
		err := session.codec.Send(msg)
		if err != nil {
			session.Close()
		}
		return err
	}

	session.sendMutex.RLock()
	select {
	case session.sendChan <- msg:
		session.sendMutex.RUnlock()
		return nil
	case <-session.closeChan:
		session.sendMutex.RUnlock()
		return SessionClosedError
	default:
		session.sendMutex.RUnlock()
		session.Close()
		return SessionBlockedError
	}
}

type closeCallback struct {
	Handler interface{}
	Func    func()
}

func (session *Session) addCloseCallback(handler interface{}, callback func()) {
	if session.IsClosed() {
		return
	}

	session.closeMutex.Lock()
	defer session.closeMutex.Unlock()

	if session.closeCallbacks == nil {
		session.closeCallbacks = list.New()
	}

	session.closeCallbacks.PushBack(closeCallback{handler, callback})
}

func (session *Session) removeCloseCallback(handler interface{}) {
	if session.IsClosed() {
		return
	}

	session.closeMutex.Lock()
	defer session.closeMutex.Unlock()

	for i := session.closeCallbacks.Front(); i != nil; i = i.Next() {
		if i.Value.(closeCallback).Handler == handler {
			session.closeCallbacks.Remove(i)
			return
		}
	}
}

func (session *Session) invokeCloseCallbacks() {
	session.closeMutex.Lock()
	defer session.closeMutex.Unlock()

	if session.closeCallbacks == nil {
		return
	}

	for i := session.closeCallbacks.Front(); i != nil; i = i.Next() {
		callback := i.Value.(closeCallback)
		callback.Func()
	}
}
