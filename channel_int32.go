// DO NOT EDIT
// GENERATE BY 'go run channel_gen.go Int32Channel int32 channel_int32.go'
package link

import (
	"sync"
)

type Int32Channel struct {
	mutex    sync.RWMutex
	sessions map[int32]*Session

	// channel state
	State interface{}
}

func NewInt32Channel() *Int32Channel {
	return &Int32Channel{
		sessions: make(map[int32]*Session),
	}
}

func (channel *Int32Channel) Len() int {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()
	return len(channel.sessions)
}

// Fetch the sessions. NOTE: Dead lock happends if invoke Exit() in fetch callback.
func (channel *Int32Channel) Fetch(callback func(*Session)) {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()
	for _, session := range channel.sessions {
		callback(session)
	}
}

func (channel *Int32Channel) Get(key int32) *Session {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()
	session, _ := channel.sessions[key]
	return session
}

func (channel *Int32Channel) Put(key int32, session *Session) {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	if session, exists := channel.sessions[key]; exists {
		channel.remove(key, session)
	}
	session.addCloseCallback(channel, func() {
		channel.Remove(key)
	})
	channel.sessions[key] = session
}

func (channel *Int32Channel) remove(key int32, session *Session) {
	session.removeCloseCallback(channel)
	delete(channel.sessions, key)
}

func (channel *Int32Channel) Remove(key int32) bool {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	session, exists := channel.sessions[key]
	if exists {
		channel.remove(key, session)
	}
	return exists
}

func (channel *Int32Channel) Close() {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	for key, session := range channel.sessions {
		channel.remove(key, session)
	}
}
