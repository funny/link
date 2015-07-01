package main

import (
	"flag"
	"fmt"
	"github.com/funny/link"
)

func main() {
	var addr string

	flag.StringVar(&addr, "addr", "127.0.0.1:10010", "echo server address")
	flag.Parse()

	session, err := link.Connect("tcp://"+addr, link.String(link.Uint16BE))
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
