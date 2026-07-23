package vm

// 地址分配器，初始地址从 0 开始
var nextStackAddr int64 = 0
var nextHeapAddr int64 = 0

type Memory struct {
	Stack map[int64]int64
	Heap  map[int64][]byte
}

func NewMemory() *Memory {
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

func (m *Memory) ReadInt64(p Ptr) (int64, bool) {
	if p.Kind != Stack {
		return 0, false
	}
	val, ok := m.Stack[p.Pointer]
	return val, ok
}

func (m *Memory) WriteInt64(p Ptr, val int64) bool {
	if p.Kind != Stack {
		return false
	}
	m.Stack[p.Pointer] = val
	return true
}

func (m *Memory) ReadBytes(p Ptr) ([]byte, bool) {
	if p.Kind != Heap {
		return nil, false
	}
	b, ok := m.Heap[p.Pointer]
	return b, ok
}

func (m *Memory) WriteBytes(p Ptr, data []byte) bool {
	if p.Kind != Heap {
		return false
	}
	b, ok := m.Heap[p.Pointer]
	if !ok || len(b) != len(data) {
		return false
	}
	copy(b, data)
	return true
}

func (m *Memory) AllocString(s string) Ptr {
	data := []byte(s)
	ptr := m.AllocHeap(int64(len(data)))
	if b, ok := m.Heap[ptr.Pointer]; ok {
		copy(b, data)
	}
	return ptr
}

func (m *Memory) AllocSlice(data []byte) Ptr {
	ptr := m.AllocHeap(int64(len(data)))
	if b, ok := m.Heap[ptr.Pointer]; ok {
		copy(b, data)
	}
	return ptr
}

func (m *Memory) Free(p Ptr) {
	switch p.Kind {
	case Stack:
		delete(m.Stack, p.Pointer)
	case Heap:
		delete(m.Heap, p.Pointer)
	}
}
