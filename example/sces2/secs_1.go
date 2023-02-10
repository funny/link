package main

import (
	"github.com/wolimst/lib-secs2-hsms-go/pkg/ast"
	"github.com/younglifestyle/link"
	"github.com/younglifestyle/link/codec"
	"log"
)

func main() {
	secsii := codec.SECSII()

	server, err := link.Listen("tcp", "0.0.0.0:5555", secsii, 0 /* sync send */, link.HandlerFunc(serverSessionLoop))
	if err != nil {
		log.Fatal(err)
	}
	//addr := server.Listener().Addr().String()
	server.Serve()

	//client, err := link.Dial("tcp", "0.0.0.0:5555", secsii, 0)
	//checkErr(err)
	//clientSessionLoop(client)
}

// SEND 0 0 0 A FF FF 00 00 00 01 7F 00 00 00   Length = 10          (Select.req  )
func serverSessionLoop(session *link.Session) {
	for {
		req, err := session.Receive()
		if err != nil {
			log.Println("recv error : ", err)
			return
		}
		message, ok := req.(ast.HSMSMessage)
		if !ok {
			log.Println("not ok")
		}

		// Receive: select.req
		log.Printf("Receive: %s", message.Type())
	}
}

func clientSessionLoop(session *link.Session) {
	for {
		rsp, err := session.Receive()
		if err != nil {
			log.Println("recv error : ", err)
			continue
		}
		log.Printf("Receive: %v", rsp)
	}
}
