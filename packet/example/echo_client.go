package main

import (
	"flag"
	"fmt"
	"github.com/funny/binary"
	"github.com/funny/link"
	"github.com/funny/link/packet"
)

var (
	addr = flag.String("addr", "127.0.0.1:10010", "echo server address")
)

func main() {
	flag.Parse()

	session, err := link.Connect("tcp", *addr, packet.New(
		binary.SplitByUint16BE, 1024, 1024, 1024,
	))
	if err != nil {
		panic(err)
	}

	go func() {
		var msg packet.RAW
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
		session.Send(packet.RAW(msg))
	}

	session.Close()
	println("bye")
}
