// File: sdk/queue/event_store.go
// Purpose: Event storage interface — abstracts persistence from engine logic

package queue

import (
	"context"

	"meshguard/sdk/engine"
)

// EventStore defines all persistence operations for MeshGuard events
type EventStore interface {
	// Create persists a new event
	Create(ctx context.Context, event *engine.MeshGuardEvent) error

	// Get retrieves an event by ID
	Get(ctx context.Context, id string) (*engine.MeshGuardEvent, error)

	// Update modifies an existing event
	Update(ctx context.Context, event *engine.MeshGuardEvent) error

	// ListByStatus returns all events with a given status
	ListByStatus(ctx context.Context, status engine.EventStatus) ([]*engine.MeshGuardEvent, error)

	// ListAll returns all events ordered by sequence descending
	ListAll(ctx context.Context, limit int) ([]*engine.MeshGuardEvent, error)

	// CountByStatus returns aggregate counts for dashboard metrics
	CountByStatus(ctx context.Context) (map[engine.EventStatus]int, error)

	// Delete removes an event (rarely used, mainly for cleanup)
	Delete(ctx context.Context, id string) error
}
