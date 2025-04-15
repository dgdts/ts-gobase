package atomic_buffer

import (
	"sync/atomic"
)

// AtomicBuffer is a thread-safe buffer
type AtomicBuffer[T any] struct {
	value atomic.Value
}

func NewAtomicBuffer[T any](init T) *AtomicBuffer[T] {
	ab := &AtomicBuffer[T]{}
	ab.value.Store(init)
	return ab
}

func (ab *AtomicBuffer[T]) Load() T {
	return ab.value.Load().(T)
}

func (ab *AtomicBuffer[T]) Store(newValue T) {
	ab.value.Store(newValue)
}
