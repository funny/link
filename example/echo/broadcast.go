package main

import (
	"encoding/binary"
	"flag"
	"io"
	"io/ioutil"
	"time"

	"github.com/funny/link"
)

// This is broadcast server demo work with the echo_client.
// usage:
//     cd src/github.com/funny/link
//     go generate channel.go
//     go run example/echo/broadcast.go
func main() {
	var addr string

	flag.StringVar(&addr, "addr", ":10010", "echo server address")
	flag.Parse()

	server, err := link.Serve("tcp", addr, link.Async(1024, link.Packet(2, 1024*1024, 1024, binary.LittleEndian, TestCodec{})))
	if err != nil {
		panic(err)
	}
	println("server start:", server.Listener().Addr().String())

	channel := link.NewUint64Channel()
	go func() {
		for range time.Tick(time.Second * 2) {
			now := "from channel: " + time.Now().Format("2006-01-02 15:04:05")
			channel.Fetch(func(session *link.Session) {
				session.Send(now)
			})
		}
	}()

	for {
		session, err := server.Accept()
		if err != nil {
			break
		}

		go func() {
			addr := session.Conn().RemoteAddr().String()
			println("client", addr, "connected")

			channel.Put(session.Id(), session)

			for {
				var msg string
				if err := session.Receive(&msg); err != nil {
					break
				}
				println(addr, "say:", msg)
				channel.Fetch(func(session *link.Session) {
					session.Send("from " + addr + ": " + string(msg))
				})
			}

			println("client", addr, "closed")
			channel.Remove(session.Id())
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
	_, err := encoder.w.Write([]byte(msg.(string)))
	return err
}

func (decoder *TestDecoder) Decode(msg interface{}) error {
	d, err := ioutil.ReadAll(decoder.r)
	if err != nil {
		return err
	}
	*(msg.(*string)) = string(d)
	return nil
}
