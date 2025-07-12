package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	args := os.Args[1:]
	for _, arg := range args {
		r, err := os.Open(arg)

		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		header, err := LoadBinaryChunkHeader(r)

		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		PrintBinaryChunkHeader(header)
		r.Close()
	}
}

type BinaryChunkHeader struct {
	HeaderSignature [4]byte // must be ESC, “Lua” or 0x1B4C7561.

	VersionNumber byte // 0x51 (81 decimal) for Lua 5.1. High hex digit is major version number. Low hex digit is minor version number.

	FormatVersion byte // Format version, 0=official version
	Endianness    byte // default 1, 0=big endian, 1=little endian.  Lua 5.1 will not load a chunk whose endianness is different from that of the platform.

	IntSize         byte // default 4 bytes
	SizetSize       byte // default 4 bytes
	InstructionSize byte //default 4 bytes
	LuaNumberSize   byte // default 8 bytes
	IntegralFlag    byte // default 0, 0=floating-point, 1=integral number type
}

func LoadBinaryChunkHeader(r io.Reader) (BinaryChunkHeader, error) {
	var buf [12]byte
	n, err := r.Read(buf[:])

	if err != nil {
		return BinaryChunkHeader{}, fmt.Errorf("failed to load the binary chunk header: %w", err)
	}

	if n < 12 {
		return BinaryChunkHeader{}, fmt.Errorf("invalid length for binary chunk header: got %d, expected 12", n)
	}

	header := BinaryChunkHeader{}

	// ESC, "Lua"
	var expectedLuaHeaderSignature = [4]byte{27, 'L', 'u', 'a'}

	for i := 0; i < len(expectedLuaHeaderSignature); i++ {
		if expectedLuaHeaderSignature[i] != buf[i] {
			return BinaryChunkHeader{},
				fmt.Errorf("invalid binary chunk header signature: got %x, expected %x",
					buf[:len(expectedLuaHeaderSignature)], expectedLuaHeaderSignature)
		}
		header.HeaderSignature[i] = expectedLuaHeaderSignature[i]
	}

	const expectedLuaVersion = 0x51
	if buf[4] != expectedLuaVersion {
		return BinaryChunkHeader{}, fmt.Errorf("unsuported binary chunk version: got %x, expected %x (Lua 5.1)",
			buf[4], expectedLuaVersion)
	}
	header.VersionNumber = expectedLuaVersion

	const expectedLuaFormat = 0
	if buf[5] != expectedLuaFormat {
		return BinaryChunkHeader{}, fmt.Errorf("unsuported binary chunk format: got %d, expected %d (official version)",
			buf[5], expectedLuaFormat)
	}
	header.FormatVersion = expectedLuaFormat

	const maxValueEndianness = 1
	if buf[6] > maxValueEndianness {
		return BinaryChunkHeader{}, fmt.Errorf("invalid binary chunk endianness: got %d, expected <= %d (0=big endian and 1=little endian)",
			buf[6], maxValueEndianness)
	}
	header.Endianness = buf[6]

	header.IntSize = buf[7]
	header.SizetSize = buf[8]
	header.InstructionSize = buf[9]
	header.LuaNumberSize = buf[10]

	const maxValueIntegralFlag = 1
	if buf[11] > maxValueIntegralFlag {
		return BinaryChunkHeader{}, fmt.Errorf("invalid binary chunk integral flag: got %d, expected <= %d (0=floating-point and 1=integral number type)",
			buf[11], maxValueIntegralFlag)
	}
	header.IntegralFlag = buf[11]

	return header, nil
}

func PrintBinaryChunkHeader(header BinaryChunkHeader) {
	fmt.Println("Binary Chunk Header:")
	fmt.Println("\tHeader signature - ESC,\"Lua\"")
	fmt.Println("\tVersion number - Lua 5.1")
	fmt.Println("\tFormat version - official version")

	const bigEndian = 0
	const littleEndian = 1
	switch header.Endianness {
	case bigEndian:
		fmt.Println("\tEndianness - big endian")
	case littleEndian:
		fmt.Println("\tEndianness - little endian")
	}

	fmt.Println("\tSize of int -", header.IntSize, "bytes")
	fmt.Println("\tSize of size_t -", header.SizetSize, "bytes")
	fmt.Println("\tSize of Instruction -", header.InstructionSize, "bytes")
	fmt.Println("\tSize of lua_Number -", header.LuaNumberSize, "bytes")

	const floatingPoing = 0
	const integralNumberType = 1
	switch header.IntegralFlag {
	case floatingPoing:
		fmt.Println("\tIntegral flag - floating-point")
	case integralNumberType:
		fmt.Println("\tIntegral flag - integral number type")
	}
}
