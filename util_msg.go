package link

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"github.com/funny/binary"
)

type JSON struct {
	V interface{}
	S binary.Spliter
}

func (j JSON) Send(w *binary.Writer) error {
	b, err := json.Marshal(j.V)
	if err != nil {
		return err
	}
	w.WritePacket(b, j.S)
	return nil
}

func (j JSON) Receive(r *binary.Reader) error {
	if limiter, ok := j.S.(binary.Limiter); ok {
		return json.NewDecoder(limiter.Limit(r)).Decode(j.V)
	}
	b := r.ReadPacket(j.S)
	return json.Unmarshal(b, j.V)
}

type GOB struct {
	V interface{}
	S binary.Spliter
}

func (g GOB) Send(w *binary.Writer) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(g.V); err != nil {
		return err
	}
	w.WritePacket(buf.Bytes(), g.S)
	return nil
}

func (g GOB) Receive(r *binary.Reader) error {
	if limiter, ok := g.S.(binary.Limiter); ok {
		return gob.NewDecoder(limiter.Limit(r)).Decode(g.V)
	}
	b := r.ReadPacket(g.S)
	rd := bytes.NewReader(b)
	return gob.NewDecoder(rd).Decode(g.V)
}
