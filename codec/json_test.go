package codec

import (
	"bytes"
	"testing"

	"github.com/funny/link"
)

type MyMessage1 struct {
	Field1 string
	Field2 int
}

type MyMessage2 struct {
	Field1 int
	Field2 string
}

func JsonTestProtocol() *JsonProtocol {
	protocol := Json()
	protocol.Register(MyMessage1{})
	protocol.RegisterName("msg2", &MyMessage2{})
	return protocol
}

func JsonTest(t *testing.T, protocol link.Protocol) {
	var stream bytes.Buffer

	codec, _ := protocol.NewCodec(&stream)

	sendMsg1 := MyMessage1{
		Field1: "abc",
		Field2: 123,
	}

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

	sendMsg3 := map[string]MyMessage1{
		"a": MyMessage1{"abc", 123},
		"b": MyMessage1{"def", 456},
	}

	err = codec.Send(&sendMsg3)
	if err != nil {
		t.Fatal(err)
	}

	recvMsg3, err := codec.Receive()
	if err != nil {
		t.Fatal(err)
	}

	recvMap3, ok := recvMsg3.(map[string]interface{})
	if !ok {
		t.Fatalf("message type not match: %#v", recvMsg3)
	}

	recvMepLevel2, ok := recvMap3["a"].(map[string]interface{})
	if !ok {
		t.Fatalf("map type not match: %v", recvMsg3)
	}

	if recvMepLevel2["Field1"].(string) != "abc" {
		t.Fatalf("map not match %v", recvMepLevel2)
	}
}

func Test_Json(t *testing.T) {
	protocol := JsonTestProtocol()
	JsonTest(t, protocol)
}
