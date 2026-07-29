package vm

// 地址分配器，初始地址从 0 开始
var nextStackAddr int64 = 0
var nextHeapAddr int64 = 0

type Memory struct {
	Stack map[int64]int64
	Heap  map[int64][]byte
}

func NewMemoryObj() *Memory {
	return &Memory{
		Stack: make(map[int64]int64),
		Heap:  make(map[int64][]byte),
	}
}

func (m *Memory) AllocStack() Ptr {
	addr := nextStackAddr
	nextStackAddr++
	m.Stack[addr] = 0
	return Ptr{Kind: Stack, Pointer: addr}
}

func (m *Memory) AllocHeap(size int64) Ptr {
	addr := nextHeapAddr
	nextHeapAddr += size
	m.Heap[addr] = make([]byte, size)
	return Ptr{Kind: Heap, Pointer: addr}
}

func (m *Memory) ReadStack(ptr Ptr) (int64, bool) {
	if ptr.Kind != Stack {
		return 0, false
	}
	i, ok := m.Stack[ptr.Pointer]
	if !ok {
		return 0, false
	}

	return i, true
}

func (m *Memory) WriteStack(data int64, ptr Ptr) bool {
	if ptr.Kind != Stack {
		return false
	}
	_, ok := m.Stack[ptr.Pointer]
	if !ok {
		// unexisted ptr
		return false
	}
	m.Stack[ptr.Pointer] = data
	return true
}

func (m *Memory) ReadHeap(ptr Ptr) ([]byte, bool) {
	if ptr.Kind != Heap {
		return nil, false
	}
	b, ok := m.Heap[ptr.Pointer]
	if !ok {
		return nil, false
	}

	return b, true
}

func (m *Memory) WriteHeap(data []byte, ptr Ptr) bool {
	if ptr.Kind != Heap {
		return false
	}
	_, ok := m.Heap[ptr.Pointer]
	if !ok {
		return false
	}
	m.Heap[ptr.Pointer] = data
	return true
}

func (m *Memory) Free(p Ptr) {
	switch p.Kind {
	case Stack:
		delete(m.Stack, p.Pointer)
	case Heap:
		delete(m.Heap, p.Pointer)
	}
}
