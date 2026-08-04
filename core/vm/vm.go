package vm

import (
	"axe-cr/core/utils"
	"bytes"
	"fmt"
	"os"
	"strconv"
)

type VM struct {
	Code      []Instruction                // 字节码
	OriginPos map[int64]map[int64]TokenPos // 原始位置，与 Code 里面的每一个 string Token 一一对应
	Debug     bool                         // 调试模式
	TraceMode bool                         // 是否启用轨迹
	Env       *Environment                 // 调用外部函数

	Int64PublicInputs  []int64  // 公共输入的 int64 数据
	BytesPublicInputs  [][]byte // 公共输入的 bytes 数据
	Int64PublicOutputs []int64  // 公共输出的 int64 数据
	BytesPublicOutputs [][]byte // 公共输出的 bytes 数据

	// 隐私输入直接在创建虚拟机时硬分配到虚拟机内，只读，内存地址实际为隐私输入参数 idx 索引
	/*
		Int64PrivateInputs  []int64  // 私有输入的 int64 数据（已经在节点端门限解密，结果需要在节点端再聚合成员节点的分片解密 proof）
		BytesPrivateInputs  [][]byte // 私有输入的 bytes 数据（已经在节点端门限解密，结果需要在节点端再聚合成员节点的分片解密 proof）
	*/

	Int64PrivateOutputs []int64  // 私有输出的 int64 数据（需在节点端加密，结果需要在节点端附带加密 proof）
	BytesPrivateOutputs [][]byte // 私有输出的 bytes 数据（需在节点端加密，结果需要在节点端附带加密 proof）

	PC    int64 // 第几行了
	Lines int64 // 真实执行命令函数行数

	Mem    *Memory // 简单的堆栈模型
	PreMem *Memory // 上一帧的内存状态

	Vars    map[string]Ptr // 变量名称 -> 内存地址
	PreVars map[string]Ptr // 上一帧的变量表

	Blocks    map[int64]Block // 代码标记块
	PreBlocks map[int64]Block // 上一帧的代码标记块

	PrivacyMem    *Memory            // 隐私内存，用于存储隐私数据，不暴露给验证者，只能直接输入 zkVM
	PrivacyMemExp map[int64][32]byte // 隐私内存表示，用于验证者比较隐私数据是否正确，使用 SHA256 计算得到

	Traces map[int64]TraceStep // 执行轨迹, vm.Traces[vm.Lines]

}

// NewExecuteVM 创建一个 VM，初始化内存与变量表
func NewExecuteVM(code []Instruction, env *Environment, debug bool, traceMode bool) *VM {
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
		Debug:     debug,
		TraceMode: traceMode,
		Traces:    make(map[int64]TraceStep),
		OriginPos: make(map[int64]map[int64]TokenPos),
	}
}

func (vm *VM) Run() (error, []TokenPos) {
	for vm.PC < int64(len(vm.Code)) {
		ins := vm.Code[vm.PC]

		if vm.Debug {
			fmt.Println(fmt.Sprintf("")) // todo
		}

		if vm.TraceMode {
			vm.PreMem = vm.Mem.copy()
			vm.PreVars = vm.copyVars()
			vm.PreBlocks = vm.copyBlocks()
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

			// todo：复杂度原因暂时禁用 RW 操作

			// uncomment to enable read operation
			/*
				pos_, ok := vm.getBytesVar(ins.ArgIdentifier[0])
				if !ok {
					sourcePos := vm.OriginPos[vm.PC][4]
					return &ErrVarNotFound{Name: ins.ArgIdentifier[0]}, []TokenPos{sourcePos}
				}

				dataPtr, ok := vm.Vars[ins.ArgIdentifier[1]]
				if !ok {
					sourcePos := vm.OriginPos[vm.PC][5]
					return &ErrPtrNotFound{Name: ins.ArgIdentifier[1]}, []TokenPos{sourcePos}
				}

				okPtr, ok := vm.Vars[ins.ArgIdentifier[2]]
				if !ok {
					sourcePos := vm.OriginPos[vm.PC][6]
					return &ErrPtrNotFound{Name: ins.ArgIdentifier[2]}, []TokenPos{sourcePos}
				}
				if okPtr.Kind != Stack {
					sourcePos := vm.OriginPos[vm.PC][6]
					return &ErrTypeMismatch{Expected: "int64", Got: "bytes"}, []TokenPos{sourcePos}
				}

				if dataPtr.Kind == Stack {
					data, err := vm.Env.ReadInt64(pos_)
					if err != nil {
						// 更新 ok 为 false
						ok := vm.updateVarBySurfaceInt64(ins.ArgIdentifier[2], 0)
						if !ok {
							sourcePos := vm.OriginPos[vm.PC][5]
							return &ErrUpdateVarBySurfaceInt64{
								VarName: ins.ArgIdentifier[2],
								Surface: 0,
							}, []TokenPos{sourcePos}
						}
					}

					err = vm.irUpdaterInt64(data, ins)
					if err != nil {
						sourcePos := vm.OriginPos[vm.PC][5]
						return &ErrRead{
							Pos: pos_,
							Err: err,
						}, []TokenPos{sourcePos}
					}
				} else {
					data, err := vm.Env.ReadBytes(pos_)
					if err != nil {
						// 更新 ok 为 false
						ok := vm.updateVarBySurfaceInt64(ins.ArgIdentifier[2], 0)
						if !ok {
							sourcePos := vm.OriginPos[vm.PC][5]
							return &ErrUpdateVarBySurfaceInt64{
								VarName: ins.ArgIdentifier[2],
								Surface: 0,
							}, []TokenPos{sourcePos}
						}
					}

					err = vm.irUpdaterBytes(data, ins)
					if err != nil {
						sourcePos := vm.OriginPos[vm.PC][5]
						return &ErrRead{
							Pos: pos_,
							Err: err,
						}, []TokenPos{sourcePos}
					}
				}

			*/

			vm.PC++
			vm.Lines++

		case OP_INPUT:
			/*
				OP_INPUT -> MAIN INS

				int64
				Ptr

				idx -> 输入参数第几个
				data -> 使用变量来存储输入的数据
			*/

			idx, ok := vm.getInt64Var(ins.ArgIdentifier[0])
			if !ok {
				sourcePos := vm.OriginPos[vm.PC][3]
				return &ErrVarNotFound{Name: ins.ArgIdentifier[0]}, []TokenPos{sourcePos}
			}
			dataPtr, ok := vm.Vars[ins.ArgIdentifier[1]]
			if !ok {
				sourcePos := vm.OriginPos[vm.PC][4]
				return &ErrPtrNotFound{Name: ins.ArgIdentifier[1]}, []TokenPos{sourcePos}
			}

			switch dataPtr.Kind {
			case PubStack:
				data := vm.Int64PublicInputs[idx]
				err := vm.irUpdaterInt64(data, ins)
				if err != nil {
					sourcePos := vm.OriginPos[vm.PC][4]
					return &InputErr{
						Index: idx,
						Err:   err,
					}, []TokenPos{sourcePos}
				}
			case PubHeap:
				data := vm.BytesPublicInputs[idx]
				err := vm.irUpdaterBytes(data, ins)
				if err != nil {
					sourcePos := vm.OriginPos[vm.PC][4]
					return &InputErr{
						Index: idx,
						Err:   err,
					}, []TokenPos{sourcePos}
				}
			case PrivStack:
				data := vm.PrivacyMem.Stack[idx]
				err := vm.irUpdaterInt64(data, ins)
				if err != nil {
					sourcePos := vm.OriginPos[vm.PC][4]
					return &InputErr{
						Index: idx,
						Err:   err,
					}, []TokenPos{sourcePos}
				}
			case PrivHeap:
				data := vm.PrivacyMem.Heap[idx]
				err := vm.irUpdaterBytes(data, ins)
				if err != nil {
					sourcePos := vm.OriginPos[vm.PC][4]
					return &InputErr{
						Index: idx,
						Err:   err,
					}, []TokenPos{sourcePos}
				}
			default:
				sourcePos := vm.OriginPos[vm.PC][4]
				return &ErrUnsupportedMemType{MemType: dataPtr.Kind}, []TokenPos{sourcePos}
			}

			vm.PC++
			vm.Lines++

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

			// todo：复杂度原因暂时禁用 RW 操作

			// uncomment to enable write operation
			/*
				pos_, ok := vm.getBytesVar(ins.ArgIdentifier[0])
				if !ok {
					sourcePos := vm.OriginPos[vm.PC][4]
					return &ErrVarNotFound{Name: ins.ArgIdentifier[0]}, []TokenPos{sourcePos}
				}

				dataPtr, ok := vm.Vars[ins.ArgIdentifier[1]]
				if !ok {
					sourcePos := vm.OriginPos[vm.PC][5]
					return &ErrVarNotFound{Name: ins.ArgIdentifier[1]}, []TokenPos{sourcePos}
				}

				if dataPtr.Kind == Stack {
					data, ok := vm.getInt64Var(ins.ArgIdentifier[1])
					if !ok {
						sourcePos := vm.OriginPos[vm.PC][5]
						return &ErrVarNotFound{Name: ins.ArgIdentifier[1]}, []TokenPos{sourcePos}
					}
					err := vm.Env.WriteInt64(pos_, data)
					if err != nil {
						// 更新 ok 为 false
						ok := vm.updateVarBySurfaceInt64(ins.ArgIdentifier[2], 0)
						if !ok {
							sourcePos := vm.OriginPos[vm.PC][5]
							return &ErrUpdateVarBySurfaceInt64{
								VarName: ins.ArgIdentifier[2],
								Surface: 0,
							}, []TokenPos{sourcePos}
						}
					}
				} else {
					data, ok := vm.getBytesVar(ins.ArgIdentifier[1])
					if !ok {
						sourcePos := vm.OriginPos[vm.PC][5]
						return &ErrVarNotFound{Name: ins.ArgIdentifier[1]}, []TokenPos{sourcePos}
					}
					err := vm.Env.WriteBytes(pos_, data)
					if err != nil {
						// 更新 ok 为 false
						ok := vm.updateVarBySurfaceInt64(ins.ArgIdentifier[2], 0)
						if !ok {
							sourcePos := vm.OriginPos[vm.PC][5]
							return &ErrUpdateVarBySurfaceInt64{
								VarName: ins.ArgIdentifier[2],
								Surface: 0,
							}, []TokenPos{sourcePos}
						}
					}
				}
			*/

			vm.PC++
			vm.Lines++

		case OP_OUTPUT:
			/*
				OP_OUTPUT -> MAIN INS

				int64
				Ptr

				idx -> 输出参数第几个
				data -> 使用变量来存储将要输出的数据
			*/

			idx, ok := vm.getInt64Var(ins.ArgIdentifier[0])
			if !ok {
				sourcePos := vm.OriginPos[vm.PC][3]
				return &ErrVarNotFound{Name: ins.ArgIdentifier[0]}, []TokenPos{sourcePos}
			}
			dataPtr, ok := vm.Vars[ins.ArgIdentifier[1]]
			if !ok {
				sourcePos := vm.OriginPos[vm.PC][4]
				return &ErrVarNotFound{Name: ins.ArgIdentifier[1]}, []TokenPos{sourcePos}
			}

			switch dataPtr.Kind {
			case PubStack:
				data, err := vm.getInt64Arg(ins.ArgIdentifier[4])
				if err != nil {
					sourcePos := vm.OriginPos[vm.PC][4]
					return &OutputErr{
						Index: idx,
						Err:   err,
					}, []TokenPos{sourcePos}
				}
				vm.Int64PublicOutputs[idx] = data
			case PubHeap:
				data, err := vm.getBytesArg(ins.ArgIdentifier[4])
				if err != nil {
					sourcePos := vm.OriginPos[vm.PC][4]
					return &OutputErr{
						Index: idx,
						Err:   err,
					}, []TokenPos{sourcePos}
				}
				vm.BytesPublicOutputs[idx] = data
			case PrivStack:
				data, err := vm.getInt64Arg(ins.ArgIdentifier[4])
				if err != nil {
					sourcePos := vm.OriginPos[vm.PC][4]
					return &OutputErr{
						Index: idx,
						Err:   err,
					}, []TokenPos{sourcePos}
				}
				vm.Int64PrivateOutputs[idx] = data
			case PrivHeap:
				data, err := vm.getBytesArg(ins.ArgIdentifier[4])
				if err != nil {
					sourcePos := vm.OriginPos[vm.PC][4]
					return &OutputErr{
						Index: idx,
						Err:   err,
					}, []TokenPos{sourcePos}
				}
				vm.BytesPrivateOutputs[idx] = data
			default:
				sourcePos := vm.OriginPos[vm.PC][4]
				return &ErrUnsupportedMemType{MemType: dataPtr.Kind}, []TokenPos{sourcePos}
			}

			vm.PC++
			vm.Lines++

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
				sourcePos := vm.OriginPos[vm.PC][5]
				return &ErrOperandType{Operation: "alloc", Detail: fmt.Sprintf("memType '%s' is not int64", ins.ArgIdentifier[1])}, []TokenPos{sourcePos}
			}
			size, err := strconv.ParseInt(ins.ArgIdentifier[2], 10, 64)
			if err != nil {
				sourcePos := vm.OriginPos[vm.PC][6]
				return &ErrOperandType{Operation: "alloc", Detail: fmt.Sprintf("size '%s' is not int64", ins.ArgIdentifier[2])}, []TokenPos{sourcePos}
			}

			vm.allocVar(varName, PtrKind(memType), size)

			vm.PC++
			vm.Lines++

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
				ok := vm.updateVarBySurfaceInt64(ins.ArgIdentifier[0], num)
				if !ok {
					sourcePos := vm.OriginPos[vm.PC][3]
					return &ErrVarNotFound{Name: ins.ArgIdentifier[0]}, []TokenPos{sourcePos}
				}
			case surfaceBytes:
				ok := vm.updateVarBySurfaceBytes(ins.ArgIdentifier[0], byt)
				if !ok {
					sourcePos := vm.OriginPos[vm.PC][3]
					return &ErrVarNotFound{Name: ins.ArgIdentifier[0]}, []TokenPos{sourcePos}
				}
			case identifier:
				ok := vm.updateVarByIdentifier(ins.ArgIdentifier[0], ins.ArgIdentifier[1])
				if !ok {
					sourcePos := vm.OriginPos[vm.PC][3]
					return &ErrVarNotFound{Name: ins.ArgIdentifier[0]}, []TokenPos{sourcePos}
				}
			default:
				sourcePos := vm.OriginPos[vm.PC][3]
				return &ErrVarNotFound{Name: ins.ArgIdentifier[0]}, []TokenPos{sourcePos}
			}

			vm.PC++
			vm.Lines++

		case OP_DROP:
			/*
				OP_DROP

				var_name

				varname
			*/

			err := vm.dropVar(ins.ArgIdentifier[0])
			if err != nil {
				sourcePos := vm.OriginPos[vm.PC][2]
				return &ErrVarDrop{Name: ins.ArgIdentifier[0]}, []TokenPos{sourcePos}
			}

			vm.PC++
			vm.Lines++

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
				sourcePosA := vm.OriginPos[vm.PC][4]
				sourcePosB := vm.OriginPos[vm.PC][5]
				return &ErrArithmetic{
					Operation: "add",
					Err:       err,
				}, []TokenPos{sourcePosA, sourcePosB}
			}

			ok := vm.updateVarBySurfaceInt64(ins.ArgIdentifier[2], aNum+bNum)
			if !ok {
				sourcePosSum := vm.OriginPos[vm.PC][6]
				return &ErrVarNotFound{Name: ins.ArgIdentifier[2]}, []TokenPos{sourcePosSum}
			}

			vm.PC++
			vm.Lines++

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
				sourcePosA := vm.OriginPos[vm.PC][4]
				sourcePosB := vm.OriginPos[vm.PC][5]
				return &ErrArithmetic{
					Operation: "sub",
					Err:       err,
				}, []TokenPos{sourcePosA, sourcePosB}
			}

			ok := vm.updateVarBySurfaceInt64(ins.ArgIdentifier[2], aNum-bNum)
			if !ok {
				sourcePosSub := vm.OriginPos[vm.PC][6]
				return &ErrVarNotFound{Name: ins.ArgIdentifier[2]}, []TokenPos{sourcePosSub}
			}

			vm.PC++
			vm.Lines++

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
				sourcePosA := vm.OriginPos[vm.PC][4]
				sourcePosB := vm.OriginPos[vm.PC][5]
				return &ErrArithmetic{
					Operation: "mul",
					Err:       err,
				}, []TokenPos{sourcePosA, sourcePosB}
			}

			ok := vm.updateVarBySurfaceInt64(ins.ArgIdentifier[2], aNum*bNum)
			if !ok {
				sourcePosProduct := vm.OriginPos[vm.PC][6]
				return &ErrVarNotFound{Name: ins.ArgIdentifier[2]}, []TokenPos{sourcePosProduct}
			}

			vm.PC++
			vm.Lines++

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
				sourcePosA := vm.OriginPos[vm.PC][4]
				sourcePosB := vm.OriginPos[vm.PC][5]
				return &ErrArithmetic{
					Operation: "div",
					Err:       err,
				}, []TokenPos{sourcePosA, sourcePosB}
			}

			ok := vm.updateVarBySurfaceInt64(ins.ArgIdentifier[2], aNum/bNum)
			if !ok {
				sourcePosDiv := vm.OriginPos[vm.PC][6]
				return &ErrVarNotFound{Name: ins.ArgIdentifier[2]}, []TokenPos{sourcePosDiv}
			}

			vm.PC++
			vm.Lines++

		case OP_EQ_INT:
			/*
				OP_EQ_INT

				int64
				int64
				var_name (int64(bool))

				num1
				num2
				sum
			*/

			aNum, bNum, err := vm.arithmeticPrepare(ins, "eq_int")
			if err != nil {
				sourcePosA := vm.OriginPos[vm.PC][4]
				sourcePosB := vm.OriginPos[vm.PC][5]
				return &ErrArithmetic{
					Operation: "eq_int",
					Err:       err,
				}, []TokenPos{sourcePosA, sourcePosB}
			}

			var boolean int64

			if aNum == bNum {
				boolean = 1
			} else {
				boolean = 0
			}

			ok := vm.updateVarBySurfaceInt64(ins.ArgIdentifier[2], boolean)
			if !ok {
				sourcePosSum := vm.OriginPos[vm.PC][6]
				return &ErrVarNotFound{Name: ins.ArgIdentifier[2]}, []TokenPos{sourcePosSum}
			}

			vm.PC++
			vm.Lines++

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
				sourcePosA := vm.OriginPos[vm.PC][4]
				sourcePosB := vm.OriginPos[vm.PC][5]
				return &ErrArithmetic{
					Operation: "eq_bytes",
					Err:       err,
				}, []TokenPos{sourcePosA, sourcePosB}
			}

			var boolean int64

			if bytes.Equal(aBytes, bBytes) {
				boolean = 1
			} else {
				boolean = 0
			}

			ok := vm.updateVarBySurfaceInt64(ins.ArgIdentifier[2], boolean)
			if !ok {
				sourcePosSum := vm.OriginPos[vm.PC][6]
				return &ErrVarNotFound{Name: ins.ArgIdentifier[2]}, []TokenPos{sourcePosSum}
			}

			vm.PC++
			vm.Lines++

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
				sourcePosA := vm.OriginPos[vm.PC][4]
				sourcePosB := vm.OriginPos[vm.PC][5]
				return &ErrArithmetic{
					Operation: "large_int",
					Err:       err,
				}, []TokenPos{sourcePosA, sourcePosB}
			}

			var boolean int64
			if aNum > bNum {
				boolean = 1
			} else {
				boolean = 0
			}

			ok := vm.updateVarBySurfaceInt64(ins.ArgIdentifier[2], boolean)
			if !ok {
				sourcePosSum := vm.OriginPos[vm.PC][6]
				return &ErrVarNotFound{Name: ins.ArgIdentifier[2]}, []TokenPos{sourcePosSum}
			}

			vm.PC++
			vm.Lines++

		case OP_JMP:
			/*
				OP_JMP

				label (surface int64)

				000...
			*/
			labelStr := ins.ArgIdentifier[0]
			label, err := strconv.ParseInt(labelStr, 10, 64)
			if err != nil {
				sourcePosJmp := vm.OriginPos[vm.PC][2]
				return &ErrBlockLabelInvalid{Label: labelStr, Reason: "not a valid integer"}, []TokenPos{sourcePosJmp}
			}

			tgtBlock, ok := vm.Blocks[label]
			if !ok {
				sourcePosJmp := vm.OriginPos[vm.PC][2]
				return &ErrLabelNotFound{Label: labelStr}, []TokenPos{sourcePosJmp}
			}
			eptBlock := Block{}
			if tgtBlock == eptBlock {
				sourcePosJmp := vm.OriginPos[vm.PC][2]
				return &ErrLabelEmpty{Label: labelStr}, []TokenPos{sourcePosJmp}
			}

			vm.PC = tgtBlock.BeginPC
			vm.Lines++

		case OP_IF:
			/*
				OP_IF

				bool (int64)
				label (int64)

				condition
				label
			*/

			conditionStr := ins.ArgIdentifier[0]
			condition, err := strconv.ParseInt(conditionStr, 10, 64)
			if err != nil {
				sourcePosIf := vm.OriginPos[vm.PC][3]
				return &ErrOperandType{Operation: "if", Detail: fmt.Sprintf("condition '%s' not a valid int64: %v", conditionStr, err)}, []TokenPos{sourcePosIf}
			}

			switch condition {
			case 0:
				// false, 跳过 label
				vm.PC++
				vm.Lines++
			case 1:
				// true, jmp to label block
				labelStr := ins.ArgIdentifier[1]
				label, err := strconv.ParseInt(labelStr, 10, 64)
				if err != nil {
					sourcePosIf := vm.OriginPos[vm.PC][4]
					return &ErrBlockLabelInvalid{Label: labelStr, Reason: "not a valid integer"}, []TokenPos{sourcePosIf}
				}
				tgtBlock, ok := vm.Blocks[label]
				if !ok {
					sourcePosIf := vm.OriginPos[vm.PC][4]
					return &ErrLabelNotFound{Label: labelStr}, []TokenPos{sourcePosIf}
				}
				eptBlock := Block{}
				if tgtBlock == eptBlock {
					sourcePosIf := vm.OriginPos[vm.PC][4]
					return &ErrLabelEmpty{Label: labelStr}, []TokenPos{sourcePosIf}
				}

				vm.PC = tgtBlock.BeginPC
				vm.Lines++
			default:
				sourcePosIf := vm.OriginPos[vm.PC][3]
				return &ErrOperandType{Operation: "if", Detail: fmt.Sprintf("condition '%s' is illegal: must be 0 or 1", conditionStr)}, []TokenPos{sourcePosIf}
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
				sourcePosBegin := vm.OriginPos[vm.PC][0]
				return &ErrBlockNotClosed{Label: label}, []TokenPos{sourcePosBegin}
			}

			// 存在，记录
			block := Block{
				BeginPC: beginPC,
				EndPC:   endPC,
			}

			labelNum, err := strconv.ParseInt(label, 10, 64)
			if err != nil {
				sourcePosBegin := vm.OriginPos[vm.PC][2]
				return &ErrBlockLabelInvalid{Label: label, Reason: "not a number"}, []TokenPos{sourcePosBegin}
			}

			vm.Blocks[labelNum] = block

			vm.PC = endPC + 2 // 跳过 block body 和 end，到达 end 后一条指令
			vm.Lines++

		case OP_END:
			// 理论上不可能到达这个位置
			// 因为已经在 OP_BEGIN 处理
			sourcePosEnd := vm.OriginPos[vm.PC][0]
			return &ErrUnexpectedEnd{Label: ins.ArgIdentifier[0]}, []TokenPos{sourcePosEnd}

		case OP_CALL_RS:
			/*
				OP_CALL_RS

				func_name
				int64
				int64
				int64 / bytes
				... n 个

				// 必须前面全部是 int64 后面全是 bytes

				funname
				inNum
				outNum
				i0
				i1
				o0
			*/

			// 此处个数不定，为了简化，出错位置统一定位在 call... 关键字
			sourcePosCall := vm.OriginPos[vm.PC][0]

			// ============================================================
			// Step 1: 获取函数运行要求
			// ============================================================
			funName := ins.ArgIdentifier[0]
			funReq, ok := vm.Env.Funcs[funName]
			if !ok {
				sourcePosCall := vm.OriginPos[vm.PC][0]
				return &ErrFuncNotFound{Name: funName}, []TokenPos{sourcePosCall}
			}

			// ============================================================
			// Step 2: 检查输入个数要求，检查输入输出类型要求
			//         约定：前面全部是 int64，后面全是 bytes
			//         限制：输入必须来自隐私 privmem 内存地址
			// ============================================================

			// 2a. 解析并校验输入/输出个数
			inNumStr := ins.ArgIdentifier[1]
			outNumStr := ins.ArgIdentifier[2]
			inNum, err := vm.getInt64Arg(inNumStr)
			if err != nil {
				return &ErrOperandType{Operation: "call_rs", Detail: fmt.Sprintf("invalid inNum '%s': %v", inNumStr, err)}, []TokenPos{sourcePosCall}
			}
			if inNum != int64(len(funReq.ReqInput)) {
				return &ErrFuncInputCount{Name: funName, Expected: len(funReq.ReqInput), Got: int(inNum)}, []TokenPos{sourcePosCall}
			}
			outNum, err := vm.getInt64Arg(outNumStr)
			if err != nil {
				return &ErrOperandType{Operation: "call_rs", Detail: fmt.Sprintf("invalid outNum '%s': %v", outNumStr, err)}, []TokenPos{sourcePosCall}
			}
			if outNum != int64(len(funReq.ReqOutput)) {
				return &ErrFuncOutputCount{Name: funName, Expected: len(funReq.ReqOutput), Got: int(outNum)}, []TokenPos{sourcePosCall}
			}

			// 2b. 校验 ReqInput 遵循约定：前面全是 int64，后面全是 bytes
			seenBytes := false
			for i, t := range funReq.ReqInput {
				switch t {
				case "int64":
					if seenBytes {
						return &ErrOperandType{Operation: "call_rs", Detail: fmt.Sprintf("func '%s' ReqInput[%d]: int64 must not appear after bytes", funName, i)}, []TokenPos{sourcePosCall}
					}
				case "bytes":
					seenBytes = true
				default:
					return &ErrUnsupportedInputType{FuncName: funName, InputIndex: i, InputType: t}, []TokenPos{sourcePosCall}
				}
			}

			// 2c. 校验 ReqOutput 遵循约定：前面全是 int64，后面全是 bytes
			seenBytes = false
			for i, t := range funReq.ReqOutput {
				switch t {
				case "int64":
					if seenBytes {
						return &ErrOperandType{Operation: "call_rs", Detail: fmt.Sprintf("func '%s' ReqOutput[%d]: int64 must not appear after bytes", funName, i)}, []TokenPos{sourcePosCall}
					}
				case "bytes":
					seenBytes = true
				default:
					return &ErrOperandType{Operation: "call_rs", Detail: fmt.Sprintf("func '%s' unsupported output type '%s'", funName, t)}, []TokenPos{sourcePosCall}
				}
			}

			// 2d. 收集输出变量名并校验（必须是已存在的变量，不能是字面值）
			argStrs := ins.ArgIdentifier[3:]
			outputVarNames := make([]string, outNum)
			for idx, argStr := range argStrs {
				if int64(idx) >= inNum {
					outIdx := int64(idx) - inNum
					// 不能是字面值
					if len(argStr) >= 2 && argStr[0] == '\'' && argStr[len(argStr)-1] == '\'' {
						return &ErrFuncOutputNotVar{FuncName: funName, OutputIndex: int(outIdx)}, []TokenPos{sourcePosCall}
					}
					if _, err := strconv.ParseInt(argStr, 10, 64); err == nil {

						return &ErrFuncOutputNotVar{FuncName: funName, OutputIndex: int(outIdx)}, []TokenPos{sourcePosCall}
					}
					// 变量必须存在且类型匹配
					ptr, exists := vm.Vars[argStr]
					if !exists {
						return &ErrVarNotFound{Name: argStr}, []TokenPos{sourcePosCall}
					}
					expectedType := funReq.ReqOutput[outIdx]
					if expectedType == "int64" && ptr.Kind != PrivStack {
						return &ErrOperandType{Operation: "call_rs", Detail: fmt.Sprintf("output %d for '%s' expects int64 but variable '%s' is not stack", outIdx, funName, argStr)}, []TokenPos{sourcePosCall}
					}
					if expectedType == "bytes" && ptr.Kind != PrivHeap {
						return &ErrOperandType{Operation: "call_rs", Detail: fmt.Sprintf("output %d for '%s' expects bytes but variable '%s' is not heap", outIdx, funName, argStr)}, []TokenPos{sourcePosCall}
					}
					outputVarNames[outIdx] = argStr
				}
			}

			// ============================================================
			// Step 3: 构建输入并调用函数
			// ============================================================
			ints := make([]int64, 0, inNum)
			rawBytes := make([][]byte, 0, inNum)

			for idx, argStr := range argStrs {
				if int64(idx) >= inNum {
					break
				}
				switch funReq.ReqInput[idx] {
				case "int64":
					num, err := vm.getInt64Arg(argStr)
					if err != nil {
						return &ErrOperandType{Operation: "call_rs", Detail: fmt.Sprintf("input %d for '%s': %v", idx, funName, err)}, []TokenPos{sourcePosCall}
					}
					ints = append(ints, num)
				case "bytes":
					b, err := vm.getBytesArg(argStr)
					if err != nil {
						return &ErrOperandType{Operation: "call_rs", Detail: fmt.Sprintf("input %d for '%s': %v", idx, funName, err)}, []TokenPos{sourcePosCall}
					}
					rawBytes = append(rawBytes, b)
				}
			}

			resultFilePath, err := vm.Env.CallRS(funName, ints, rawBytes)
			if err != nil {
				return &ErrOperandType{Operation: "call_rs", Detail: fmt.Sprintf("func '%s' failed: %v", funName, err)}, []TokenPos{sourcePosCall}
			}

			hexData, err := os.ReadFile(resultFilePath.PvPath)
			if err != nil {
				return &ErrOperandType{Operation: "call_rs", Detail: fmt.Sprintf("read '%s' failed: %v", resultFilePath, err)}, []TokenPos{sourcePosCall}
			}

			res, err := utils.SerializeSP1output(string(hexData))
			if err != nil {
				return &ErrOperandType{Operation: "call_rs", Detail: fmt.Sprintf("parse '%s' failed: %v", resultFilePath, err)}, []TokenPos{sourcePosCall}
			}

			// ============================================================
			// Step 4: 获取函数公共输出并解析，写入变量
			//         获取函数隐私内存哈希计算结果并解析，写入变量
			// ============================================================
			public

			vm.PC++
			vm.Lines++

		case OP_CALL_C:
			// todo 忽略
			vm.PC++
			vm.Lines++

		case OP_CALL_CPP:
			// todo 忽略
			vm.PC++
			vm.Lines++

		default:
			sourcePosCall := vm.OriginPos[vm.PC][0]
			return &ErrUnknownOpcode{Opcode: int(ins.Op)}, []TokenPos{sourcePosCall}
		}

		if vm.TraceMode {
			// generate trace step
			traceStep := vm.Mem.getDiff(vm.PreMem)
			traceStep.VarsChanges = vm.getVarsDiff()
			traceStep.BlocksChanges = vm.getBlocksDiff()
			vm.Traces[vm.Lines] = traceStep
		}
	}
	return nil, []TokenPos{}
}

// allocVar 声明一个变量并分配栈空间，返回指针; size only be needed for heap
func (vm *VM) allocVar(name string, memType PtrKind, size int64) Ptr {
	if _, exists := vm.Vars[name]; !exists {
		var ptr Ptr

		switch memType {
		case PubStack:
			ptr = vm.Mem.allocStack()
		case PubHeap:
			ptr = vm.Mem.allocHeap(size)
		case PrivStack:
			// 隐私内存由调用者提供，不支持 VM 层更新，支持在外部函数中获取副本可变性
		case PrivHeap:
			// 隐私内存由调用者提供，不支持 VM 层更新，支持在外部函数中获取副本可变性
			ptr = vm.PrivacyMem.allocHeap(size)
		default:
			panic(fmt.Sprintf("unsupported mem type: %v", memType))
		}
		vm.Vars[name] = ptr
		return ptr
	}
	return vm.Vars[name] // 已存在则返回原指针
}

// getInt64Var 获取 Stack 变量值（int64）
func (vm *VM) getInt64Var(name string) (int64, bool) {
	ptr, ok := vm.Vars[name]
	if !ok || ptr.Kind != PubStack && ptr.Kind != PrivStack {
		return 0, false
	}
	return vm.Mem.readStack(ptr)
}

// getInt64VarOnlyPriv 获取隐私内存变量值（int64）
func (vm *VM) getInt64VarOnlyPriv(name string) (int64, bool) {
	ptr, ok := vm.Vars[name]
	if !ok || ptr.Kind != PrivStack {
		return 0, false
	}
	return vm.Mem.readStack(ptr)
}

// getBytesVar 获取 Heap 变量值（[]byte）
func (vm *VM) getBytesVar(name string) ([]byte, bool) {
	ptr, ok := vm.Vars[name]
	if !ok || ptr.Kind != PubHeap && ptr.Kind != PrivHeap {
		return nil, false
	}
	return vm.Mem.readHeap(ptr)
}

// getBytesVarOnlyPriv 获取隐私内存变量值（[]byte）
func (vm *VM) getBytesVarOnlyPriv(name string) ([]byte, bool) {
	ptr, ok := vm.Vars[name]
	if !ok || ptr.Kind != PrivHeap {
		return nil, false
	}
	return vm.Mem.readHeap(ptr)
}

// updateVarByIdentifier 设置变量值
func (vm *VM) updateVarByIdentifier(name string, valIdentifier string) bool {
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
	case PubStack:
		data, ok := vm.Mem.readStack(dataPtr)
		if !ok {
			return false
		}
		return vm.Mem.writeStack(data, targetPtr)
	case PubHeap:
		data, ok := vm.Mem.readHeap(dataPtr)
		if !ok {
			return false
		}
		return vm.Mem.writeHeap(data, targetPtr)
	case PrivStack:
		// 隐私内存由调用者提供，不支持 VM 层更新，支持在外部函数中获取副本可变性
		return false
	case PrivHeap:
		// 隐私内存由调用者提供，不支持 VM 层更新，支持在外部函数中获取副本可变性
		return false
	default:
		panic(fmt.Sprintf("unsupported ptr kind: %v", dataPtr.Kind))
	}
}

func (vm *VM) updateVarBySurfaceInt64(name string, i int64) bool {
	ptr, ok := vm.Vars[name]
	if !ok {
		return false
	}
	if ptr.Kind != PubStack {
		return false
	}

	return vm.Mem.writeStack(i, ptr)
}

func (vm *VM) updateVarBySurfaceBytes(name string, data []byte) bool {
	ptr, ok := vm.Vars[name]
	if !ok {
		return false
	}
	if ptr.Kind != PubHeap {
		return false
	}

	return vm.Mem.writeHeap(data, ptr)
}

// dropVar 删除变量并释放对应的栈/堆内存
func (vm *VM) dropVar(name string) error {
	ptr, ok := vm.Vars[name]
	if !ok {
		return &ErrVarNotFound{Name: name}
	}
	vm.Mem.free(ptr)
	delete(vm.Vars, name)
	return nil
}

func (vm *VM) irUpdaterInt64(data int64, ins Instruction) error {
	// 更新 data
	ok := vm.updateVarBySurfaceInt64(ins.ArgIdentifier[1], data)
	if !ok {
		// 更新 ok 为 false
		ok := vm.updateVarBySurfaceInt64(ins.ArgIdentifier[2], 0)
		if !ok {
			// 更新 ok 时 false
			return &ErrUpdateVarBySurfaceInt64{
				VarName: ins.ArgIdentifier[2],
				Surface: 0,
			}
		}
	}
	// 更新 ok 为 true
	ok = vm.updateVarBySurfaceInt64(ins.ArgIdentifier[2], 1)
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
	ok := vm.updateVarBySurfaceBytes(ins.ArgIdentifier[1], data)
	if !ok {
		// 更新 ok 为 false
		ok := vm.updateVarBySurfaceInt64(ins.ArgIdentifier[2], 0)
		if !ok {
			// 更新 ok 时 false
			return &ErrUpdateVarBySurfaceInt64{
				VarName: ins.ArgIdentifier[2],
				Surface: 0,
			}
		}
	}
	// 更新 ok 为 true
	ok = vm.updateVarBySurfaceInt64(ins.ArgIdentifier[2], 1)
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
	_, ok := vm.getInt64Var(ins.ArgIdentifier[2])
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
	_, ok := vm.getInt64Var(ins.ArgIdentifier[2])
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
	num, ok := vm.getInt64Var(identifier)
	if !ok {
		return 0, &ErrVarNotFound{Name: identifier}
	}
	return num, nil
}

func (vm *VM) getInt64ArgExcludeSurface(identifier string) (int64, error) {
	// 检查是否为 []byte 字面量（以单引号包围）
	if len(identifier) >= 2 && identifier[0] == '\'' && identifier[len(identifier)-1] == '\'' {
		return 0, &ErrOperandType{Operation: "getInt64ArgExcludeSurface", Detail: fmt.Sprintf("var '%s' is surface []byte", identifier)}
	}
	// 尝试解析为十进制 int64
	if _, err := strconv.ParseInt(identifier, 10, 64); err == nil {
		return 0, &ErrOperandType{Operation: "getInt64ArgExcludeSurface", Detail: fmt.Sprintf("var '%s' is surface int64", identifier)}
	}
	// 从变量中读取
	num, ok := vm.getInt64Var(identifier)
	if !ok {
		return 0, &ErrVarNotFound{Name: identifier}
	}
	return num, nil
}

// getBytesArg 尝试从 identifier 中提取 []byte 值。
func (vm *VM) getBytesArg(identifier string) ([]byte, error) {
	// 检查是否为 []byte 字面量
	if len(identifier) >= 2 && identifier[0] == '\'' && identifier[len(identifier)-1] == '\'' {
		return []byte(identifier), nil
	}
	// 检查是否为 int64 字面量
	_, err := strconv.ParseInt(identifier, 10, 64)
	if err == nil {
		return nil, &ErrOperandType{Operation: "getBytesArg", Detail: fmt.Sprintf("var '%s' is int64", identifier)}
	}
	// 从变量中读取
	b, ok := vm.getBytesVar(identifier)
	if !ok {
		return nil, &ErrVarNotFound{Name: identifier}
	}
	return b, nil
}

// getBytesArgExcludeSurface 尝试从 identifier 中提取 []byte 值，排除表面变量。
func (vm *VM) getBytesArgExcludeSurface(identifier string) ([]byte, error) {
	// 检查是否为 []byte 字面量
	if len(identifier) >= 2 && identifier[0] == '\'' && identifier[len(identifier)-1] == '\'' {
		return nil, nil
	}
	// 检查是否为 int64 字面量
	_, err := strconv.ParseInt(identifier, 10, 64)
	if err == nil {
		return nil, &ErrOperandType{Operation: "getBytesArg", Detail: fmt.Sprintf("var '%s' is int64", identifier)}
	}
	// 从变量中读取
	b, ok := vm.getBytesVar(identifier)
	if !ok {
		return nil, &ErrVarNotFound{Name: identifier}
	}
	return b, nil
}

func (vm *VM) getVarsDiff() []VarsDiff {
	var diff []VarsDiff
	for name, addr := range vm.Vars {
		if addr != vm.PreVars[name] {
			diff = append(diff, VarsDiff{
				Name: name,
				Pre:  vm.PreVars[name],
				Now:  addr,
			})
			continue
		}
	}
	return diff
}

func (vm *VM) applyVarsDiff(diff []VarsDiff) {
	for _, d := range diff {
		vm.Vars[d.Name] = d.Now
	}
}

func (vm *VM) copyVars() map[string]Ptr {
	var newVars map[string]Ptr
	for name, addr := range vm.Vars {
		newVars[name] = addr
	}
	return newVars
}

func (vm *VM) getBlocksDiff() []BlocksDiff {
	var diff []BlocksDiff
	for label, block := range vm.Blocks {
		if block != vm.PreBlocks[label] {
			diff = append(diff, BlocksDiff{
				Label: label,
				Pre:   vm.PreBlocks[label],
				Now:   block,
			})
			continue
		}
	}
	return diff
}

func (vm *VM) applyBlocksDiff(diff []BlocksDiff) {
	for _, d := range diff {
		vm.Blocks[d.Label] = d.Now
	}
}

func (vm *VM) copyBlocks() map[int64]Block {
	var newBlocks map[int64]Block
	for label, block := range vm.Blocks {
		newBlocks[label] = block
	}
	return newBlocks
}
