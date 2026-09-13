package vm

import (
	"bytes"
	"errors"
	"fmt"
)

func NewReplayVM(
	code []Instruction,
	originPos map[int64]map[int64]TokenPos,
	env *Environment,
	pubIn PublicInput,
	privIn PrivateInput,
	privInExpr []PrivateInputExpr,
	traces map[int64]TraceStep,
	isolationMem *IsolationMemory,
) *VM {
	return &VM{
		Code:         code,
		OriginPos:    originPos,
		TraceMode:    true,
		Env:          env,
		PubIn:        pubIn,
		PrivIn:       privIn,
		PrivInExpr:   privInExpr,
		PC:           0,
		Lines:        0,
		Mem:          NewMemoryObj(),
		Vars:         make(map[string]Ptr),
		Blocks:       make(map[int64]Block),
		IsolationMem: isolationMem,
		Traces:       traces,
	}
}

// Replay 回放 VM 执行轨迹, 需要一个新 VM 实例和一个新内存状态, 返回不一致错误类型和错误位置
func (vm *VM) Replay() (error, []TokenPos) {
	if !vm.TraceMode {
		return nil, []TokenPos{}
	}

	// 同步内存状态
	vm.Mem.applyDiff(vm.Traces[vm.Lines])

	// 同步变量状态
	vm.applyVarsDiff(vm.Traces[vm.Lines].VarsChanges)

	// 同步块状态
	vm.applyBlocksDiff(vm.Traces[vm.Lines].BlocksChanges)

	for vm.PC < int64(len(vm.Code)) {
		ins := vm.Code[vm.PC]

		switch ins.Op {
		case OP_READ:
			// todo：复杂度原因暂时禁用 RW 操作

			vm.PC++
			vm.Lines++

		case OP_INPUT:
			// 开头标志性语句，无实际语义

			vm.PC++
			vm.Lines++

		case OP_WRITE:
			// todo：复杂度原因暂时禁用 RW 操作

			vm.PC++
			vm.Lines++

		case OP_OUTPUT:
			// 公共输出由重放计算逻辑保证，隐私输出由 zkp 保证，这里需要确认上一步（还没有output的line）的输出变量和写入的是否一样

			for i, varName := range ins.ArgIdentifier[1:] {
				// 判断是否为 surface
				if isSurfaceInt64(varName) {
					sourcePos := vm.OriginPos[vm.PC][int64(i+1)]
					return &ErrExecShouldFoundItUnmatchMemType{VarName: varName, Got: "surface int64", Expect: "PubStack or PubHeap (Pub: int64 or bytes)"}, []TokenPos{sourcePos}
				}
				if isSurfaceBytes(varName) {
					sourcePos := vm.OriginPos[vm.PC][int64(i+1)]
					return &ErrExecShouldFoundItUnmatchMemType{VarName: varName, Got: "surface bytes", Expect: "PubStack or PubHeap (Pub: int64 or bytes)"}, []TokenPos{sourcePos}
				}

				// 判断是否为 Priv 和 Isolation Memory
				if vm.Vars[varName].Kind == PrivStack || vm.Vars[varName].Kind == PrivHeap || vm.Vars[varName].Kind == IsolationHeap || vm.Vars[varName].Kind == IsolationStack {
					sourcePos := vm.OriginPos[vm.PC][int64(i+1)]
					return &ErrExecShouldFoundItUnmatchMemType{VarName: varName, Got: "Priv or Isolation", Expect: "PubStack or PubHeap (Pub: int64 or bytes)"}, []TokenPos{sourcePos}
				}

				// 获取重定向后的名称，以及原始指针
				varNameThis := fmt.Sprintf("__internal_pub_out_%d", i)
				ptrOfOrig := vm.Vars[varName]

				// 读取写入的 IsolationMem 对比上一帧内存对应变量 eq
				switch ptrOfOrig.Kind {
				case PubStack:
					contentFromIsolationMem, ok := vm.IsolationMem.readStackOnlyForReplay(vm.Vars[varNameThis])
					if !ok {
						sourcePos := vm.OriginPos[vm.PC][int64(i+1)]
						return &ErrIsolationMemInaccessible{VarName: varNameThis}, []TokenPos{sourcePos}
					}

				}

			}

			vm.PC++
			vm.Lines++

		case OP_ALLOC:
			// 纯内存变更, 不处理
			vm.PC++

		case OP_UPDATE:
			// 纯内存变更, 不处理
			vm.PC++

		case OP_DROP:
			// 纯内存变更, 不处理
			vm.PC++

		case OP_ADD:
			// replay add
			aNum, bNum, err := vm.arithmeticPrepare(ins, "add")
			if err != nil {
				return err
			}
			sum := aNum + bNum

			vm.PC++

			// verify that replay res and given res are the same
			givenSum, err := vm.getInt64Arg(ins.ArgIdentifier[2]) // 已经进入下一行了，所以这里获取和是合法的
			if err != nil {
				return err
			}
			if givenSum != sum {
				return &ErrArithmetic{Operation: "add", Err: errors.New("add result not match")}
			}

		case OP_SUB:
			// replay sub
			aNum, bNum, err := vm.arithmeticPrepare(ins, "sub")
			if err != nil {
				return err
			}
			sub := aNum - bNum

			vm.PC++

			// verify that replay res and given res are the same
			givenSub, err := vm.getInt64Arg(ins.ArgIdentifier[2]) // 已经进入下一行了，所以这里获取差是合法的
			if err != nil {
				return err
			}
			if givenSub != sub {
				return &ErrArithmetic{Operation: "sub", Err: errors.New("sub result not match")}
			}

		case OP_MUL:
			// replay mul
			aNum, bNum, err := vm.arithmeticPrepare(ins, "mul")
			if err != nil {
				return err
			}
			mul := aNum * bNum

			vm.PC++

			// verify that replay res and given res are the same
			givenMul, err := vm.getInt64Arg(ins.ArgIdentifier[2]) // 已经进入下一行了，所以这里获取积是合法的
			if err != nil {
				return err
			}
			if givenMul != mul {
				return &ErrArithmetic{Operation: "mul", Err: errors.New("mul result not match")}
			}

		case OP_DIV:
			// replay div
			aNum, bNum, err := vm.arithmeticPrepare(ins, "div")
			if err != nil {
				return err
			}
			div := aNum / bNum

			vm.PC++

			// verify that replay res and given res are the same
			givenDiv, err := vm.getInt64Arg(ins.ArgIdentifier[2]) // 已经进入下一行了，所以这里获取商是合法的
			if err != nil {
				return err
			}
			if givenDiv != div {
				return &ErrArithmetic{Operation: "div", Err: errors.New("div result not match")}
			}

		case OP_EQ_INT:
			// replay eq int
			aNum, bNum, err := vm.arithmeticPrepare(ins, "eq_int")
			if err != nil {
				return err
			}
			eqIntBool := aNum == bNum
			var eqInt int64
			if eqIntBool {
				eqInt = 1
			} else {
				eqInt = 0
			}

			vm.PC++

			// verify that replay res and given res are the same
			givenInt, err := vm.getInt64Arg(ins.ArgIdentifier[2]) // 已经进入下一行了，所以这里获取是否相等是合法的
			if err != nil {
				return err
			}
			if givenInt != eqInt {
				return &ErrArithmetic{Operation: "eq_int", Err: errors.New("eq_int result not match")}
			}

		case OP_EQ_BYTES:
			// replay eq bytes
			aBytes, bBytes, err := vm.extractBytesForEq(ins)
			if err != nil {
				return err
			}
			var boolean int64
			if bytes.Equal(aBytes, bBytes) {
				boolean = 1
			} else {
				boolean = 0
			}

			vm.PC++

			// verify that replay res and given res are the same
			givenBoolean, err := vm.getInt64Arg(ins.ArgIdentifier[2]) // 已经进入下一行了，所以这里获取是否相等是合法的
			if err != nil {
				return err
			}
			if givenBoolean != boolean {
				return &ErrArithmetic{Operation: "eq_bytes", Err: errors.New("eq_bytes result not match")}
			}

		case OP_LARGE_INT:
			// replay large int
			aNum, bNum, err := vm.arithmeticPrepare(ins, "large_int")
			if err != nil {
				return err
			}
			var boolean int64
			if aNum > bNum {
				boolean = 1
			} else {
				boolean = 0
			}

			vm.PC++

			// verify that replay res and given res are the same
			givenBoolean, err := vm.getInt64Arg(ins.ArgIdentifier[2]) // 已经进入下一行了，所以这里获取是否大于是合法的
			if err != nil {
				return err
			}
			if givenBoolean != boolean {
				return &ErrArithmetic{Operation: "large_int", Err: errors.New("large_int result not match")}
			}

		case OP_JMP:

		}
	}
}
