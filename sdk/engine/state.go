// File: sdk/engine/state.go
// Purpose: Defines MeshGuardEvent structure and state machine transitions

package engine

import (
	"time"
)

// EventType identifies the kind of financial operation
type EventType string

const (
	EventTypePayment EventType = "PAYMENT"
	EventTypeInvoice EventType = "INVOICE"
)

// EventStatus tracks where an event is in the lifecycle
type EventStatus string

const (
	StatusCreated     EventStatus = "CREATED"
	StatusOffline     EventStatus = "OFFLINE"
	StatusQueued      EventStatus = "QUEUED"
	StatusReconciling EventStatus = "RECONCILING"
	StatusSettled     EventStatus = "SETTLED"
	StatusFailed      EventStatus = "FAILED"
)

// MeshGuardEvent is the core data structure for all offline-capable operations
type MeshGuardEvent struct {
	ID          string      `json:"id"`
	Type        EventType   `json:"type"`
	Status      EventStatus `json:"status"`
	FromNode    string      `json:"from_node"`
	ToNode      string      `json:"to_node"`
	AmountSats  int64       `json:"amount_sats"`
	ChannelID   string      `json:"channel_id,omitempty"`
	Sequence    uint64      `json:"sequence"`
	Timestamp   time.Time   `json:"timestamp"`
	Payload     []byte      `json:"payload"`
	Signature   []byte      `json:"signature"`
	HTLCHash    string      `json:"htlc_hash,omitempty"`
	Invoice     string      `json:"invoice,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// CanTransition checks if a status change is valid
func (e *MeshGuardEvent) CanTransition(to EventStatus) bool {
	valid := map[EventStatus][]EventStatus{
		StatusCreated:     {StatusOffline, StatusQueued, StatusSettled, StatusFailed},
		StatusOffline:     {StatusQueued, StatusReconciling},
		StatusQueued:      {StatusReconciling, StatusFailed},
		StatusReconciling: {StatusSettled, StatusFailed},
		StatusSettled:     {},
		StatusFailed:      {},
	}
	allowed, ok := valid[e.Status]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// Transition changes status if valid, returns error otherwise
func (e *MeshGuardEvent) Transition(to EventStatus) error {
	if !e.CanTransition(to) {
		return ErrInvalidTransition
	}
	e.Status = to
	e.UpdatedAt = time.Now()
	return nil
}

// ErrInvalidTransition is returned when a state change is not allowed
var ErrInvalidTransition = &StateError{"invalid state transition"}

type StateError struct {
	Msg string
}

func (e *StateError) Error() string {
	return e.Msg
}
