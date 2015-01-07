package link

import (
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
)

// Convert to bytes message.
func Bytes(v []byte) Encoder {
	return func(buffer *OutBuffer) error {
		buffer.WriteBytes(v)
		return nil
	}
}

// Convert to string message.
func String(v string) Encoder {
	return func(buffer *OutBuffer) error {
		buffer.WriteString(v)
		return nil
	}
}

// Create a json message.
func Json(v interface{}) Encoder {
	return func(buffer *OutBuffer) error {
		return json.NewEncoder(buffer).Encode(v)
	}
}

// Create a gob message.
func Gob(v interface{}) Encoder {
	return func(buffer *OutBuffer) error {
		return gob.NewEncoder(buffer).Encode(v)
	}
}

// Create a xml message.
func Xml(v interface{}) Encoder {
	return func(buffer *OutBuffer) error {
		return xml.NewEncoder(buffer).Encode(v)
	}
}
