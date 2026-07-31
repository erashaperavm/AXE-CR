package vm

import (
	"bytes"
	"fmt"
	"strconv"
)

type VM struct {
	ExeDir  string
	WorkDir string

	Code      []Instruction   // 字节码
	PC        int64           // 第几行了
	Mem       *Memory         // 简单的堆栈模型
	Vars      map[string]Ptr  // 变量名称 -> 内存地址
	Blocks    map[int64]Block // 代码标记块
	Env       *Environment    // 调用外部函数
	Debug     bool            // 调试模式
	TraceMode bool            // 是否启用轨迹
	Traces    []Trace
}

type Trace struct {
	// todo
}

// NewVM 创建一个 VM，初始化内存与变量表
func NewVM(code []Instruction, env *Environment, debug bool, traceMode bool) *VM {
	return &VM{
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
func (vm *VM) AllocVar(name string, memType PtrKind, size int64) Ptr {
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

// GetInt64Var 获取 Stack 变量值（int64）
func (vm *VM) GetInt64Var(name string) (int64, bool) {
	ptr, ok := vm.Vars[name]
	if !ok || ptr.Kind != Stack {
		return 0, false
	}
	return vm.Mem.ReadStack(ptr)
}

// GetBytesVar 获取 Heap 变量值（[]byte）
func (vm *VM) GetBytesVar(name string) ([]byte, bool) {
	ptr, ok := vm.Vars[name]
	if !ok || ptr.Kind != Heap {
		return nil, false
	}
	return vm.Mem.ReadHeap(ptr)
}

// UpdateVarByIdentifier 设置变量值
func (vm *VM) UpdateVarByIdentifier(name string, valIdentifier string) bool {
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
		return vm.Mem.WriteStack(data, targetPtr)
	case Heap:
		data, ok := vm.Mem.ReadHeap(dataPtr)
		if !ok {
			return false
		}
		return vm.Mem.WriteHeap(data, targetPtr)
	default:
		panic(fmt.Sprintf("unsupported ptr kind: %v", dataPtr.Kind))
	}
}

func (vm *VM) UpdateVarBySurfaceInt64(name string, i int64) bool {
	ptr, ok := vm.Vars[name]
	if !ok {
		return false
	}
	if ptr.Kind != Stack {
		return false
	}

	return vm.Mem.WriteStack(i, ptr)
}

func (vm *VM) UpdateVarBySurfaceBytes(name string, data []byte) bool {
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
func (vm *VM) DropVar(name string) error {
	ptr, ok := vm.Vars[name]
	if !ok {
		return &ErrVarNotFound{Name: name}
	}
	vm.Mem.Free(ptr)
	delete(vm.Vars, name)
	return nil
}

func (vm *VM) Run() error {
	for vm.PC < int64(len(vm.Code)) {
		ins := vm.Code[vm.PC]

		if vm.Debug {
			fmt.Println(fmt.Sprintf("")) // todo
		}

		if vm.TraceMode {

		}

		// todo : fmt 和 types 由 compiler 保证
		switch ins.Op {
		case OP_READ:
			/*
				OP_READ -> MAIN INS

				[]byte
				Ptr
				Ptr

				pos -> 链上位置
				data -> 使用变量来存储读取的数据
				ok -> 是否读取成功
			*/

			pos_, ok := vm.GetBytesVar(ins.ArgIdentifier[0])
			if !ok {
				return &ErrVarNotFound{Name: ins.ArgIdentifier[0]}
			}

			dataPtr, ok := vm.Vars[ins.ArgIdentifier[1]]
			if !ok {
				return &ErrPtrNotFound{Name: ins.ArgIdentifier[1]}
			}

			okPtr, ok := vm.Vars[ins.ArgIdentifier[2]]
			if !ok {
				return &ErrPtrNotFound{Name: ins.ArgIdentifier[2]}
			}
			if okPtr.Kind != Stack {
				return &ErrTypeMismatch{Expected: "int64", Got: "bytes"}
			}

			if dataPtr.Kind == Stack {
				data, err := vm.Env.ReadInt64(pos_)
				if err != nil {
					// 更新 ok 为 false
					ok := vm.UpdateVarBySurfaceInt64(ins.ArgIdentifier[2], 0)
					if !ok {
						return &ErrUpdateVarBySurfaceInt64{
							VarName: ins.ArgIdentifier[2],
							Surface: 0,
						}
					}
				}

				err = vm.irUpdaterInt64(data, ins)
				if err != nil {
					return &ErrRead{
						Pos: pos_,
						Err: err,
					}
				}
			} else {
				data, err := vm.Env.ReadBytes(pos_)
				if err != nil {
					// 更新 ok 为 false
					ok := vm.UpdateVarBySurfaceInt64(ins.ArgIdentifier[2], 0)
					if !ok {
						return &ErrUpdateVarBySurfaceInt64{
							VarName: ins.ArgIdentifier[2],
							Surface: 0,
						}
					}
				}

				err = vm.irUpdaterBytes(data, ins)
				if err != nil {
					return &ErrRead{
						Pos: pos_,
						Err: err,
					}
				}
			}

			vm.PC++

		case OP_INPUT:
			/*
				OP_INPUT -> MAIN INS

				int64
				Ptr

				idx -> 输入参数第几个
				data -> 使用变量来存储输入的数据
			*/

			idx, ok := vm.GetInt64Var(ins.ArgIdentifier[0])
			if !ok {
				return &ErrVarNotFound{Name: ins.ArgIdentifier[0]}
			}
			dataPtr, ok := vm.Vars[ins.ArgIdentifier[1]]
			if !ok {
				return &ErrPtrNotFound{Name: ins.ArgIdentifier[1]}
			}

			if dataPtr.Kind == Stack {
				data := vm.Env.InputInt64(idx)
				err := vm.irUpdaterInt64(data, ins)
				if err != nil {
					return &InputErr{
						Index: idx,
						Err:   err,
					}
				}
			} else {
				data := vm.Env.InputBytes(idx)
				err := vm.irUpdaterBytes(data, ins)
				if err != nil {
					return &InputErr{
						Index: idx,
						Err:   err,
					}
				}
			}

			vm.PC++

		case OP_WRITE:
			/*
				OP_WRITE -> MAIN INS

				[]byte
				Ptr
				Ptr

				pos -> 链上位置
				data -> 使用变量来存储将要写入的数据
				ok -> 是否写入成功
			*/

			pos_, ok := vm.GetBytesVar(ins.ArgIdentifier[0])
			if !ok {
				return &ErrVarNotFound{Name: ins.ArgIdentifier[0]}
			}

			dataPtr, ok := vm.Vars[ins.ArgIdentifier[1]]
			if !ok {
				return &ErrVarNotFound{Name: ins.ArgIdentifier[1]}
			}

			if dataPtr.Kind == Stack {
				data, ok := vm.GetInt64Var(ins.ArgIdentifier[1])
				if !ok {
					return &ErrVarNotFound{Name: ins.ArgIdentifier[1]}
				}
				err := vm.Env.WriteInt64(pos_, data)
				if err != nil {
					// 更新 ok 为 false
					ok := vm.UpdateVarBySurfaceInt64(ins.ArgIdentifier[2], 0)
					if !ok {
						return &ErrUpdateVarBySurfaceInt64{
							VarName: ins.ArgIdentifier[2],
							Surface: 0,
						}
					}
				}
			} else {
				data, ok := vm.GetBytesVar(ins.ArgIdentifier[1])
				if !ok {
					return &ErrVarNotFound{Name: ins.ArgIdentifier[1]}
				}
				err := vm.Env.WriteBytes(pos_, data)
				if err != nil {
					// 更新 ok 为 false
					ok := vm.UpdateVarBySurfaceInt64(ins.ArgIdentifier[2], 0)
					if !ok {
						return &ErrUpdateVarBySurfaceInt64{
							VarName: ins.ArgIdentifier[2],
							Surface: 0,
						}
					}
				}
			}

			vm.PC++

		case OP_OUTPUT:
			/*
				OP_OUTPUT -> MAIN INS

				int64
				Ptr

				idx -> 输出参数第几个
				data -> 使用变量来存储将要输出的数据
			*/

			idx, ok := vm.GetInt64Var(ins.ArgIdentifier[0])
			if !ok {
				return &ErrVarNotFound{Name: ins.ArgIdentifier[0]}
			}
			dataPtr, ok := vm.Vars[ins.ArgIdentifier[1]]
			if !ok {
				return &ErrVarNotFound{Name: ins.ArgIdentifier[1]}
			}

			if dataPtr.Kind == Stack {
				data, ok := vm.GetInt64Var(ins.ArgIdentifier[1])
				if !ok {
					return &ErrVarNotFound{Name: ins.ArgIdentifier[1]}
				}
				vm.Env.OutputInt64(idx, data)
			} else {
				data, ok := vm.GetBytesVar(ins.ArgIdentifier[1])
				if !ok {
					return &ErrVarNotFound{Name: ins.ArgIdentifier[1]}
				}
				vm.Env.OutputBytes(idx, data)
			}

			vm.PC++

		case OP_ALLOC:
			/*
				OP_ALLOC

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
				return &ErrOperandType{Operation: "alloc", Detail: fmt.Sprintf("memType '%s' is not int64", ins.ArgIdentifier[1])}
			}
			size, err := strconv.ParseInt(ins.ArgIdentifier[2], 10, 64)
			if err != nil {
				return &ErrOperandType{Operation: "alloc", Detail: fmt.Sprintf("size '%s' is not int64", ins.ArgIdentifier[2])}
			}

			vm.AllocVar(varName, PtrKind(memType), size)

			vm.PC++

		case OP_UPDATE:
			/*
				OP_UPDATE

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
					return &ErrVarNotFound{Name: ins.ArgIdentifier[0]}
				}
			case surfaceBytes:
				ok := vm.UpdateVarBySurfaceBytes(ins.ArgIdentifier[0], byt)
				if !ok {
					return &ErrVarNotFound{Name: ins.ArgIdentifier[0]}
				}
			case identifier:
				ok := vm.UpdateVarByIdentifier(ins.ArgIdentifier[0], ins.ArgIdentifier[1])
				if !ok {
					return &ErrVarNotFound{Name: ins.ArgIdentifier[0]}
				}
			default:
				return &ErrVarNotFound{Name: ins.ArgIdentifier[0]}
			}

			vm.PC++

		case OP_DROP:
			/*
				OP_DROP

				var_name

				varname
			*/

			err := vm.DropVar(ins.ArgIdentifier[0])
			if err != nil {
				return err
			}

			vm.PC++

		case OP_ADD:
			/*
				OP_ADD

				int64
				int64
				var_name

				num1
				num2
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

				int64
				int64
				var_name

				num1
				num2
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

				int64
				int64
				var_name

				num1
				num2
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

				int64
				int64
				var_name

				num1
				num2
				sum
			*/

			aNum, bNum, err := vm.arithmeticPrepare(ins, "div")
			if err != nil {
				return err
			}

			vm.UpdateVarBySurfaceInt64(ins.ArgIdentifier[2], aNum/bNum)

			vm.PC++

		case OP_EQ_INT:
			/*
				OP_CMP_INT

				int64
				int64
				var_name (int64(bool))

				num1
				num2
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
				return &ErrVarNotFound{Name: ins.ArgIdentifier[2]}
			}

			vm.PC++

		case OP_EQ_BYTES:
			/*
				OP_CMP_BYTES

				[]byte
				[]byte
				var_name (int64(bool))

				bytes1
				bytes2
				sum
			*/

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

			ok := vm.UpdateVarBySurfaceInt64(ins.ArgIdentifier[2], boolean)
			if !ok {
				return &ErrVarNotFound{Name: ins.ArgIdentifier[2]}
			}

			vm.PC++

		case OP_LARGE_INT:
			/*
				OP_LARGE_INT

				int64
				int64
				var_name (int64(bool))

				num1
				num2
				res
			*/

			aNum, bNum, err := vm.arithmeticPrepare(ins, "large_int")
			if err != nil {
				return err // arithmeticPrepare already returns *ErrArithmetic
			}

			var boolean int64
			if aNum > bNum {
				boolean = 1
			} else {
				boolean = 0
			}

			ok := vm.UpdateVarBySurfaceInt64(ins.ArgIdentifier[2], boolean)
			if !ok {
				return &ErrVarNotFound{Name: ins.ArgIdentifier[2]}
			}

			vm.PC++

		case OP_JMP:
			/*
				OP_JMP

				label (surface int64)

				000...
			*/
			labelStr := ins.ArgIdentifier[0]
			label, err := strconv.ParseInt(labelStr, 10, 64)
			if err != nil {
				return &ErrBlockLabelInvalid{Label: labelStr, Reason: "not a valid integer"}
			}

			tgtBlock, ok := vm.Blocks[label]
			if !ok {
				return &ErrLabelNotFound{Label: labelStr}
			}
			eptBlock := Block{}
			if tgtBlock == eptBlock {
				return &ErrLabelEmpty{Label: labelStr}
			}
			vm.PC = tgtBlock.BeginPC

		case OP_IF:
			/*
				OP_IF

				bool (int64)
				label (int64)

				condition
			*/

			conditionStr := ins.ArgIdentifier[0]
			condition, err := strconv.ParseInt(conditionStr, 10, 64)
			if err != nil {
				return &ErrOperandType{Operation: "if", Detail: fmt.Sprintf("condition '%s' not a valid int64: %v", conditionStr, err)}
			}

			switch condition {
			case 0:
				// false, 跳过 label
				vm.PC++
			case 1:
				// true, jmp to label block
				labelStr := ins.ArgIdentifier[1]
				label, err := strconv.ParseInt(labelStr, 10, 64)
				if err != nil {
					return &ErrBlockLabelInvalid{Label: labelStr, Reason: "not a valid integer"}
				}
				tgtBlock, ok := vm.Blocks[label]
				if !ok {
					return &ErrLabelNotFound{Label: labelStr}
				}
				eptBlock := Block{}
				if tgtBlock == eptBlock {
					return &ErrLabelEmpty{Label: labelStr}
				}
				vm.PC = tgtBlock.BeginPC
			default:
				return &ErrOperandType{Operation: "if", Detail: fmt.Sprintf("condition '%s' is illegal: must be 0 or 1", conditionStr)}
			}

		case OP_BEGIN:
			/*
				OP_BEGIN

				label (surface int64)

				000...
			*/

			beginPC := vm.PC - 1
			var endPC int64
			label := ins.ArgIdentifier[0]
			isEndExists := false

			for nowPC, nowIns := range vm.Code[beginPC:] {
				switch nowIns.Op {
				case OP_END:
					// 遇到 end, 检查 label
					if nowIns.ArgIdentifier[0] != label {
						// 不是当前 label 的 end
						continue
					}
					// 是当前 label 的 end,记录
					isEndExists = true
					endPC = int64(nowPC) - 1
					break
				default:
					// 这也意味着不能嵌套 block
					continue
				}
			}

			if !isEndExists {
				// do not found end
				return &ErrBlockNotClosed{Label: label}
			}

			// 存在，记录
			block := Block{
				BeginPC: beginPC,
				EndPC:   endPC,
			}

			labelNum, err := strconv.ParseInt(label, 10, 64)
			if err != nil {
				return &ErrBlockLabelInvalid{Label: label, Reason: "not a number"}
			}

			vm.Blocks[labelNum] = block

			vm.PC = endPC + 2 // 跳过 block body 和 end，到达 end 后一条指令

		case OP_END:
			// 理论上不可能到达这个位置
			// 因为已经在 OP_BEGIN 处理

			return &ErrUnexpectedEnd{Label: ins.ArgIdentifier[0]}

		case OP_CALL_RS:
			/*
				OP_CALL_RS

				func_name
				int64
				int64
				int64 / []byte
				... n 个

				funname
				inNum
				outNum
				i0
				i1
				o0
			*/

			// 获取要求
			funName := ins.ArgIdentifier[0]
			funReq, ok := vm.Env.Funcs[funName]
			if !ok {
				return &ErrFuncNotFound{Name: funName}
			}

			// 检查个数要求
			inNumStr := ins.ArgIdentifier[1]
			outNumStr := ins.ArgIdentifier[2]
			inNum, err := vm.getInt64Arg(inNumStr)
			if err != nil {
				return &ErrOperandType{Operation: "call_rs", Detail: fmt.Sprintf("invalid inNum '%s': %v", inNumStr, err)}
			}
			if inNum != int64(len(funReq.ReqInput)) {
				return &ErrFuncInputCount{Name: funName, Expected: len(funReq.ReqInput), Got: int(inNum)}
			}
			outNum, err := vm.getInt64Arg(outNumStr)
			if err != nil {
				return &ErrOperandType{Operation: "call_rs", Detail: fmt.Sprintf("invalid outNum '%s': %v", outNumStr, err)}
			}
			if outNum != int64(len(funReq.ReqOutput)) {
				return &ErrFuncOutputCount{Name: funName, Expected: len(funReq.ReqOutput), Got: int(outNum)}
			}

			// Build input values and output pointers from argument identifiers
			argStrs := ins.ArgIdentifier[3:]
			inputs := make([][]byte, inNum)
			outputVarNames := make([]string, outNum)

			for idx, argStr := range argStrs {
				if int64(idx) < inNum {
					// Input argument: must match ReqInput type and be encoded to []byte
					expectedType := funReq.ReqInput[idx]
					switch expectedType {
					case "int64":
						num, err := vm.getInt64Arg(argStr)
						if err != nil {
							// getInt64Arg already returns a structured error; wrap context
							return &ErrOperandType{Operation: "call_rs", Detail: fmt.Sprintf("input %d for '%s': %v", idx, funName, err)}
						}
						inputs[idx] = []byte(strconv.FormatInt(num, 10))
					case "bytes":
						b, err := vm.getBytesArg(argStr)
						if err != nil {
							return &ErrOperandType{Operation: "call_rs", Detail: fmt.Sprintf("input %d for '%s': %v", idx, funName, err)}
						}
						inputs[idx] = b
					default:
						return &ErrUnsupportedInputType{FuncName: funName, InputIndex: idx, InputType: expectedType}
					}
				} else {
					// Output argument: must match ReqOutput type and be a variable name (not literal)
					outIdx := int64(idx) - inNum
					expectedType := funReq.ReqOutput[outIdx]

					// Validate that argStr is a variable, not a literal
					if len(argStr) >= 2 && argStr[0] == '\'' && argStr[len(argStr)-1] == '\'' {
						return &ErrFuncOutputNotVar{FuncName: funName, OutputIndex: int(outIdx)}
					}
					if _, err := strconv.ParseInt(argStr, 10, 64); err == nil {
						return &ErrFuncOutputNotVar{FuncName: funName, OutputIndex: int(outIdx)}
					}

					// Check that the variable exists and has the correct type
					ptr, exists := vm.Vars[argStr]
					if !exists {
						return &ErrVarNotFound{Name: argStr}
					}
					if expectedType == "int64" && ptr.Kind != Stack {
						return &ErrOperandType{Operation: "call_rs", Detail: fmt.Sprintf("output %d for '%s' expects int64 (stack) but variable '%s' is not stack", outIdx, funName, argStr)}
					}
					if expectedType == "bytes" && ptr.Kind != Heap {
						return &ErrOperandType{Operation: "call_rs", Detail: fmt.Sprintf("output %d for '%s' expects bytes (heap) but variable '%s' is not heap", outIdx, funName, argStr)}
					}

					// Store variable name for later write-back (CallRS now returns values directly)
					outputVarNames[outIdx] = argStr
				}
			}

			// Call the RS function with encoded inputs
			outs, err := vm.Env.CallRS(funName, inputs)
			if err != nil {
				return &ErrOperandType{Operation: "call_rs", Detail: fmt.Sprintf("func '%s' failed: %v", funName, err)}
			}

			// Write back output values to the corresponding VM variables
			for i, varName := range outputVarNames {
				expectedType := funReq.ReqOutput[i]
				switch expectedType {
				case "int64":
					num, err := strconv.ParseInt(string(outs[i]), 10, 64)
					if err != nil {
						return &ErrOperandType{Operation: "call_rs", Detail: fmt.Sprintf("output %d for '%s': invalid int64: %v", i, funName, err)}
					}
					ok := vm.UpdateVarBySurfaceInt64(varName, num)
					if !ok {
						return &ErrVarNotFound{Name: varName}
					}
				case "bytes":
					ok := vm.UpdateVarBySurfaceBytes(varName, outs[i])
					if !ok {
						return &ErrVarNotFound{Name: varName}
					}
				default:
					return &ErrOperandType{Operation: "call_rs", Detail: fmt.Sprintf("unsupported output type for variable '%s'", varName)}
				}
			}

			vm.PC++

		case OP_CALL_C:
			// todo 忽略
			vm.PC++
		case OP_CALL_CPP:
			// todo 忽略
			vm.PC++

		default:
			return &ErrUnknownOpcode{Opcode: int(ins.Op)}
		}
	}
	return nil
}

func (vm *VM) irUpdaterInt64(data int64, ins Instruction) error {
	// 更新 data
	ok := vm.UpdateVarBySurfaceInt64(ins.ArgIdentifier[1], data)
	if !ok {
		// 更新 ok 为 false
		ok := vm.UpdateVarBySurfaceInt64(ins.ArgIdentifier[2], 0)
		if !ok {
			// 更新 ok 时 false
			return &ErrUpdateVarBySurfaceInt64{
				VarName: ins.ArgIdentifier[2],
				Surface: 0,
			}
		}
	}
	// 更新 ok 为 true
	ok = vm.UpdateVarBySurfaceInt64(ins.ArgIdentifier[2], 1)
	if !ok {
		// 更新 ok 时 false
		return &ErrUpdateVarBySurfaceInt64{
			VarName: ins.ArgIdentifier[2],
			Surface: 1,
		}
	}
	return nil
}

func (vm *VM) irUpdaterBytes(data []byte, ins Instruction) error {
	// 更新 data
	ok := vm.UpdateVarBySurfaceBytes(ins.ArgIdentifier[1], data)
	if !ok {
		// 更新 ok 为 false
		ok := vm.UpdateVarBySurfaceInt64(ins.ArgIdentifier[2], 0)
		if !ok {
			// 更新 ok 时 false
			return &ErrUpdateVarBySurfaceInt64{
				VarName: ins.ArgIdentifier[2],
				Surface: 0,
			}
		}
	}
	// 更新 ok 为 true
	ok = vm.UpdateVarBySurfaceInt64(ins.ArgIdentifier[2], 1)
	if !ok {
		// 更新 ok 时 false
		return &ErrUpdateVarBySurfaceInt64{
			VarName: ins.ArgIdentifier[2],
			Surface: 1,
		}
	}
	return nil
}

func (vm *VM) arithmeticPrepare(ins Instruction, opName string) (int64, int64, error) {
	// 使用 getInt64Arg 提取前两个操作数（只接受变量，拒绝字面量）
	aNum, err := vm.getInt64Arg(ins.ArgIdentifier[0])
	if err != nil {
		return 0, 0, &ErrArithmetic{Operation: opName, Err: err}
	}
	bNum, err := vm.getInt64Arg(ins.ArgIdentifier[1])
	if err != nil {
		return 0, 0, &ErrArithmetic{Operation: opName, Err: err}
	}

	// 第三个参数必须为 int64 变量，拒绝所有字面量
	if len(ins.ArgIdentifier[2]) >= 2 && ins.ArgIdentifier[2][0] == '\'' && ins.ArgIdentifier[2][len(ins.ArgIdentifier[2])-1] == '\'' {
		return 0, 0, &ErrArithmetic{Operation: opName, Err: &ErrOperandType{Operation: opName, Detail: fmt.Sprintf("sum var '%s' is []byte", ins.ArgIdentifier[2])}}
	}
	if _, err := strconv.ParseInt(ins.ArgIdentifier[2], 10, 64); err == nil {
		return 0, 0, &ErrArithmetic{Operation: opName, Err: &ErrOperandType{Operation: opName, Detail: fmt.Sprintf("sum var '%s' is int64", ins.ArgIdentifier[2])}}
	}
	// 必须是已存在的 int64 变量
	_, ok := vm.GetInt64Var(ins.ArgIdentifier[2])
	if !ok {
		return 0, 0, &ErrArithmetic{Operation: opName, Err: &ErrVarNotFound{Name: ins.ArgIdentifier[2]}}
	}
	return aNum, bNum, nil
}

func (vm *VM) extractBytesForEq(ins Instruction) ([]byte, []byte, error) {
	// 使用 getBytesArg 提取前两个操作数
	aBytes, err := vm.getBytesArg(ins.ArgIdentifier[0])
	if err != nil {
		return nil, nil, &ErrArithmetic{Operation: "eq_bytes", Err: err}
	}
	bBytes, err := vm.getBytesArg(ins.ArgIdentifier[1])
	if err != nil {
		return nil, nil, &ErrArithmetic{Operation: "eq_bytes", Err: err}
	}

	// 第三个参数必须为 int64 变量，拒绝所有字面量
	if len(ins.ArgIdentifier[2]) >= 2 && ins.ArgIdentifier[2][0] == '\'' && ins.ArgIdentifier[2][len(ins.ArgIdentifier[2])-1] == '\'' {
		return nil, nil, &ErrArithmetic{Operation: "eq_bytes", Err: &ErrOperandType{Operation: "eq_bytes", Detail: fmt.Sprintf("sum var '%s' is []byte", ins.ArgIdentifier[2])}}
	}
	if _, err := strconv.ParseInt(ins.ArgIdentifier[2], 10, 64); err == nil {
		return nil, nil, &ErrArithmetic{Operation: "eq_bytes", Err: &ErrOperandType{Operation: "eq_bytes", Detail: fmt.Sprintf("sum var '%s' is int64", ins.ArgIdentifier[2])}}
	}
	// 必须是已存在的 int64 变量
	_, ok := vm.GetInt64Var(ins.ArgIdentifier[2])
	if !ok {
		return nil, nil, &ErrArithmetic{Operation: "eq_bytes", Err: &ErrVarNotFound{Name: ins.ArgIdentifier[2]}}
	}
	return aBytes, bBytes, nil
}

// getInt64Arg 尝试从 identifier 中提取 int64 值。
// 若 identifier 为 []byte 字面量（'...'），则返回错误；
// 否则尝试解析为十进制 int64，若失败则从变量中读取。
func (vm *VM) getInt64Arg(identifier string) (int64, error) {
	// 检查是否为 []byte 字面量（以单引号包围）
	if len(identifier) >= 2 && identifier[0] == '\'' && identifier[len(identifier)-1] == '\'' {
		return 0, &ErrOperandType{Operation: "getInt64Arg", Detail: fmt.Sprintf("var '%s' is []byte", identifier)}
	}
	// 尝试解析为十进制 int64
	if num, err := strconv.ParseInt(identifier, 10, 64); err == nil {
		return num, nil
	}
	// 从变量中读取
	num, ok := vm.GetInt64Var(identifier)
	if !ok {
		return 0, &ErrVarNotFound{Name: identifier}
	}
	return num, nil
}

// getBytesArg 尝试从 identifier 中提取 []byte 值。
// 若 identifier 为 []byte 字面量（'...'），则直接返回其字节表示（包含引号）；
// 否则从变量中读取。
func (vm *VM) getBytesArg(identifier string) ([]byte, error) {
	// 检查是否为 []byte 字面量
	if len(identifier) >= 2 && identifier[0] == '\'' && identifier[len(identifier)-1] == '\'' {
		return []byte(identifier), nil
	}
	// 从变量中读取
	b, ok := vm.GetBytesVar(identifier)
	if !ok {
		return nil, &ErrVarNotFound{Name: identifier}
	}
	return b, nil
}
