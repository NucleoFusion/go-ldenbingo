package neto

import (
	"encoding/hex"
	"testing"
)

func TestLz4BlockDecode(t *testing.T) {
	extData, _ := hex.DecodeString("d2000003ebdfdc00649305a3666f6fa36261720a00ffffffc9506fa3626172")
	out, err := decompressLz4Block(extData)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(out) != 1003 {
		t.Fatalf("expected 1003 bytes, got %d", len(out))
	}
}
