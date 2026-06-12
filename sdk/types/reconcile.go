// File: sdk/types/reconcile.go
// Purpose: Reconciliation result types
// Connects to: engine.Reconciler

package types

import "time"

// ReconcileResult summarizes a reconciliation run
type ReconcileResult struct {
	Processed int       `json:"processed"`
	Failed    int       `json:"failed"`
	Remaining int       `json:"remaining"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}
