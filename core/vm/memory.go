package vm

import (
	"bytes"
)

// 地址分配器，初始地址从 0 开始
var nextPubStackAddr int64 = 0
var nextPubHeapAddr int64 = 0
var nextPrivStackAddr int64 = 0
var nextPrivHeapAddr int64 = 0
var nextIsoStackAddr int64 = 0
var nextIsoHeapAddr int64 = 0

type PubMemory struct {
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

type IsolatedMemory struct {
	Stack map[int64]struct {
		Detail  int64
		Counter int8
	}
	Heap map[int64]struct {
		Detail  []byte
		Counter int8
	}
}

func NewMemoryObj() *PubMemory {
	return &PubMemory{
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

func NewIsolationMemoryObj() *IsolatedMemory {
	return &IsolatedMemory{
		Heap: make(map[int64]struct {
			Detail  []byte
			Counter int8
		}),
	}
}

func (m *PubMemory) allocStack() Ptr {
	addr := nextPubStackAddr
	nextPubStackAddr++
	m.Stack[addr] = 0
	return Ptr{Kind: PUB_STACK, Pointer: addr}
}

func (m *PubMemory) allocHeap(size int64) Ptr {
	addr := nextPubHeapAddr
	nextPubHeapAddr++
	m.Heap[addr] = make([]byte, size)
	return Ptr{Kind: PUB_HEAP, Pointer: addr}
}

func (m *PubMemory) readStack(ptr Ptr) (int64, bool) {
	if ptr.Kind != PUB_STACK && ptr.Kind != PRIV_STACK {
		return 0, false
	}
	i, ok := m.Stack[ptr.Pointer]
	if !ok {
		return 0, false
	}

	return i, true
}

func (m *PubMemory) writeStack(data int64, ptr Ptr) bool {
	if ptr.Kind != PUB_STACK {
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

func (m *PubMemory) readHeap(ptr Ptr) ([]byte, bool) {
	if ptr.Kind != PUB_HEAP && ptr.Kind != PRIV_HEAP {
		return nil, false
	}
	b, ok := m.Heap[ptr.Pointer]
	if !ok {
		return nil, false
	}

	return b, true
}

func (m *PubMemory) writeHeap(data []byte, ptr Ptr) bool {
	if ptr.Kind != PUB_HEAP {
		return false
	}
	_, ok := m.Heap[ptr.Pointer]
	if !ok {
		return false
	}
	m.Heap[ptr.Pointer] = data
	return true
}

func (m *PubMemory) free(p Ptr) {
	switch p.Kind {
	case PUB_STACK:
		delete(m.Stack, p.Pointer)
	case PUB_HEAP:
		delete(m.Heap, p.Pointer)
	case PRIV_STACK:
		panic("PrivStack is not supported")
	case PRIV_HEAP:
		panic("PrivHeap is not supported")
	default:
		panic("unimplemented ptr kind")
	}
}

func (m *PubMemory) copy() *PubMemory {
	var newMem *PubMemory

	for addr, val := range m.Stack {
		newMem.Stack[addr] = val
	}
	for addr, val := range m.Heap {
		newMem.Heap[addr] = val
	}
	return newMem
}

func (m *PubMemory) getDiff(oldMem *PubMemory) TraceStep {
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

func (m *PubMemory) applyDiff(diff TraceStep) {
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
	addr := nextPrivStackAddr
	nextPrivStackAddr++
	m.Stack[addr] = struct {
		Detail  int64
		Counter int8
	}{Detail: 0, Counter: 0}
	return Ptr{Kind: PRIV_STACK, Pointer: addr}
}

func (m *PrivacyMemory) allocHeap(size int64) Ptr {
	addr := nextPubHeapAddr
	nextPrivHeapAddr++
	m.Heap[addr] = struct {
		Detail  []byte
		Counter int8
	}{Detail: make([]byte, size), Counter: 0}
	return Ptr{Kind: PRIV_HEAP, Pointer: addr}
}

func (m *PrivacyMemory) readStack(ptr Ptr) (int64, bool) {
	if ptr.Kind != PRIV_STACK {
		return 0, false
	}
	i, ok := m.Stack[ptr.Pointer]
	if !ok {
		return 0, false
	}

	return i.Detail, true
}

func (m *PrivacyMemory) writeStack(data int64, ptr Ptr) bool {
	if ptr.Kind != PRIV_STACK {
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
	if ptr.Kind != PRIV_HEAP {
		return nil, false
	}
	b, ok := m.Heap[ptr.Pointer]
	if !ok {
		return nil, false
	}

	return b.Detail, true
}

func (m *PrivacyMemory) writeHeap(data []byte, ptr Ptr) bool {
	if ptr.Kind != PRIV_HEAP {
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
	case PRIV_STACK:
		// 隐私栈不支持 VM 层级释放
	case PRIV_HEAP:
		// 隐私堆不支持 VM 层级释放
	default:
		panic("unimplemented ptr kind")
	}
}

func (m *IsolatedMemory) allocStack() Ptr {
	addr := nextIsoStackAddr
	nextIsoStackAddr++
	m.Stack[addr] = struct {
		Detail  int64
		Counter int8
	}{Detail: 0, Counter: 0}
	return Ptr{
		Kind:    ISOLATED_STACK,
		Pointer: addr,
	}
}

func (m *IsolatedMemory) allocHeap(size int64) Ptr {
	addr := nextIsoHeapAddr
	nextIsoHeapAddr++
	m.Heap[addr] = struct {
		Detail  []byte
		Counter int8
	}{Detail: make([]byte, size), Counter: 0}
	return Ptr{Kind: ISOLATED_STACK, Pointer: addr}
}

// readHeapOnlyForReplay func OnlyForReplay !!!
func (m *IsolatedMemory) readHeapOnlyForReplay(ptr Ptr) ([]byte, bool) {
	if ptr.Kind != ISOLATED_STACK {
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

func (m *IsolatedMemory) writeHeap(data []byte, ptr Ptr) bool {
	if ptr.Kind != ISOLATED_STACK {
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
func (m *IsolatedMemory) readStackOnlyForReplay(ptr Ptr) (int64, bool) {
	if ptr.Kind != ISOLATED_HEAP {
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

func (m *IsolatedMemory) writeStack(data int64, ptr Ptr) bool {
	if ptr.Kind != ISOLATED_HEAP {
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
