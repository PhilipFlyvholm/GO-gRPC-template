package shared

import (
	"math"
	"sync"
)

type SafeTimestamp struct {
	value int32
	mu    sync.Mutex
}

func (s *SafeTimestamp) MaxInc(otherTime int32) {
	timestamp := math.Max(float64(s.value), float64(otherTime)) + 1
	s.mu.Lock()
	defer s.mu.Unlock()
	s.value = int32(timestamp)
}

func (s *SafeTimestamp) Increment() {
	s.mu.Lock()
	s.value++
	s.mu.Unlock()
}

func (s *SafeTimestamp) IncrementAndGet() int32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.value++
	return s.value
}

func (s *SafeTimestamp) Value() int32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.value
}
