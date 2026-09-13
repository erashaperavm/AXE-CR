package vm

import "fmt"

type ErrExecShouldFoundItUnmatchMemType struct {
	VarName string
	Got     string
	Expect  string
}

func (e *ErrExecShouldFoundItUnmatchMemType) Error() string {
	return fmt.Sprintf("var %s is %s type, expected %s, this should be found in exec-stage", e.VarName, e.Got, e.Expect)
}

type ErrIsolationMemInaccessible struct {
	VarName string
}

func (e *ErrIsolationMemInaccessible) Error() string {
	return fmt.Sprintf("replayer can not access isolation memory: %s", e.VarName)
}
