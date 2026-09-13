package vm

import (
	"axe-cr/core/utils"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strconv"
)

type VM struct {
	Code      []Instruction                // 字节码
	OriginPos map[int64]map[int64]TokenPos // 原始位置，与 Code 里面的每一个 string Token 一一对应
	DebugMode bool                         // 调试模式
	TraceMode bool                         // 是否启用轨迹
	Env       *Environment                 // 调用外部函数

	// 公共输入
	PubIn PublicInput

	// 隐私输入
	PrivIn PrivateInput

	// 隐私输入表示
	PrivInExpr []PrivateInputExpr

	PC    int64 // 第几行了
	Lines int64 // 真实执行命令函数行数

	Mem    *Memory // 简单的堆栈模型
	PreMem *Memory // 上一帧的内存状态

	Vars    map[string]Ptr // 变量名称 -> 内存地址
	PreVars map[string]Ptr // 上一帧的变量表

	Blocks    map[int64]Block // 代码标记块
	PreBlocks map[int64]Block // 上一帧的代码标记块

	PrivacyMem   *PrivacyMemory   // 隐私内存，用于存储隐私数据，不暴露给验证者，只能直接输入 zkVM
	IsolationMem *IsolationMemory // 隔离内存，用于存储程序不可访问内容，如 report 等

	Traces map[int64]TraceStep // 执行轨迹, vm.Traces[vm.Lines]

}

// NewExecuteVM 创建一个 VM，初始化内存与变量表
func NewExecuteVM(
	code []Instruction,
	originPos map[int64]map[int64]TokenPos,
	env *Environment,
	debug bool,
	traceMode bool,
	pubIn PublicInput,
	privIn PrivateInput,
	privInExpr []PrivateInputExpr,
) *VM {
	return &VM{
		Code:         code,
		OriginPos:    originPos,
		DebugMode:    debug,
		TraceMode:    traceMode,
		Env:          env,
		PubIn:        pubIn,
		PrivIn:       privIn,
		PrivInExpr:   privInExpr,
		PC:           0,
		Lines:        0,
		Mem:          NewMemoryObj(),
		PreMem:       NewMemoryObj(),
		Vars:         make(map[string]Ptr),
		PreVars:      make(map[string]Ptr),
		Blocks:       make(map[int64]Block),
		PreBlocks:    make(map[int64]Block),
		PrivacyMem:   NewPrivacyMemoryObj(),
		IsolationMem: NewIsolationMemoryObj(),
		Traces:       make(map[int64]TraceStep),
	}
}

func (vm *VM) Run() (error, []TokenPos) {
	// 将输入映射到内存
	// 处理公共输入 Bytes
	for i, pubIn := range vm.PubIn.Bytes {
		name := fmt.Sprintf("__internal_pub_in_%d", i)
		_ = vm.allocVar(name, PubHeap, int64(len(pubIn)))
		vm.updateVarBySurfaceBytesPub(name, pubIn)
	}
	// 处理公共输入 int64
	for i, pubIn := range vm.PubIn.Int64 {
		name := fmt.Sprintf("__internal_pub_in_%d", i)
		_ = vm.allocVar(name, PubStack, 0)
		vm.updateVarBySurfaceInt64Pub(name, pubIn)
	}
	// 处理隐私输入 Bytes
	for i, privIn := range vm.PrivIn.Bytes {
		name := fmt.Sprintf("__internal_priv_in_%d", i)
		_ = vm.allocVar(name, PrivHeap, int64(len(privIn)))
		vm.updateVarBySurfaceBytesPriv(name, privIn)
	}
	// 处理隐私输入 int64
	for i, privIn := range vm.PrivIn.Int64 {
		name := fmt.Sprintf("__internal_priv_in_%d", i)
		_ = vm.allocVar(name, PrivStack, 0)
		vm.updateVarBySurfaceInt64Priv(name, privIn)
	}
	// 处理隐私输入内存表示（同时表示公共+隐私）
	for i, privIn := range vm.PrivInExpr {
		name := fmt.Sprintf("__internal_expr_%d", i)
		_ = vm.allocVar(name, PrivHeap, 32)
		vm.updateVarBySurfaceBytesPub(name, privIn.HashSum[:])
	}

	for vm.PC < int64(len(vm.Code)) {
		ins := vm.Code[vm.PC]

		if vm.DebugMode {
			// todo
		}

		if vm.TraceMode {
			vm.PreMem = vm.Mem.copy()
			vm.PreVars = vm.copyVars()
			vm.PreBlocks = vm.copyBlocks()
		}

		// 检查是否最后一行了
		if vm.PC == int64(len(vm.Code)-1) {
			// 创建输出
			// todo
		}

		// 先检查参数和类型数量匹配
		if len(ins.ArgType) != len(ins.ArgIdentifier) {
			sourcePos := vm.OriginPos[vm.PC][0]
			return &ErrTypeArgNumMismatch{GetType: len(ins.ArgType), GetArg: len(ins.ArgIdentifier)}, []TokenPos{sourcePos}
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
				OP_INPUT
			*/

			// 开头标志性语句，无实际语义

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
				var_name,
				var_name,
				...

				num -> 输出几个公共输出参数
				varname-> 使用变量来存储将要输出的数据，公共内存; 隐私表示和隐私内存在独立内存分区中，不需要手动输出
				varname
				...
			*/

			for i, varName := range ins.ArgIdentifier[1:] {
				// 判断是否为 surface
				if isSurfaceInt64(varName) {
					sourcePos := vm.OriginPos[vm.PC][int64(i+1)]
					return &ErrOutputIsSurface{VarName: varName}, []TokenPos{sourcePos}
				}
				if isSurfaceBytes(varName) {
					sourcePos := vm.OriginPos[vm.PC][int64(i+1)]
					return &ErrOutputIsSurface{VarName: varName}, []TokenPos{sourcePos}
				}

				// 判断是否为 Priv 和 Isolation Memory
				if vm.Vars[varName].Kind == PrivStack || vm.Vars[varName].Kind == PrivHeap || vm.Vars[varName].Kind == IsolationHeap || vm.Vars[varName].Kind == IsolationStack {
					sourcePos := vm.OriginPos[vm.PC][int64(i+1)]
					return &ErrOutputIsSurface{VarName: varName}, []TokenPos{sourcePos}
				}

				// 重定向为新名称，并为新名称分配变量、内存
				varNameThis := fmt.Sprintf("__internal_pub_out_%d", i)
				ptrOfOrig := vm.Vars[varName]
				var contentBytes []byte
				var contentNum int64
				var ok bool
				switch ptrOfOrig.Kind {
				case PubStack:
					contentNum, ok = vm.getInt64VarPub(varName)
					if !ok {
						sourcePos := vm.OriginPos[vm.PC][int64(i+1)]
						return &ErrVarNotFound{Name: varName}, []TokenPos{sourcePos}
					}
					ptr := vm.allocVar(varNameThis, IsolationStack, 0)
					ok = vm.IsolationMem.writeStack(contentNum, ptr)
					if !ok {
						sourcePos := vm.OriginPos[vm.PC][int64(i+1)]
						return &ErrUpdateVarBySurfaceInt64{VarName: varName}, []TokenPos{sourcePos}
					}
				case PubHeap:
					contentBytes, ok = vm.getBytesVarPub(varName)
					if !ok {
						sourcePos := vm.OriginPos[vm.PC][int64(i+1)]
						return &ErrVarNotFound{Name: varName}, []TokenPos{sourcePos}
					}
					ptr := vm.allocVar(varNameThis, IsolationHeap, int64(len(contentBytes)))
					ok = vm.IsolationMem.writeHeap(contentBytes, ptr)
					if !ok {
						sourcePos := vm.OriginPos[vm.PC][int64(i+1)]
						return &ErrUpdateVarBySurfaceBytes{VarName: varName}, []TokenPos{sourcePos}
					}
				default:
					sourcePos := vm.OriginPos[vm.PC][int64(i+1)]
					return &ErrTypeArgNumMismatch{GetType: int(ptrOfOrig.Kind), GetArg: 1}, []TokenPos{sourcePos}
				}
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
				mem_type -> 内存类型（PubStack, PrivStack, PubHeap, PrivHeap)
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
				switch vm.Vars[ins.ArgIdentifier[0]].Kind {
				case PubStack:
					ok := vm.updateVarBySurfaceInt64Pub(ins.ArgIdentifier[0], num)
					if !ok {
						sourcePos := vm.OriginPos[vm.PC][3]
						return &ErrVarNotFound{Name: ins.ArgIdentifier[0]}, []TokenPos{sourcePos}
					}
				case PrivStack:
					// Priv 数据修改只允许在 Call 分支中处理
					sourcePos := vm.OriginPos[vm.PC][3]
					return &ErrPrivacyMemUpdateInNoCallIns{ThisIns: "update"}, []TokenPos{sourcePos}
				case PubHeap:
					// PubHeap 数据修改只允许用 Heap 类型数据替代
					sourcePos := vm.OriginPos[vm.PC][3]
					return &ErrTypeMismatch{
						Expected: "PubStack",
						Got:      "PubHeap",
					}, []TokenPos{sourcePos}
				case PrivHeap:
					// Priv 数据修改只允许在 Call 分支中处理
					sourcePos := vm.OriginPos[vm.PC][3]
					return &ErrPrivacyMemUpdateInNoCallIns{ThisIns: "update"}, []TokenPos{sourcePos}
				case IsolationHeap:
					// IsolationHeap 数据修改只允许用 Heap 类型数据替代
					sourcePos := vm.OriginPos[vm.PC][3]
					return &ErrTypeMismatch{
						Expected: "IsolationHeap",
						Got:      "PubHeap",
					}, []TokenPos{sourcePos}
				case IsolationStack:
					ok := vm.IsolationMem.writeStack(num, vm.Vars[ins.ArgIdentifier[0]])
					if !ok {
						sourcePos := vm.OriginPos[vm.PC][3]
						return &ErrUpdateVarBySurfaceInt64{VarName: ins.ArgIdentifier[0]}, []TokenPos{sourcePos}
					}
				}
			case surfaceBytes:
				switch vm.Vars[ins.ArgIdentifier[0]].Kind {
				case PubStack:
					// PubStack 数据修改只允许用 Stack 类型数据替代
					sourcePos := vm.OriginPos[vm.PC][3]
					return &ErrTypeMismatch{
						Expected: "PubHeap",
						Got:      "PubStack",
					}, []TokenPos{sourcePos}
				case PrivStack:
					// Priv 数据修改只允许在 Call 分支中处理
					sourcePos := vm.OriginPos[vm.PC][3]
					return &ErrPrivacyMemUpdateInNoCallIns{ThisIns: "update"}, []TokenPos{sourcePos}
				case PubHeap:
					ok := vm.updateVarBySurfaceBytesPub(ins.ArgIdentifier[0], byt)
					if !ok {
						sourcePos := vm.OriginPos[vm.PC][3]
						return &ErrVarNotFound{Name: ins.ArgIdentifier[0]}, []TokenPos{sourcePos}
					}
				case PrivHeap:
					// Priv 数据修改只允许在 Call 分支中处理
					sourcePos := vm.OriginPos[vm.PC][3]
					return &ErrPrivacyMemUpdateInNoCallIns{ThisIns: "update"}, []TokenPos{sourcePos}
				case IsolationHeap:
					// IsolationHeap 数据修改只允许用 Heap 类型数据替代
					sourcePos := vm.OriginPos[vm.PC][3]
					return &ErrTypeMismatch{
						Expected: "IsolationHeap",
						Got:      "PubHeap",
					}, []TokenPos{sourcePos}
				case IsolationStack:
					ok := vm.IsolationMem.writeHeap(byt, vm.Vars[ins.ArgIdentifier[0]])
					if !ok {
						sourcePos := vm.OriginPos[vm.PC][3]
						return &ErrUpdateVarBySurfaceBytes{VarName: ins.ArgIdentifier[0]}, []TokenPos{sourcePos}
					}
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

			ok := vm.updateVarBySurfaceInt64Pub(ins.ArgIdentifier[2], aNum+bNum)
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

			ok := vm.updateVarBySurfaceInt64Pub(ins.ArgIdentifier[2], aNum-bNum)
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

			ok := vm.updateVarBySurfaceInt64Pub(ins.ArgIdentifier[2], aNum*bNum)
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

			ok := vm.updateVarBySurfaceInt64Pub(ins.ArgIdentifier[2], aNum/bNum)
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

			ok := vm.updateVarBySurfaceInt64Pub(ins.ArgIdentifier[2], boolean)
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

			ok := vm.updateVarBySurfaceInt64Pub(ins.ArgIdentifier[2], boolean)
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

			ok := vm.updateVarBySurfaceInt64Pub(ins.ArgIdentifier[2], boolean)
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

			// 匹配 "__internal_数字" 在字符串末尾, 这是编译器用于判定统一函数调用次数
			re := regexp.MustCompile(`__internal_(\d+)$`)
			matches := re.FindStringSubmatch(ins.ArgIdentifier[0])

			if len(matches) < 2 {
				// 没有匹配到，返回原字符串和错误
				return &ErrFuncNameInvalid{Name: ins.ArgIdentifier[0]}, []TokenPos{sourcePosCall}
			}

			// 提取数字
			funCallIdx, err := strconv.Atoi(matches[1])
			if err != nil {
				return &ErrFuncNameInvalid{Name: ins.ArgIdentifier[0]}, []TokenPos{sourcePosCall}
			}

			// 删除后缀，返回剩余部分
			funName := re.ReplaceAllString(ins.ArgIdentifier[0], "")

			// 获取函数运行要求
			req := vm.Env.Funcs[ins.ArgIdentifier[0]]
			if req == nil {
				return &ErrFuncNotFound{Name: ins.ArgIdentifier[0]}, []TokenPos{sourcePosCall}
			}

			// 检查参数是否符合前 int64 后 bytes
			nowBytes := false
			var splitPoint int
			for i, a := range ins.ArgIdentifier {
				if isSurfaceInt64(a) {
					// refuse surface int64 as param
					return &ErrOperandType{Operation: "call_rs", Detail: fmt.Sprintf("param %d is surface int64", i)}, []TokenPos{sourcePosCall}
				}
				if isSurfaceBytes(a) {
					// refuse surface bytes as param
					return &ErrOperandType{
						Operation: "call_rs",
						Detail:    fmt.Sprintf("param %d is surface bytes", i),
					}, []TokenPos{sourcePosCall}
				}

				switch nowBytes {
				case true:
					// 已经有 bytes ，必须是 bytes
					_, ok := vm.getBytesVarPriv(a)
					if !ok {
						return &ErrTypeMismatch{
							Expected: "priv int64 or bytes",
							Got:      "pub int64 or bytes",
						}, []TokenPos{sourcePosCall}
					}
					continue
				case false:
					// now ensure it is variable
					// get it via privmem visit method
					// we try int64 first
					_, ok := vm.getInt64VarPriv(a)
					if !ok {
						// try bytes
						_, ok := vm.getBytesVarPriv(a)
						if !ok {
							return &ErrTypeMismatch{
								Expected: "priv int64 or bytes",
								Got:      "pub int64 or bytes",
							}, []TokenPos{sourcePosCall}
						}

						// bytes
						splitPoint = i
						nowBytes = true
					}
					// int64
					continue
				}
			}

			// 检查参数(没有输出参数，因为它是隐私内存，不需要显式赋值)是否符合函数调用要求
			for i, r := range req.ReqInput {
				kind := vm.Vars[ins.ArgIdentifier[i+1]].Kind // 第一个参数是函数名
				if r == "bytes" {
					// require for bytes
					if kind != PrivHeap {
						return &ErrTypeMismatch{
							Expected: "priv bytes",
							Got:      fmt.Sprintf("%s", kind),
						}, []TokenPos{sourcePosCall}
					}
					continue
				} else if r == "int64" {
					// require for int64
					if kind != PrivStack {
						return &ErrTypeMismatch{
							Expected: "priv int64",
							Got:      fmt.Sprintf("%s", kind),
						}, []TokenPos{sourcePosCall}
					}
					continue
				} else {
					return &ErrTypeMismatch{
						Expected: "priv int64 or bytes",
						Got:      fmt.Sprintf("%s", kind),
					}, []TokenPos{sourcePosCall}
				}
			}

			// 解析参数
			int64Slice := make([]int64, splitPoint-1)
			for i, content := range ins.ArgIdentifier[1 : splitPoint+1] {
				num, err := strconv.ParseInt(content, 10, 64)
				if err != nil {
					return &ErrOperandType{
						Operation: "call_rs",
						Detail:    fmt.Sprintf("param %d is not int64", i),
					}, []TokenPos{sourcePosCall}
				}
				int64Slice[i-1] = num
				continue
			}
			bytesSlice := make([][]byte, len(ins.ArgIdentifier)-splitPoint)
			for i, content := range ins.ArgIdentifier[splitPoint+1:] {
				bytesSlice[i-splitPoint] = []byte(content)
				continue
			}

			// 运行函数
			result, err := vm.Env.CallRS(funName, int64Slice, bytesSlice)
			if err != nil {
				return &ErrCallFunctionFailed{
					FunName:  ins.ArgIdentifier[0],
					Detailed: fmt.Sprintf("%v", err),
				}, []TokenPos{sourcePosCall}
			}

			// 解析文件
			// 读取公共输出（公共内存）、加密输出（隐私内存）、原结果承诺（隐私内存表示，公共内存）
			pvContent, err := os.ReadFile(result.PvPath)
			if err != nil {
				return &ErrReadFileFailed{
					Path:     result.PvPath,
					Detailed: fmt.Sprintf("Failed to read pv file: %v", err),
				}, []TokenPos{sourcePosCall}
			}
			// 解析
			parseRes, err := utils.SerializeSP1output(string(pvContent))
			if err != nil {
				return &ErrParsePvFailed{
					Path:     result.PvPath,
					Detailed: fmt.Sprintf("Failed to parse pv file: %v", err),
				}, []TokenPos{sourcePosCall}
			}
			// 加载到内存中
			// 公共输出（第一次 commit）
			pubOutsFroCall := parseRes.FirstCommit
			for i, val := range pubOutsFroCall.Values {
				varNameThis := fmt.Sprintf("__internal_%s_%d_pub_out_%d", funName, funCallIdx, i)
				switch val.Type {
				case utils.TypeInt64:
					_ = vm.allocVar(varNameThis, PubStack, 0)
					n, ok := val.Value.(int64)
					if !ok {
						return &ErrCallInt64Assert{
							VarName: varNameThis,
							Surface: val.Value,
						}, []TokenPos{sourcePosCall}
					}
					ok = vm.updateVarBySurfaceInt64Pub(varNameThis, n)
					if !ok {
						return &ErrUpdateVarBySurfaceInt64{
							VarName: varNameThis,
							Surface: n,
						}, []TokenPos{sourcePosCall}
					}
				case utils.TypeBytes:
					_ = vm.allocVar(varNameThis, PubHeap, int64(len(val.Value.([]byte))))
					b, ok := val.Value.([]byte)
					if !ok {
						return &ErrCallBytesAssert{
							VarName: varNameThis,
							Surface: val.Value,
						}, []TokenPos{sourcePosCall}
					}
					ok = vm.updateVarBySurfaceBytesPub(varNameThis, b)
					if !ok {
						return &ErrUpdateVarBySurfaceBytes{
							VarName: varNameThis,
							Surface: b,
						}, []TokenPos{sourcePosCall}
					}
				default:
					return &ErrTypeMismatch{
						Expected: "pub int64 or bytes",
						Got:      fmt.Sprintf("%s", val.Type),
					}, []TokenPos{sourcePosCall}
				}
			}
			// 隐私输出（加密）
			encOutsFroCall := parseRes.SecondCommit
			for i, ct := range encOutsFroCall.EncryptedItems {
				varNameThis := fmt.Sprintf("__internal_%s_%d_priv_out_%d", funName, funCallIdx, i)
				_ = vm.allocVar(varNameThis, PrivHeap, int64(len(ct)))
				ok := vm.updateVarBySurfaceBytesPriv(varNameThis, ct)
				if !ok {
					return &ErrUpdateVarBySurfaceBytes{
						VarName: varNameThis,
						Surface: ct,
					}, []TokenPos{sourcePosCall}
				}
			}
			// 验证结果承诺（隐私内存表示，公共内存）
			exprOutsFroCall := parseRes.ThirdCommit
			for i, val := range exprOutsFroCall.HashValues {
				varNameThis := fmt.Sprintf("__internal_%s_%d_expr_%d", funName, funCallIdx, i)
				_ = vm.allocVar(varNameThis, PubHeap, 0)
				ok := vm.updateVarBySurfaceBytesPub(varNameThis, val[:])
				if !ok {
					return &ErrUpdateVarBySurfaceBytes{
						VarName: varNameThis,
						Surface: val[:],
					}, []TokenPos{sourcePosCall}
				}
			}

			// 读取 report 并加载到 IsolationHeap 中
			reportContent, err := os.ReadFile(result.RpPath)
			if err != nil {
				return &ErrReadFileFailed{
					Path:     result.RpPath,
					Detailed: fmt.Sprintf("Failed to read report file: %v", err),
				}, []TokenPos{sourcePosCall}
			}
			// 载入
			varNameThis := fmt.Sprintf("__internal_%s_%d_report", funName, funCallIdx)
			ptr := vm.allocVar(varNameThis, IsolationHeap, int64(len(reportContent)))
			vm.IsolationMem.writeHeap(reportContent, ptr)

			// 加载 proof 到 IsolationHeap 中
			proofContent, err := os.ReadFile(result.PfPath)
			if err != nil {
				return &ErrReadFileFailed{
					Path:     result.PfPath,
					Detailed: fmt.Sprintf("Failed to read proof file: %v", err),
				}, []TokenPos{sourcePosCall}
			}
			// 载入
			varNameThis = fmt.Sprintf("__internal_%s_%d_proof", funName, funCallIdx)
			ptr = vm.allocVar(varNameThis, IsolationHeap, int64(len(proofContent)))
			vm.IsolationMem.writeHeap(proofContent, ptr)

			// 加载 vkey 到 IsolationHeap 中
			vkeyContent, err := os.ReadFile(result.VkeyPath)
			if err != nil {
				return &ErrReadFileFailed{
					Path:     result.VkeyPath,
					Detailed: fmt.Sprintf("Failed to read vkey file: %v", err),
				}, []TokenPos{sourcePosCall}
			}
			// 载入
			varNameThis = fmt.Sprintf("__internal_%s_%d_vkey", funName, funCallIdx)
			ptr = vm.allocVar(varNameThis, IsolationHeap, int64(len(vkeyContent)))
			vm.IsolationMem.writeHeap(vkeyContent, ptr)

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
			ptr = vm.PrivacyMem.allocStack()
		case PrivHeap:
			// 隐私内存由调用者提供，不支持 VM 层更新，支持在外部函数中获取副本可变性
			ptr = vm.PrivacyMem.allocHeap(size)
		case IsolationHeap:
			ptr = vm.IsolationMem.allocHeap(size)
		default:
			panic(fmt.Sprintf("unsupported mem type: %v", memType))
		}
		vm.Vars[name] = ptr
		return ptr
	}
	return vm.Vars[name] // 已存在则返回原指针
}

// getInt64VarPub 获取 Stack 变量值（int64）
func (vm *VM) getInt64VarPub(name string) (int64, bool) {
	ptr, ok := vm.Vars[name]
	if !ok || ptr.Kind != PubStack {
		return 0, false
	}
	return vm.Mem.readStack(ptr)
}

// getInt64VarPriv 获取隐私内存变量值（int64）
func (vm *VM) getInt64VarPriv(name string) (int64, bool) {
	ptr, ok := vm.Vars[name]
	if !ok || ptr.Kind != PrivStack {
		return 0, false
	}
	return vm.PrivacyMem.readStack(ptr)
}

// getBytesVarPub 获取 Heap 变量值（[]byte）
func (vm *VM) getBytesVarPub(name string) ([]byte, bool) {
	ptr, ok := vm.Vars[name]
	if !ok || ptr.Kind != PubHeap {
		return nil, false
	}
	return vm.Mem.readHeap(ptr)
}

// getBytesVarPriv 获取隐私内存变量值（[]byte）
func (vm *VM) getBytesVarPriv(name string) ([]byte, bool) {
	ptr, ok := vm.Vars[name]
	if !ok || ptr.Kind != PrivHeap {
		return nil, false
	}
	return vm.PrivacyMem.readHeap(ptr)
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
		// 隐私内存不支持映射到公共内存
		return false
	case PrivHeap:
		// 隐私内存不支持映射到公共内存
		return false
	case IsolationHeap:
		// 隔离内存不支持映射到公共内存
		return false
	default:
		panic(fmt.Sprintf("unsupported ptr kind: %v", dataPtr.Kind))
	}
}

func (vm *VM) updateVarBySurfaceInt64Pub(name string, i int64) bool {
	ptr, ok := vm.Vars[name]
	if !ok {
		return false
	}
	if ptr.Kind != PubStack {
		return false
	}

	return vm.Mem.writeStack(i, ptr)
}

func (vm *VM) updateVarBySurfaceBytesPub(name string, data []byte) bool {
	ptr, ok := vm.Vars[name]
	if !ok {
		return false
	}
	if ptr.Kind != PubHeap {
		return false
	}

	return vm.Mem.writeHeap(data, ptr)
}

func (vm *VM) updateVarBySurfaceInt64Priv(name string, i int64) bool {
	ptr, ok := vm.Vars[name]
	if !ok {
		return false
	}
	if ptr.Kind != PrivStack {
		return false
	}

	return vm.PrivacyMem.writeStack(i, ptr)
}

func (vm *VM) updateVarBySurfaceBytesPriv(name string, data []byte) bool {
	ptr, ok := vm.Vars[name]
	if !ok {
		return false
	}
	if ptr.Kind != PrivHeap {
		return false
	}

	return vm.PrivacyMem.writeHeap(data, ptr)
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

func (vm *VM) arithmeticPrepare(ins Instruction, opName string) (int64, int64, error) {
	// 使用 getInt64Arg 提取前两个操作数（只接受变量，拒绝字面量）
	// todo: 计划接受字面量，但目前不支持
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
	_, ok := vm.getInt64VarPub(ins.ArgIdentifier[2])
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
	_, ok := vm.getInt64VarPub(ins.ArgIdentifier[2])
	if !ok {
		return nil, nil, &ErrArithmetic{Operation: "eq_bytes", Err: &ErrVarNotFound{Name: ins.ArgIdentifier[2]}}
	}
	return aBytes, bBytes, nil
}

// isSurfaceBytes 判断 identifier 是否为 []byte 字面量（以单引号包围）。
func isSurfaceBytes(identifier string) bool {
	return len(identifier) >= 2 && identifier[0] == '\'' && identifier[len(identifier)-1] == '\''
}

// isSurfaceInt64 判断 identifier 是否为 int64 字面量（十进制）。
func isSurfaceInt64(identifier string) bool {
	_, err := strconv.ParseInt(identifier, 10, 64)
	return err == nil
}

// getInt64Arg 尝试从 identifier 中提取 int64 值。
// 若 identifier 为 []byte 字面量（'...'），则返回错误；
// 否则尝试解析为十进制 int64，若失败则从变量中读取。
func (vm *VM) getInt64Arg(identifier string) (int64, error) {
	if isSurfaceBytes(identifier) {
		return 0, &ErrOperandType{Operation: "getInt64Arg", Detail: fmt.Sprintf("var '%s' is []byte", identifier)}
	}
	if num, err := strconv.ParseInt(identifier, 10, 64); err == nil {
		return num, nil
	}
	num, ok := vm.getInt64VarPub(identifier)
	if !ok {
		return 0, &ErrVarNotFound{Name: identifier}
	}
	return num, nil
}

// getBytesArg 尝试从 identifier 中提取 []byte 值。
func (vm *VM) getBytesArg(identifier string) ([]byte, error) {
	if isSurfaceBytes(identifier) {
		return []byte(identifier), nil
	}
	if isSurfaceInt64(identifier) {
		return nil, &ErrOperandType{Operation: "getBytesArg", Detail: fmt.Sprintf("var '%s' is int64", identifier)}
	}
	b, ok := vm.getBytesVarPub(identifier)
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
