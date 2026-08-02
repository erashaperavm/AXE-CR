package vm

import (
	"axe-cr/core/bridge"
	"axe-cr/core/config"
	"axe-cr/core/execute"
	"errors"
)

type Environment struct {
	// 函数区域
	// idx -> input slice idx
	InputInt64  func(idx int64) int64
	InputBytes  func(idx int64) []byte
	OutputInt64 func(idx int64, data int64)
	OutputBytes func(idx int64, data []byte)
	ReadInt64   func(dataPosOnChain []byte) (int64, error)
	ReadBytes   func(dataPosOnChain []byte) ([]byte, error)
	WriteInt64  func(dataPosOnChain []byte, data int64) error
	WriteBytes  func(dataPosOnChain []byte, data []byte) error
	CallRS      func(funName string, ints []int64, bytes [][]byte) (*execute.RsSp1Result, error)
	CallC       func(funName string, input [][]byte) ([][]byte, error)
	CallCpp     func(funName string, input [][]byte) ([][]byte, error)

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

	e.InputInt64 = func(idx int64) int64 {
		return bridge.InputInt64(idx)
	}
	e.InputBytes = func(idx int64) []byte {
		return bridge.InputBytes(idx)
	}

	e.OutputInt64 = func(idx int64, data int64) {
		bridge.OutputInt64(idx, data)
	}
	e.OutputBytes = func(idx int64, data []byte) {
		bridge.OutputBytes(idx, data)
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
