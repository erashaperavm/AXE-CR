package vm

const INTERNAL_PUB_VAR_IN_SIGN string = "__internal_pub_in_"
const INTERNAL_PRIV_VAR_IN_SIGN string = "__internal_priv_in_"
const INTERNAL_EXPR_SIGN string = "__internal_expr_"
const INTERNAL_PUB_VAR_OUT_SIGN string = "__internal_pub_out_"

type Code struct {
	OpCode       OpWrap
	ArgType      []ArgTypeWrap
	IsArgSurface []IsSurfaceWrap // to assert is surface value
	ArgContent   []ArgContWrap   // if IsArgSurface is true, it's surface value; else it's var name
}

type OpWrap struct {
	Code     Op
	TokenPos TokPos
}

type ArgTypeWrap struct {
	Type     ArgT
	TokenPos TokPos
}

type ArgContWrap struct {
	ArgContent               string
	ArgContentInt64IfSurface int64
	TokenPos                 TokPos
}

type IsSurfaceWrap struct {
	IsSurface bool
	TokenPos  TokPos
}

type ArgT byte

type Op byte

const (
	/*
		Interact with Chain token
	*/
	OP_READ Op = iota
	OP_INPUT
	OP_WRITE
	OP_OUTPUT

	/*
		Manage PubMemory Token
	*/
	OP_ALLOC
	OP_UPDATE
	OP_DROP

	/*
		Arithmetic Operation Token
	*/
	OP_ADD
	OP_SUB
	OP_MUL
	OP_DIV

	/*
		Equal Operation Token
	*/
	OP_EQ_INT
	OP_EQ_BYTES

	/*
		Compare Operation Token
	*/
	OP_LARGE_INT

	/*
		Jump and Judge Token
	*/
	OP_JMP
	OP_IF

	/*
		Block Token
	*/
	OP_BEGIN
	OP_END

	/*
		Call function Token
	*/
	OP_CALL_RS
	OP_CALL_C
	OP_CALL_CPP
)

const (
	/*
		Public Type
	*/
	PUB_I64 ArgT = iota
	PUB_BYTES

	/*
		Private Type
	*/
	PRIV_I64
	PRIV_BYTES

	/*
		Isolation Type
	*/
	ISO_I64
	ISO_BYTES
)

type PtrKind byte

const (
	PUB_STACK PtrKind = iota
	PUB_HEAP
	PRIV_STACK // 隐私内存由调用者提供，不支持 VM 层更新，支持在外部函数中获取副本可变性
	PRIV_HEAP  // 隐私内存由调用者提供，不支持 VM 层更新，支持在外部函数中获取副本可变性
	ISOLATED_STACK
	ISOLATED_HEAP // 隔离内存，用于存储程序不可访问内容，如 report 等
)

type Ptr struct {
	Kind    PtrKind
	Pointer int64
}

type PublicInput struct {
	Int64 []int64
	Bytes [][]byte
}

// PrivateInput is the same as the PublicInput,
// the only difference is PrivateInput only exists in executors
type PrivateInput = PublicInput

type PrivateInputExpr struct {
	Kind    PtrKind
	Idx     int64
	HashSum [32]byte
}

type PublicOutput struct {
	Int64 []int64
	Bytes [][]byte
}

type PrivateOutput struct {
	Kind          PtrKind
	Idx           int64
	EncryptedData []byte
}

type PrivateOriginOutputExpr struct {
	Kind    PtrKind
	Idx     int64
	HashSum [32]byte
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

type TokPos struct {
	Line int64
	Col  int64
}
