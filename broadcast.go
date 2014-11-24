package link

import "sync"

// A broadcast sender. The broadcast message only encoded once
// so the performance it's better then send message one by one.
type Broadcaster struct {
	mutex  sync.Mutex
	writer PacketWriter
	buffer OutBuffer
}

// Craete a broadcaster.
func NewBroadcaster(protocol PacketProtocol) *Broadcaster {
	return &Broadcaster{
		writer: protocol.NewWriter(),
		buffer: protocol.BufferFactory().NewOutBuffer(),
	}
}

func (b *Broadcaster) packet(message Message) error {
	b.buffer.Prepare(message.RecommendBufferSize())
	return message.WriteBuffer(b.buffer)
}

// The session collection use to fetch session and send broadcast.
type SessionCollection interface {
	Fetch(func(*Session))
}

// Broadcast to sessions. The message only encoded once
// so the performance it's better then send message one by one.
func (b *Broadcaster) Broadcast(sessions SessionCollection, message Message) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if err := b.packet(message); err != nil {
		return err
	}
	sessions.Fetch(func(session *Session) {
		session.TrySendPacket(b.buffer, 0)
	})
	return nil
}

// Broadcast to sessions. The message only encoded once
// so the performance it's better then send message one by one.
func (b *Broadcaster) MustBroadcast(sessions SessionCollection, message Message) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if err := b.packet(message); err != nil {
		return err
	}
	sessions.Fetch(func(session *Session) {
		session.SendPacket(b.buffer)
	})
	return nil
}
