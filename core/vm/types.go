package vm

type Opcode byte

const (
	// operate chain
	OP_READ = iota
	OP_INPUT
	OP_WRITE
	OP_OUTPUT

	// memory manage
	OP_CREATE
	OP_UPDATE
	OP_DROP

	// block token
	OP_BEGIN
	OP_END

	// condition
	OP_LOOP
	OP_JMP
	OP_IF
	OP_ELSE
	OP_ELSIF
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
	InNum         int64
	OuNum         int64
	InType        []string
	OutType       []string
	InIdentifier  []string
	OutIdentifier []string
}

type Block struct {
	BeginPC int64
	EndPc   int64
}
