package main

import (
	"flag"
	"fmt"
	"github.com/funny/link"
	"github.com/funny/link/protocol/fixhead"
)

var (
	serverAddr = flag.String("addr", "127.0.0.1:10010", "echo server address")
)

// This is an echo client demo work with the echo_server.
// usage:
//     go run main.go
//     go run main.go -addr="127.0.0.1:10010"
func main() {
	flag.Parse()

	client, err := link.Dial("tcp", *serverAddr, fixhead.Uint16BE, nil)
	if err != nil {
		panic(err)
	}

	go client.Process(link.DecodeFunc(func(buf *link.Buffer) (link.Request, error) {
		println(string(buf.Data))
		return nil, nil
	}))

	for {
		var input string
		if _, err := fmt.Scanf("%s\n", &input); err != nil {
			break
		}
		client.Send(link.String(input))
	}

	client.Close()
	println("bye")
}
