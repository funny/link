package link

import (
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
)

// Message.
type Message interface {
	// Get a recommend buffer size.
	RecommendBufferSize() int

	// Write the message to the packet buffer and returns the new buffer like append() function.
	WriteBuffer(buffer *OutBuffer) error
}

// Convert to bytes message.
func Bytes(v []byte) Message {
	return rawMsg(v)
}

// Convert to string message.
func String(v string) Message {
	return rawMsg(v)
}

// Create a json message.
func Json(v interface{}) Message {
	return jsonMsg{v}
}

// Create a gob message.
func Gob(v interface{}) Message {
	return gobMsg{v}
}

// Create a xml message.
func Xml(v interface{}) Message {
	return xmlMsg{v}
}

// Binary message
type rawMsg []byte

// Implement the Message interface.
func (raw rawMsg) RecommendBufferSize() int {
	return len(raw)
}

// Implement the Message interface.
func (raw rawMsg) WriteBuffer(buffer *OutBuffer) error {
	buffer.Append([]byte(raw)...)
	return nil
}

// JSON message
type jsonMsg struct {
	v interface{}
}

// Implement the Message interface.
func (j jsonMsg) RecommendBufferSize() int {
	return 1024
}

// Implement the Message interface.
func (j jsonMsg) WriteBuffer(buffer *OutBuffer) error {
	return json.NewEncoder(buffer).Encode(j.v)
}

// GOB message
type gobMsg struct {
	v interface{}
}

// Implement the Message interface.
func (g gobMsg) RecommendBufferSize() int {
	return 1024
}

// Implement the Message interface.
func (g gobMsg) WriteBuffer(buffer *OutBuffer) error {
	return gob.NewEncoder(buffer).Encode(g.v)
}

// XML message
type xmlMsg struct {
	v interface{}
}

// Implement the Message interface.
func (x xmlMsg) RecommendBufferSize() int {
	return 1024
}

// Implement the Message interface.
func (x xmlMsg) WriteBuffer(buffer *OutBuffer) error {
	return xml.NewEncoder(buffer).Encode(x.v)
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
func (q *SendQueue) WriteBuffer(buffer *OutBuffer) error {
	var err error
	for _, message := range q.messages {
		if err = message.WriteBuffer(buffer); err != nil {
			return err
		}
	}
	return nil
}
