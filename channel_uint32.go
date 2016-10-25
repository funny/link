// DO NOT EDIT
// GENERATE BY 'go run channel_gen.go Uint32Channel uint32 channel_uint32.go'
package link

import (
	"sync"
)

type Uint32Channel struct {
	mutex    sync.RWMutex
	sessions map[uint32]*Session

	// channel state
	State interface{}
}

func NewUint32Channel() *Uint32Channel {
	return &Uint32Channel{
		sessions: make(map[uint32]*Session),
	}
}

func (channel *Uint32Channel) Len() int {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()
	return len(channel.sessions)
}

// Fetch the sessions. NOTE: Dead lock happends if invoke Exit() in fetch callback.
func (channel *Uint32Channel) Fetch(callback func(*Session)) {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()
	for _, session := range channel.sessions {
		callback(session)
	}
}

func (channel *Uint32Channel) Get(key uint32) *Session {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()
	session, _ := channel.sessions[key]
	return session
}

func (channel *Uint32Channel) Put(key uint32, session *Session) {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	if session, exists := channel.sessions[key]; exists {
		channel.remove(key, session)
	}
	session.AddCloseCallback(channel, key, func() {
		channel.Remove(key)
	})
	channel.sessions[key] = session
}

func (channel *Uint32Channel) remove(key uint32, session *Session) {
	session.RemoveCloseCallback(channel, key)
	delete(channel.sessions, key)
}

func (channel *Uint32Channel) Remove(key uint32) bool {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	session, exists := channel.sessions[key]
	if exists {
		channel.remove(key, session)
	}
	return exists
}

func (channel *Uint32Channel) Close() {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	for key, session := range channel.sessions {
		channel.remove(key, session)
	}
}
