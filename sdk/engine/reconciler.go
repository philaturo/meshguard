// File: sdk/engine/reconciler.go
// Purpose: Conflict resolution and state reconciliation engine
// Connects to: types (MeshGuardEvent, ReconcileResult), queue (EventStore)
// Used by: api/handlers.go

package engine

import (
	"context"
	"fmt"
	"time"

	"meshguard/sdk/queue"
	"meshguard/sdk/types"
)

// Reconciler processes pending events when connectivity returns
type Reconciler struct {
	store  queue.EventStore
	clock  *types.SequenceClock
	active bool
}

// NewReconciler creates a reconciler bound to an event store
func NewReconciler(store queue.EventStore, clock *types.SequenceClock) *Reconciler {
	return &Reconciler{
		store:  store,
		clock:  clock,
		active: true,
	}
}

// Reconcile processes all QUEUED and OFFLINE events
func (r *Reconciler) Reconcile(ctx context.Context) (*types.ReconcileResult, error) {
	if !r.active {
		return nil, fmt.Errorf("reconciler paused")
	}

	pending, err := r.store.ListByStatus(ctx, types.StatusQueued)
	if err != nil {
		return nil, fmt.Errorf("list queued: %w", err)
	}

	offline, err := r.store.ListByStatus(ctx, types.StatusOffline)
	if err != nil {
		return nil, fmt.Errorf("list offline: %w", err)
	}

	pending = append(pending, offline...)

	result := &types.ReconcileResult{
		Processed: 0,
		Failed:    0,
		Remaining: len(pending),
		StartTime: time.Now(),
	}

	for _, event := range pending {
		if err := event.Transition(types.StatusReconciling); err != nil {
			result.Failed++
			continue
		}

		if err := r.store.Update(ctx, event); err != nil {
			result.Failed++
			continue
		}

		result.Processed++
		result.Remaining--
	}

	result.EndTime = time.Now()
	return result, nil
}

func (r *Reconciler) Pause() {
	r.active = false
}

func (r *Reconciler) Resume() {
	r.active = true
}

func (r *Reconciler) IsActive() bool {
	return r.active
}
