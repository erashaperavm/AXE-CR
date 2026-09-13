package vm

import "fmt"

// ============== Variable & Pointer Errors ==============

// ErrVarNotFound indicates a variable was not found in the VM.
type ErrVarNotFound struct{ Name string }

func (e *ErrVarNotFound) Error() string { return fmt.Sprintf("variable '%s' not found", e.Name) }

// ErrPtrNotFound indicates a pointer variable was not found.
type ErrPtrNotFound struct{ Name string }

func (e *ErrPtrNotFound) Error() string { return fmt.Sprintf("ptr '%s' not found", e.Name) }

// ErrVarDrop indicates a variable drop failed.
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

// ErrPrivacyMemUpdateInNoCallIns indicates a privacy memory update is unsupported in a non-call instruction.
type ErrPrivacyMemUpdateInNoCallIns struct {
	ThisIns string
}

func (e *ErrPrivacyMemUpdateInNoCallIns) Error() string {
	return fmt.Sprintf("privacy mem update is unsupported in: %s, but call instruction", e.ThisIns)
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

// ErrTypeArgNumMismatch indicates a mismatch in the number of type and arguments.
type ErrTypeArgNumMismatch struct {
	GetType int
	GetArg  int
}

func (e *ErrTypeArgNumMismatch) Error() string {
	return fmt.Sprintf("type arg num mismatch: expected %d, got %d", e.GetType, e.GetArg)
}

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

// ErrUpdateVarBySurfaceInt64 indicates a variable update by surface int64 failed.
type ErrUpdateVarBySurfaceInt64 struct {
	VarName string
	Surface int64
}

func (e *ErrUpdateVarBySurfaceInt64) Error() string {
	return fmt.Sprintf("update var '%s' by surface int64 '%d' failed", e.VarName, e.Surface)
}

// ErrUpdateVarBySurfaceBytes indicates a variable update by surface bytes failed.
type ErrUpdateVarBySurfaceBytes struct {
	VarName string
	Surface []byte
}

func (e *ErrUpdateVarBySurfaceBytes) Error() string {
	return fmt.Sprintf("update var '%s' by surface bytes '%v' failed", e.VarName, e.Surface)
}

// ErrOutputIsSurface indicates the output variable is a surface value
type ErrOutputIsSurface struct {
	VarName string
}

func (e *ErrOutputIsSurface) Error() string {
	return fmt.Sprintf("variable '%s' is a surface value, can not be used as output parameter", e.VarName)
}

// ============== Function Call (RS) Errors ==============

// ErrCallInt64Assert indicates a variable from call assert int64 failed.
type ErrCallInt64Assert struct {
	VarName string
	Surface interface{}
}

func (e *ErrCallInt64Assert) Error() string {
	return fmt.Sprintf("call int64 assert failed: %s", e.VarName)
}

// ErrCallBytesAssert indicates a variable from call assert bytes failed.
type ErrCallBytesAssert struct {
	VarName string
	Surface interface{}
}

func (e *ErrCallBytesAssert) Error() string {
	return fmt.Sprintf("call bytes assert failed: %s", e.VarName)
}

// ErrFuncNameInvalid indicates the RS function name is invalid.
type ErrFuncNameInvalid struct{ Name string }

func (e *ErrFuncNameInvalid) Error() string {
	return fmt.Sprintf("function name '%s' is invalid", e.Name)
}

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

// ErrCallFunctionFailed indicates a function call failed.
type ErrCallFunctionFailed struct {
	FunName  string
	Detailed string
}

func (e ErrCallFunctionFailed) Error() string {
	return fmt.Sprintf("call function '%s' failed: %s", e.FunName, e.Detailed)
}

// ErrReadFileFailed indicates a file read failed.
type ErrReadFileFailed struct {
	Path     string
	Detailed string
}

func (e ErrReadFileFailed) Error() string {
	return fmt.Sprintf("read file '%s' failed: %s", e.Path, e.Detailed)
}

type ErrParsePvFailed struct {
	Path     string
	Detailed string
}

func (e ErrParsePvFailed) Error() string {
	return fmt.Sprintf("parse pv file '%s' failed: %s", e.Path, e.Detailed)
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
