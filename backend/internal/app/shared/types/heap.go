package sharedtypes

import (
	"container/heap"
	"sync"
)

type Heap[T any] struct {
	mu     sync.Mutex
	closed bool
	h      *heapWrapper[T]
}

func NewHeap[T any](items []T, sortFunc func(i, j T) bool) *Heap[T] {
	return &Heap[T]{
		h: newHeapWrapper(items, sortFunc),
	}
}

func (h *Heap[T]) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.closed = true
}

func (h *Heap[T]) Pop() T {
	h.mu.Lock()
	defer h.mu.Unlock()

	var zero T
	if h.closed || h.h.Len() == 0 {
		return zero
	}

	return heap.Pop(h.h).(T)
}

func (h *Heap[T]) Push(x T) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return
	}

	heap.Push(h.h, x)
}

func (h *Heap[T]) Len() int {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return 0
	}

	return h.h.Len()
}

func (h *Heap[T]) Peek() T {
	h.mu.Lock()
	defer h.mu.Unlock()

	var zero T
	if h.closed || h.h.Len() == 0 {
		return zero
	}

	return h.h.items[0]
}

type heapWrapper[T any] struct {
	items    []T
	SortFunc func(i, j T) bool
}

func newHeapWrapper[T any](items []T, sortFunc func(i, j T) bool) *heapWrapper[T] {
	h := &heapWrapper[T]{
		items:    []T{},
		SortFunc: sortFunc,
	}

	for _, item := range items {
		heap.Push(h, item)
	}

	return h
}

func (h heapWrapper[T]) Len() int           { return len(h.items) }
func (h heapWrapper[T]) Less(i, j int) bool { return h.SortFunc(h.items[i], h.items[j]) }
func (h heapWrapper[T]) Swap(i, j int) {
	h.items[i], h.items[j] = h.items[j], h.items[i]
}

func (h *heapWrapper[T]) Push(x any) {
	h.items = append(h.items, x.(T))
}

func (h *heapWrapper[T]) Pop() any {
	old := h.items
	n := len(old)
	x := old[n-1]
	h.items = old[0 : n-1]
	return x
}
