package vm

import "fmt"

// ============== Variable & Pointer Errors ==============

// ErrVarNotFound indicates a variable was not found in the VM.
type ErrVarNotFound struct{ Name string }

func (e *ErrVarNotFound) Error() string { return fmt.Sprintf("variable '%s' not found", e.Name) }

// ErrPtrNotFound indicates a pointer variable was not found.
type ErrPtrNotFound struct{ Name string }

func (e *ErrPtrNotFound) Error() string { return fmt.Sprintf("ptr '%s' not found", e.Name) }

type ErrVarDrop struct{ Name string }

func (e *ErrVarDrop) Error() string { return fmt.Sprintf("variable '%s' dropped failed", e.Name) }

// ErrTypeMismatch indicates a type mismatch.
type ErrTypeMismatch struct {
	Expected string
	Got      string
}

func (e *ErrTypeMismatch) Error() string {
	return fmt.Sprintf("type mismatch: expected %s, got %s", e.Expected, e.Got)
}

// ============== Memory Errors ==============

// ErrUnsupportedMemType indicates an unsupported memory type.
type ErrUnsupportedMemType struct{ MemType interface{} }

func (e *ErrUnsupportedMemType) Error() string {
	return fmt.Sprintf("unsupported mem type: %v", e.MemType)
}

// ErrUnsupportedPtrKind indicates an unsupported pointer kind.
type ErrUnsupportedPtrKind struct{ Kind interface{} }

func (e *ErrUnsupportedPtrKind) Error() string {
	return fmt.Sprintf("unsupported ptr kind: %v", e.Kind)
}

// ============== Block/Label Errors ==============

// ErrLabelNotFound indicates a label was not found in the Blocks map.
type ErrLabelNotFound struct{ Label string }

func (e *ErrLabelNotFound) Error() string { return fmt.Sprintf("label '%s' not found", e.Label) }

// ErrLabelEmpty indicates a label's corresponding block is empty.
type ErrLabelEmpty struct{ Label string }

func (e *ErrLabelEmpty) Error() string {
	return fmt.Sprintf("label '%s' block is empty", e.Label)
}

// ErrBlockNotClosed indicates a BEGIN instruction never found its matching END.
type ErrBlockNotClosed struct{ Label string }

func (e *ErrBlockNotClosed) Error() string {
	return fmt.Sprintf("block with label '%s' not closed", e.Label)
}

// ErrBlockLabelInvalid indicates a block label is not a valid integer.
type ErrBlockLabelInvalid struct {
	Label  string
	Reason string
}

func (e *ErrBlockLabelInvalid) Error() string {
	return fmt.Sprintf("invalid block label '%s': %s", e.Label, e.Reason)
}

// ============== Instruction Errors ==============

// ErrUnknownOpcode indicates an unknown opcode.
type ErrUnknownOpcode struct{ Opcode int }

func (e *ErrUnknownOpcode) Error() string {
	return fmt.Sprintf("unknown opcode: %d", e.Opcode)
}

// ErrUnexpectedEnd indicates an END instruction that should have been handled.
type ErrUnexpectedEnd struct{ Label string }

func (e *ErrUnexpectedEnd) Error() string {
	return fmt.Sprintf("unexpected end: with label '%s'", e.Label)
}

type ErrUpdateVarBySurfaceInt64 struct {
	VarName string
	Surface int64
}

func (e *ErrUpdateVarBySurfaceInt64) Error() string {
	return fmt.Sprintf("update var '%s' by surface int64 '%d' failed", e.VarName, e.Surface)
}

type ErrUpdateVarBySurfaceBytes struct {
	VarName string
	Surface []byte
}

func (e *ErrUpdateVarBySurfaceBytes) Error() string {
	return fmt.Sprintf("update var '%s' by surface bytes '%v' failed", e.VarName, e.Surface)
}

type ErrRead struct {
	Pos []byte
	Err error
}

func (e *ErrRead) Error() string {
	return fmt.Sprintf("read failed at position '%v': %v", e.Pos, e.Err)
}

type InputErr struct {
	Index int64
	Err   error
}

func (e *InputErr) Error() string {
	return fmt.Sprintf("input at index '%d' failed: %v", e.Index, e.Err)
}

type OutputErr struct {
	Index int64
	Data  []byte
	Err   error
}

func (e *OutputErr) Error() string {
	return fmt.Sprintf("output at index '%d' with data '%v' failed: %v", e.Index, e.Data, e.Err)
}

// ============== Function Call (RS) Errors ==============

// ErrFuncNotFound indicates the RS function was not found in the environment.
type ErrFuncNotFound struct{ Name string }

func (e *ErrFuncNotFound) Error() string {
	return fmt.Sprintf("function '%s' not found", e.Name)
}

// ErrFuncInputCount indicates a mismatch in the number of input arguments.
type ErrFuncInputCount struct {
	Name     string
	Expected int
	Got      int
}

func (e *ErrFuncInputCount) Error() string {
	return fmt.Sprintf("function '%s' requires %d inputs, but %d given", e.Name, e.Expected, e.Got)
}

// ErrFuncOutputCount indicates a mismatch in the number of output arguments.
type ErrFuncOutputCount struct {
	Name     string
	Expected int
	Got      int
}

func (e *ErrFuncOutputCount) Error() string {
	return fmt.Sprintf("function '%s' requires %d outputs, but %d given", e.Name, e.Expected, e.Got)
}

// ErrFuncOutputNotVar indicates an output argument that should be a variable name is a literal.
type ErrFuncOutputNotVar struct {
	FuncName    string
	OutputIndex int
}

func (e *ErrFuncOutputNotVar) Error() string {
	return fmt.Sprintf("call_rs: output %d for '%s' must be a variable, not a literal", e.OutputIndex, e.FuncName)
}

// ErrUnsupportedInputType indicates an input type that is not supported for RS function.
type ErrUnsupportedInputType struct {
	FuncName   string
	InputIndex int
	InputType  string
}

func (e *ErrUnsupportedInputType) Error() string {
	return fmt.Sprintf("call_rs: input %d for '%s' unsupported type '%s'", e.InputIndex, e.FuncName, e.InputType)
}

// ============== Operand & Arithmetic Errors ==============

// ErrOperandType indicates an operand is of the wrong type (e.g., []byte where int64 expected).
type ErrOperandType struct {
	Operation string
	Detail    string
}

func (e *ErrOperandType) Error() string {
	return fmt.Sprintf("operand type error in %s: %s", e.Operation, e.Detail)
}

// ErrArithmetic wraps a lower-level error for arithmetic operations.
type ErrArithmetic struct {
	Operation string
	Err       error
}

func (e *ErrArithmetic) Error() string {
	return fmt.Sprintf("unable to %s: %v", e.Operation, e.Err)
}
