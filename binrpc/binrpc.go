package binrpc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math/rand"
)

/*
	Kamailio's binrpc protocol consists of a variable-byte header
	with an attached payload.

	HEADER:

	|   4 bits    |    4 bits     | 4 bits | 2b | 2b | variable    | variable |
	| BinRpcMagic | BinRpcVersion | Flags  | LL | CL | data length | cookie   |

	BinRpcMagic and BinRpcVersion are constants
	Flags ???
	LL = (size in bytes of the data length) - 1
	CL = (size in bytes of the cookie ) - 1
	data length = total number of bytes of the payload (excluding this header)
	cookie = arbitrary identifier

	PAYLOAD:

	| 1 bit     | 3 bits | 4 bits | variable              | variable       |
	| Size flag |  Size  |  Type  | optional value length | optional value |

	Size flag:
		0 = Size is the direct size, in bytes of the value
		1 = Size is the size (in bytes) of the "optional value length"
		(hence, if the size in bytes of the payload is larger than 2^3 bytes,
		 you must use size flag = 1)
	Type:  the type code of the data
		0 = Integer
		1 = String (null-terminated)
		2 = Double
		3 = Struct
		4 = Array
		5 = AVP
		6 = Byte array (not null-terminated)

*/

const (
	BinRpcMagic        uint = 0xA
	BinRpcVersion      uint = 0x1
	BinRpcMagicVersion uint = BinRpcMagic<<4 + BinRpcVersion

	BinRpcTypeInt    uint = 0x0
	BinRpcTypeString uint = 0x1 // Null-terminated string
	BinRpcTypeDouble uint = 0x2
	BinRpcTypeStruct uint = 0x3
	BinRpcTypeArray  uint = 0x4
	BinRpcTypeAVP    uint = 0x5
	BinRpcTypeBytes  uint = 0x6 // Byte array without null terminator
	BinRpcTypeAll    uint = 0xF // Wildcard; matches any record
)

type BinRpcEncoder interface {
	Encode(io.Writer) error
}

type BinRpcInt int32

func (s BinRpcInt) Encode(w io.Writer) error {
	return WritePacket(w, BinRpcTypeInt, s)
}

type BinRpcString string

func (s BinRpcString) Encode(w io.Writer) error {
	// Strings must be NUL-terminated in binrpc
	val := append([]byte(s), 0x0)
	return WritePacket(w, BinRpcTypeString, val)
}

// ConstructHeader takes the payload length and cookie
// and returns a byte array header
func ConstructHeader(header *bytes.Buffer, payloadLength uint64, cookie uint32) error {
	// Add the Magic/Version
	err := header.WriteByte(byte(BinRpcMagicVersion))
	if err != nil {
		return fmt.Errorf("Failed to write magic/version to header: %s", err.Error())
	}

	// Find the size (in bytes) of the payload length
	plSize := uint8(payloadLength / 256)
	if payloadLength%256 > 0 {
		plSize += 1
	}

	//log.Printf("Payload length is %d, and the length of that value in bytes is %d", payloadLength, plSize)

	// Find the size of the cookie
	cookieSize := binary.Size(cookie)
	if cookieSize < 0 {
		return fmt.Errorf("failed to determine byte length of cookie")
	}

	// Write the Flags/LL/CL byte (flags hard-coded to 0x0 for now)
	err = header.WriteByte(byte(0x0<<4 | uint(plSize-1)<<2 | uint(cookieSize-1)))
	if err != nil {
		return fmt.Errorf("failed to write flags byte: %w", err)
	}

	// Write the payload length
	err = binary.Write(header, binary.BigEndian, uint8(payloadLength))
	if err != nil {
		return fmt.Errorf("failed to append payload length: %w", err)
	}

	// Write the cookie
	err = binary.Write(header, binary.BigEndian, cookie)
	if err != nil {
		return fmt.Errorf("failed to append cookie: %w", err)
	}

	return nil
}

// ConstructPayload takes a value and encodes it
// into a BinRpc payload
func ConstructPayload(payload *bytes.Buffer, valType uint, val interface{}) error {
	// Calculate the minimum byte-size of the value
	valueLength := int8(binary.Size(val))
	if valueLength < 0 {
		return fmt.Errorf("failed to determine byte-size of value")
	}

	// If the minimum byte-size is larger than will
	// fit in three bits, set the size flag = 1
	var sflag uint
	var size uint
	if valueLength > 8 { // 2^3 = 8
		sflag = 1
		// If sflag = 1, size now describes the byte size
		// of the _length_ of the value instead of the value itself
		if temp_size := binary.Size(valueLength); temp_size < 0 {
			log.Println("binary size of", valueLength, "is", temp_size)
			return fmt.Errorf("failed to determine byte-size of value length")
		} else {
			size = uint(temp_size)
		}
	} else {
		// Otherwise, the size is (directly) the byte-length of the value
		size = uint(valueLength)
	}

	// Write the payload header
	err := payload.WriteByte(byte(sflag<<7 | size<<4 | valType))
	if err != nil {
		return fmt.Errorf("failed to write payload header: %w", err)
	}

	// Write the optional value length if our size is too large
	// to fit in `size`
	if sflag == 1 {
		err = binary.Write(payload, binary.BigEndian, uint8(valueLength))
		if err != nil {
			return fmt.Errorf("failed to append optional value length: %w", err)
		}
	}

	// Append the value itself
	err = binary.Write(payload, binary.BigEndian, val)
	if err != nil {
		return fmt.Errorf("failed to append payload value: %w", err)
	}

	return nil
}

func WritePacket(w io.Writer, valType uint, val interface{}) error {
	buf := new(bytes.Buffer)

	// Construct the payload
	payload := new(bytes.Buffer)
	err := ConstructPayload(payload, valType, val)
	if err != nil {
		return fmt.Errorf("failed to construct payload: %w", err)
	}

	// Generate a cookie
	cookie := uint32(rand.Int63())

	// Add the header
	err = ConstructHeader(buf, uint64(payload.Len()), cookie)
	if err != nil {
		return fmt.Errorf("failed to construct header: %w", err)
	}

	// Add the payload
	_, err = payload.WriteTo(buf)
	if err != nil {
		return fmt.Errorf("failed to append payload: %w", err)
	}

	// Write the buffer to the io.Writer
	_, err = buf.WriteTo(w)
	return err
}
