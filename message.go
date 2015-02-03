package link

import (
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
)

type Message interface {
	OutBufferSize() int
	WriteOutBuffer(*OutBuffer) error
}

// A func implement the Message interface.
type MessageFunc func(*OutBuffer) error

func (e MessageFunc) OutBufferSize() int {
	return 1024
}

func (e MessageFunc) WriteOutBuffer(out *OutBuffer) error {
	return e(out)
}

// Convert to bytes message.
func Bytes(v []byte) Message {
	return MessageFunc(func(buffer *OutBuffer) error {
		buffer.WriteBytes(v)
		return nil
	})
}

// Convert to string message.
func String(v string) Message {
	return MessageFunc(func(buffer *OutBuffer) error {
		buffer.WriteString(v)
		return nil
	})
}

// Create a json message.
func Json(v interface{}) Message {
	return MessageFunc(func(buffer *OutBuffer) error {
		return json.NewEncoder(buffer).Encode(v)
	})
}

// Create a gob message.
func Gob(v interface{}) Message {
	return MessageFunc(func(buffer *OutBuffer) error {
		return gob.NewEncoder(buffer).Encode(v)
	})
}

// Create a xml message.
func Xml(v interface{}) Message {
	return MessageFunc(func(buffer *OutBuffer) error {
		return xml.NewEncoder(buffer).Encode(v)
	})
}
