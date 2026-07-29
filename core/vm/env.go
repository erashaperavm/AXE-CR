package vm

type Environment[T NativeType] struct {
	// idx -> input slice idx
	// dataPtr -> ptr need to be written or read
	Input   func(idx int64, dataPtr Ptr)
	Output  func(idx int64, dataPtr Ptr)
	Read    func(dataPosOnChain []byte, dataPtr Ptr, okPtr Ptr)
	Write   func(dataPosOnChain []byte, dataPtr Ptr, okPtr Ptr)
	CallC   func(funName string, input [][]T, output [][]T)
	CallRS  func(funName string, input [][]T, output [][]T)
	CallCpp func(funName string, input [][]T, output [][]T)
	Mem     *Memory
}

// todo
func newEnvironment[T NativeType]() *Environment[T] {
	return &Environment[T]{
		// impl
		// mem
	}
}
