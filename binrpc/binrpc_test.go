package binrpc

import (
	"bytes"
	"testing"
)

func TestRouting(t *testing.T) {
}

func TestEncodeString(t *testing.T) {
	out := new(bytes.Buffer)
	var myString = BinRpcString("dispatcher.list")
	err := myString.Encode(out)
	if err != nil {
		t.Log("Failed to encode string:", err)
		t.Fail()
	}
	t.Logf("Output: %x", out.Bytes())
}
