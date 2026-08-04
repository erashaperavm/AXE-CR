package vm

import (
	"axe-cr/core/bridge"
	"axe-cr/core/config"
	"axe-cr/core/execute"
	"errors"
)

type Environment struct {
	// 执行者函数区域
	// idx -> input slice idx
	ReadInt64  func(dataPosOnChain []byte) (int64, error)
	ReadBytes  func(dataPosOnChain []byte) ([]byte, error)
	WriteInt64 func(dataPosOnChain []byte, data int64) error
	WriteBytes func(dataPosOnChain []byte, data []byte) error
	CallRS     func(funName string, ints []int64, bytes [][]byte) (*execute.RsSp1Result, error)
	CallC      func(funName string, input [][]byte) ([][]byte, error)
	CallCpp    func(funName string, input [][]byte) ([][]byte, error)

	// todo 验证者函数区域（无法解密，只能根据提供是 hash sum 来比较）
	// Input 验证哈希 proof 和哈希链上比较
	// Output 验证加密 proof

	// 数据区域
	Funcs       map[string]config.FunctionMeta
	InputTypes  map[int64]string
	OutputTypes map[int64]string
	ExeDir      string
	WorkDir     string
	Mode        string
}

func NewEnvironment(
	inputTypes,
	outputTypes map[int64]string,
	funcs map[string]config.FunctionMeta,
	exeDir,
	workDir,
	mode string,
) *Environment {
	e := &Environment{
		InputTypes:  inputTypes,
		OutputTypes: outputTypes,
		Funcs:       funcs,
	}

	e.ReadInt64 = func(dataPosOnChain []byte) (int64, error) {
		return bridge.ReadInt64(dataPosOnChain)
	}
	e.ReadBytes = func(dataPosOnChain []byte) ([]byte, error) {
		return bridge.ReadBytes(dataPosOnChain)
	}

	e.WriteInt64 = func(dataPosOnChain []byte, data int64) error {
		return bridge.WriteInt64(dataPosOnChain, data)
	}
	e.WriteBytes = func(dataPosOnChain []byte, data []byte) error {
		return bridge.WriteBytes(dataPosOnChain, data)
	}

	e.CallRS = func(funName string, ints []int64, bytes [][]byte) (*execute.RsSp1Result, error) {
		tv := funcs[funName].TargetVerify
		res, err := bridge.CallRS(exeDir, workDir, funName, tv, ints, bytes, mode)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
	e.CallC = func(funName string, input [][]byte) ([][]byte, error) {
		return nil, errors.New("coming soon")
	}
	e.CallCpp = func(funName string, input [][]byte) ([][]byte, error) {
		return nil, errors.New("coming soon")
	}
	return e
}
