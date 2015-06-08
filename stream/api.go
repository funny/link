package stream

import (
	"encoding/gob"

	"github.com/funny/binary"
)

type InMessage interface {
	Unmarshal(*binary.Reader) error
}

type OutMessage interface {
	Marshal(*binary.Writer) error
}

type GOB struct{ V interface{} }

func (msg GOB) Marshal(w *binary.Writer) error {
	return gob.NewEncoder(w).Encode(msg.V)
}

func (msg GOB) Unmarshal(r *binary.Reader) error {
	return gob.NewDecoder(r).Decode(msg.V)
}
