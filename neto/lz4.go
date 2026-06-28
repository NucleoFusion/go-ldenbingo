package neto

import (
	"bytes"
	"fmt"

	"github.com/pierrec/lz4/v4"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/vmihailenco/msgpack/v5/msgpcode"
)

const (
	ExtLz4BlockArray = 98
	ExtLz4Block      = 99
)

// Decompress decompresses a MessagePack payload if it is compressed using C# MessagePack-CSharp's
// Lz4Block (99) or Lz4BlockArray (98) format. Otherwise, it returns the payload as-is.
func Decompress(payload []byte) ([]byte, error) {
	if len(payload) == 0 {
		return payload, nil
	}

	firstByte := payload[0]

	// Check if array type (0x90 to 0x9f, 0xdc, 0xdd)
	isArray := (firstByte >= 0x90 && firstByte <= 0x9f) || firstByte == 0xdc || firstByte == 0xdd
	// Check if extension type (0xd4 to 0xd8, 0xc7 to 0xc9)
	isExt := (firstByte >= 0xd4 && firstByte <= 0xd8) || firstByte == 0xc7 || firstByte == 0xc8 || firstByte == 0xc9

	if isArray {
		dec := msgpack.NewDecoder(bytes.NewReader(payload))
		arrLen, err := dec.DecodeArrayLen()
		if err != nil {
			return nil, fmt.Errorf("failed to decode array len: %w", err)
		}
		if arrLen > 1 {
			// Peek at the type of the next array element without consuming it
			code, err := dec.PeekCode()
			if err != nil {
				return nil, fmt.Errorf("failed to peek array element: %w", err)
			}
			isNextExt := msgpcode.IsExt(code)
			if isNextExt {
				extID, extLen, err := dec.DecodeExtHeader()
				if err != nil {
					return nil, fmt.Errorf("failed to decode ext header: %w", err)
				}
				if extID == ExtLz4BlockArray {
					// NEED TO CONFIRM: how to read exactly extLen raw bytes here.
					// dec.DecodeBytes() is wrong — that expects a bin-format header
					// (0xc4/0xc5/0xc6), not raw post-ext-header payload bytes.
					extData, err := readExtPayload(dec, extLen)
					if err != nil {
						return nil, fmt.Errorf("failed to read ext payload: %w", err)
					}

					// Parse the metadata inside the extension payload
					reader := bytes.NewReader(extData)
					metaDec := msgpack.NewDecoder(reader)
					totalUncompressedLen, err := metaDec.DecodeInt32()
					if err != nil {
						return nil, fmt.Errorf("failed to decode total uncompressed length: %w", err)
					}
					chunkCount := arrLen - 1
					uncompressedChunkLengths := make([]int32, chunkCount)
					for i := 0; i < chunkCount; i++ {
						uncompressedLen, err := metaDec.DecodeInt32()
						if err != nil {
							return nil, fmt.Errorf("failed to decode chunk uncompressed length at index %d: %w", i, err)
						}
						uncompressedChunkLengths[i] = uncompressedLen
					}
					// Now read the remaining binary elements of the array and decompress each chunk
					decompressedPayload := make([]byte, totalUncompressedLen)
					offset := 0
					for i := 0; i < chunkCount; i++ {
						binBytes, err := dec.DecodeBytes()
						if err != nil {
							return nil, fmt.Errorf("failed to decode compressed chunk %d: %w", i, err)
						}
						chunkUncompressedLen := uncompressedChunkLengths[i]
						decompressedChunk := make([]byte, chunkUncompressedLen)
						n, err := lz4.UncompressBlock(binBytes, decompressedChunk)
						if err != nil {
							return nil, fmt.Errorf("failed to decompress chunk %d using LZ4: %w", i, err)
						}
						if int32(n) != chunkUncompressedLen {
							return nil, fmt.Errorf("decompressed size mismatch for chunk %d: expected %d, got %d", i, chunkUncompressedLen, n)
						}
						copy(decompressedPayload[offset:], decompressedChunk)
						offset += n
					}
					return decompressedPayload, nil
				}
			}
		}
	} else if isExt {
		// Try parsing as Lz4Block (99)
		payloadReader := bytes.NewReader(payload)
		dec := msgpack.NewDecoder(payloadReader)
		extID, extLen, err := dec.DecodeExtHeader()
		if err == nil && extID == ExtLz4Block {
			extData := make([]byte, extLen)
			if err := dec.ReadFull(extData); err != nil {
				return nil, fmt.Errorf("failed to read ext payload: %w", err)
			}

			reader := bytes.NewReader(extData)
			metaDec := msgpack.NewDecoder(reader)
			uncompressedLen, err := metaDec.DecodeInt32()
			if err != nil {
				return nil, fmt.Errorf("failed to decode uncompressed length: %w", err)
			}
			// The remaining bytes of the extension payload is the compressed data
			offset := len(extData) - reader.Len()
			compressedBytes := extData[offset:]
			decompressedPayload := make([]byte, uncompressedLen)
			n, err := lz4.UncompressBlock(compressedBytes, decompressedPayload)
			if err != nil {
				return nil, fmt.Errorf("failed to decompress LZ4 block: %w", err)
			}
			if int32(n) != uncompressedLen {
				return nil, fmt.Errorf("decompressed size mismatch: expected %d, got %d", uncompressedLen, n)
			}
			return decompressedPayload, nil
		}
	}

	// Not compressed
	return payload, nil
}

func readExtPayload(dec *msgpack.Decoder, extLen int) ([]byte, error) {
	buf := make([]byte, extLen)
	if err := dec.ReadFull(buf); err != nil {
		return nil, fmt.Errorf("failed to read ext payload: %w", err)
	}
	return buf, nil
}
