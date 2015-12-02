package main

import (
	"encoding/binary"
	"flag"
	"io"
	"io/ioutil"

	"github.com/funny/link"
)

func main() {
	var addr string

	flag.StringVar(&addr, "addr", ":10010", "echo server address")
	flag.Parse()

	server, err := link.Serve("tcp", addr, link.Packet(2, 1024*1024, 4096, binary.LittleEndian, TestCodec{}))
	if err != nil {
		panic(err)
	}

	println("server start:", server.Listener().Addr().String())
	for {
		session, err := server.Accept()
		if err != nil {
			break
		}
		go func() {
			//addr := session.Conn().RemoteAddr().String()
			//println("client", addr, "connected")
			for {
				var msg []byte
				if err = session.Receive(&msg); err != nil {
					break
				}
				if err = session.Send(msg); err != nil {
					break
				}
			}
			//println("client", addr, "closed")
		}()
	}
}

type TestCodec struct {
}

type TestEncoder struct {
	w io.Writer
}

type TestDecoder struct {
	r io.Reader
}

func (codec TestCodec) NewEncoder(w io.Writer) link.Encoder {
	return &TestEncoder{w}
}

func (codec TestCodec) NewDecoder(r io.Reader) link.Decoder {
	return &TestDecoder{r}
}

func (encoder *TestEncoder) Encode(msg interface{}) error {
	_, err := encoder.w.Write(msg.([]byte))
	return err
}

func (decoder *TestDecoder) Decode(msg interface{}) error {
	// We use ReadAll() here because we know the reader is a buffer object not a real net.Conn
	d, err := ioutil.ReadAll(decoder.r)
	if err != nil {
		return err
	}
	*(msg.(*[]byte)) = d
	return nil
}
