package link

import (
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
)

var SessionClosedError = errors.New("Session Closed")
var SessionBlockedError = errors.New("Session Blocked")

var globalSessionId uint64

type Session struct {
	id        uint64
	codec     Codec
	manager   *Manager
	sendChan  chan interface{}
	recvMutex sync.Mutex
	sendMutex sync.RWMutex

	closeFlag          int32
	closeChan          chan int
	closeMutex         sync.Mutex
	firstCloseCallback *closeCallback
	lastCloseCallback  *closeCallback

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

		go func() {
			session.invokeCloseCallbacks()

			if session.manager != nil {
				session.manager.delSession(session)
			}
		}()
		return err
	}
	return SessionClosedError
}

func (session *Session) Codec() Codec {
	return session.codec
}

func (session *Session) Receive() (interface{}, error) {
	session.recvMutex.Lock()

	msg, err := session.codec.Receive()
	if err != nil {
		if err0 := session.Close(); err0 != nil {
			session.recvMutex.Unlock()
			return msg, errors.Errorf("receive message failed. error: %s\n, and close session also failed. error: %s\n",
				err.Error(),
				err0.Error())
		}
	}
	session.recvMutex.Unlock()
	return msg, err
}

// 1.  if session closes failed, why the error message doesn't return
// 2.  if session sends  message failed, why the error message doesn't return
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
	if session.sendChan == nil {
		if session.IsClosed() {
			return SessionClosedError
		}

		session.sendMutex.Lock()

		err := session.codec.Send(msg)
		if err != nil {
			session.Close()
		}
		session.sendMutex.Unlock()
		return err
	}

	session.sendMutex.RLock()
	if session.IsClosed() {
		session.sendMutex.RUnlock()
		return SessionClosedError
	}

	select {
	case session.sendChan <- msg:
		session.sendMutex.RUnlock()
		return nil
	default:
		session.sendMutex.RUnlock()
		session.Close()
		return SessionBlockedError
	}
}

type closeCallback struct {
	Handler interface{}
	Key     interface{}
	Func    func()
	Next    *closeCallback
}

func (session *Session) AddCloseCallback(handler, key interface{}, callback func()) {
	if session.IsClosed() {
		return
	}

	newItem := &closeCallback{handler, key, callback, nil}

	session.closeMutex.Lock()
	if session.firstCloseCallback == nil {
		session.firstCloseCallback = newItem
	} else {
		session.lastCloseCallback.Next = newItem
	}
	session.lastCloseCallback = newItem
	session.closeMutex.Unlock()
}

func (session *Session) RemoveCloseCallback(handler, key interface{}) {
	if session.IsClosed() {
		return
	}

	session.closeMutex.Lock()

	var prev *closeCallback
	for callback := session.firstCloseCallback; callback != nil; prev, callback = callback, callback.Next {
		if callback.Handler == handler && callback.Key == key {
			if session.firstCloseCallback == callback {
				session.firstCloseCallback = callback.Next
			} else {
				prev.Next = callback.Next
			}
			if session.lastCloseCallback == callback {
				session.lastCloseCallback = prev
			}
			session.closeMutex.Unlock()
			return
		}
	}
	session.closeMutex.Unlock()
	return
}

func (session *Session) invokeCloseCallbacks() {
	session.closeMutex.Lock()

	for callback := session.firstCloseCallback; callback != nil; callback = callback.Next {
		callback.Func()
	}
	session.closeMutex.Unlock()
}
