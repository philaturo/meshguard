// File: sdk/types/clock.go
// Purpose: Monotonic sequence counter for event ordering and replay protection
// Connects to: MeshGuardEvent.Sequence

package types

import "sync/atomic"

// SequenceClock provides strictly increasing sequence numbers
type SequenceClock struct {
	current uint64
}

// NewSequenceClock creates a clock starting from a given value
func NewSequenceClock(start uint64) *SequenceClock {
	return &SequenceClock{current: start}
}

// Next returns the next sequence number atomically
func (c *SequenceClock) Next() uint64 {
	return atomic.AddUint64(&c.current, 1)
}

// Current returns the last issued sequence without incrementing
func (c *SequenceClock) Current() uint64 {
	return atomic.LoadUint64(&c.current)
}
