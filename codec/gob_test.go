package codec

import (
	"bytes"
	"testing"
)

func Test_Gob(t *testing.T) {
	type MyMessage1 struct {
		Field1 string
		Field2 int
	}

	type MyMessage2 struct {
		Field1 int
		Field2 string
	}

	protocol := Gob()
	protocol.Register(&MyMessage1{})
	protocol.RegisterName("msg2", &MyMessage2{})

	var stream bytes.Buffer

	sendMsg1 := MyMessage1{
		Field1: "abc",
		Field2: 123,
	}

	codec := protocol.NewCodec(&stream)
	err := codec.Send(&sendMsg1)
	if err != nil {
		t.Fatal(err)
	}

	recvMsg1, err := codec.Receive()
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := recvMsg1.(*MyMessage1); !ok {
		t.Fatalf("message type not match: %#v", recvMsg1)
	}

	if sendMsg1 != *(recvMsg1.(*MyMessage1)) {
		t.Fatalf("message not match: %v, %v", sendMsg1, recvMsg1)
	}

	sendMsg2 := MyMessage2{
		Field1: 123,
		Field2: "abc",
	}

	err = codec.Send(&sendMsg2)
	if err != nil {
		t.Fatal(err)
	}

	recvMsg2, err := codec.Receive()
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := recvMsg2.(*MyMessage2); !ok {
		t.Fatalf("message type not match: %#v", recvMsg2)
	}

	if sendMsg2 != *(recvMsg2.(*MyMessage2)) {
		t.Fatalf("message not match %v, %v", sendMsg1, recvMsg1)
	}
}
