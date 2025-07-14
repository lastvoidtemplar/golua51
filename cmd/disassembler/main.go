package main

import (
	"fmt"
	"math"
	"os"
)

func main() {
	args := os.Args[1:]
	// args = append(args, "/home/deyan/Programming/CLI/golua51/lua/file1.out")
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
	pos = PrintBinaryChunkString(functionBlock.SourceName, header, pos)
	fmt.Println("\t\t\t\tsource name: ", string(functionBlock.SourceName.Data))
	pos = PrintBinaryChunkInt(functionBlock.LineDefined, fmt.Sprintf("line defined (%d)", functionBlock.LineDefined), header, pos)
	pos = PrintBinaryChunkInt(functionBlock.LastLineDefined, fmt.Sprintf("last line defined (%d)", functionBlock.LastLineDefined), header, pos)

	fmt.Printf("%06X\t%02X\t\t\tnups (%d)\n", pos, functionBlock.UpvaluesCount, functionBlock.UpvaluesCount)
	fmt.Printf("%06X\t%02X\t\t\tnumparams (%d)\n", pos+1, functionBlock.ParametersCount, functionBlock.ParametersCount)
	fmt.Printf("%06X\t%02X\t\t\tis_vararg (%d)\n", pos+2, functionBlock.IsVararg, functionBlock.IsVararg)
	fmt.Printf("%06X\t%02X\t\t\tmaxstacksize (%d)\n", pos+3, functionBlock.MaximumStackSize, functionBlock.MaximumStackSize)
	pos += 4

	pos = PrintInstructionList(functionBlock.InstructionList, header, pos)
	pos = PrintConstantList(functionBlock.ConstantList, header, pos)
	pos = PrintFunctionPrototypeList(functionBlock.FunctionPrototypeList, header, pos, functionLevel)
	return pos
}

func PrintBinaryChunkSizet(n uint64, note string, header BinaryChunkHeader, pos int) int {
	// TODO: add big endian display
	fmt.Printf("%06X\t", pos)
	for i := 0; i < int(header.SizetSize); i++ {
		b := (n >> (8 * i)) & ((1 << 8) - 1)
		fmt.Printf("%02X", b)
	}
	fmt.Printf("\t%s\n", note)
	return pos + int(header.SizetSize)
}

func PrintBinaryChunkInt(n int64, note string, header BinaryChunkHeader, pos int) int {
	// TODO: add big endian display
	fmt.Printf("%06X\t", pos)
	for i := 0; i < int(header.IntSize); i++ {
		b := (n >> (8 * i)) & ((1 << 8) - 1)
		fmt.Printf("%02X", b)
	}
	fmt.Printf("\t\t%s\n", note)
	return pos + int(header.IntSize)
}

func PrintBinaryChunkString(str BinaryChunkString, header BinaryChunkHeader, pos int) int {
	pos = PrintBinaryChunkSizet(str.Size, fmt.Sprintf("string size (%d)", str.Size), header, pos)
	const bytesOnOneline = 8
	data := str.Data
	for i := 0; i < len(data); i += bytesOnOneline {
		endBound := 0
		if i+bytesOnOneline < int(str.Size) {
			endBound = i + bytesOnOneline
		} else {
			endBound = len(data)
		}
		display := string(data[i:endBound])

		fmt.Printf("%06X\t%02X", pos, data[i:endBound])
		t := 2 * (endBound - i) / (bytesOnOneline)

		const maxTabs = 3

		for j := 0; j < maxTabs-t; j++ {
			fmt.Printf("\t")
		}

		fmt.Printf("\"%s\"\n", display)
		pos += bytesOnOneline
	}
	return pos
}

var InstructionNames = []string{
	"move",      //0
	"loadk",     //1
	"loadbool",  //2
	"loadnil",   //3
	"getupval",  //4
	"getglobal", //5
	"gettable",  //6
	"setglobal", //7
	"setupval",  //8
	"settable",  //9
	"newtable",  //10
	"self",      //11
	"add",       //12
	"sub",       //13
	"mul",       //14
	"div",       //15
	"mod",       //16
	"pow",       //17
	"unm",       //18
	"not",       //19
	"len",       //20
	"concat",    //21
	"jmp",       //22
	"eq",        //23
	"lt",        //24,
	"le",        //25
	"test",      //26
	"testset",   //27
	"call",      //28
	"tailcall",  //29
	"return",    //30
	"forloop",   //31
	"forprep",   //32
	"tforloop",  //33
	"setlist",   //34
	"close",     //35
	"closure",   //36
	"vararg",    // 37
}

func PrintBinaryChunkInstruction(ins uint32, header BinaryChunkHeader, pos int, insInd int) int {
	// TODO: add big endian display
	fmt.Printf("%06X\t", pos)
	for i := 0; i < int(header.IntSize); i++ {
		b := (int(ins) >> (8 * i)) & ((1 << 8) - 1)
		fmt.Printf("%02X", b)
	}

	//TODO diff instruction types iABC, iABx, iAsBx
	opcode := int(ins & ((1 << 6) - 1))
	a := int(((((1 << 8) - 1) << 6) & ins) >> 6)
	b := int(((((1 << 18) - 1) << 14) & ins) >> 14)
	fmt.Printf("\t\t[%d] %s %d %d\n", insInd, InstructionNames[opcode], a, b)

	return pos + int(header.InstructionSize)
}

func PrintInstructionList(instructions InstructionList, header BinaryChunkHeader, pos int) int {
	fmt.Println("\t\t\t\t* code:")
	pos = PrintBinaryChunkInt(instructions.Size, fmt.Sprintf("sizecode (%d)", instructions.Size), header, pos)

	for i, v := range instructions.Instructions {
		pos = PrintBinaryChunkInstruction(v, header, pos, i+1)
	}
	return pos
}

func PrintBinaryChunkLuaNumber(n uint64, header BinaryChunkHeader, pos int, conInd int) int {
	// TODO: add big endian display
	fmt.Printf("%06X\t", pos)
	for i := 0; i < int(header.LuaNumberSize); i++ {
		b := (n >> (8 * i)) & ((1 << 8) - 1)
		fmt.Printf("%02X", b)
	}

	const float = 0
	const integral = 1

	fmt.Printf("\tconst [%d]: ", conInd)

	switch header.IntegralFlag {
	case float:
		switch header.LuaNumberSize {
		case 4:
			fmt.Printf("(%f)\n", math.Float32frombits(uint32(n)))
		case 8:
			fmt.Printf("(%f)\n", math.Float64frombits(n))
		}
	case integral:
		fmt.Printf("(%d)\n", n)
	}

	return pos + int(header.LuaNumberSize)
}

var constTypeString = [5]string{"nil", "bool", "", "number", "string"}
var constBoolString = [2]string{"false", "true"}

func PrintBinaryChunkConstant(con BinaryChunkConstant, header BinaryChunkHeader, pos int, conInd int) int {
	fmt.Printf("%06X\t%02X\t\t\tconst type %d (%s)\n", pos, con.Type, con.Type, constTypeString[con.Type])
	pos++

	switch con.Type {
	case LUA_TNIL:
		fmt.Printf("\t\t\t\tconst [%d]: (nil)\n", conInd)
	case LUA_TBOOLEAN:
		if b, ok := con.Value.(byte); ok && b < 2 {
			fmt.Printf("%06X\t%02X\t\t\tconst [%d]: (%s)\n", pos, b, conInd, constBoolString[b])
			pos++
		}
	case LUA_TNUMBER:
		if num, ok := con.Value.(uint64); ok {
			pos = PrintBinaryChunkLuaNumber(num, header, pos, conInd)
		}
	case LUA_TSTRING:
		if str, ok := con.Value.(BinaryChunkString); ok {
			pos = PrintBinaryChunkString(str, header, pos)
			fmt.Printf("\t\t\t\tconst [%d]: \"%s\"\n", conInd, string(str.Data))
		}
	}

	return pos
}

func PrintConstantList(constants ConstantList, header BinaryChunkHeader, pos int) int {
	fmt.Println("\t\t\t\t* constants:")
	pos = PrintBinaryChunkInt(constants.Size, fmt.Sprintf("sizek (%d)", constants.Size), header, pos)

	for i, v := range constants.Constants {
		pos = PrintBinaryChunkConstant(v, header, pos, i)
	}
	return pos
}

func PrintFunctionPrototypeList(functionBlocks FunctionPrototypeList, header BinaryChunkHeader, pos int, functionLevel int) int {
	fmt.Println("\t\t\t\t* functions:")
	pos = PrintBinaryChunkInt(functionBlocks.Size, fmt.Sprintf("sizep (%d)", functionBlocks.Size), header, pos)

	for i, v := range functionBlocks.FunctionPrototypes {
		fmt.Println()
		pos = PrintBinaryChunkHeaderFunctionBlock(v, header, pos, functionLevel+1, i)
	}
	return pos
}
