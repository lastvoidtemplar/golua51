package main

type BinaryChunkHeader struct {
	HeaderSignature [4]byte // must be ESC, “Lua” or 0x1B4C7561.

	VersionNumber byte // 0x51 (81 decimal) for Lua 5.1. High hex digit is major version number. Low hex digit is minor version number.

	FormatVersion byte // format version, 0=official version
	Endianness    byte // default 1, 0=big endian, 1=little endian.  Lua 5.1 will not load a chunk whose endianness is different from that of the platform.

	IntSize         byte // default 4 bytes
	SizetSize       byte // default 4 bytes
	InstructionSize byte // default 4 bytes
	LuaNumberSize   byte // default 8 bytes
	IntegralFlag    byte // default 0, 0=floating-point, 1=integral number type
}

type BinaryChunkFunctionBlock struct {
	SourceName      BinaryChunkString //the source name is specified only in the top-levelfunction; in other functions, this field consists only of a size_t with the value 0.
	LineDefined     int64             // int, is the line number where the function prototype starts in the source file, for the main chunk the value is 0
	LastLineDefined int64             // int, is the line number where the function prototype ends in the source file, for the main chunk the value is 0
	UpvaluesCount   byte
	ParametersCount byte

	// bit mask, 1=VARARG_HASARG, 2=VARARG_ISVARARG, 4=VARARG_NEEDSARG
	// count of registers used, minimum=2, VARARG_ISVARARG (2) is always set for vararg functions,
	// if LUA_COMPAT_VARARG is defined, VARARG_HASARG (1) is also set,
	// if ... is not used within the function, then VARARG_NEEDSARG (4) is set.
	// normal function always has an IsVararg flag value of 0, while the main chunk always has an IsVararg flag value of 2
	IsVararg byte

	MaximumStackSize byte

	InstructionList       InstructionList
	ConstantList          ConstantList
	FunctionPrototypeList FunctionPrototypeList

	SourceLinePositionList SourceLinePositionList //optinal
	LocalList              LocalList              //optinal
	UpvalueList            UpvalueList            //optinal
}

type BinaryChunkString struct {
	Size uint64 // size_t
	Data []byte // includes a NUL (ASCII 0) at the end
}

type InstructionList struct {
	Size         int64 // int
	Instructions []uint32
}

type BinaryChunkConstantType byte

const (
	LUA_TNIL     BinaryChunkConstantType = 0
	LUA_TBOOLEAN BinaryChunkConstantType = 1
	LUA_TNUMBER  BinaryChunkConstantType = 3
	LUA_TSTRING  BinaryChunkConstantType = 4
)

type BinaryChunConstant struct {
	Type  BinaryChunkConstantType
	Value any
}

type ConstantList struct {
	Size      int64 // int
	Constants []BinaryChunConstant
}

type FunctionPrototypeList struct {
	Size               int64 // int
	FunctionPrototypes []BinaryChunkFunctionBlock
}

type SourceLinePositionList struct {
	Size                int64 // int
	SourceLinePositions []int64
}

type BinaryChunkLocal struct {
	VariableName       BinaryChunkString
	StartVariableScope int64
	EndVariableScope   int64
}

type LocalList struct {
	Size                int64 // int
	SourceLinePositions []BinaryChunkLocal
}

type UpvalueList struct {
	Size                int64 // int
	SourceLinePositions []BinaryChunkString
}
