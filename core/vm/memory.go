package vm

import (
	"bytes"
)

// 地址分配器，初始地址从 0 开始
var nextStackAddr int64 = 0
var nextHeapAddr int64 = 0

type Memory struct {
	Stack map[int64]int64
	Heap  map[int64][]byte
}

// PrivacyMemory keeps a counter that ensures data can be only updated one time
type PrivacyMemory struct {
	Stack map[int64]struct {
		Detail  int64
		Counter int8
	}
	Heap map[int64]struct {
		Detail  []byte
		Counter int8
	}
}

type IsolationMemory struct {
	Stack map[int64]struct {
		Detail  int64
		Counter int8
	}
	Heap map[int64]struct {
		Detail  []byte
		Counter int8
	}
}

func NewMemoryObj() *Memory {
	return &Memory{
		Stack: make(map[int64]int64),
		Heap:  make(map[int64][]byte),
	}
}

func NewPrivacyMemoryObj() *PrivacyMemory {
	return &PrivacyMemory{
		Stack: make(map[int64]struct {
			Detail  int64
			Counter int8
		}),
		Heap: make(map[int64]struct {
			Detail  []byte
			Counter int8
		}),
	}
}

func NewIsolationMemoryObj() *IsolationMemory {
	return &IsolationMemory{
		Heap: make(map[int64]struct {
			Detail  []byte
			Counter int8
		}),
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
		panic("PrivStack is not supported")
	case PrivHeap:
		panic("PrivHeap is not supported")
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

func (m *PrivacyMemory) allocStack() Ptr {
	addr := nextStackAddr
	nextStackAddr++
	m.Stack[addr] = struct {
		Detail  int64
		Counter int8
	}{Detail: 0, Counter: 0}
	return Ptr{Kind: PrivStack, Pointer: addr}
}

func (m *PrivacyMemory) allocHeap(size int64) Ptr {
	addr := nextHeapAddr
	nextHeapAddr += size
	m.Heap[addr] = struct {
		Detail  []byte
		Counter int8
	}{Detail: make([]byte, size), Counter: 0}
	return Ptr{Kind: PrivHeap, Pointer: addr}
}

func (m *PrivacyMemory) readStack(ptr Ptr) (int64, bool) {
	if ptr.Kind != PrivStack {
		return 0, false
	}
	i, ok := m.Stack[ptr.Pointer]
	if !ok {
		return 0, false
	}

	return i.Detail, true
}

func (m *PrivacyMemory) writeStack(data int64, ptr Ptr) bool {
	if ptr.Kind != PrivStack {
		return false
	}
	_, ok := m.Stack[ptr.Pointer]
	if !ok {
		// unexisted ptr
		return false
	}
	if m.Stack[ptr.Pointer].Counter != 0 {
		return false
	}
	m.Stack[ptr.Pointer] = struct {
		Detail  int64
		Counter int8
	}{Detail: data, Counter: 1}
	return true
}

func (m *PrivacyMemory) readHeap(ptr Ptr) ([]byte, bool) {
	if ptr.Kind != PrivHeap {
		return nil, false
	}
	b, ok := m.Heap[ptr.Pointer]
	if !ok {
		return nil, false
	}

	return b.Detail, true
}

func (m *PrivacyMemory) writeHeap(data []byte, ptr Ptr) bool {
	if ptr.Kind != PrivHeap {
		return false
	}
	_, ok := m.Heap[ptr.Pointer]
	if !ok {
		return false
	}
	if m.Heap[ptr.Pointer].Counter != 0 {
		return false
	}
	m.Heap[ptr.Pointer] = struct {
		Detail  []byte
		Counter int8
	}{Detail: data, Counter: 1}
	return true
}

func (m *PrivacyMemory) free(p Ptr) {
	switch p.Kind {
	case PrivStack:
		// 隐私栈不支持 VM 层级释放
	case PrivHeap:
		// 隐私堆不支持 VM 层级释放
	default:
		panic("unimplemented ptr kind")
	}
}

func (m *IsolationMemory) allocHeap(size int64) Ptr {
	addr := nextHeapAddr
	nextHeapAddr += size
	m.Heap[addr] = struct {
		Detail  []byte
		Counter int8
	}{Detail: make([]byte, size), Counter: 0}
	return Ptr{Kind: IsolationHeap, Pointer: addr}
}

// readHeapOnlyForReplay func OnlyForReplay !!!
func (m *IsolationMemory) readHeapOnlyForReplay(ptr Ptr) ([]byte, bool) {
	if ptr.Kind != IsolationHeap {
		return nil, false
	}

	b, ok := m.Heap[ptr.Pointer]
	if !ok {
		return nil, false
	}

	if b.Counter != 1 {
		return nil, false
	}

	return b.Detail, true
}

func (m *IsolationMemory) writeHeap(data []byte, ptr Ptr) bool {
	if ptr.Kind != IsolationHeap {
		return false
	}
	_, ok := m.Heap[ptr.Pointer]
	if !ok {
		return false
	}
	if m.Heap[ptr.Pointer].Counter != 0 {
		return false
	}
	m.Heap[ptr.Pointer] = struct {
		Detail  []byte
		Counter int8
	}{Detail: data, Counter: 1}
	return true
}

// readStackOnlyForReplay func OnlyForReplay !!!
func (m *IsolationMemory) readStackOnlyForReplay(ptr Ptr) (int64, bool) {
	if ptr.Kind != IsolationStack {
		return 0, false
	}

	n, ok := m.Stack[ptr.Pointer]
	if !ok {
		return 0, false
	}

	if n.Counter != 1 {
		return 0, false
	}

	return n.Detail, true
}

func (m *IsolationMemory) writeStack(data int64, ptr Ptr) bool {
	if ptr.Kind != IsolationStack {
		return false
	}
	_, ok := m.Stack[ptr.Pointer]
	if !ok {
		return false
	}
	if m.Stack[ptr.Pointer].Counter != 0 {
		return false
	}
	m.Stack[ptr.Pointer] = struct {
		Detail  int64
		Counter int8
	}{Detail: data, Counter: 1}
	return true
}
