package vm

import "bytes"

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

func (m *Memory) allocStack() Ptr {
	addr := nextStackAddr
	nextStackAddr++
	m.Stack[addr] = 0
	return Ptr{Kind: PubStack, Pointer: addr}
}

func (m *Memory) allocHeap(size int64) Ptr {
	addr := nextHeapAddr
	nextHeapAddr += size
	m.Heap[addr] = make([]byte, size)
	return Ptr{Kind: PubHeap, Pointer: addr}
}

func (m *Memory) readStack(ptr Ptr) (int64, bool) {
	if ptr.Kind != PubStack && ptr.Kind != PrivStack {
		return 0, false
	}
	i, ok := m.Stack[ptr.Pointer]
	if !ok {
		return 0, false
	}

	return i, true
}

func (m *Memory) writeStack(data int64, ptr Ptr) bool {
	if ptr.Kind != PubStack {
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

func (m *Memory) readHeap(ptr Ptr) ([]byte, bool) {
	if ptr.Kind != PubHeap && ptr.Kind != PrivHeap {
		return nil, false
	}
	b, ok := m.Heap[ptr.Pointer]
	if !ok {
		return nil, false
	}

	return b, true
}

func (m *Memory) writeHeap(data []byte, ptr Ptr) bool {
	if ptr.Kind != PubHeap {
		return false
	}
	_, ok := m.Heap[ptr.Pointer]
	if !ok {
		return false
	}
	m.Heap[ptr.Pointer] = data
	return true
}

func (m *Memory) free(p Ptr) {
	switch p.Kind {
	case PubStack:
		delete(m.Stack, p.Pointer)
	case PubHeap:
		delete(m.Heap, p.Pointer)
	case PrivStack:
		// 隐私内存由调用者提供，不支持 VM 层更新，支持在外部函数中获取副本可变性
	case PrivHeap:
		// 隐私内存由调用者提供，不支持 VM 层更新，支持在外部函数中获取副本可变性
	default:
		panic("unimplemented ptr kind")
	}
}

func (m *Memory) copy() *Memory {
	var newMem *Memory

	for addr, val := range m.Stack {
		newMem.Stack[addr] = val
	}
	for addr, val := range m.Heap {
		newMem.Heap[addr] = val
	}
	return newMem
}

func (m *Memory) getDiff(oldMem *Memory) TraceStep {
	var stackChanges []StackDiff
	var heapChanges []HeapDiff
	for addr, pre := range oldMem.Stack {
		now, ok := m.Stack[addr]
		if !ok {
			continue
		}
		if pre != now {
			stackChanges = append(stackChanges, StackDiff{
				Addr: addr,
				Pre:  pre,
				Now:  now,
			})
		}
	}
	for addr, pre := range oldMem.Heap {
		now, ok := m.Heap[addr]
		if !ok {
			continue
		}
		if !bytes.Equal(pre, now) {
			heapChanges = append(heapChanges, HeapDiff{
				Addr: addr,
				Pre:  pre,
				Now:  now,
			})
		}
	}

	return TraceStep{
		StackChanges: stackChanges,
		HeapChanges:  heapChanges,
	}
}

func (m *Memory) applyDiff(diff TraceStep) {
	for _, stackDiff := range diff.StackChanges {
		if stackDiff.Now != stackDiff.Pre {
			m.Stack[stackDiff.Addr] = stackDiff.Now
		}
	}
	for _, heapDiff := range diff.HeapChanges {
		if !bytes.Equal(heapDiff.Now, heapDiff.Pre) {
			m.Heap[heapDiff.Addr] = heapDiff.Now
		}
	}
	return
}
