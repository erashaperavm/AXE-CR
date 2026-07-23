package vm

type Environment struct {
	Input   func(local []byte)
	Output  func(local []byte)
	ReadFn  func(local []byte, chain []byte) error
	WriteFn func(local []byte, chain []byte) error
}
