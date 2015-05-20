package link

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"io"
)

type Limiter interface {
	Limit(conn *Conn) *io.LimitedReader
}

type JSON struct {
	V interface{}
	S Spliter
}

func (j JSON) Send(conn *Conn) error {
	if limiter, ok := j.S.(Limiter); ok {
		return json.NewDecoder(limiter.Limit(conn)).Decode(j.V)
	}
	b, err := json.Marshal(j.V)
	if err != nil {
		return err
	}
	conn.WritePacket(b, j.S)
	return nil
}

func (j JSON) Receive(conn *Conn) error {
	b := conn.ReadPacket(j.S)
	return json.Unmarshal(b, j.V)
}

type GOB struct {
	V interface{}
	S Spliter
}

func (g GOB) Send(conn *Conn) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(g.V); err != nil {
		return err
	}
	conn.WritePacket(buf.Bytes(), g.S)
	return nil
}

func (g GOB) Receive(conn *Conn) error {
	if limiter, ok := g.S.(Limiter); ok {
		return gob.NewDecoder(limiter.Limit(conn)).Decode(g.V)
	}
	b := conn.ReadPacket(g.S)
	r := bytes.NewReader(b)
	return gob.NewDecoder(r).Decode(g.V)
}
