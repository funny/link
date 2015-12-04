// DO NOT EDIT
// GENERATE BY 'go run channel_gen.go Int64Channel int64 channel_int64.go'
package link

import (
	"sync"
)

type Int64Channel struct {
	mutex    sync.RWMutex
	sessions map[int64]*Session

	// channel state
	State interface{}
}

func NewInt64Channel() *Int64Channel {
	return &Int64Channel{
		sessions: make(map[int64]*Session),
	}
}

func (channel *Int64Channel) Len() int {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()
	return len(channel.sessions)
}

// Fetch the sessions. NOTE: Dead lock happends if invoke Exit() in fetch callback.
func (channel *Int64Channel) Fetch(callback func(*Session)) {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()
	for _, session := range channel.sessions {
		callback(session)
	}
}

func (channel *Int64Channel) Get(key int64) *Session {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()
	session, _ := channel.sessions[key]
	return session
}

func (channel *Int64Channel) Put(key int64, session *Session) {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	if session, exists := channel.sessions[key]; exists {
		channel.remove(key, session)
	}
	session.AddCloseCallback(channel, func() {
		channel.Remove(key)
	})
	channel.sessions[key] = session
}

func (channel *Int64Channel) remove(key int64, session *Session) {
	session.RemoveCloseCallback(channel)
	delete(channel.sessions, key)
}

func (channel *Int64Channel) Remove(key int64) bool {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	session, exists := channel.sessions[key]
	if exists {
		channel.remove(key, session)
	}
	return exists
}

func (channel *Int64Channel) Close() {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	for key, session := range channel.sessions {
		channel.remove(key, session)
	}
}
