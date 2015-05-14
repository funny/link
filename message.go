package link

import (
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
)

type SingleFrame struct {
	Message
}

func (frame SingleFrame) IsFinalFrame() bool {
	return true
}

func (frame SingleFrame) NextFrame() FrameMessage {
	return nil
}

// A func implement the Message interface.
type MessageFunc func(*Buffer) error

func (e MessageFunc) BufferSize() int {
	return 1024
}

func (e MessageFunc) WriteBuffer(buf *Buffer) error {
	return e(buf)
}

// Convert to bytes message.
func Bytes(v []byte) Message {
	return MessageFunc(func(buf *Buffer) error {
		buf.WriteBytes(v)
		return nil
	})
}

// Convert to string message.
func String(v string) Message {
	return MessageFunc(func(buf *Buffer) error {
		buf.WriteString(v)
		return nil
	})
}

// Create a json message.
func Json(v interface{}) Message {
	return MessageFunc(func(buf *Buffer) error {
		return json.NewEncoder(buf).Encode(v)
	})
}

// Create a gob message.
func Gob(v interface{}) Message {
	return MessageFunc(func(buf *Buffer) error {
		return gob.NewEncoder(buf).Encode(v)
	})
}

// Create a xml message.
func Xml(v interface{}) Message {
	return MessageFunc(func(buf *Buffer) error {
		return xml.NewEncoder(buf).Encode(v)
	})
}
