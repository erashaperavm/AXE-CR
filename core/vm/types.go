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
	PubStack PtrKind = iota
	PubHeap
	PrivStack // 隐私内存由调用者提供，不支持 VM 层更新，支持在外部函数中获取副本可变性
	PrivHeap  // 隐私内存由调用者提供，不支持 VM 层更新，支持在外部函数中获取副本可变性
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

type StackDiff struct {
	Addr int64 // 栈地址
	Pre  int64 // 变化前的值
	Now  int64 // 变化后的值
}

type HeapDiff struct {
	Addr int64  // 堆地址
	Pre  []byte // 变化前的数据
	Now  []byte // 变化后的数据
}

type TraceStep struct {
	StackChanges  []StackDiff
	HeapChanges   []HeapDiff
	VarsChanges   []VarsDiff
	BlocksChanges []BlocksDiff
}

type VarsDiff struct {
	Name string
	Pre  Ptr
	Now  Ptr
}

type BlocksDiff struct {
	Label int64
	Pre   Block
	Now   Block
}

type TokenPos struct {
	Line int64
	Col  int64
}
