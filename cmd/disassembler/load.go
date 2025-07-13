package main

import (
	"fmt"
	"io"
	"slices"
)

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

	const maxSupportedIntSize = 8
	if buf[7] > maxSupportedIntSize {
		return BinaryChunkHeader{}, fmt.Errorf("unsuported binary chunk int size: got %d, expected <= %d (bytes)",
			buf[7], maxSupportedIntSize)
	}
	header.IntSize = buf[7]

	const maxSupportedSizetSize = 8
	if buf[8] > maxSupportedSizetSize {
		return BinaryChunkHeader{}, fmt.Errorf("unsuported binary chunk size_t size: got %d, expected <= %d (bytes)",
			buf[8], maxSupportedSizetSize)
	}
	header.SizetSize = buf[8]

	const expectedInstructionSize = 4
	if buf[9] != expectedInstructionSize {
		return BinaryChunkHeader{}, fmt.Errorf("unsuported binary chunk instruction size: got %d, expected %d (bytes)",
			buf[9], expectedInstructionSize)
	}
	header.InstructionSize = buf[9]

	const maxSupportedLuaNumberSize = 8
	if buf[10] > maxSupportedLuaNumberSize {
		return BinaryChunkHeader{}, fmt.Errorf("unsuported binary chunk lua number size: got %d, expected <= %d (bytes)",
			buf[10], maxSupportedLuaNumberSize)
	}
	header.LuaNumberSize = buf[10]

	const maxValueIntegralFlag = 1
	if buf[11] > maxValueIntegralFlag {
		return BinaryChunkHeader{}, fmt.Errorf("invalid binary chunk integral flag: got %d, expected <= %d (0=floating-point and 1=integral number type)",
			buf[11], maxValueIntegralFlag)
	}
	header.IntegralFlag = buf[11]

	return header, nil
}

func LoadBinaryChunkSizet(r io.Reader, header BinaryChunkHeader) (uint64, error) {
	const bigEndian = 0
	const littleEndian = 1

	buf := make([]byte, header.SizetSize)
	n, err := r.Read(buf)

	if err != nil {
		return 0, fmt.Errorf("failed to load size_t: %w", err)
	}

	if n < int(header.SizetSize) {
		return 0, fmt.Errorf("invalid size_t size: got %d, expected %d", n, header.SizetSize)
	}

	var result uint64
	b := 0
	switch header.Endianness {
	case bigEndian:
		for i := n; i >= 0; i-- {
			result = result | (uint64(buf[i]) << b)
			b += 8
		}
	case littleEndian:
		for i := 0; i < n; i++ {
			result = result | (uint64(buf[i]) << b)
			b += 8
		}
	}

	return result, nil
}

func LoadBinaryChunkString(r io.Reader, header BinaryChunkHeader) (BinaryChunkString, error) {
	size, err := LoadBinaryChunkSizet(r, header)

	if err != nil {
		return BinaryChunkString{}, fmt.Errorf("failed to load string size: %w", err)
	}

	buf := make([]byte, size)
	n, err := r.Read(buf)

	if err != nil {
		return BinaryChunkString{}, fmt.Errorf("failed to load string data: %w", err)
	}

	if n < int(size) {
		return BinaryChunkString{}, fmt.Errorf("invalid read bytes for a string data: got %d, expected %d", n, size)
	}

	if buf[size-1] != 0 {
		return BinaryChunkString{}, fmt.Errorf("invalid last byte for a string data: got %d, expected ASCII 0", buf[size-1])
	}

	str := BinaryChunkString{
		Size: size,
		Data: slices.Clone(buf),
	}
	return str, nil
}

func LoadBinaryChunkInt(r io.Reader, header BinaryChunkHeader) (int64, error) {
	const bigEndian = 0
	const littleEndian = 1

	buf := make([]byte, header.IntSize)
	n, err := r.Read(buf)

	if err != nil {
		return 0, fmt.Errorf("failed to load size_t: %w", err)
	}

	if n < int(header.IntSize) {
		return 0, fmt.Errorf("invalid size_t size: got %d, expected %d", n, header.IntSize)
	}

	var result int64
	b := 0
	switch header.Endianness {
	case bigEndian:
		for i := n; i >= 0; i-- {
			result = result | (int64(buf[i]) << b)
			b += 8
		}
	case littleEndian:
		for i := 0; i < n; i++ {
			result = result | (int64(buf[i]) << b)
			b += 8
		}
	}

	return result, nil
}

func LoadBinaryChunkFunctionBlock(r io.Reader, header BinaryChunkHeader) (BinaryChunkFunctionBlock, error) {
	functionBlock := BinaryChunkFunctionBlock{}

	sourceName, err := LoadBinaryChunkString(r, header)
	if err != nil {
		return BinaryChunkFunctionBlock{}, fmt.Errorf("failed to load source name for a function block: %w", err)
	}
	functionBlock.SourceName = sourceName

	lineDefined, err := LoadBinaryChunkInt(r, header)
	if err != nil {
		return BinaryChunkFunctionBlock{}, fmt.Errorf("failed to load line defined for a function block: %w", err)
	}
	functionBlock.LineDefined = lineDefined

	lastLineDefined, err := LoadBinaryChunkInt(r, header)
	if err != nil {
		return BinaryChunkFunctionBlock{}, fmt.Errorf("failed to load last line defined for a function block: %w", err)
	}
	functionBlock.LastLineDefined = lastLineDefined

	var buf [4]byte
	n, err := r.Read(buf[:])

	if err != nil {
		return BinaryChunkFunctionBlock{}, fmt.Errorf("failed to load upvalues count, parameter count, is_vararg and maximum stack size for a function block: %w", err)
	}

	if n < 4 {
		return BinaryChunkFunctionBlock{},
			fmt.Errorf("failed to load upvalues count, parameter count, is_vararg and maximum stack size for a function block, invalid length: got %d, expected %d", n, 4)
	}
	functionBlock.UpvaluesCount = buf[0]
	functionBlock.ParametersCount = buf[1]
	functionBlock.IsVararg = buf[2]
	functionBlock.MaximumStackSize = buf[3]

	instructions, err := LoadBinaryChunkInstructionList(r, header)

	if err != nil {
		return BinaryChunkFunctionBlock{}, fmt.Errorf("failed to load instruction list for a function block: %w", err)
	}
	functionBlock.InstructionList = instructions

	return functionBlock, nil
}

func LoadBinaryChunkInstruction(r io.Reader, header BinaryChunkHeader) (uint32, error) {
	const bigEndian = 0
	const littleEndian = 1

	buf := make([]byte, header.InstructionSize)
	n, err := r.Read(buf)

	if err != nil {
		return 0, fmt.Errorf("failed to load size_t: %w", err)
	}

	if n < int(header.InstructionSize) {
		return 0, fmt.Errorf("invalid size_t size: got %d, expected %d", n, header.InstructionSize)
	}

	var result uint32
	b := 0
	switch header.Endianness {
	case bigEndian:
		for i := n; i >= 0; i-- {
			result = result | (uint32(buf[i]) << b)
			b += 8
		}
	case littleEndian:
		for i := 0; i < n; i++ {
			result = result | (uint32(buf[i]) << b)
			b += 8
		}
	}

	return result, nil
}

func LoadBinaryChunkInstructionList(r io.Reader, header BinaryChunkHeader) (InstructionList, error) {
	instructions := InstructionList{}

	size, err := LoadBinaryChunkInt(r, header)
	if err != nil {
		return InstructionList{}, fmt.Errorf("failed to load size for a instruction list: %w", err)
	}
	instructions.Size = size

	list := make([]uint32, size)
	for i := 0; i < int(size); i++ {
		ins, err := LoadBinaryChunkInstruction(r, header)
		if err != nil {
			return InstructionList{}, fmt.Errorf("failed to load virtual machine instruction (ind=%d): %w", i, err)
		}

		list[i] = ins
	}
	instructions.Instructions = list

	return instructions, nil
}
