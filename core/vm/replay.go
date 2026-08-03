package vm

import (
	"bytes"
	"errors"
)

func NewReplayVM(code []Instruction, env *Environment) *VM {
	return &VM{
		Code:      code,
		PC:        0,
		Lines:     0,
		Mem:       NewMemoryObj(),
		PreMem:    NewMemoryObj(),
		Vars:      make(map[string]Ptr),
		PreVars:   make(map[string]Ptr),
		Blocks:    make(map[int64]Block),
		PreBlocks: make(map[int64]Block),
		Env:       env,
		TraceMode: true,
		Traces:    make(map[int64]TraceStep),
		OriginPos: make(map[int64]map[int64]TokenPos),
	}
}

// Replay 回放 VM 执行轨迹, 需要一个新 VM 实例和一个新内存状态
func (vm *VM) Replay() error {
	if !vm.TraceMode {
		return nil
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
			// todo bridge read hash compute
			vm.PC++

		case OP_INPUT:
			// todo bridge input hash compute
			vm.PC++

		case OP_WRITE:
			// todo bridge write hash compute
			vm.PC++

		case OP_OUTPUT:
			// todo bridge output hash compute
			vm.PC++

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
