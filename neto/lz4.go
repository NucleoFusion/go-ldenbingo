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
// Lz4Block (ext 99) or Lz4BlockArray (ext 98) format. Otherwise, it returns the payload as-is.
//
// Both formats are verified directly against MessagePack-CSharp's actual source
// (MessagePackSerializer.cs: ToLZ4BinaryCore / TryDecompress) and against a real captured
// packet from a live EldenBingoServerStandalone instance, decompressed successfully.
//
// Lz4Block (ext 99):
//
//	ext payload = [msgpack-encoded int32: uncompressed_length] + [raw LZ4 block bytes]
//
// Lz4BlockArray (ext 98):
//
//	[ ext(98, <data>), bin, bin, ... ]
//
//	Confirmed quirk (verified against real C# source and real bytes): the ext header's
//	own data bytes ARE the FIRST chunk's msgpack-encoded uncompressed length — there is
//	no separate "ext payload to extract". DecodeExtHeader consumes those bytes as part
//	of the header per the msgpack spec, so they must be read via ReadFull and decoded
//	as an int directly, NOT re-read via a subsequent DecodeInt32 call on the main stream
//	(that would read past them into the next array element and fail). Any additional
//	chunks beyond the first (when arrLen > 2) have their lengths read as plain ints
//	directly off the continuing stream. Each declared chunk is followed by a `bin`
//	element containing that chunk's actual LZ4-compressed bytes (block format, not
//	frame format — see github.com/lz4/lz4/blob/dev/doc/lz4_Block_format.md).
//
// In both cases the "raw LZ4 block" is LZ4 *block* format, decompressed via
// github.com/pierrec/lz4/v4's UncompressBlock, which requires the destination
// buffer to be pre-sized to the (now known) uncompressed length.
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
		return decompressLz4BlockArray(payload)
	} else if isExt {
		return decompressLz4Block(payload)
	}

	// Not compressed, or not an ext/array we recognize — return as-is.
	return payload, nil
}

// decompressLz4BlockArray handles the ext-98 (Lz4BlockArray) case. payload's first byte
// is already known to be an array header.
func decompressLz4BlockArray(payload []byte) ([]byte, error) {
	dec := msgpack.NewDecoder(bytes.NewReader(payload))

	arrLen, err := dec.DecodeArrayLen()
	if err != nil {
		return nil, fmt.Errorf("failed to decode array len: %w", err)
	}
	if arrLen <= 1 {
		return payload, nil
	}

	code, err := dec.PeekCode()
	if err != nil {
		return nil, fmt.Errorf("failed to peek array element: %w", err)
	}
	if !msgpcode.IsExt(code) {
		return payload, nil
	}

	extID, extLen, err := dec.DecodeExtHeader()
	if err != nil {
		return nil, fmt.Errorf("failed to decode ext header: %w", err)
	}
	if extID != ExtLz4BlockArray {
		return payload, nil
	}

	sequenceCount := arrLen - 1
	uncompressedLengths := make([]int32, sequenceCount)

	// First length comes from the ext header's own data bytes (see doc comment above).
	firstLenBytes := make([]byte, extLen)
	if err := dec.ReadFull(firstLenBytes); err != nil {
		return nil, fmt.Errorf("failed to read ext header data bytes: %w", err)
	}
	firstLen, err := decodeMsgpackInt32FromBytes(firstLenBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode first chunk length from ext data: %w", err)
	}
	uncompressedLengths[0] = firstLen

	// Any remaining lengths are plain msgpack ints read directly off the continuing stream.
	for i := 1; i < sequenceCount; i++ {
		length, err := dec.DecodeInt32()
		if err != nil {
			return nil, fmt.Errorf("failed to decode chunk uncompressed length at index %d: %w", i, err)
		}
		uncompressedLengths[i] = length
	}

	decompressedChunks := make([][]byte, sequenceCount)
	totalLen := 0
	for i := 0; i < sequenceCount; i++ {
		compressedChunk, err := dec.DecodeBytes()
		if err != nil {
			return nil, fmt.Errorf("failed to decode compressed chunk %d: %w", i, err)
		}
		uncompressedLen := uncompressedLengths[i]
		decompressedChunk := make([]byte, uncompressedLen)
		n, err := lz4.UncompressBlock(compressedChunk, decompressedChunk)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress chunk %d via LZ4: %w", i, err)
		}
		if int32(n) != uncompressedLen {
			return nil, fmt.Errorf("decompressed size mismatch for chunk %d: expected %d, got %d", i, uncompressedLen, n)
		}
		decompressedChunks[i] = decompressedChunk
		totalLen += n
	}

	result := make([]byte, 0, totalLen)
	for _, chunk := range decompressedChunks {
		result = append(result, chunk...)
	}
	return result, nil
}

// decompressLz4Block handles the ext-99 (Lz4Block) case. payload's first byte is already
// known to be an ext header.
func decompressLz4Block(payload []byte) ([]byte, error) {
	dec := msgpack.NewDecoder(bytes.NewReader(payload))

	extID, extLen, err := dec.DecodeExtHeader()
	if err != nil {
		// Not a valid ext header at all — treat as uncompressed passthrough.
		return payload, nil
	}
	if extID != ExtLz4Block {
		return payload, nil
	}

	extData := make([]byte, extLen)
	if err := dec.ReadFull(extData); err != nil {
		return nil, fmt.Errorf("failed to read ext payload: %w", err)
	}

	// ext payload = [msgpack int32: uncompressed_length][raw LZ4 block]
	reader := bytes.NewReader(extData)
	metaDec := msgpack.NewDecoder(reader)

	uncompressedLen, err := metaDec.DecodeInt32()
	if err != nil {
		return nil, fmt.Errorf("failed to decode uncompressed length: %w", err)
	}

	remainingLen := reader.Len()
	compressedBytes := make([]byte, remainingLen)
	if _, err := reader.Read(compressedBytes); err != nil {
		return nil, fmt.Errorf("failed to read remaining LZ4 block bytes: %w", err)
	}

	decompressed := make([]byte, uncompressedLen)
	n, err := lz4.UncompressBlock(compressedBytes, decompressed)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress LZ4 block: %w", err)
	}
	if int32(n) != uncompressedLen {
		return nil, fmt.Errorf("decompressed size mismatch: expected %d, got %d", uncompressedLen, n)
	}

	return decompressed, nil
}

// decodeMsgpackInt32FromBytes decodes a msgpack-encoded integer from a raw byte slice
// (e.g. ext header data bytes that were consumed separately from the main stream).
func decodeMsgpackInt32FromBytes(b []byte) (int32, error) {
	dec := msgpack.NewDecoder(bytes.NewReader(b))
	return dec.DecodeInt32()
}
