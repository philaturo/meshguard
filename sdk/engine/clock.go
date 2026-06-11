// File: sdk/engine/clock.go
// Purpose: Sequence validation and monotonic counter for event ordering

package engine

import (
	"sync"
)

// SequenceClock provides thread-safe monotonic sequence numbers
type SequenceClock struct {
	mu      sync.Mutex
	current uint64
}

// NewSequenceClock creates a clock starting from a given value
func NewSequenceClock(start uint64) *SequenceClock {
	return &SequenceClock{current: start}
}

// Next returns the next sequence number and increments the counter
func (sc *SequenceClock) Next() uint64 {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.current++
	return sc.current
}

// Current returns the current sequence without incrementing
func (sc *SequenceClock) Current() uint64 {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.current
}

// ValidateSequence checks if a sequence is strictly greater than current
func (sc *SequenceClock) ValidateSequence(seq uint64) bool {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return seq > sc.current
}
