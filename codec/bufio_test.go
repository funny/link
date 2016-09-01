package codec

import (
	"encoding/binary"
	"testing"
)

func Test_Bufio(t *testing.T) {
	JsonTest(t, Bufio(FixLen(JsonTestProtocol(), 2, binary.LittleEndian, 64*1024, 64*1024), 1024, 1024))
}
