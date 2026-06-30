package neto

import (
	"github.com/vmihailenco/msgpack/v5"
)

func DecodeCompressedValue(data []byte, dst any) error {
	decompressed, err := Decompress(data)
	if err != nil {
		return err
	}

	return msgpack.Unmarshal(decompressed, dst)
}
