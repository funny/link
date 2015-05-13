package link

import (
	"sync"
	"sync/atomic"
)

type SessionFetcher func(*int32, func(*Session))

// Broadcaster.
type Broadcaster struct {
	codec   Codec
	pool    *MemPool
	fetcher SessionFetcher
}

// Broadcast work.
type BroadcastWork struct {
	Session *Session
	c       <-chan error
}

// Wait work done. Returns error when work failed.
func (bw BroadcastWork) Wait() error {
	return <-bw.c
}

type broadcast struct {
	*Buffer
	refNum int32
}

func (bc *broadcast) Free() {
	if atomic.AddInt32(&bc.refNum, -1) == 0 {
		bc.Buffer.free()
	}
}

// Create a broadcaster.
func NewBroadcaster(codec Codec, pool *MemPool, fetcher SessionFetcher) *Broadcaster {
	return &Broadcaster{
		codec:   codec,
		pool:    pool,
		fetcher: fetcher,
	}
}

// Broadcast to sessions. The response only encoded once
// so the performance is better than send response one by one.
func (b *Broadcaster) Broadcast(msg Message) ([]BroadcastWork, error) {
	buffer := NewPoolBuffer(0, 1024, b.pool)
	b.codec.Prepend(buffer, msg)
	msg.WriteBuffer(buffer)

	bc := &broadcast{buffer, 0}
	works := make([]BroadcastWork, 0, 10)
	b.fetcher(&bc.refNum, func(session *Session) {
		works = append(works, session.asyncBroadcast(bc))
	})

	return works, nil
}

// The channel type. Used to maintain a group of session.
// Normally used for broadcast classify purpose.
type Channel struct {
	mutex       sync.RWMutex
	sessions    map[uint64]channelSession
	broadcaster *Broadcaster

	// channel state
	State interface{}
}

type channelSession struct {
	*Session
	KickCallback func()
}

// Create a channel instance.
func NewChannel(protocol Protocol, pool *MemPool) *Channel {
	channel := &Channel{
		sessions: make(map[uint64]channelSession),
	}
	codec := protocol.NewCodec()
	channel.broadcaster = NewBroadcaster(codec, pool, channel.sessionFetcher)
	return channel
}

// Broadcast.
func (channel *Channel) Broadcast(msg Message) ([]BroadcastWork, error) {
	return channel.broadcaster.Broadcast(msg)
}

// How mush sessions in this channel.
func (channel *Channel) Len() int {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()

	return len(channel.sessions)
}

// Join the channel. The kickCallback will called when the session kick out from the channel.
func (channel *Channel) Join(session *Session, kickCallback func()) {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	session.AddCloseCallback(channel, func() {
		channel.Exit(session)
	})
	channel.sessions[session.Id()] = channelSession{session, kickCallback}
}

// Exit the channel.
func (channel *Channel) Exit(session *Session) {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	session.RemoveCloseCallback(channel)
	delete(channel.sessions, session.Id())
}

// Kick out a session from the channel.
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

func (channel *Channel) sessionFetcher(num *int32, callback func(*Session)) {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()

	*num = int32(len(channel.sessions))

	for _, sesssion := range channel.sessions {
		callback(sesssion.Session)
	}
}

// Fetch the sessions. NOTE: Invoke Kick() or Exit() in fetch callback will dead lock.
func (channel *Channel) Fetch(callback func(*Session)) {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()

	for _, sesssion := range channel.sessions {
		callback(sesssion.Session)
	}
}
