// File: sdk/queue/store.go
// Purpose: Event storage interface — abstracts persistence from engine logic
// Connects to: types (MeshGuardEvent), sqlite_store.go (implementation)
// Used by: engine.Reconciler, api handlers

package queue

import (
	"context"

	"meshguard/sdk/types"
)

// EventStore defines all persistence operations for MeshGuard events
type EventStore interface {
	Create(ctx context.Context, event *types.MeshGuardEvent) error
	Get(ctx context.Context, id string) (*types.MeshGuardEvent, error)
	Update(ctx context.Context, event *types.MeshGuardEvent) error
	ListByStatus(ctx context.Context, status types.EventStatus) ([]*types.MeshGuardEvent, error)
	ListAll(ctx context.Context, limit int) ([]*types.MeshGuardEvent, error)
	CountByStatus(ctx context.Context) (map[types.EventStatus]int, error)
	Delete(ctx context.Context, id string) error
}
