package link

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
)

// Message.
type Message interface {
	// Get a recommend packet size for buffer initialization.
	RecommendBufferSize() int

	// Write the message to the packet buffer and returns the new buffer like append() function.
	WriteBuffer(buffer []byte) ([]byte, error)
}

// Binary message
type Binary []byte

// Implement the Message interface.
func (bin Binary) RecommendBufferSize() int {
	return len(bin)
}

// Implement the Message interface.
func (bin Binary) WriteBuffer(buffer []byte) ([]byte, error) {
	return append(buffer, []byte(bin)...), nil
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
func (j JSON) WriteBuffer(buffer []byte) ([]byte, error) {
	var w bytes.Buffer
	if err := json.NewEncoder(&w).Encode(j.V); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
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
func (g GOB) WriteBuffer(buffer []byte) ([]byte, error) {
	var w bytes.Buffer
	if err := gob.NewEncoder(&w).Encode(g.V); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
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
func (x XML) WriteBuffer(buffer []byte) ([]byte, error) {
	var w bytes.Buffer
	if err := xml.NewEncoder(&w).Encode(x.V); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
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
func (q *SendQueue) WriteBuffer(buffer []byte) ([]byte, error) {
	var err error
	for _, message := range q.messages {
		if buffer, err = message.WriteBuffer(buffer); err != nil {
			return nil, err
		}
	}
	return buffer, nil
}
