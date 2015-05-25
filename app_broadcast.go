package link

import (
	"sync"
)

type SessionFetcher func(func(*Session))

type BroadcastWork struct {
	*Session
	AsyncWork
}

func Broadcast(msg OutMessage, fetcher SessionFetcher) ([]BroadcastWork, error) {
	works := make([]BroadcastWork, 0, 10)
	fetcher(func(session *Session) {
		works = append(works, BroadcastWork{
			session,
			session.AsyncSend(msg),
		})
	})
	return works, nil
}

type Channel struct {
	mutex    sync.RWMutex
	sessions map[uint64]channelSession
}

type channelSession struct {
	*Session
	KickCallback func()
}

func NewChannel() *Channel {
	channel := &Channel{
		sessions: make(map[uint64]channelSession),
	}
	return channel
}

func (channel *Channel) Broadcast(msg OutMessage) ([]BroadcastWork, error) {
	return Broadcast(msg, channel.Fetch)
}

func (channel *Channel) Len() int {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()

	return len(channel.sessions)
}

func (channel *Channel) Join(session *Session, kickCallback func()) {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	session.AddCloseCallback(channel, func() { channel.Exit(session) })
	channel.sessions[session.Id()] = channelSession{session, kickCallback}
}

func (channel *Channel) Exit(session *Session) {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	session.RemoveCloseCallback(channel)
	delete(channel.sessions, session.Id())
}

func (channel *Channel) Kick(sessionId uint64) {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	if session, exists := channel.sessions[sessionId]; exists {
		delete(channel.sessions, sessionId)
		if session.KickCallback != nil {
			session.KickCallback()
		}
	}
}

// Fetch the sessions. NOTE: Dead lock happends if invoke Kick() or Exit() in fetch callback.
func (channel *Channel) Fetch(callback func(*Session)) {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()

	for _, sesssion := range channel.sessions {
		callback(sesssion.Session)
	}
}
