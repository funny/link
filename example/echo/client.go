package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/funny/link"
)

func main() {
	var addr string

	flag.StringVar(&addr, "addr", "127.0.0.1:10010", "echo server address")
	flag.Parse()

	session, err := link.Connect("tcp", addr, link.Packet(2, 1024*1024, 1024, binary.LittleEndian, TestCodec{}))
	if err != nil {
		panic(err)
	}

	go func() {
		var msg string
		for {
			if err := session.Receive(&msg); err != nil {
				break
			}
			fmt.Printf("%s\n", msg)
		}
	}()

	for {
		var msg string
		if _, err := fmt.Scanf("%s\n", &msg); err != nil {
			break
		}
		if err = session.Send(msg); err != nil {
			break
		}
	}

	session.Close()
	println("bye")
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
