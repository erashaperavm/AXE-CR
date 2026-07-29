package vm

import (
	"bytes"
	"fmt"
	"strconv"
)

type VM[T NativeType] struct {
	Code      []Instruction   // 字节码
	PC        int64           // 第几行了
	Mem       *Memory         // 简单的堆栈模型
	Vars      map[string]Ptr  // 变量名称 -> 内存地址
	Blocks    map[int64]Block // 代码标记块
	Env       *Environment[T] // 调用外部函数
	Debug     bool            // 调试模式
	TraceMode bool            // 是否启用轨迹
}

// NewVM 创建一个 VM，初始化内存与变量表
func NewVM[T NativeType](code []Instruction, env *Environment[T], debug bool, traceMode bool) *VM[T] {
	return &VM[T]{
		Code:      code,
		PC:        0,
		Mem:       NewMemoryObj(),
		Vars:      make(map[string]Ptr),
		Blocks:    make(map[int64]Block),
		Env:       env,
		Debug:     debug,
		TraceMode: traceMode,
	}
}

// AllocVar 声明一个变量并分配栈空间，返回指针; size only be needed for heap
func (vm *VM[T]) AllocVar(name string, memType PtrKind, size int64) Ptr {
	if _, exists := vm.Vars[name]; !exists {
		var ptr Ptr

		switch memType {
		case Stack:
			ptr = vm.Mem.AllocStack()
		case Heap:
			ptr = vm.Mem.AllocHeap(size)
		default:
			panic(fmt.Sprintf("unsupported mem type: %v", memType))
		}
		vm.Vars[name] = ptr
		return ptr
	}
	return vm.Vars[name] // 已存在则返回原指针
}

// GetVar 获取变量值
func (vm *VM[T]) GetVar(name string) (T, bool) {
	ptr, ok := vm.Vars[name]
	if !ok {
		var zero T
		return zero, false
	}
	var readVal interface{}
	switch ptr.Kind {
	case Stack:
		v, ok := vm.Mem.ReadStack(ptr)
		if !ok {
			var zero T
			return zero, false
		}
		readVal = v
	case Heap:
		v, ok := vm.Mem.ReadHeap(ptr)
		if !ok {
			var zero T
			return zero, false
		}
		readVal = v
	default:
		panic(fmt.Sprintf("unsupported ptr kind: %v", ptr.Kind))
	}
	return readVal.(T), true
}

// UpdateVarByIdentifier 设置变量值
func (vm *VM[T]) UpdateVarByIdentifier(name string, valIdentifier string) bool {
	targetPtr, ok := vm.Vars[name]
	if !ok {
		return false
	}
	dataPtr, ok := vm.Vars[valIdentifier]
	if !ok {
		return false
	}

	if targetPtr.Kind != dataPtr.Kind {
		return false
	}

	switch dataPtr.Kind {
	case Stack:
		data, ok := vm.Mem.ReadStack(dataPtr)
		if !ok {
			return false
		}
		v, ok := any(data).(int64)
		if !ok {
			panic(fmt.Sprintf("type mismatch: expected int64 for Stack, got %T", data))
		}
		return vm.Mem.WriteStack(v, targetPtr)
	case Heap:
		data, ok := vm.Mem.ReadHeap(dataPtr)
		if !ok {
			return false
		}
		v, ok := any(data).([]byte)
		if !ok {
			panic(fmt.Sprintf("type mismatch: expected []byte for Heap, got %T", data))
		}
		return vm.Mem.WriteHeap(v, targetPtr)
	default:
		panic(fmt.Sprintf("unsupported ptr kind: %v", dataPtr.Kind))
	}
}

func (vm *VM[T]) UpdateVarBySurfaceInt64(name string, i int64) bool {
	ptr, ok := vm.Vars[name]
	if !ok {
		return false
	}
	if ptr.Kind != Stack {
		return false
	}

	return vm.Mem.WriteStack(i, ptr)
}

func (vm *VM[T]) UpdateVarBySurfaceBytes(name string, data []byte) bool {
	ptr, ok := vm.Vars[name]
	if !ok {
		return false
	}
	if ptr.Kind != Heap {
		return false
	}

	return vm.Mem.WriteHeap(data, ptr)
}

// DropVar 删除变量并释放对应的栈/堆内存
func (vm *VM[T]) DropVar(name string) error {
	ptr, ok := vm.Vars[name]
	if !ok {
		return fmt.Errorf("variable '%s' not found", name)
	}
	vm.Mem.Free(ptr)
	delete(vm.Vars, name)
	return nil
}

func (vm *VM[T]) Run() error {
	for vm.PC < int64(len(vm.Code)) {
		ins := vm.Code[vm.PC]

		if vm.Debug {
			fmt.Println(fmt.Sprintf("")) // todo
		}

		// todo : fmt 和 types 由 compiler 保证
		switch ins.Op {
		case OP_READ:
			/*
				OP_READ -> MAIN INS
				3 -> 3 args

				[]byte
				Ptr
				Ptr

				pos -> 链上位置
				data -> 使用变量来存储读取的数据
				ok -> 是否读取成功
			*/

			pos, ok := vm.GetVar(ins.ArgIdentifier[0])
			if !ok {
				return fmt.Errorf("unable to read: variable '%s' not found", ins.ArgIdentifier[0])
			}
			pos_, ok := any(pos).([]byte)
			if !ok {
				return fmt.Errorf("type mismatch: expected int64 for Stack, got %T", pos)
			}

			dataPtr, ok := vm.Vars[ins.ArgIdentifier[1]]
			if !ok {
				return fmt.Errorf("unable to read: ptr '%s' not found", ins.ArgIdentifier[1])
			}

			okPtr, ok := vm.Vars[ins.ArgIdentifier[2]]
			if !ok {
				return fmt.Errorf("unable to read: okPtr '%s' not found", ins.ArgIdentifier[2])
			}

			vm.Env.Read(pos_, dataPtr, okPtr)

			vm.PC++

		case OP_INPUT:
			/*
				OP_INPUT -> MAIN INS
				2 -> 2 args

				int64
				Ptr

				idx -> 输入参数第几个
				data -> 使用变量来存储输入的数据
			*/

			idx, ok := any(vm.Vars[ins.ArgIdentifier[0]]).(int64)
			if !ok {
				return fmt.Errorf("variable '%s' not found", ins.ArgIdentifier[0])
			}
			dataPtr, ok := vm.Vars[ins.ArgIdentifier[1]]
			if !ok {
				return fmt.Errorf("unable to read: ptr '%s' not found", ins.ArgIdentifier[1])
			}

			vm.Env.Input(idx, dataPtr)

			vm.PC++

		case OP_WRITE:
			/*
				OP_WRITE -> MAIN INS
				3 -> 3 args

				[]byte
				Ptr
				Ptr

				pos -> 链上位置
				data -> 使用变量来存储将要写入的数据
				ok -> 是否写入成功
			*/

			pos, ok := vm.GetVar(ins.ArgIdentifier[0])
			if !ok {
				return fmt.Errorf("unable to read: variable '%s' not found", ins.ArgIdentifier[0])
			}
			pos_, ok := any(pos).([]byte)
			if !ok {
				return fmt.Errorf("type mismatch: expected int64 for Stack, got %T", pos)
			}

			dataPtr, ok := vm.Vars[ins.ArgIdentifier[1]]
			if !ok {
				return fmt.Errorf("unable to read: ptr '%s' not found", ins.ArgIdentifier[1])
			}

			okPtr, ok := vm.Vars[ins.ArgIdentifier[2]]
			if !ok {
				return fmt.Errorf("unable to read: okPtr '%s' not found", ins.ArgIdentifier[2])
			}

			vm.Env.Write(pos_, dataPtr, okPtr)

			vm.PC++

		case OP_OUTPUT:
			/*
				OP_OUTPUT -> MAIN INS
				2 -> 2 args

				int64
				Ptr

				idx -> 输出参数第几个
				data -> 使用变量来存储将要输出的数据
			*/

			idx, ok := any(vm.Vars[ins.ArgIdentifier[0]]).(int64)
			if !ok {
				return fmt.Errorf("variable '%s' not found", ins.ArgIdentifier[0])
			}
			dataPtr, ok := vm.Vars[ins.ArgIdentifier[1]]
			if !ok {
				return fmt.Errorf("unable to read: ptr '%s' not found", ins.ArgIdentifier[1])
			}

			vm.Env.Output(idx, dataPtr)

			vm.PC++

		case OP_ALLOC:
			/*
				OP_ALLOC
				3 -> 3 args

				var_name
				int64
				int64

				varname _> 变量名称
				mem_type -> 内存类型（编译器自动根据数据类型判断：int64 -> stack; []byte -> heap）
				size -> 分配内存大小 (如果为 heap）
			*/

			varName := ins.ArgIdentifier[0]
			memType, err := strconv.ParseInt(ins.ArgIdentifier[1], 10, 64)
			if err != nil {
				return fmt.Errorf("variable '%s' not found", ins.ArgIdentifier[1])
			}
			size, err := strconv.ParseInt(ins.ArgIdentifier[2], 10, 64)
			if err != nil {
				return fmt.Errorf("unable to read: size '%s' not found", ins.ArgIdentifier[2])
			}

			vm.AllocVar(varName, PtrKind(memType), size)

			vm.PC++

		case OP_UPDATE:
			/*
				OP_UPDATE
				2 -> 2 args

				var_name
				NativeType / varname -> 字面值（非变量，int64 数字或 []byte '' 数据）

				varname -> 变量名称
				surface / varname -> 字面值
			*/

			const (
				surfaceBytes = iota
				surfaceInt64
				identifier
			)

			var kind byte
			var num int64
			var byt []byte

			if ins.ArgIdentifier[1][0] == '\'' || ins.ArgIdentifier[1][len(ins.ArgIdentifier[1])] == '\'' {
				// 是 []byte
				kind = surfaceBytes
				byt = []byte(ins.ArgIdentifier[1])
			} else {
				num_, err := strconv.ParseInt(ins.ArgIdentifier[1], 10, 64)
				if err != nil {
					// 不是 int64
					kind = identifier
				}
				// 是 int64
				kind = surfaceInt64
				num = num_
			}

			switch kind {
			case surfaceInt64:
				ok := vm.UpdateVarBySurfaceInt64(ins.ArgIdentifier[0], num)
				if !ok {
					return fmt.Errorf("unable to update: var '%s' not found", ins.ArgIdentifier[0])
				}
			case surfaceBytes:
				ok := vm.UpdateVarBySurfaceBytes(ins.ArgIdentifier[0], byt)
				if !ok {
					return fmt.Errorf("unable to update: var '%s' not found", ins.ArgIdentifier[0])
				}
			case identifier:
				ok := vm.UpdateVarByIdentifier(ins.ArgIdentifier[0], ins.ArgIdentifier[1])
				if !ok {
					return fmt.Errorf("unable to update: var '%s' not found", ins.ArgIdentifier[0])
				}
			default:
				return fmt.Errorf("unable to update: var '%s' not found", ins.ArgIdentifier[0])
			}

			vm.PC++

		case OP_DROP:
			/*
				OP_DROP
				1 -> 1 arg

				var_name
			*/

			err := vm.DropVar(ins.ArgIdentifier[0])
			if err != nil {
				return err
			}

			vm.PC++

		case OP_ADD:
			/*
				OP_ADD
				3 -> 3 args

				int64
				int64
				var_name

				add1
				add2
				sum
			*/

			aNum, bNum, err := vm.arithmeticPrepare(ins, "add")
			if err != nil {
				return err
			}

			vm.UpdateVarBySurfaceInt64(ins.ArgIdentifier[2], aNum+bNum)

			vm.PC++

		case OP_SUB:
			/*
				OP_SUB
				3 -> 3 args

				int64
				int64
				var_name

				add1
				add2
				sum
			*/

			aNum, bNum, err := vm.arithmeticPrepare(ins, "sub")
			if err != nil {
				return err
			}

			vm.UpdateVarBySurfaceInt64(ins.ArgIdentifier[2], aNum-bNum)

			vm.PC++

		case OP_MUL:
			/*
				OP_MUL
				3 -> 3 args

				int64
				int64
				var_name

				add1
				add2
				sum
			*/

			aNum, bNum, err := vm.arithmeticPrepare(ins, "mul")
			if err != nil {
				return err
			}

			vm.UpdateVarBySurfaceInt64(ins.ArgIdentifier[2], aNum*bNum)

			vm.PC++

		case OP_DIV:
			/*
				OP_DIV
				3 -> 3 args

				int64
				int64
				var_name

				add1
				add2
				sum
			*/

			aNum, bNum, err := vm.arithmeticPrepare(ins, "div")
			if err != nil {
				return err
			}

			vm.UpdateVarBySurfaceInt64(ins.ArgIdentifier[2], aNum/bNum)

			vm.PC++

		case OP_CMP_INT:
			/*
				OP_CMP_INT
				3 -> 3 args

				int64
				int64
				var_name (int64(bool))

				add1
				add2
				sum
			*/

			aNum, bNum, err := vm.arithmeticPrepare(ins, "cmp_int")
			if err != nil {
				return err
			}

			var boolean int64

			if aNum == bNum {
				boolean = 1
			} else {
				boolean = 0
			}

			ok := vm.UpdateVarBySurfaceInt64(ins.ArgIdentifier[2], boolean)
			if !ok {
				return fmt.Errorf("cmp_int: var_name not found")
			}

			vm.PC++

		case OP_CMP_BYTES:
			/*
				OP_CMP_BYTES
				3 -> 3 args

				int64
				int64
				var_name (int64(bool))

				add1
				add2
				sum
			*/

			aBytes, bBytes, err := vm.extractBytesForCMP(ins)
			if err != nil {
				return err
			}

			var boolean int64

			if bytes.Equal(aBytes, bBytes) {
				boolean = 1
			} else {
				boolean = 0
			}

			ok := vm.UpdateVarBySurfaceInt64(ins.ArgIdentifier[2], boolean)
			if !ok {
				return fmt.Errorf("cmp_bytes: var_name not found")
			}

			vm.PC++

		case OP_JMP:
		case OP_IF:
		case OP_BEGIN:
		case OP_END:
		case OP_CALL_RS:
		case OP_CALL_C:
			// todo
		case OP_CALL_CPP:
			// todo

		default:
			return fmt.Errorf("unknown opcode: %d", ins.Op)
		}
	}
	return nil
}

func (vm *VM[T]) arithmeticPrepare(ins Instruction, opName string) (int64, int64, error) {
	var aNum int64
	var bNum int64
	var err error

	if ins.ArgIdentifier[0][0] == '\'' || ins.ArgIdentifier[0][len(ins.ArgIdentifier[0])] == '\'' {
		// 是 []byte
		return 0, 0, fmt.Errorf("unable to %v: var '%s' is []byte", opName, ins.ArgIdentifier[0])
	}
	aNum, err = strconv.ParseInt(ins.ArgIdentifier[0], 10, 64)
	if err != nil {
		// 不是 int64
		aNum_, ok := vm.GetVar(ins.ArgIdentifier[0])
		if !ok {
			return 0, 0, fmt.Errorf("unable to read: var '%s' not found", ins.ArgIdentifier[0])
		}
		aNum, ok = any(aNum_).(int64)
		if !ok {
			return 0, 0, fmt.Errorf("unable to read: var '%s' is not int64", ins.ArgIdentifier[0])
		}
	}
	// 是 int64
	// 不做

	if ins.ArgIdentifier[1][0] == '\'' || ins.ArgIdentifier[1][len(ins.ArgIdentifier[1])] == '\'' {
		// 是 []byte
		return 0, 0, fmt.Errorf("unable to %v: var '%s' is []byte", opName, ins.ArgIdentifier[1])
	}
	bNum, err = strconv.ParseInt(ins.ArgIdentifier[1], 10, 64)
	if err != nil {
		// 不是 int64
		bNum_, ok := vm.GetVar(ins.ArgIdentifier[1])
		if !ok {
			return 0, 0, fmt.Errorf("unable to read: var '%s' not found", ins.ArgIdentifier[1])
		}
		bNum, ok = any(bNum_).(int64)
		if !ok {
			return 0, 0, fmt.Errorf("unable to read: var '%s' is not int64", ins.ArgIdentifier[1])
		}
	}
	// 是 int64
	// 不做

	if ins.ArgIdentifier[2][0] == '\'' || ins.ArgIdentifier[2][len(ins.ArgIdentifier[2])] == '\'' {
		// 是 []byte
		return 0, 0, fmt.Errorf("unable to %v: var '%s' is []byte", opName, ins.ArgIdentifier[2])
	}
	_, err = strconv.ParseInt(ins.ArgIdentifier[2], 10, 64)
	if err != nil {
		// 不是 int64
		// 不做
	} else {
		// 是 int64
		return 0, 0, fmt.Errorf("unable to %v: sum var '%s' is int64", opName, ins.ArgIdentifier[2])
	}
	return aNum, bNum, nil
}

func (vm *VM[T]) extractBytesForCMP(ins Instruction) ([]byte, []byte, error) {
	var aBytes []byte
	var bBytes []byte
	var err error

	if ins.ArgIdentifier[0][0] == '\'' || ins.ArgIdentifier[0][len(ins.ArgIdentifier[0])] == '\'' {
		// 是 []byte
		aBytes = []byte(ins.ArgIdentifier[0])
	} else {
		return nil, nil, fmt.Errorf("unable to cmp_bytes: var '%s' is not []byte", ins.ArgIdentifier[0])
	}

	if ins.ArgIdentifier[1][0] == '\'' || ins.ArgIdentifier[1][len(ins.ArgIdentifier[1])] == '\'' {
		// 是 []byte
		bBytes = []byte(ins.ArgIdentifier[1])
	} else {
		return nil, nil, fmt.Errorf("unable to cmp_bytes: var '%s' is not []byte", ins.ArgIdentifier[0])
	}

	if ins.ArgIdentifier[2][0] == '\'' || ins.ArgIdentifier[2][len(ins.ArgIdentifier[2])] == '\'' {
		// 是 []byte
		return nil, nil, fmt.Errorf("unable to cmp_bytes: var '%s' is []byte", ins.ArgIdentifier[2])
	}
	_, err = strconv.ParseInt(ins.ArgIdentifier[2], 10, 64)
	if err != nil {
		// 不是 int64
		// 不做
	} else {
		// 是 int64
		return nil, nil, fmt.Errorf("unable to cmp_bytes: sum var '%s' is int64")
	}
	return aBytes, bBytes, nil
}
