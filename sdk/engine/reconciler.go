// File: sdk/engine/reconciler.go
// Purpose: Conflict resolution and state reconciliation engine

package engine

import (
	"context"
	"fmt"
	"time"

	"meshguard/sdk/queue"
)

// Reconciler processes pending events when connectivity returns
type Reconciler struct {
	store  queue.EventStore
	clock  *SequenceClock
	active bool
}

// NewReconciler creates a reconciler bound to an event store
func NewReconciler(store queue.EventStore, clock *SequenceClock) *Reconciler {
	return &Reconciler{
		store:  store,
		clock:  clock,
		active: true,
	}
}

// Reconcile processes all QUEUED and OFFLINE events
func (r *Reconciler) Reconcile(ctx context.Context) (*ReconcileResult, error) {
	if !r.active {
		return nil, fmt.Errorf("reconciler paused")
	}

	pending, err := r.store.ListByStatus(ctx, StatusQueued)
	if err != nil {
		return nil, fmt.Errorf("list queued: %w", err)
	}

	offline, err := r.store.ListByStatus(ctx, StatusOffline)
	if err != nil {
		return nil, fmt.Errorf("list offline: %w", err)
	}

	pending = append(pending, offline...)

	result := &ReconcileResult{
		Processed: 0,
		Failed:    0,
		Remaining: len(pending),
		StartTime: time.Now(),
	}

	for _, event := range pending {
		if err := event.Transition(StatusReconciling); err != nil {
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

// Pause stops the reconciler (simulates offline mode)
func (r *Reconciler) Pause() {
	r.active = false
}

// Resume restarts the reconciler (simulates reconnect)
func (r *Reconciler) Resume() {
	r.active = true
}

// IsActive returns current reconciler state
func (r *Reconciler) IsActive() bool {
	return r.active
}

// ReconcileResult summarizes a reconciliation run
type ReconcileResult struct {
	Processed int       `json:"processed"`
	Failed    int       `json:"failed"`
	Remaining int       `json:"remaining"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}
