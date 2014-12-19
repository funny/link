package link

import (
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
)

// Message.
type Message interface {
	// Get a recommend packet size for buffer initialization.
	RecommendBufferSize() int

	// Write the message to the packet buffer and returns the new buffer like append() function.
	WriteBuffer(buffer *Buffer) error
}

// Binary message
type Binary []byte

// Implement the Message interface.
func (bin Binary) RecommendBufferSize() int {
	return len(bin)
}

// Implement the Message interface.
func (bin Binary) WriteBuffer(buffer *Buffer) error {
	buffer.Append([]byte(bin)...)
	return nil
}

// JSON message
type JSON struct {
	V interface{}
}

// Implement the Message interface.
func (j JSON) RecommendBufferSize() int {
	return 0
}

// Implement the Message interface.
func (j JSON) WriteBuffer(buffer *Buffer) error {
	return json.NewEncoder(buffer).Encode(j.V)
}

// GOB message
type GOB struct {
	V interface{}
}

// Implement the Message interface.
func (g GOB) RecommendBufferSize() int {
	return 0
}

// Implement the Message interface.
func (g GOB) WriteBuffer(buffer *Buffer) error {
	return gob.NewEncoder(buffer).Encode(g.V)
}

// XML message
type XML struct {
	V interface{}
}

// Implement the Message interface.
func (x XML) RecommendBufferSize() int {
	return 0
}

// Implement the Message interface.
func (x XML) WriteBuffer(buffer *Buffer) error {
	return xml.NewEncoder(buffer).Encode(x.V)
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
func (q *SendQueue) RecommendBufferSize() int {
	size := 0
	for _, message := range q.messages {
		size += message.RecommendBufferSize()
	}
	return size
}

// Implement the Message interface.
func (q *SendQueue) WriteBuffer(buffer *Buffer) error {
	var err error
	for _, message := range q.messages {
		if err = message.WriteBuffer(buffer); err != nil {
			return err
		}
	}
	return nil
}
