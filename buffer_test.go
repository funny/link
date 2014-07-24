package packnet

import "bytes"
import "testing"
import "encoding/binary"

func Test_Buffer(t *testing.T) {
	buffer := NewBuffer(nil, binary.BigEndian)

	buffer.WriteUint8(1)
	if buffer.ReadUint8() != 1 {
		t.Fatal("buffer.ReadUint8() != 1")
	}

	buffer.WriteByte(99)
	if buffer.ReadByte() != 99 {
		t.Fatal("buffer.ReadByte() != 99")
	}

	buffer.WriteInt8(-2)
	if buffer.ReadInt8() != -2 {
		t.Fatal("buffer.ReadInt8() != -2")
	}

	buffer.WriteUint16(0xFFEE)
	if buffer.ReadUint16() != 0xFFEE {
		t.Fatal("buffer.ReadUint16() != 0xFFEE")
	}

	buffer.WriteInt16(0x7FEE)
	if buffer.ReadInt16() != 0x7FEE {
		t.Fatal("buffer.ReadInt16() != 0x7FEE")
	}

	buffer.WriteUint32(0xFFEEDDCC)
	if buffer.ReadUint32() != 0xFFEEDDCC {
		t.Fatal("buffer.ReadUint32() != 0xFFEEDDCC")
	}

	buffer.WriteInt32(0x7FEEDDCC)
	if buffer.ReadInt32() != 0x7FEEDDCC {
		t.Fatal("buffer.ReadInt32() != 0x7FEEDDCC")
	}

	buffer.WriteUint64(0xFFEEDDCCBBAA9988)
	if buffer.ReadUint64() != 0xFFEEDDCCBBAA9988 {
		t.Fatal("buffer.ReadUint64() != 0xFFEEDDCCBBAA9988")
	}

	buffer.WriteInt64(0x7FEEDDCCBBAA9988)
	if buffer.ReadInt64() != 0x7FEEDDCCBBAA9988 {
		t.Fatal("buffer.ReadInt64() != 0x7FEEDDCCBBAA9988")
	}

	buffer.WriteRune('好')
	if buffer.ReadRune() != '好' {
		t.Fatal(`buffer.ReadRune() != '好'`)
	}

	buffer.WriteString("Hello")
	if buffer.ReadString(5) != "Hello" {
		t.Fatal(`buffer.ReadString() != "Hello"`)
	}

	buffer.WriteBytes([]byte("Hello"))
	if bytes.Equal(buffer.ReadBytes(5), []byte("Hello")) != true {
		t.Fatal(`bytes.Equal(buffer.ReadBytes(5), []byte("Hello")) != true`)
	}

	buffer.WriteBytes([]byte("Hello"))
	if bytes.Equal(buffer.ReadSlice(5), []byte("Hello")) != true {
		t.Fatal(`bytes.Equal(buffer.ReadSlice(5), []byte("Hello")) != true`)
	}
}
