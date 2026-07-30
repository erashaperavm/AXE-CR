package vm

import "axe-cr/core/config"

type Environment[T NativeType] struct {
	// idx -> input slice idx
	// dataPtr -> ptr need to be written or read
	Input   func(idx int64, dataPtr Ptr)
	Output  func(idx int64, dataPtr Ptr)
	Read    func(dataPosOnChain []byte, dataPtr Ptr, okPtr Ptr)
	Write   func(dataPosOnChain []byte, dataPtr Ptr, okPtr Ptr)
	CallC   func(funName string, input []T, output []*T) error
	CallRS  func(funName string, input []T, output []*T) error
	CallCpp func(funName string, input []T, output []*T) error
	Mem     *Memory
	Funcs   map[string]config.FunctionMeta
}

func newEnvironment[T NativeType]() *Environment[T] {
	e := &Environment[T]{
		Mem:   NewMemoryObj(),
		Funcs: make(map[string]config.FunctionMeta),
	}

	// Mock: write a default value to memory (stack)
	e.Input = func(idx int64, dataPtr Ptr) {
		inputs := []int64{10, 20, 30}
		val := int64(42)
		if idx >= 0 && idx < int64(len(inputs)) {
			val = inputs[idx]
		}
		if dataPtr.Kind == Stack {
			e.Mem.WriteStack(val, dataPtr)
		}
	}

	// Mock: read from memory (stack) and ignore
	e.Output = func(idx int64, dataPtr Ptr) {
		if dataPtr.Kind == Stack {
			e.Mem.ReadStack(dataPtr) // discard
		}
	}

	// Mock: write default bytes to memory (heap), set ok to true
	e.Read = func(dataPosOnChain []byte, dataPtr Ptr, okPtr Ptr) {
		data := []byte("default read data")
		if dataPtr.Kind == Heap {
			e.Mem.WriteHeap(data, dataPtr)
		}
		if okPtr.Kind == Stack {
			e.Mem.WriteStack(1, okPtr)
		}
	}

	// Mock: read from memory (heap) and ignore, set ok to true
	e.Write = func(dataPosOnChain []byte, dataPtr Ptr, okPtr Ptr) {
		if dataPtr.Kind == Heap {
			e.Mem.ReadHeap(dataPtr) // discard
		}
		if okPtr.Kind == Stack {
			e.Mem.WriteStack(1, okPtr)
		}
	}
	return e
}
