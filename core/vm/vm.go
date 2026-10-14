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
	Codes     []Code       // 字节码
	DebugMode bool         // 调试模式
	TraceMode bool         // 是否启用轨迹
	Env       *Environment // 调用外部函数

	// 公共输入
	PubIn PublicInput

	// 隐私输入
	PrivIn PrivateInput

	// 隐私输入表示
	PrivInExpr []PrivateInputExpr

	PC    int64 // 第几行了
	Lines int64 // 真实执行命令函数行数

	Mem    *PubMemory // 简单的堆栈模型
	PreMem *PubMemory // 上一帧的内存状态

	Vars    map[string]Ptr // 变量名称 -> 内存地址
	PreVars map[string]Ptr // 上一帧的变量表

	Blocks    map[int64]Block // 代码标记块
	PreBlocks map[int64]Block // 上一帧的代码标记块

	PrivacyMem  *PrivacyMemory  // 隐私内存，用于存储隐私数据，不暴露给验证者，只能直接输入 zkVM
	IsolatedMem *IsolatedMemory // 隔离内存，用于存储程序不可访问内容，如 report 等

	Traces map[int64]TraceStep // 执行轨迹, vm.Traces[vm.Lines]

}

// NewExecuteVM 创建一个 VM，初始化内存与变量表
func NewExecuteVM(
	code []Code,
	env *Environment,
	debug bool,
	traceMode bool,
	pubIn PublicInput,
	privIn PrivateInput,
	privInExpr []PrivateInputExpr,
) *VM {
	return &VM{
		Codes:       code,
		DebugMode:   debug,
		TraceMode:   traceMode,
		Env:         env,
		PubIn:       pubIn,
		PrivIn:      privIn,
		PrivInExpr:  privInExpr,
		PC:          0,
		Lines:       0,
		Mem:         NewMemoryObj(),
		PreMem:      NewMemoryObj(),
		Vars:        make(map[string]Ptr),
		PreVars:     make(map[string]Ptr),
		Blocks:      make(map[int64]Block),
		PreBlocks:   make(map[int64]Block),
		PrivacyMem:  NewPrivacyMemoryObj(),
		IsolatedMem: NewIsolationMemoryObj(),
		Traces:      make(map[int64]TraceStep),
	}
}

func (vm *VM) Run() (error, []TokPos) {
	// 将输入映射到内存

	// 处理公共输入 Bytes
	for i, pubIn := range vm.PubIn.Bytes {
		name := fmt.Sprintf("%s%d", INTERNAL_PUB_VAR_IN_SIGN, i)
		_ = vm.allocVar(name, PUB_HEAP, int64(len(pubIn)))
		vm.updateVarBySurfaceBytesPub(name, pubIn)
	}
	// 处理公共输入 int64
	for i, pubIn := range vm.PubIn.Int64 {
		name := fmt.Sprintf("%s%d", INTERNAL_PUB_VAR_IN_SIGN, i)
		_ = vm.allocVar(name, PUB_STACK, 0)
		vm.updateVarBySurfaceInt64Pub(name, pubIn)
	}

	// 处理隐私输入 Bytes
	for i, privIn := range vm.PrivIn.Bytes {
		name := fmt.Sprintf("%s%d", INTERNAL_PRIV_VAR_IN_SIGN, i)
		_ = vm.allocVar(name, PRIV_HEAP, int64(len(privIn)))
		vm.updateVarBySurfaceBytesPriv(name, privIn)
	}
	// 处理隐私输入 int64
	for i, privIn := range vm.PrivIn.Int64 {
		name := fmt.Sprintf("%s%d", INTERNAL_PRIV_VAR_IN_SIGN, i)
		_ = vm.allocVar(name, PRIV_STACK, 0)
		vm.updateVarBySurfaceInt64Priv(name, privIn)
	}

	// 处理隐私输入内存表示（同时表示公共+隐私）
	for i, privIn := range vm.PrivInExpr {
		name := fmt.Sprintf("%s%d", INTERNAL_EXPR_SIGN, i)
		_ = vm.allocVar(name, PRIV_HEAP, 32)
		vm.updateVarBySurfaceBytesPub(name, privIn.HashSum[:])
	}

	for vm.PC < int64(len(vm.Codes)) {
		ins := vm.Codes[vm.PC]

		if vm.DebugMode {
			// todo
		}

		if vm.TraceMode {
			vm.PreMem = vm.Mem.copy()
			vm.PreVars = vm.copyVars()
			vm.PreBlocks = vm.copyBlocks()
		}

		// 检查是否最后一行了
		if vm.PC == int64(len(vm.Codes)-1) {
			// 创建输出
			// todo
		}

		// 先检查 Arg Type, Content, IsSurface 数量匹配
		lenArgTypeAsBaseLine := len(ins.ArgType)
		lenArgCont := len(ins.ArgContent)
		if lenArgTypeAsBaseLine != lenArgCont {
			var errTokPos []TokPos
			for _, tokPos := range ins.ArgType {
				errTokPos = append(errTokPos, tokPos.TokenPos)
			}
			for _, tokPos := range ins.ArgContent {
				errTokPos = append(errTokPos, tokPos.TokenPos)
			}
			return &ErrTypeContentLenMismatch{
				ArgTypeHave: lenArgTypeAsBaseLine,
				ArgContHave: lenArgCont,
			}, errTokPos
		}
		lenArgIsSurface := len(ins.IsArgSurface)
		if lenArgTypeAsBaseLine != lenArgIsSurface {
			var errTokPos []TokPos
			for _, tokPos := range ins.ArgType {
				errTokPos = append(errTokPos, tokPos.TokenPos)
			}
			for _, tokPos := range ins.IsArgSurface {
				errTokPos = append(errTokPos, tokPos.TokenPos)
			}
			return &ErrTypeIsSurfaceLenMismatch{
				ArgTypeHave:      lenArgTypeAsBaseLine,
				ArgIsSurfaceHave: lenArgIsSurface,
			}, errTokPos
		}

		// todo : fmt 和 types 由 compiler 保证
		switch ins.OpCode.Code {
		case OP_READ:
			/*
				Code{
					OpCode:       OpWrap{OP_READ,TokPos{..}},
					ArgType:      []ArgTypeWrap{..},
					IsArgSurface: []IsSurfaceWrap{..},
					ArgContent:   []ArgContWrap{..},
				}
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
				Code{
					OpCode:       OpWrap{OP_INPUT,TokPos{..}},
					ArgType:      nil,
					IsArgSurface: nil,
					ArgContent:   nil,
				}
			*/

			// 开头标志性语句，无实际语义

			vm.PC++
			vm.Lines++

		case OP_WRITE:
			/*
				Code{
					OpCode:       OpWrap{OP_WRITE,TokPos{..}},
					ArgType:      []ArgTypeWrap{..},
					IsArgSurface: []IsSurfaceWrap{..},
					ArgContent:   []ArgContWrap{..},
				}
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
				Code{
					OpCode:       OpWrap{OP_OUTPUT,TokPos{..}},
					ArgType:      []ArgTypeWrap{..},
					IsArgSurface: []IsSurfaceWrap{..},
					ArgContent:   []ArgContWrap{..},
				}
			*/

			// 只允许返回 Public 内存（Privacy 内存都是隐私输入的内容，而 Isolation 内存是运行时 VM 元数据）
			// 字面值也不允许返回，因为这是提前知道的，根本不需要计算，属于程序错误，由编译期保证
			// 类型安全也在编译期保证
			// OUTPUT 是最后一句，编译期保证

			for idx, contVar := range ins.ArgContent {
				// create isolation memory
				nicknameVar := fmt.Sprintf("%v%d", INTERNAL_PUB_VAR_OUT_SIGN, idx)
				memType := ins.ArgType[idx].Type
				switch memType {
				case PUB_I64:
					// 位于 PUB_STACK 上，转储到 ISOLATED_STACK
					cont, ok := vm.getInt64VarPub(contVar.ArgContent)
					if !ok {
						return &ErrVarNotFound{Name: contVar.ArgContent}, []TokPos{contVar.TokenPos}
					}
					ptr := vm.allocVar(nicknameVar, ISOLATED_STACK, 0)
					// 写入
					ok = vm.IsolatedMem.writeStack(cont, ptr)
					if !ok {
						return &ErrIsolationMemInaccessible{nicknameVar}, []TokPos{contVar.TokenPos}
					}
				case PUB_BYTES:
					// 位于 PUB_HEAP 上，转储到 ISOLATED_STACK
					cont, ok := vm.getBytesVarPub(contVar.ArgContent)
					if !ok {
						return &ErrVarNotFound{Name: contVar.ArgContent}, []TokPos{contVar.TokenPos}
					}
					ptr := vm.allocVar(nicknameVar, ISOLATED_HEAP, int64(len(cont)))
					// 写入
					ok = vm.IsolatedMem.writeHeap(cont, ptr)
					if !ok {
						return &ErrIsolationMemInaccessible{nicknameVar}, []TokPos{contVar.TokenPos}
					}
				case PRIV_I64:
					// 不可返回（若编译期无误，这里永远不会到达，写这里违背了编译期严格确定类型和内存问题的原则，但起到消除警告的作用）
					return &ErrOutputIsPrivMem{
						VarName:   contVar.ArgContent,
						ReturnIdx: idx,
					}, []TokPos{contVar.TokenPos}
				case PRIV_BYTES:
					// 不可返回（若编译期无误，这里永远不会到达，写这里违背了编译期严格确定类型和内存问题的原则，但起到消除警告的作用）
					return &ErrOutputIsPrivMem{
						VarName:   contVar.ArgContent,
						ReturnIdx: idx,
					}, []TokPos{contVar.TokenPos}
				case ISO_I64:
					// 作为返回类型，在返回内存中，不能返回自身（若编译期无误，这里永远不会到达，写这里违背了编译期严格确定类型和内存问题的原则，但起到消除警告的作用）
					return &ErrOutputIsIsoMem{
						VarName:   contVar.ArgContent,
						ReturnIdx: idx,
					}, []TokPos{contVar.TokenPos}
				case ISO_BYTES:
					// 作为返回类型，在返回内存中，不能返回自身（若编译期无误，这里永远不会到达，写这里违背了编译期严格确定类型和内存问题的原则，但起到消除警告的作用）
					return &ErrOutputIsIsoMem{
						VarName:   contVar.ArgContent,
						ReturnIdx: idx,
					}, []TokPos{contVar.TokenPos}
				default:
					return &ErrUnsupportedMemType{MemType: PtrKind(memType)}, []TokPos{contVar.TokenPos}
				}
			}

			vm.PC++
			vm.Lines++

		case OP_ALLOC:
			/*
				ArgType:
					var_name
					int64
					int64

				ArgContent
					varname _> 变量名称
					mem_type -> 内存类型（PubStack, PrivStack, PubHeap, PrivHeap, IsolatedStack, IsolatedStack)
					size -> 分配内存大小 (如果为 heap）

				Code{
					OpCode:       OpWrap{OP_OUTPUT,TokPos{..}},
					ArgType:      []ArgTypeWrap{..},
					IsArgSurface: []IsSurfaceWrap{..},
					ArgContent:   []ArgContWrap{..},
				}
			*/

			// 编译期保证：ins.IsArgSurface[0].IsSurface==0（变量名不为字面值）
			varName := ins.ArgContent[0].ArgContent

			// 编译期保证：ins.ArgContent[1].ArgContent 为合法类型
			// 编译期保证：ins.IsArgSurface[1~2].IsSurface==1（MemType 和 Size 为字面值）
			memType := PtrKind(ins.ArgContent[1].ArgContentInt64IfSurface)
			size := ins.ArgContent[2].ArgContentInt64IfSurface

			vm.allocVar(varName, memType, size)

			vm.PC++
			vm.Lines++

		case OP_UPDATE:
			/*
				Code{
					OpCode:       OpWrap{OP_UPDATE,TokPos{..}},
					ArgType:      []ArgTypeWrap{..}, // 这里两个类型应该相等
					IsArgSurface: []IsSurfaceWrap{..}, // 第一个应该是变量名，第二个可以是表面值
					ArgContent:   []ArgContWrap{..},
				}
			*/

			varName := ins.ArgContent[0].ArgContent
			varType := ins.ArgType[0].Type
			switch ins.IsArgSurface[1].IsSurface {
			case true:
				// targetVal is surface val
				switch varType {
				case PUB_I64:
					targetVal := ins.ArgContent[1].ArgContentInt64IfSurface
					ok := vm.updateVarBySurfaceInt64Pub(varName, targetVal)
					if !ok {
						return &ErrUpdateVarBySurfaceInt64{
							VarName: varName,
							Surface: targetVal,
						}, []TokPos{ins.ArgContent[0].TokenPos}
					}
				case PUB_BYTES:
					targetVal := ins.ArgContent[1].ArgContentBytesIfSurface
					ok := vm.updateVarBySurfaceBytesPub(varName, targetVal)
					if !ok {
						return &ErrUpdateVarBySurfaceBytes{
							VarName: varName,
							Surface: targetVal,
						}, []TokPos{ins.ArgContent[0].TokenPos}
					}
				case PRIV_I64:
					// priv mem only write at the program started to load to priv input
					// （若编译期无误，这里永远不会到达，写这里违背了编译期严格确定类型和内存问题的原则，但起到消除警告的作用）
					return &ErrUpdatePrivVarBySurface{
						VarName: varName,
					}, []TokPos{ins.ArgContent[0].TokenPos}
				case PRIV_BYTES:
					// priv mem only write at the program started to load to priv input
					// （若编译期无误，这里永远不会到达，写这里违背了编译期严格确定类型和内存问题的原则，但起到消除警告的作用）
					return &ErrUpdatePrivVarBySurface{
						VarName: varName,
					}, []TokPos{ins.ArgContent[0].TokenPos}
				case ISO_I64:
					// iso mem only write at rust / c / cpp function call with zkvm to storage vm runtime data
					// （若编译期无误，这里永远不会到达，写这里违背了编译期严格确定类型和内存问题的原则，但起到消除警告的作用）
					return &ErrIsolationMemInaccessible{VarName: varName}, []TokPos{ins.ArgContent[0].TokenPos}
				case ISO_BYTES:
					// iso mem only write at rust / c / cpp function call with zkvm to storage vm runtime data
					// （若编译期无误，这里永远不会到达，写这里违背了编译期严格确定类型和内存问题的原则，但起到消除警告的作用）
					return &ErrIsolationMemInaccessible{VarName: varName}, []TokPos{ins.ArgContent[0].TokenPos}
				default:
					return &ErrUnsupportedMemType{MemType: PtrKind(varType)}, []TokPos{ins.ArgContent[0].TokenPos}
				}
			case false:
				// targetVal is not surface val, but a var
				targetVarName := ins.ArgContent[1].ArgContent
				targetVarType := ins.ArgType[1].Type
				switch varType {
				case PUB_I64:
					switch targetVarType {
					case PUB_I64:
						targetVal, ok := vm.getInt64VarPub(targetVarName)
						if !ok {
							return &ErrVarNotFound{Name: targetVarName}, []TokPos{ins.ArgContent[0].TokenPos}
						}
						ok = vm.updateVarBySurfaceInt64Pub(varName, targetVal)
						if !ok {
							return &ErrUpdateVarBySurfaceInt64{
								VarName: varName,
								Surface: targetVal,
							}, []TokPos{ins.ArgContent[0].TokenPos}
						}
					case PUB_BYTES:
						return &ErrTypeMismatch{
							Expected: "PUB_I64",
							Got:      "PUB_BYTES",
						}, []TokPos{ins.ArgContent[1].TokenPos}
					case PRIV_I64:
						return &ErrTypeMismatch{
							Expected: "PUB_I64",
							Got:      "PRIV_I64",
						}, []TokPos{ins.ArgContent[1].TokenPos}
					case PRIV_BYTES:
						return &ErrTypeMismatch{
							Expected: "PUB_I64",
							Got:      "PRIV_BYTES",
						}, []TokPos{ins.ArgContent[1].TokenPos}
					case ISO_I64:
						return &ErrTypeMismatch{
							Expected: "PUB_I64",
							Got:      "ISO_I64",
						}, []TokPos{ins.ArgContent[1].TokenPos}
					case ISO_BYTES:
						return &ErrTypeMismatch{
							Expected: "PUB_I64",
							Got:      "ISO_BYTES",
						}, []TokPos{ins.ArgContent[1].TokenPos}
					default:
						return &ErrUnsupportedMemType{MemType: PtrKind(targetVarType)}, []TokPos{ins.ArgContent[1].TokenPos}
					}

				case PUB_BYTES:
					switch ins.ArgType[1].Type {
					case PUB_I64:
						return &ErrTypeMismatch{
							Expected: "PUB_BYTES",
							Got:      "PUB_I64",
						}, []TokPos{ins.ArgContent[1].TokenPos}
					case PUB_BYTES:
						targetVal, ok := vm.getBytesVarPub(targetVarName)
						if !ok {
							return &ErrVarNotFound{Name: targetVarName}, []TokPos{ins.ArgContent[1].TokenPos}
						}
						ok = vm.updateVarBySurfaceBytesPub(varName, targetVal)
						if !ok {
							return &ErrUpdateVarBySurfaceBytes{
								VarName: varName,
								Surface: targetVal,
							}, []TokPos{ins.ArgContent[1].TokenPos}
						}
					case PRIV_I64:
						return &ErrTypeMismatch{
							Expected: "PUB_BYTES",
							Got:      "PRIV_I64",
						}, []TokPos{ins.ArgContent[1].TokenPos}
					case PRIV_BYTES:
						return &ErrTypeMismatch{
							Expected: "PUB_BYTES",
							Got:      "PRIV_BYTES",
						}, []TokPos{ins.ArgContent[1].TokenPos}
					case ISO_I64:
						return &ErrTypeMismatch{
							Expected: "PUB_BYTES",
							Got:      "ISO_I64",
						}, []TokPos{ins.ArgContent[1].TokenPos}
					case ISO_BYTES:
						return &ErrTypeMismatch{
							Expected: "PUB_BYTES",
							Got:      "ISO_BYTES",
						}, []TokPos{ins.ArgContent[1].TokenPos}
					}

				case PRIV_I64:
					// priv mem only write at the program started to load to priv input
					// （若编译期无误，这里永远不会到达，写这里违背了编译期严格确定类型和内存问题的原则，但起到消除警告的作用）
					return &ErrUpdatePrivVarBySurface{
						VarName: varName,
					}, []TokPos{ins.ArgContent[0].TokenPos}
				case PRIV_BYTES:
					// priv mem only write at the program started to load to priv input
					// （若编译期无误，这里永远不会到达，写这里违背了编译期严格确定类型和内存问题的原则，但起到消除警告的作用）
					return &ErrUpdatePrivVarBySurface{
						VarName: varName,
					}, []TokPos{ins.ArgContent[0].TokenPos}
				case ISO_I64:
					// iso mem only write at rust / c / cpp function call with zkvm to storage vm runtime data
					// （若编译期无误，这里永远不会到达，写这里违背了编译期严格确定类型和内存问题的原则，但起到消除警告的作用）
					return &ErrIsolationMemInaccessible{VarName: varName}, []TokPos{ins.ArgContent[0].TokenPos}
				case ISO_BYTES:
					// iso mem only write at rust / c / cpp function call with zkvm to storage vm runtime data
					// （若编译期无误，这里永远不会到达，写这里违背了编译期严格确定类型和内存问题的原则，但起到消除警告的作用）
					return &ErrIsolationMemInaccessible{VarName: varName}, []TokPos{ins.ArgContent[0].TokenPos}
				default:
					return &ErrUnsupportedMemType{MemType: PtrKind(varType)}, []TokPos{ins.ArgContent[0].TokenPos}
				}
			}

			vm.PC++
			vm.Lines++

		case OP_DROP:
			/*
				Code{
					OpCode:       OpWrap{OP_UPDATE,TokPos{..}},
					ArgType:      []ArgTypeWrap{..}, // 这里类型任意
					IsArgSurface: []IsSurfaceWrap{..}, // 不应该是字面值，由编译期保证
					ArgContent:   []ArgContWrap{..}, // 任意存在变量
				}
			*/

			varName := ins.ArgContent[0].ArgContent

			err := vm.dropVar(varName)
			if err != nil {
				return &ErrVarDrop{Name: varName}, []TokPos{ins.ArgContent[0].TokenPos}
			}

			vm.PC++
			vm.Lines++

		case OP_ADD:
			/*
				OP_ADD

				ArgType:
					PUB_I64
					PUB_I64

				Code{
					OpCode:       OpWrap{OP_ADD,TokPos{..}},
					ArgType:      []ArgTypeWrap{..}, // 类型都为 PUB_I64
					IsArgSurface: []IsSurfaceWrap{..}, // 第三个一定为非字面值，编译期保证
					ArgContent:   []ArgContWrap{..}, // 第三个一定为变量
				}
			*/

			var aNum, bNum, res int64
			var err error
			var errPos []TokPos
			var ok bool

			// 先提取 a
			aNum, err, errPos, ok = vm.extractArithmeticToken(ins, 0)
			if !ok {
				return err, errPos
			}

			// 再提取 b
			bNum, err, errPos, ok = vm.extractArithmeticToken(ins, 1)
			if !ok {
				return err, errPos
			}

			res = aNum + bNum

			resVarName := ins.ArgContent[2].ArgContent

			ok = vm.updateVarBySurfaceInt64Pub(resVarName, res)
			if !ok {
				return &ErrVarNotFound{Name: resVarName}, []TokPos{ins.ArgContent[2].TokenPos}
			}

			vm.PC++
			vm.Lines++

		case OP_SUB:
			/*
				OP_SUB

				ArgType:
					PUB_I64
					PUB_I64

				Code{
					OpCode:       OpWrap{OP_SUB,TokPos{..}},
					ArgType:      []ArgTypeWrap{..}, // 类型都为 PUB_I64
					IsArgSurface: []IsSurfaceWrap{..}, // 第三个一定为非字面值，编译期保证
					ArgContent:   []ArgContWrap{..}, // 第三个一定为变量
				}
			*/

			var aNum, bNum, res int64
			var err error
			var errPos []TokPos
			var ok bool

			// 先提取 a
			aNum, err, errPos, ok = vm.extractArithmeticToken(ins, 0)
			if !ok {
				return err, errPos
			}

			// 再提取 b
			bNum, err, errPos, ok = vm.extractArithmeticToken(ins, 1)
			if !ok {
				return err, errPos
			}

			res = aNum - bNum

			resVarName := ins.ArgContent[2].ArgContent

			ok = vm.updateVarBySurfaceInt64Pub(resVarName, res)
			if !ok {
				return &ErrVarNotFound{Name: resVarName}, []TokPos{ins.ArgContent[2].TokenPos}
			}

			vm.PC++
			vm.Lines++

		case OP_MUL:
			/*
				OP_MUL

				ArgType:
					PUB_I64
					PUB_I64

				Code{
					OpCode:       OpWrap{OP_MUL,TokPos{..}},
					ArgType:      []ArgTypeWrap{..}, // 类型都为 PUB_I64
					IsArgSurface: []IsSurfaceWrap{..}, // 第三个一定为非字面值，编译期保证
					ArgContent:   []ArgContWrap{..}, // 第三个一定为变量
				}
			*/

			var aNum, bNum, res int64
			var err error
			var errPos []TokPos
			var ok bool

			// 先提取 a
			aNum, err, errPos, ok = vm.extractArithmeticToken(ins, 0)
			if !ok {
				return err, errPos
			}

			// 再提取 b
			bNum, err, errPos, ok = vm.extractArithmeticToken(ins, 1)
			if !ok {
				return err, errPos
			}

			res = aNum * bNum

			resVarName := ins.ArgContent[2].ArgContent

			ok = vm.updateVarBySurfaceInt64Pub(resVarName, res)
			if !ok {
				return &ErrVarNotFound{Name: resVarName}, []TokPos{ins.ArgContent[2].TokenPos}
			}

			vm.PC++
			vm.Lines++

		case OP_DIV:
			/*
				OP_DIV

				ArgType:
					PUB_I64
					PUB_I64

				Code{
					OpCode:       OpWrap{OP_DIV,TokPos{..}},
					ArgType:      []ArgTypeWrap{..}, // 类型都为 PUB_I64
					IsArgSurface: []IsSurfaceWrap{..}, // 第三个一定为非字面值，编译期保证
					ArgContent:   []ArgContWrap{..}, // 第三个一定为变量
				}
			*/

			var aNum, bNum, res int64
			var err error
			var errPos []TokPos
			var ok bool

			// 先提取 a
			aNum, err, errPos, ok = vm.extractArithmeticToken(ins, 0)
			if !ok {
				return err, errPos
			}

			// 再提取 b
			bNum, err, errPos, ok = vm.extractArithmeticToken(ins, 1)
			if !ok {
				return err, errPos
			}

			res = aNum / bNum

			resVarName := ins.ArgContent[2].ArgContent

			ok = vm.updateVarBySurfaceInt64Pub(resVarName, res)
			if !ok {
				return &ErrVarNotFound{Name: resVarName}, []TokPos{ins.ArgContent[2].TokenPos}
			}

			vm.PC++
			vm.Lines++

		case OP_EQ_INT:
			/*
				OP_EQ_INT

				ArgType:
					PUB_I64
					PUB_I64

				Code{
					OpCode:       OpWrap{OP_EQ_INT,TokPos{..}},
					ArgType:      []ArgTypeWrap{..}, // 类型都为 PUB_I64
					IsArgSurface: []IsSurfaceWrap{..}, // 第三个一定为非字面值，编译期保证
					ArgContent:   []ArgContWrap{..}, // 第三个一定为变量
				}
			*/

			var aNum, bNum, res int64
			var err error
			var errPos []TokPos
			var ok bool

			// 先提取 a
			aNum, err, errPos, ok = vm.extractArithmeticToken(ins, 0)
			if !ok {
				return err, errPos
			}

			// 再提取 b
			bNum, err, errPos, ok = vm.extractArithmeticToken(ins, 1)
			if !ok {
				return err, errPos
			}

			resBool := aNum == bNum
			switch resBool {
			case true:
				res = 1
			case false:
				res = 0
			}

			resVarName := ins.ArgContent[2].ArgContent

			ok = vm.updateVarBySurfaceInt64Pub(resVarName, res)
			if !ok {
				return &ErrVarNotFound{Name: resVarName}, []TokPos{ins.ArgContent[2].TokenPos}
			}

			vm.PC++
			vm.Lines++

		case OP_EQ_BYTES:
			/*
				OP_EQ_BYTES

				ArgType:
					PUB_I64
					PUB_I64

				Code{
					OpCode:       OpWrap{OP_EQ_BYTES,TokPos{..}},
					ArgType:      []ArgTypeWrap{..}, // 前两个类型为 PUB_BYTES, 第三个是 PUB_I64
					IsArgSurface: []IsSurfaceWrap{..}, // 第三个一定为非字面值，编译期保证
					ArgContent:   []ArgContWrap{..}, // 第三个一定为变量
				}
			*/

			var aBytes, bBytes []byte
			var res int64
			var err error
			var errPos []TokPos
			var ok bool

			// 先提取 a
			aBytes, err, errPos, ok = vm.extractArithmeticTokenBytes(ins, 0)
			if !ok {
				return err, errPos
			}

			// 再提取 b
			bBytes, err, errPos, ok = vm.extractArithmeticTokenBytes(ins, 1)
			if !ok {
				return err, errPos
			}

			resBool := bytes.Equal(aBytes, bBytes)
			switch resBool {
			case true:
				res = 1
			case false:
				res = 0
			}

			resVarName := ins.ArgContent[2].ArgContent

			ok = vm.updateVarBySurfaceInt64Pub(resVarName, res)
			if !ok {
				return &ErrVarNotFound{Name: resVarName}, []TokPos{ins.ArgContent[2].TokenPos}
			}

			vm.PC++
			vm.Lines++

		case OP_LARGE_INT:
			/*
				OP_LARGE_INT

				ArgType:
					PUB_I64
					PUB_I64

				Code{
					OpCode:       OpWrap{OP_LARGE_INT,TokPos{..}},
					ArgType:      []ArgTypeWrap{..}, // 类型都为 PUB_I64
					IsArgSurface: []IsSurfaceWrap{..}, // 第三个一定为非字面值，编译期保证
					ArgContent:   []ArgContWrap{..}, // 第三个一定为变量
				}
			*/

			var aNum, bNum, res int64
			var err error
			var errPos []TokPos
			var ok bool

			// 先提取 a
			aNum, err, errPos, ok = vm.extractArithmeticToken(ins, 0)
			if !ok {
				return err, errPos
			}

			// 再提取 b
			bNum, err, errPos, ok = vm.extractArithmeticToken(ins, 1)
			if !ok {
				return err, errPos
			}

			resBool := aNum == bNum
			switch resBool {
			case true:
				res = 1
			case false:
				res = 0
			}

			resVarName := ins.ArgContent[2].ArgContent

			ok = vm.updateVarBySurfaceInt64Pub(resVarName, res)
			if !ok {
				return &ErrVarNotFound{Name: resVarName}, []TokPos{ins.ArgContent[2].TokenPos}
			}

			vm.PC++
			vm.Lines++

		case OP_JMP:
			/*
				OP_JMP

				ArgType
					PUB_I64

				ArgContent
					LabelSign

				Code{
						OpCode:       OpWrap{OP_JMP,TokPos{..}},
						ArgType:      []ArgTypeWrap{..}, // 一个 PUB_I64
						IsArgSurface: []IsSurfaceWrap{..}, // 字面值，编译期保证
						ArgContent:   []ArgContWrap{..},
				}
			*/

			labelSign := ins.ArgContent[0].ArgContentInt64IfSurface

			tgtBlock, ok := vm.Blocks[labelSign]
			if !ok {
				return &ErrLabelNotFound{Label: labelSign}, []TokPos{ins.ArgContent[0].TokenPos}
			}
			if &tgtBlock.BeginPC == nil {
				return &ErrBlockLabelInvalid{
					Label:  labelSign,
					Reason: "begin pc ptr is nil",
				}, []TokPos{ins.ArgContent[0].TokenPos}
			}
			if &tgtBlock.EndPC == nil {
				return &ErrBlockNotClosed{
					Label: labelSign,
				}, []TokPos{
					ins.ArgContent[0].TokenPos,
				}
			}
			if tgtBlock.EndPC < tgtBlock.BeginPC {
				// （若编译期无误，这里永远不会到达）
				return &ErrBlockLabelInvalid{
					Label:  labelSign,
					Reason: "end pc even comes before begin pc, this should be found in compiling",
				}, []TokPos{ins.ArgContent[0].TokenPos}
			}

			vm.PC = tgtBlock.BeginPC
			vm.Lines++

		case OP_IF:
			/*
				OP_IF

				ArgType
					PUB_I64
					PUB_I64

				ArgContent
					ConditionBool
					LabelSign

				Code{
					OpCode:       OpWrap{OP_IF,TokPos{..}},
					ArgType:      []ArgTypeWrap{..}, // 两个 PUB_I64
					IsArgSurface: []IsSurfaceWrap{..}, //第一个不做要求，第二个字面值，编译期保证
					ArgContent:   []ArgContWrap{..},
				}
			*/

			var condition bool
			rawConditionType := ins.ArgType[0].Type
			switch rawConditionType {
			case PUB_I64:
				switch ins.IsArgSurface[0].IsSurface {
				case true:
					conditionInt64 := ins.ArgContent[0].ArgContentInt64IfSurface
					switch conditionInt64 {
					case 0:
						condition = false
					case 1:
						condition = true
					default:
						return &ErrIfConditionIsNotBoolInt64{Get: conditionInt64}, []TokPos{ins.ArgContent[0].TokenPos}
					}
				case false:
					conditionVarName := ins.ArgContent[0].ArgContent
					conditionInt64, ok := vm.getInt64VarPub(conditionVarName)
					if !ok {
						return &ErrVarNotFound{Name: conditionVarName}, []TokPos{ins.ArgContent[0].TokenPos}
					}
					switch conditionInt64 {
					case 0:
						condition = false
					case 1:
						condition = true
					default:
						return &ErrIfConditionIsNotBoolInt64{Get: conditionInt64}, []TokPos{ins.ArgContent[0].TokenPos}
					}
				}
			case PUB_BYTES:
				return &ErrTypeMismatch{Expected: "PUB_I64", Got: "PUB_BYTES"}, []TokPos{ins.ArgType[0].TokenPos}
			case PRIV_I64:
				return &ErrTypeMismatch{Expected: "PUB_I64", Got: "PRIV_I64"}, []TokPos{ins.ArgType[0].TokenPos}
			case PRIV_BYTES:
				return &ErrTypeMismatch{Expected: "PUB_I64", Got: "PRIV_BYTES"}, []TokPos{ins.ArgType[0].TokenPos}
			case ISO_I64:
				return &ErrTypeMismatch{Expected: "PUB_I64", Got: "ISO_I64"}, []TokPos{ins.ArgType[0].TokenPos}
			case ISO_BYTES:
				return &ErrTypeMismatch{Expected: "PUB_I64", Got: "ISO_BYTES"}, []TokPos{ins.ArgType[0].TokenPos}
			default:
				return &ErrUnsupportedMemType{MemType: PtrKind(rawConditionType)}, []TokPos{ins.ArgType[0].TokenPos}
			}

			switch condition {
			case false:
				// false, 跳过 label
				vm.PC++
				vm.Lines++
			case true:
				labelSign := ins.ArgContent[1].ArgContentInt64IfSurface

				tgtBlock, ok := vm.Blocks[labelSign]
				if !ok {
					return &ErrLabelNotFound{Label: labelSign}, []TokPos{ins.ArgContent[0].TokenPos}
				}
				if &tgtBlock.BeginPC == nil {
					return &ErrBlockLabelInvalid{
						Label:  labelSign,
						Reason: "begin pc ptr is nil",
					}, []TokPos{ins.ArgContent[0].TokenPos}
				}
				if &tgtBlock.EndPC == nil {
					return &ErrBlockNotClosed{
						Label: labelSign,
					}, []TokPos{
						ins.ArgContent[0].TokenPos,
					}
				}
				if tgtBlock.EndPC < tgtBlock.BeginPC {
					// （若编译期无误，这里永远不会到达）
					return &ErrBlockLabelInvalid{
						Label:  labelSign,
						Reason: "end pc even comes before begin pc, this should be found in compiling",
					}, []TokPos{ins.ArgContent[0].TokenPos}
				}

				vm.PC = tgtBlock.BeginPC
				vm.Lines++
			}

		case OP_BEGIN:
			/*
				OP_BEGIN

				ArgType
					PUB_I64

				ArgContent
					LabelSign

				Code{
					OpCode:       OpWrap{OP_BEGIN,TokPos{..}},
					ArgType:      []ArgTypeWrap{..}, // 一个 PUB_I64
					IsArgSurface: []IsSurfaceWrap{..}, // 第一个字面值，编译期保证
					ArgContent:   []ArgContWrap{..},
				}
			*/

			beginPC := vm.PC - 1
			var endPC int64
			labelSign := ins.ArgContent[0].ArgContentInt64IfSurface
			isEndExists := false

			for nowPC, nowIns := range vm.Codes[beginPC:] {
				switch nowIns.OpCode.Code {
				case OP_END:
					// 遇到 end, 检查 label
					if nowIns.ArgContent[0].ArgContentInt64IfSurface != labelSign {
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
				return &ErrBlockNotClosed{Label: labelSign}, []TokPos{ins.ArgContent[0].TokenPos}
			}

			// 存在，记录
			block := Block{
				BeginPC: beginPC,
				EndPC:   endPC,
			}

			vm.Blocks[labelSign] = block

			vm.PC = endPC + 2 // 跳过 block body 和 end，到达 end 后一条指令
			vm.Lines++

		case OP_END:
			// 理论上不可能到达这个位置
			// 因为已经在 OP_BEGIN 处理
			return &ErrUnexpectedEnd{Label: ins.ArgContent[0].ArgContentInt64IfSurface}, []TokPos{ins.ArgContent[0].TokenPos}

		case OP_CALL_RS:
			/*
				OP_CALL_RS

				ArgType (example)
					func_name
					int64
					int64 / bytes
					... n 个

				ArgContent (example)
					funname
					inNum
					outNum
					i0
					i1
					o0

				Code{
					OpCode:       OpWrap{OP_CALL_RS,TokPos{..}},
					ArgType:      []ArgTypeWrap{..}, // 一个 PUB_BYTES, 后面不限，由编译期保证
					IsArgSurface: []IsSurfaceWrap{..}, // 第一个是字面值，后面的不限，由编译期保证
					ArgContent:   []ArgContWrap{..},
				}
			*/

			// 此处个数不定，为了简化，出错位置统一定位在 call... 关键字
			sourcePosCall := ins.OpCode.TokenPos

			// 匹配 "__internal_数字" 在字符串末尾, 这是编译器用于判定统一函数调用次数
			re := regexp.MustCompile(`__internal_(\d+)$`)
			matches := re.FindStringSubmatch(string(ins.ArgContent[0].ArgContentBytesIfSurface))

			if len(matches) < 2 {
				// 没有匹配到，返回原字符串和错误
				return &ErrFuncNameInvalid{Name: string(ins.ArgContent[0].ArgContentBytesIfSurface)}, []TokPos{sourcePosCall}
			}

			// 提取数字
			funCallIdx, err := strconv.Atoi(matches[1])
			if err != nil {
				return &ErrFuncNameInvalid{Name: string(ins.ArgContent[0].ArgContentBytesIfSurface)}, []TokPos{sourcePosCall}
			}

			// 删除后缀，返回剩余部分
			funName := re.ReplaceAllString(string(ins.ArgContent[0].ArgContentBytesIfSurface), "")

			// 获取函数运行要求
			req := vm.Env.Funcs[string(ins.ArgContent[0].ArgContentBytesIfSurface)]
			if req == nil {
				return &ErrFuncNotFound{Name: string(ins.ArgContent[0].ArgContentBytesIfSurface)}, []TokPos{sourcePosCall}
			}

			// 检查参数(没有输出参数，因为它是隐私内存，不需要显式赋值)是否符合函数调用要求
			for i, r := range req.ReqInput {
				kind := vm.Vars[ins.ArgContent[i].ArgContent].Kind
				if r == "bytes" {
					// require for bytes (pub and priv is acceptable)
					if kind != PUB_HEAP {
						return &ErrTypeMismatch{
							Expected: "PUB_BYTES",
							Got:      fmt.Sprintf("%s", kind),
						}, []TokPos{sourcePosCall}
					}
					continue
				} else if r == "int64" {
					// require for int64 (pub and priv is acceptable)
					if kind != PUB_STACK {
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
					_ = vm.allocVar(varNameThis, PUB_STACK, 0)
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
					_ = vm.allocVar(varNameThis, PUB_HEAP, int64(len(val.Value.([]byte))))
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
				_ = vm.allocVar(varNameThis, PRIV_HEAP, int64(len(ct)))
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
				_ = vm.allocVar(varNameThis, PUB_HEAP, 0)
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
			ptr := vm.allocVar(varNameThis, ISOLATED_STACK, int64(len(reportContent)))
			vm.IsolatedMem.writeHeap(reportContent, ptr)

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
			ptr = vm.allocVar(varNameThis, ISOLATED_STACK, int64(len(proofContent)))
			vm.IsolatedMem.writeHeap(proofContent, ptr)

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
			ptr = vm.allocVar(varNameThis, ISOLATED_STACK, int64(len(vkeyContent)))
			vm.IsolatedMem.writeHeap(vkeyContent, ptr)

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
			return &ErrUnknownOpcode{Opcode: int(ins.OpCode.Code)}, []TokPos{ins.OpCode.TokenPos}
		}

		if vm.TraceMode {
			// generate trace step
			traceStep := vm.Mem.getDiff(vm.PreMem)
			traceStep.VarsChanges = vm.getVarsDiff()
			traceStep.BlocksChanges = vm.getBlocksDiff()
			vm.Traces[vm.Lines] = traceStep
		}
	}
	return nil, []TokPos{}
}

func (vm *VM) extractArithmeticToken(ins Code, position int64) (int64, error, []TokPos, bool) {
	thisType := ins.ArgType[position].Type
	var thisNum int64
	switch thisType {
	case PUB_I64:
		switch ins.IsArgSurface[position].IsSurface {
		case true:
			// 字面 PUB_I64
			thisNum = ins.ArgContent[position].ArgContentInt64IfSurface
		case false:
			// 变量 PUB_I64
			varName := ins.ArgContent[position].ArgContent
			var ok bool
			thisNum, ok = vm.getInt64VarPub(varName)
			if !ok {
				return 0, &ErrVarNotFound{Name: varName}, []TokPos{ins.ArgContent[position].TokenPos}, false
			}
		}
	case PUB_BYTES:
		// （若编译期无误，这里永远不会到达，写这里违背了编译期严格确定类型和内存问题的原则，但起到消除警告的作用）
		return 0, &ErrTypeMismatch{Expected: "PUB_I64", Got: "PUB_BYTES"}, []TokPos{ins.ArgContent[position].TokenPos}, false
	case PRIV_I64:
		// （若编译期无误，这里永远不会到达，写这里违背了编译期严格确定类型和内存问题的原则，但起到消除警告的作用）
		return 0, &ErrTypeMismatch{
			Expected: "PUB_I64",
			Got:      "PRIV_I64",
		}, []TokPos{ins.ArgContent[position].TokenPos}, false
	case PRIV_BYTES:
		// （若编译期无误，这里永远不会到达，写这里违背了编译期严格确定类型和内存问题的原则，但起到消除警告的作用）
		return 0, &ErrTypeMismatch{
			Expected: "PUB_I64",
			Got:      "PRIV_BYTES",
		}, []TokPos{ins.ArgContent[position].TokenPos}, false
	case ISO_I64:
		// （若编译期无误，这里永远不会到达，写这里违背了编译期严格确定类型和内存问题的原则，但起到消除警告的作用）
		return 0, &ErrTypeMismatch{
			Expected: "PUB_I64",
			Got:      "ISO_I64",
		}, []TokPos{ins.ArgContent[position].TokenPos}, false
	case ISO_BYTES:
		// （若编译期无误，这里永远不会到达，写这里违背了编译期严格确定类型和内存问题的原则，但起到消除警告的作用）
		return 0, &ErrTypeMismatch{
			Expected: "PUB_I64",
			Got:      "ISO_BYTES",
		}, []TokPos{ins.ArgContent[position].TokenPos}, false
	default:
		return 0, &ErrUnsupportedMemType{MemType: PtrKind(thisType)}, []TokPos{ins.ArgContent[position].TokenPos}, true
	}
	return thisNum, nil, nil, true
}

func (vm *VM) extractArithmeticTokenBytes(ins Code, position int64) ([]byte, error, []TokPos, bool) {
	thisType := ins.ArgType[position].Type
	var thisBytes []byte
	switch thisType {
	case PUB_I64:
		// （若编译期无误，这里永远不会到达，写这里违背了编译期严格确定类型和内存问题的原则，但起到消除警告的作用）
		return nil, &ErrTypeMismatch{Expected: "PUB_BYTES", Got: "PUB_I64"}, []TokPos{ins.ArgContent[position].TokenPos}, false
	case PUB_BYTES:
		switch ins.IsArgSurface[position].IsSurface {
		case true:
			// 字面 PUB_BYTES
			thisBytes = ins.ArgContent[position].ArgContentBytesIfSurface
		case false:
			// 变量 PUB_I64
			varName := ins.ArgContent[position].ArgContent
			var ok bool
			thisBytes, ok = vm.getBytesVarPub(varName)
			if !ok {
				return nil, &ErrVarNotFound{Name: varName}, []TokPos{ins.ArgContent[position].TokenPos}, false
			}
		}
	case PRIV_I64:
		// （若编译期无误，这里永远不会到达，写这里违背了编译期严格确定类型和内存问题的原则，但起到消除警告的作用）
		return nil, &ErrTypeMismatch{
			Expected: "PUB_BYTES",
			Got:      "PRIV_I64",
		}, []TokPos{ins.ArgContent[position].TokenPos}, false
	case PRIV_BYTES:
		// （若编译期无误，这里永远不会到达，写这里违背了编译期严格确定类型和内存问题的原则，但起到消除警告的作用）
		return nil, &ErrTypeMismatch{
			Expected: "PUB_BYTES",
			Got:      "PRIV_BYTES",
		}, []TokPos{ins.ArgContent[position].TokenPos}, false
	case ISO_I64:
		// （若编译期无误，这里永远不会到达，写这里违背了编译期严格确定类型和内存问题的原则，但起到消除警告的作用）
		return nil, &ErrTypeMismatch{
			Expected: "PUB_BYTES",
			Got:      "ISO_I64",
		}, []TokPos{ins.ArgContent[position].TokenPos}, false
	case ISO_BYTES:
		// （若编译期无误，这里永远不会到达，写这里违背了编译期严格确定类型和内存问题的原则，但起到消除警告的作用）
		return nil, &ErrTypeMismatch{
			Expected: "PUB_BYTES",
			Got:      "ISO_BYTES",
		}, []TokPos{ins.ArgContent[position].TokenPos}, false
	default:
		return nil, &ErrUnsupportedMemType{MemType: PtrKind(thisType)}, []TokPos{ins.ArgContent[position].TokenPos}, true
	}
	return thisBytes, nil, nil, true
}

// allocVar 声明一个变量并分配栈空间，返回指针; size only be needed for heap
func (vm *VM) allocVar(name string, memType PtrKind, size int64) Ptr {
	if _, exists := vm.Vars[name]; !exists {
		var ptr Ptr

		switch memType {
		case PUB_STACK:
			ptr = vm.Mem.allocStack()
		case PUB_HEAP:
			ptr = vm.Mem.allocHeap(size)
		case PRIV_STACK:
			// 隐私内存由调用者提供，不支持 VM 层更新，支持在外部函数中获取副本可变性
			ptr = vm.PrivacyMem.allocStack()
		case PRIV_HEAP:
			// 隐私内存由调用者提供，不支持 VM 层更新，支持在外部函数中获取副本可变性
			ptr = vm.PrivacyMem.allocHeap(size)
		case ISOLATED_STACK:
			ptr = vm.IsolatedMem.allocStack()
		case ISOLATED_HEAP:
			ptr = vm.IsolatedMem.allocHeap(size)
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
	if !ok || ptr.Kind != PUB_STACK {
		return 0, false
	}
	return vm.Mem.readStack(ptr)
}

// getInt64VarPriv 获取隐私内存变量值（int64）
func (vm *VM) getInt64VarPriv(name string) (int64, bool) {
	ptr, ok := vm.Vars[name]
	if !ok || ptr.Kind != PRIV_STACK {
		return 0, false
	}
	return vm.PrivacyMem.readStack(ptr)
}

// getBytesVarPub 获取 Heap 变量值（[]byte）
func (vm *VM) getBytesVarPub(name string) ([]byte, bool) {
	ptr, ok := vm.Vars[name]
	if !ok || ptr.Kind != PUB_HEAP {
		return nil, false
	}
	return vm.Mem.readHeap(ptr)
}

// getBytesVarPriv 获取隐私内存变量值（[]byte）
func (vm *VM) getBytesVarPriv(name string) ([]byte, bool) {
	ptr, ok := vm.Vars[name]
	if !ok || ptr.Kind != PRIV_HEAP {
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
	case PUB_STACK:
		data, ok := vm.Mem.readStack(dataPtr)
		if !ok {
			return false
		}
		return vm.Mem.writeStack(data, targetPtr)
	case PUB_HEAP:
		data, ok := vm.Mem.readHeap(dataPtr)
		if !ok {
			return false
		}
		return vm.Mem.writeHeap(data, targetPtr)
	case PRIV_STACK:
		// 隐私内存不支持映射到公共内存
		return false
	case PRIV_HEAP:
		// 隐私内存不支持映射到公共内存
		return false
	case ISOLATED_STACK:
		// 隔离内存不支持高级封装
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
	if ptr.Kind != PUB_STACK {
		return false
	}

	return vm.Mem.writeStack(i, ptr)
}

func (vm *VM) updateVarBySurfaceBytesPub(name string, data []byte) bool {
	ptr, ok := vm.Vars[name]
	if !ok {
		return false
	}
	if ptr.Kind != PUB_HEAP {
		return false
	}

	return vm.Mem.writeHeap(data, ptr)
}

func (vm *VM) updateVarBySurfaceInt64Priv(name string, i int64) bool {
	ptr, ok := vm.Vars[name]
	if !ok {
		return false
	}
	if ptr.Kind != PRIV_STACK {
		return false
	}

	return vm.PrivacyMem.writeStack(i, ptr)
}

func (vm *VM) updateVarBySurfaceBytesPriv(name string, data []byte) bool {
	ptr, ok := vm.Vars[name]
	if !ok {
		return false
	}
	if ptr.Kind != PRIV_HEAP {
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
