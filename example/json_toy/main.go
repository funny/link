package main

import (
	"log"

	"github.com/funny/link"
	"github.com/funny/link/codec"
)

type AddReq struct {
	A, B int
}

type AddRsp struct {
	C int
}

func main() {
	json := codec.Json()
	json.Register(AddReq{})
	json.Register(AddRsp{})

	server, err := link.Listen("tcp", "0.0.0.0:0", json, 0 /* sync send */, link.HandlerFunc(serverSessionLoop))
	checkErr(err)
	addr := server.Listener().Addr().String()
	go server.Serve()

	client, err := link.Dial("tcp", addr, json, 0)
	checkErr(err)
	clientSessionLoop(client)
}

func serverSessionLoop(session *link.Session) {
	for {
		req, err := session.Receive()
		checkErr(err)

		err = session.Send(&AddRsp{
			req.(*AddReq).A + req.(*AddReq).B,
		})
		checkErr(err)
	}
}

func clientSessionLoop(session *link.Session) {
	for i := 0; i < 10; i++ {
		err := session.Send(&AddReq{
			i, i,
		})
		checkErr(err)
		log.Printf("Send: %d + %d", i, i)

		rsp, err := session.Receive()
		checkErr(err)
		log.Printf("Receive: %d", rsp.(*AddRsp).C)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
