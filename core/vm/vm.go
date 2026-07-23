package vm

import (
	"fmt"
	"strconv"
)

type VM struct {
	Code   []Instruction
	PC     int64          // 第几行了
	Mem    *Memory        // 简单的堆栈模型
	Vars   map[string]Ptr // 变量名称 -> 内存地址
	Blocks map[int64]Block
	Env    *Environment // 调用外部函数
	Debug  bool         // 调试模式
}

// NewVM 创建一个 VM，初始化内存与变量表
func NewVM(code []Instruction, env *Environment) *VM {
	return &VM{
		Code:   code,
		PC:     0,
		Mem:    NewMemory(),
		Vars:   make(map[string]Ptr),
		Blocks: make(map[int64]Block),
		Env:    env,
	}
}

// DeclareVar 声明一个变量并分配栈空间，返回栈指针
func (vm *VM) DeclareVar(name string) Ptr {
	if _, exists := vm.Vars[name]; !exists {
		ptr := vm.Mem.AllocStack()
		vm.Vars[name] = ptr
		return ptr
	}
	return vm.Vars[name] // 已存在则返回原指针
}

// GetVar 获取变量值
func (vm *VM) GetVar(name string) (int64, bool) {
	ptr, ok := vm.Vars[name]
	if !ok {
		return 0, false
	}
	return vm.Mem.ReadInt64(ptr)
}

// SetVar 设置变量值
func (vm *VM) SetVar(name string, val int64) bool {
	ptr, ok := vm.Vars[name]
	if !ok {
		return false
	}
	return vm.Mem.WriteInt64(ptr, val)
}

// CreateVar 创建变量并分配内存，若变量已存在则返回错误
func (vm *VM) CreateVar(name, varType string, size int64) error {
	if _, exists := vm.Vars[name]; exists {
		return fmt.Errorf("variable '%s' already exists", name)
	}
	var ptr Ptr
	switch varType {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64", "":
		ptr = vm.Mem.AllocStack()
	case "string":
		ptr = vm.Mem.AllocString("")
	case "slice":
		ptr = vm.Mem.AllocSlice([]byte{})
	case "heap":
		if size <= 0 {
			return fmt.Errorf("invalid heap size for variable '%s'", name)
		}
		ptr = vm.Mem.AllocHeap(size)
	default:
		return fmt.Errorf("unsupported variable type '%s'", varType)
	}
	vm.Vars[name] = ptr
	return nil
}

// UpdateVar 更新栈变量（int64）的值，变量不存在时返回错误
func (vm *VM) UpdateVar(name string, value int64) error {
	ptr, ok := vm.Vars[name]
	if !ok {
		return fmt.Errorf("variable '%s' not found", name)
	}
	if !vm.Mem.WriteInt64(ptr, value) {
		return fmt.Errorf("failed to write to variable '%s'", name)
	}
	return nil
}

// DropVar 删除变量并释放对应的栈/堆内存
func (vm *VM) DropVar(name string) error {
	ptr, ok := vm.Vars[name]
	if !ok {
		return fmt.Errorf("variable '%s' not found", name)
	}
	vm.Mem.Free(ptr)
	delete(vm.Vars, name)
	return nil
}

// UpdateVarByType 根据类型更新变量值
func (vm *VM) UpdateVarByType(name, varType, valueStr string) error {
	switch varType {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64", "":
		val, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid integer value '%s': %v", valueStr, err)
		}
		return vm.UpdateVar(name, val)
	case "string":
		return vm.UpdateVarString(name, valueStr)
	case "slice":
		return vm.UpdateVarSlice(name, []byte(valueStr))
	default:
		return fmt.Errorf("unsupported type '%s' for update", varType)
	}
}

// UpdateVarString 更新字符串变量
func (vm *VM) UpdateVarString(name, value string) error {
	ptr, ok := vm.Vars[name]
	if !ok {
		return fmt.Errorf("variable '%s' not found", name)
	}
	vm.Mem.Free(ptr)
	newPtr := vm.Mem.AllocString(value)
	vm.Vars[name] = newPtr
	return nil
}

// UpdateVarSlice 更新切片变量
func (vm *VM) UpdateVarSlice(name string, data []byte) error {
	ptr, ok := vm.Vars[name]
	if !ok {
		return fmt.Errorf("variable '%s' not found", name)
	}
	vm.Mem.Free(ptr)
	newPtr := vm.Mem.AllocSlice(data)
	vm.Vars[name] = newPtr
	return nil
}

func (vm *VM) Run() error {
	for vm.PC < int64(len(vm.Code)) {
		ins := vm.Code[vm.PC]

		if vm.Debug {
			fmt.Println(fmt.Sprintf("")) // todo
		}

		switch ins.Op {
		case OP_READ:

		case OP_INPUT:

		case OP_WRITE:

		case OP_OUTPUT:

		case OP_CREATE:
			// 格式: CREATE var_type, var_name [, size] (heap 时需要 size)
			if len(ins.InIdentifier) < 1 || len(ins.InType) < 1 {
				return fmt.Errorf("CREATE: insufficient input arguments")
			}
			varType := ins.InType[0]
			varName := ins.InIdentifier[0]
			var size int64 = 0
			if varType == "heap" {
				if len(ins.InIdentifier) < 2 {
					return fmt.Errorf("CREATE heap: missing size argument")
				}
				var err error
				size, err = strconv.ParseInt(ins.InIdentifier[1], 10, 64)
				if err != nil {
					return fmt.Errorf("CREATE heap: invalid size '%s': %v", ins.InIdentifier[1], err)
				}
			}
			if err := vm.CreateVar(varName, varType, size); err != nil {
				return err
			}
			vm.PC++

		case OP_UPDATE:
			// 格式: UPDATE var_type, var_name, value
			if len(ins.InIdentifier) < 2 || len(ins.InType) < 1 {
				return fmt.Errorf("UPDATE: insufficient input arguments")
			}
			varType := ins.InType[0]
			varName := ins.InIdentifier[0]
			if err := vm.UpdateVarByType(varName, varType, ins.InIdentifier[1]); err != nil {
				return err
			}
			vm.PC++

		case OP_DROP:
			// 格式: DROP var_name
			if len(ins.InIdentifier) < 1 {
				return fmt.Errorf("DROP: missing variable name")
			}
			varName := ins.InIdentifier[0]
			if err := vm.DropVar(varName); err != nil {
				return err
			}
			vm.PC++

		case OP_BEGIN:
			// 格式： OP_BEGIN block_id
			if len(ins.InIdentifier) < 1 {
				return fmt.Errorf("OP_BEGIN: missing block label")
			}
			blockIdString := ins.InIdentifier[0]
			blockId, err := strconv.ParseInt(blockIdString, 10, 64)
			if err != nil {
				return fmt.Errorf("OP_BEGIN: illegal block id")
			}
			if _, exists := vm.Blocks[blockId]; exists {
				return fmt.Errorf("OP_BEGIN: existed block")
			}

			beginPC := vm.PC // 记录当前 BEGIN 的地址
			privPC := vm.PC
			foundEnd := false
			for privPC < int64(len(vm.Code)) {
				if vm.Code[privPC].Op == OP_END {
					vm.Blocks[blockId] = Block{
						BeginPC: beginPC,
						EndPc:   privPC,
					}
					foundEnd = true
					break
				}
				privPC++
			}
			if !foundEnd {
				return fmt.Errorf("OP_BEGIN: unclosed block")
			}
			vm.PC++ // 进入循环体第一条指令

		case OP_LOOP:
			// 格式：OP_LOOP block_id
			if len(ins.InIdentifier) < 1 {
				return fmt.Errorf("OP_LOOP: missing block_id")
			}
			blockIdString := ins.InIdentifier[0]
			blockId, err := strconv.ParseInt(blockIdString, 10, 64)
			if err != nil {
				return fmt.Errorf("OP_LOOP: illegal block id")
			}
			block, exists := vm.Blocks[blockId]
			if !exists {
				return fmt.Errorf("OP_LOOP: undefined block %d", blockId)
			}
			vm.PC = block.BeginPC
			continue // 跳过末尾的 vm.PC++

		case OP_JMP:
			// 格式：JMP target_pc
			if len(ins.InIdentifier) < 1 {
				return fmt.Errorf("JMP: missing target address")
			}
			target, err := strconv.ParseInt(ins.InIdentifier[0], 10, 64)
			if err != nil {
				return fmt.Errorf("JMP: invalid target '%s': %v", ins.InIdentifier[0], err)
			}
			vm.PC = target
			continue // 跳过 vm.PC++，避免多移一次

		case OP_IF:
			// 格式：IF var_name, target_pc
			if len(ins.InIdentifier) < 2 {
				return fmt.Errorf("IF: insufficient arguments (var_name, target_pc)")
			}
			varName := ins.InIdentifier[0]
			target, err := strconv.ParseInt(ins.InIdentifier[1], 10, 64)
			if err != nil {
				return fmt.Errorf("IF: invalid target '%s': %v", ins.InIdentifier[1], err)
			}
			val, ok := vm.GetVar(varName)
			if !ok {
				return fmt.Errorf("IF: variable '%s' not found", varName)
			}
			if val == 0 {
				vm.PC = target
				continue // 跳过 vm.PC++
			}
			vm.PC++

		case OP_ELSIF:
			// 格式：ELSIF var_name, target_pc
			if len(ins.InIdentifier) < 2 {
				return fmt.Errorf("ELSIF: insufficient arguments (var_name, target_pc)")
			}
			varName := ins.InIdentifier[0]
			target, err := strconv.ParseInt(ins.InIdentifier[1], 10, 64)
			if err != nil {
				return fmt.Errorf("ELSIF: invalid target '%s': %v", ins.InIdentifier[1], err)
			}
			val, ok := vm.GetVar(varName)
			if !ok {
				return fmt.Errorf("ELSIF: variable '%s' not found", varName)
			}
			if val == 0 {
				vm.PC = target
				continue // 跳过 vm.PC++
			}
			vm.PC++

		case OP_ELSE:
			// ELSE 作为标记，直接顺序执行
			vm.PC++

		default:
			return fmt.Errorf("unknown opcode: %d", ins.Op)
		}
	}
	return nil
}
