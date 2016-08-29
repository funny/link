package codec

import (
	"encoding/binary"
	"testing"
)

func Test_FixLen(t *testing.T) {
	base := JsonTestProtocol()
	protocol := FixLen(base, 2, binary.LittleEndian, 1024, 1024)
	JsonTest(t, protocol)
}
