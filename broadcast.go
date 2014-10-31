package link

// A broadcast sender. The broadcast message only encoded once
// so the performance it's better then send message one by one.
type Broadcaster struct {
	writer PacketWriter
}

// Craete a broadcaster.
func NewBroadcaster(protocol PacketProtocol) *Broadcaster {
	return &Broadcaster{
		writer: protocol.NewWriter(),
	}
}

func (b *Broadcaster) packet(message Message) (packet []byte, err error) {
	size := message.RecommendPacketSize()
	packet = b.writer.BeginPacket(size, nil)
	packet, err = message.AppendToPacket(packet)
	if err != nil {
		return nil, err
	}
	packet = b.writer.EndPacket(packet)
	return
}

// The session collection use to fetch session and send broadcast.
type SessionCollection interface {
	Fetch(func(*Session))
}

// Broadcast to sessions. The message only encoded once
// so the performance it's better then send message one by one.
func (b *Broadcaster) Broadcast(sessions SessionCollection, message Message) error {
	packet, err := b.packet(message)
	if err != nil {
		return err
	}
	sessions.Fetch(func(session *Session) {
		session.TrySendPacket(packet, 0)
	})
	return nil
}

// Broadcast to sessions. The message only encoded once
// so the performance it's better then send message one by one.
func (b *Broadcaster) MustBroadcast(sessions SessionCollection, message Message) error {
	packet, err := b.packet(message)
	if err != nil {
		return err
	}
	sessions.Fetch(func(session *Session) {
		session.SendPacket(packet)
	})
	return nil
}
