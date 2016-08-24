package codec

import (
	"encoding/json"
	"io"
	"reflect"

	"github.com/funny/link"
)

type JsonProtocol struct {
	types map[string]reflect.Type
	names map[reflect.Type]string
}

func Json() *JsonProtocol {
	return &JsonProtocol{
		types: make(map[string]reflect.Type),
		names: make(map[reflect.Type]string),
	}
}

func (j *JsonProtocol) Register(t interface{}) {
	rt := reflect.TypeOf(t)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	name := rt.PkgPath() + "/" + rt.Name()
	j.types[name] = rt
	j.names[rt] = name
}

func (j *JsonProtocol) RegisterName(name string, t interface{}) {
	rt := reflect.TypeOf(t)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	j.types[name] = rt
	j.names[rt] = name
}

func (j *JsonProtocol) NewCodec(rw io.ReadWriter) link.Codec {
	codec := &jsonCodec{
		p:       j,
		encoder: json.NewEncoder(rw),
		decoder: json.NewDecoder(rw),
	}
	codec.closer, _ = rw.(io.Closer)
	return codec
}

type jsonIn struct {
	Type    string           `json:"t"`
	Message *json.RawMessage `json:"m"`
}

type jsonOut struct {
	Type    string      `json:"t"`
	Message interface{} `json:"m"`
}

type jsonCodec struct {
	p       *JsonProtocol
	closer  io.Closer
	encoder *json.Encoder
	decoder *json.Decoder
}

func (c *jsonCodec) Receive() (interface{}, error) {
	var in jsonIn
	err := c.decoder.Decode(&in)
	if err != nil {
		return nil, err
	}
	var message interface{}
	if in.Type != "" {
		if t, exists := c.p.types[in.Type]; exists {
			message = reflect.New(t).Interface()
		}
	}
	err = json.Unmarshal(*in.Message, &message)
	if err != nil {
		return nil, err
	}
	return message, nil
}

func (c *jsonCodec) Send(msg interface{}) error {
	var out jsonOut
	t := reflect.TypeOf(msg)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if name, exists := c.p.names[t]; exists {
		out.Type = name
	}
	out.Message = msg
	return c.encoder.Encode(&out)
}

func (c *jsonCodec) Close() error {
	if c.closer != nil {
		return c.closer.Close()
	}
	return nil
}
