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

	server, err := link.Listen("tcp", "0.0.0.0:57278", json, 0 /* sync send */, link.HandlerFunc(serverSessionLoop))
	checkErr(err)
	server.Listener()
	go server.Serve()
	select {}
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

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
