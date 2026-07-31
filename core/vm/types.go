package vm

type Opcode byte

const (
	// operate chain
	OP_READ = iota
	OP_INPUT
	OP_WRITE
	OP_OUTPUT

	// memory manage
	OP_ALLOC
	OP_UPDATE
	OP_DROP

	// arithmetic
	OP_ADD
	OP_SUB
	OP_MUL
	OP_DIV

	// equal
	OP_EQ_INT
	OP_EQ_BYTES

	// compare
	OP_LARGE_INT

	// jump
	OP_JMP
	OP_IF

	// block token
	OP_BEGIN
	OP_END

	// call function
	OP_CALL_RS
	OP_CALL_C
	OP_CALL_CPP
)

type PtrKind byte

const (
	Stack PtrKind = iota
	Heap
)

type Ptr struct {
	Kind    PtrKind
	Pointer int64
}

type Instruction struct {
	Op            Opcode
	ArgType       []string
	ArgIdentifier []string
}

type Block struct {
	BeginPC int64
	EndPC   int64
}
