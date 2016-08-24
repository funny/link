package codec

import (
	"encoding/gob"
	"errors"
	"io"
	"reflect"

	"github.com/funny/link"
)

var ErrGobUnknow = errors.New("GoProtocol: unknow message type")

type GobProtocol struct {
	types map[string]reflect.Type
	names map[reflect.Type]string
}

func Gob() *GobProtocol {
	return &GobProtocol{
		types: make(map[string]reflect.Type),
		names: make(map[reflect.Type]string),
	}
}

func (g *GobProtocol) Register(t interface{}) {
	rt := reflect.TypeOf(t)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	name := rt.PkgPath() + "/" + rt.Name()
	g.types[name] = rt
	g.names[rt] = name
}

func (g *GobProtocol) RegisterName(name string, t interface{}) {
	rt := reflect.TypeOf(t)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	g.types[name] = rt
	g.names[rt] = name
}

func (g *GobProtocol) NewCodec(rw io.ReadWriter) link.Codec {
	codec := &gobCodec{
		p:       g,
		encoder: gob.NewEncoder(rw),
		decoder: gob.NewDecoder(rw),
	}
	codec.closer, _ = rw.(io.Closer)
	return codec
}

type gobMsg struct {
	Type string
}

type gobCodec struct {
	p       *GobProtocol
	closer  io.Closer
	encoder *gob.Encoder
	decoder *gob.Decoder
}

func (c *gobCodec) Receive() (interface{}, error) {
	var in gobMsg
	err := c.decoder.Decode(&in)
	if err != nil {
		return nil, err
	}
	var message interface{}
	if t, exists := c.p.types[in.Type]; exists {
		message = reflect.New(t).Interface()
	} else {
		return nil, ErrGobUnknow
	}
	err = c.decoder.Decode(message)
	if err != nil {
		return nil, err
	}
	return message, nil
}

func (c *gobCodec) Send(msg interface{}) error {
	var out gobMsg
	t := reflect.TypeOf(msg)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if name, exists := c.p.names[t]; exists {
		out.Type = name
	} else {
		return ErrGobUnknow
	}
	err := c.encoder.Encode(&out)
	if err != nil {
		return err
	}
	return c.encoder.Encode(msg)
}
func (c *gobCodec) Close() error {
	if c.closer != nil {
		return c.closer.Close()
	}
	return nil
}
