package main

import (
	"fmt"
	"github.com/funny/link"
)

// This is an echo client demo work with the echo_server.
// usage:
//     go run echo_client/main.go
func main() {
	client, err := link.Dial("tcp", "127.0.0.1:10010")
	if err != nil {
		panic(err)
	}
	go client.Handle(func(msg *link.InBuffer) {
		println(string(msg.Data))
	})

	for {
		var input string
		if _, err := fmt.Scanf("%s\n", &input); err != nil {
			break
		}
		client.Send(link.String(input))
	}

	client.Close(nil)

	println("bye")
}
