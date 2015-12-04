// DO NOT EDIT
// GENERATE BY 'go run channel_gen.go Uint64Channel uint64 channel_uint64.go'
package link

import (
	"sync"
)

type Uint64Channel struct {
	mutex    sync.RWMutex
	sessions map[uint64]*Session

	// channel state
	State interface{}
}

func NewUint64Channel() *Uint64Channel {
	return &Uint64Channel{
		sessions: make(map[uint64]*Session),
	}
}

func (channel *Uint64Channel) Len() int {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()
	return len(channel.sessions)
}

// Fetch the sessions. NOTE: Dead lock happends if invoke Exit() in fetch callback.
func (channel *Uint64Channel) Fetch(callback func(*Session)) {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()
	for _, session := range channel.sessions {
		callback(session)
	}
}

func (channel *Uint64Channel) Get(key uint64) *Session {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()
	session, _ := channel.sessions[key]
	return session
}

func (channel *Uint64Channel) Put(key uint64, session *Session) {
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

func (channel *Uint64Channel) remove(key uint64, session *Session) {
	session.RemoveCloseCallback(channel)
	delete(channel.sessions, key)
}

func (channel *Uint64Channel) Remove(key uint64) bool {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	session, exists := channel.sessions[key]
	if exists {
		channel.remove(key, session)
	}
	return exists
}

func (channel *Uint64Channel) Close() {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	for key, session := range channel.sessions {
		channel.remove(key, session)
	}
}
