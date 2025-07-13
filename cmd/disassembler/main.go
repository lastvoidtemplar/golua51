package main

import (
	"fmt"
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
		functionBlock, err := LoadBinaryChunkFunctionBlock(r, header)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		PrintBinaryChunkAssembly(arg, header, functionBlock)
		r.Close()
	}
}

func PrintBinaryChunkAssembly(source string, header BinaryChunkHeader, functionBlock BinaryChunkFunctionBlock) {
	pos := 0
	functionLevel := 1
	fmt.Println("Pos\tHex\t\t\tData Description or Code")
	fmt.Println("------------------------------------------------------------------------")
	fmt.Printf("%06X\t\t\t\t** source chunk: %s\n", pos, source)
	PrintBinaryChunkHeader(header)
	pos += 12
	PrintBinaryChunkHeaderFunctionBlock(functionBlock, header, pos, functionLevel, 0)
}

func PrintBinaryChunkHeader(header BinaryChunkHeader) {
	fmt.Println("\t\t\t\t** global header start **")
	fmt.Printf("%06X\t%X\t\theader signature: \"\\27Lua\"\n", 0, header.HeaderSignature)
	fmt.Printf("%06X\t%02X\t\t\tversion (major:minor hex digits)\n", 4, header.VersionNumber)
	fmt.Printf("%06X\t%02X\t\t\tformat (0=official)\n", 5, header.FormatVersion)
	fmt.Printf("%06X\t%02X\t\t\tendianness (0=big endian, 1=little endian)\n", 6, header.Endianness)
	fmt.Printf("%06X\t%02X\t\t\tsize of int (bytes)\n", 7, header.IntSize)
	fmt.Printf("%06X\t%02X\t\t\tsize of size_t (bytes)\n", 8, header.SizetSize)
	fmt.Printf("%06X\t%02X\t\t\tsize of instruction (bytes)\n", 9, header.InstructionSize)
	fmt.Printf("%06X\t%02X\t\t\tsize of number (bytes)\n", 10, header.LuaNumberSize)
	fmt.Printf("%06X\t%02X\t\t\tintegral (0=double, 1=integral)\n", 11, header.IntegralFlag)
	fmt.Println("\t\t\t\t** global header end **")
}

func PrintBinaryChunkHeaderFunctionBlock(functionBlock BinaryChunkFunctionBlock, header BinaryChunkHeader, pos int, functionLevel int, functionInd int) int {
	fmt.Printf("%06X\t\t\t\t** function [%d] definition (level %d)\n", pos, functionInd, functionLevel)
	fmt.Println("\t\t\t\t** start of function **")
	pos = PrintBinaryCHunckString(functionBlock.SourceName, header, pos)
	fmt.Println("\t\t\t\tsource name: ", string(functionBlock.SourceName.Data))
	pos = PrintBinaryChunckInt(functionBlock.LineDefined, fmt.Sprintf("line defined (%d)", functionBlock.LineDefined), header, pos)
	pos = PrintBinaryChunckInt(functionBlock.LastLineDefined, fmt.Sprintf("last line defined (%d)", functionBlock.LastLineDefined), header, pos)

	fmt.Printf("%06X\t%02X\t\t\tnups (%d)\n", pos, functionBlock.UpvaluesCount, functionBlock.UpvaluesCount)
	fmt.Printf("%06X\t%02X\t\t\tnumparams (%d)\n", pos+1, functionBlock.ParametersCount, functionBlock.ParametersCount)
	fmt.Printf("%06X\t%02X\t\t\tis_vararg (%d)\n", pos+2, functionBlock.IsVararg, functionBlock.IsVararg)
	fmt.Printf("%06X\t%02X\t\t\tmaxstacksize (%d)\n", pos+3, functionBlock.MaximumStackSize, functionBlock.MaximumStackSize)
	pos += 4

	return pos
}

func PrintBinaryChunckSizet(n uint64, note string, header BinaryChunkHeader, pos int) int {
	// TODO: add big endian display
	fmt.Printf("%06X\t", pos)
	for i := 0; i < int(header.SizetSize); i++ {
		b := (n >> (8 * i)) & ((1 << 8) - 1)
		fmt.Printf("%02X", b)
	}
	fmt.Printf("\t%s\n", note)
	return pos + int(header.SizetSize)
}

func PrintBinaryChunckInt(n int64, note string, header BinaryChunkHeader, pos int) int {
	// TODO: add big endian display
	fmt.Printf("%06X\t", pos)
	for i := 0; i < int(header.IntSize); i++ {
		b := (n >> (8 * i)) & ((1 << 8) - 1)
		fmt.Printf("%02X", b)
	}
	fmt.Printf("\t\t%s\n", note)
	return pos + int(header.IntSize)
}

func PrintBinaryCHunckString(str BinaryChunkString, header BinaryChunkHeader, pos int) int {
	pos = PrintBinaryChunckSizet(str.Size, fmt.Sprintf("string size (%d)", str.Size), header, pos)
	const bytesOnOneline = 8
	data := str.Data
	for i := 0; i < len(data); i += bytesOnOneline {
		endBound := 0
		if i+bytesOnOneline < int(str.Size) {
			endBound = i + bytesOnOneline
			display := string(data[i:endBound])
			fmt.Printf("%06X\t%X\t\"%s\"\n", pos, data[i:endBound], display)
		} else {
			endBound = len(data)
			display := string(data[i:endBound])
			fmt.Printf("%06X\t%X\t\t\"%s\"\n", pos, data[i:endBound], display)
		}
		pos += bytesOnOneline
	}
	return pos
}
