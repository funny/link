package link

import "github.com/funny/sync"

// Broadcaster.
type Broadcaster struct {
	protocol ProtocolState
	fetcher  func(func(*Session))
}

// Broadcast work.
type BroadcastWork struct {
	Session *Session
	AsyncWork
}

// Create a broadcaster.
func NewBroadcaster(protocol Protocol, fetcher func(func(*Session))) *Broadcaster {
	return &Broadcaster{
		protocol: protocol.New(nil),
		fetcher:  fetcher,
	}
}

// Broadcast to sessions. The message only encoded once
// so the performance is better than send message one by one.
func (b *Broadcaster) Broadcast(message Message) ([]BroadcastWork, error) {
	packet, err := b.protocol.Packet(message)
	if err != nil {
		return nil, err
	}
	packet.isBroadcast = true
	works := make([]BroadcastWork, 0, 10)
	b.fetcher(func(session *Session) {
		packet.broadcastUse()
		works = append(works, BroadcastWork{
			session,
			session.asyncSendPacket(packet),
		})
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
func NewChannel(protocol Protocol) *Channel {
	channel := &Channel{
		sessions: make(map[uint64]channelSession),
	}
	channel.broadcaster = NewBroadcaster(protocol, channel.Fetch)
	return channel
}

// Broadcast to channel. The message only encoded once
// so the performance is better than send message one by one.
func (channel *Channel) Broadcast(message Message) ([]BroadcastWork, error) {
	return channel.broadcaster.Broadcast(message)
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

// Fetch the sessions. NOTE: Invoke Kick() or Exit() in fetch callback will dead lock.
func (channel *Channel) Fetch(callback func(*Session)) {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()

	for _, sesssion := range channel.sessions {
		callback(sesssion.Session)
	}
}
