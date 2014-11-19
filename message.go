package link

import (
	"bytes"
	"encoding/json"
)

// Message.
type Message interface {
	// Get a recommend packet size for packet buffer initialization.
	RecommendPacketSize() uint

	// Append the message to the packet buffer and returns the new buffer like append() function.
	AppendToPacket(buffer *OutMessage) error
}

// Binary message
type Binary []byte

// Implement the Message interface.
func (bin Binary) RecommendPacketSize() uint {
	return uint(len(bin))
}

// Implement the Message interface.
func (bin Binary) AppendToPacket(buffer *OutMessage) error {
	buffer.AppendBytes([]byte(bin))
	return nil
}

// JSON message
type JSON struct {
	V    interface{}
	Size uint
}

// Implement the Message interface.
func (j JSON) RecommendPacketSize() uint {
	return j.Size
}

// Implement the Message interface.
func (j JSON) AppendToPacket(buffer *OutMessage) error {
	w := bytes.NewBuffer(*buffer)
	e := json.NewEncoder(w)
	err := e.Encode(j.V)
	if err != nil {
		return err
	}
	*buffer = OutMessage(w.Bytes())
	return nil
}

// A simple send queue. Can used for buffered send.
// For example, sometimes you have many Session.Send() call during a request processing.
// You can use the send queue to buffer those messages then call Session.Send() once after request processing done.
// The send queue type implemented Message interface. So you can pass it as the Session.Send() method argument.
type SendQueue struct {
	messages []Message
}

// Push a message into send queue but not send it immediately.
func (q *SendQueue) Push(message Message) {
	q.messages = append(q.messages, message)
}

// Implement the Message interface.
func (q *SendQueue) RecommendPacketSize() uint {
	size := uint(0)
	for _, message := range q.messages {
		size += message.RecommendPacketSize()
	}
	return size
}

// Implement the Message interface.
func (q *SendQueue) AppendToPacket(buffer *OutMessage) error {
	for _, message := range q.messages {
		if err := message.AppendToPacket(buffer); err != nil {
			return err
		}
	}
	return nil
}
