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

type encoder func(*OutBuffer) error

func (e encoder) OutBufferSize() int {
	return 1024
}

func (e encoder) WriteOutBuffer(out *OutBuffer) error {
	return e(out)
}

// Convert to bytes message.
func Bytes(v []byte) Message {
	return encoder(func(buffer *OutBuffer) error {
		buffer.WriteBytes(v)
		return nil
	})
}

// Convert to string message.
func String(v string) Message {
	return encoder(func(buffer *OutBuffer) error {
		buffer.WriteString(v)
		return nil
	})
}

// Create a json message.
func Json(v interface{}) Message {
	return encoder(func(buffer *OutBuffer) error {
		return json.NewEncoder(buffer).Encode(v)
	})
}

// Create a gob message.
func Gob(v interface{}) Message {
	return encoder(func(buffer *OutBuffer) error {
		return gob.NewEncoder(buffer).Encode(v)
	})
}

// Create a xml message.
func Xml(v interface{}) Message {
	return encoder(func(buffer *OutBuffer) error {
		return xml.NewEncoder(buffer).Encode(v)
	})
}
