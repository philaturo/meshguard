// File: apps/api/handlers.go
// Purpose: REST API endpoints — all dashboard data sources and demo controls
// Connects to: main.go (ServerDeps), websocket.go (broadcasts updates)
// Routes:
//   GET  /api/bitcoin/status    -> Bitcoin Core block height, mempool, network
//   GET  /api/nodes/status      -> Alice and Bob node info (live or waiting)
//   GET  /api/channels          -> Active channel state
//   GET  /api/events            -> Event queue (pending, settled, failed)
//   GET  /api/sync/status       -> Reconciler state (active/paused)
//   POST /api/offline           -> Pause sync, disconnect Alice
//   POST /api/payment           -> Create payment event (queues if offline)
//   POST /api/reconnect         -> Resume sync, reconnect, process queue

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"meshguard/drivers/bitcoin"
	"meshguard/drivers/lightning"
	"meshguard/sdk/engine"
)

// Server holds all HTTP handlers with injected dependencies
type Server struct {
	deps ServerDeps
	hub  *Hub // WebSocket hub for real-time broadcasts
}

// NewServer creates a server with the given dependencies
func NewServer(deps ServerDeps) *Server {
	return &Server{
		deps: deps,
		hub:  NewHub(),
	}
}

// Router sets up all API routes
func (s *Server) Router() *mux.Router {
	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api").Subrouter()

	api.HandleFunc("/bitcoin/status", s.handleBitcoinStatus).Methods("GET")
	api.HandleFunc("/nodes/status", s.handleNodesStatus).Methods("GET")
	api.HandleFunc("/channels", s.handleChannels).Methods("GET")
	api.HandleFunc("/events", s.handleEvents).Methods("GET")
	api.HandleFunc("/sync/status", s.handleSyncStatus).Methods("GET")

	api.HandleFunc("/offline", s.handleGoOffline).Methods("POST")
	api.HandleFunc("/payment", s.handleCreatePayment).Methods("POST")
	api.HandleFunc("/reconnect", s.handleReconnect).Methods("POST")

	// WebSocket endpoint
	r.HandleFunc("/ws", s.handleWebSocket)

	// Start WebSocket hub
	go s.hub.Run()

	return r
}

// handleBitcoinStatus returns live Bitcoin Core data
func (s *Server) handleBitcoinStatus(w http.ResponseWriter, r *http.Request) {
	info, err := s.deps.Bitcoin.GetBlockchainInfo()
	if err != nil {
		respondJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"status":  "waiting",
			"error":   err.Error(),
			"message": "Bitcoin Core not reachable at localhost:18443",
		})
		return
	}

	mempool, err := s.deps.Bitcoin.GetMempoolInfo()
	if err != nil {
		mempool = &bitcoin.MempoolInfo{Size: 0, Bytes: 0}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":          "online",
		"height":          info.Blocks,
		"headers":         info.Headers,
		"best_block_hash": info.BestBlockHash,
		"network":         info.Chain,
		"mempool_size":    mempool.Size,
		"mempool_bytes":   mempool.Bytes,
	})
}

// handleNodesStatus returns Alice and Bob node state
func (s *Server) handleNodesStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	alice := s.getNodeStatus(ctx, s.deps.Alice, "Alice")
	bob := s.getNodeStatus(ctx, s.deps.Bob, "Bob")

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"alice": alice,
		"bob":   bob,
	})
}

// getNodeStatus fetches live node data or returns waiting state
func (s *Server) getNodeStatus(ctx context.Context, client lightning.LightningDriver, name string) map[string]interface{} {
	if !client.IsConnected() {
		if err := client.Connect(); err != nil {
			return map[string]interface{}{
				"status":  "waiting",
				"alias":   name,
				"error":   err.Error(),
				"message": fmt.Sprintf("%s node not reachable", name),
			}
		}
	}

	info, err := client.GetInfo(ctx)
	if err != nil {
		return map[string]interface{}{
			"status": "error",
			"alias":  name,
			"error":  err.Error(),
		}
	}

	balance, err := client.GetWalletBalance(ctx)
	if err != nil {
		balance = &lightning.WalletBalance{}
	}

	return map[string]interface{}{
		"status":              "online",
		"alias":               info.Alias,
		"pubkey":              info.Pubkey,
		"channels":            info.Channels,
		"balance_total":       balance.TotalBalance,
		"balance_confirmed":   balance.ConfirmedBalance,
		"balance_unconfirmed": balance.UnconfirmedBalance,
	}
}

// handleChannels returns active channel state
func (s *Server) handleChannels(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	channels, err := s.deps.Alice.ListChannels(ctx)
	if err != nil || len(channels) == 0 {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"status":   "waiting",
			"channels": []interface{}{},
			"message":  "No active channels. Complete bootcamp Day 3 to open Alice-Bob channel.",
		})
		return
	}

	var result []map[string]interface{}
	for _, ch := range channels {
		result = append(result, map[string]interface{}{
			"channel_id":     ch.ChannelID,
			"capacity":       ch.Capacity,
			"local_balance":  ch.LocalBalance,
			"remote_balance": ch.RemoteBalance,
			"active":         ch.Active,
			"remote_pubkey":  ch.RemotePubkey,
		})
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "active",
		"channels": result,
		"count":    len(result),
	})
}

// handleEvents returns MeshGuard event queue
func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	counts, err := s.deps.Store.CountByStatus(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "count events: %v", err)
		return
	}

	events, err := s.deps.Store.ListAll(ctx, 50)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "list events: %v", err)
		return
	}

	var formatted []map[string]interface{}
	for _, e := range events {
		formatted = append(formatted, map[string]interface{}{
			"id":         e.ID,
			"type":       e.Type,
			"status":     e.Status,
			"from":       e.FromNode,
			"to":         e.ToNode,
			"amount":     e.AmountSats,
			"sequence":   e.Sequence,
			"timestamp":  e.Timestamp,
			"created_at": e.CreatedAt,
			"updated_at": e.UpdatedAt,
		})
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"counts": map[string]interface{}{
			"pending": counts[engine.StatusQueued] + counts[engine.StatusOffline],
			"settled": counts[engine.StatusSettled],
			"failed":  counts[engine.StatusFailed],
		},
		"events": formatted,
	})
}

// handleSyncStatus returns reconciler state
func (s *Server) handleSyncStatus(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"reconciler_active": s.deps.Reconciler.IsActive(),
		"status":            map[bool]string{true: "online", false: "offline"}[s.deps.Reconciler.IsActive()],
	})
}

// handleGoOffline simulates network partition
func (s *Server) handleGoOffline(w http.ResponseWriter, r *http.Request) {
	s.deps.Reconciler.Pause()

	if err := s.deps.Alice.Disconnect(); err != nil {
		log.Printf("alice disconnect: %v", err)
	}

	s.hub.Broadcast(map[string]interface{}{
		"type":   "node_status",
		"node":   "Alice",
		"status": "offline",
	})

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "offline",
		"message": "Alice disconnected. Sync engine paused. New events will be queued.",
	})
}

// handleCreatePayment creates a payment event
func (s *Server) handleCreatePayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		FromNode string `json:"from_node"`
		ToNode   string `json:"to_node"`
		Amount   int64  `json:"amount_sats"`
		Invoice  string `json:"invoice,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "decode request: %v", err)
		return
	}

	event := &engine.MeshGuardEvent{
		ID:         fmt.Sprintf("evt-%d", s.deps.Clock.Next()),
		Type:       engine.EventTypePayment,
		Status:     engine.StatusCreated,
		FromNode:   req.FromNode,
		ToNode:     req.ToNode,
		AmountSats: req.Amount,
		Invoice:    req.Invoice,
		Sequence:   s.deps.Clock.Next(),
		Timestamp:  time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if !s.deps.Reconciler.IsActive() || !s.deps.Alice.IsConnected() {
		event.Status = engine.StatusOffline
	} else {
		event.Status = engine.StatusQueued
	}

	if err := s.deps.Store.Create(ctx, event); err != nil {
		respondError(w, http.StatusInternalServerError, "create event: %v", err)
		return
	}

	s.hub.Broadcast(map[string]interface{}{
		"type":    "new_event",
		"event":   event,
		"message": fmt.Sprintf("Payment %s created: %d sats %s -> %s", event.ID, event.AmountSats, event.FromNode, event.ToNode),
	})

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"event":   event,
		"status":  event.Status,
		"message": fmt.Sprintf("Payment queued. Status: %s", event.Status),
	})
}

// handleReconnect restores connectivity and processes queue
func (s *Server) handleReconnect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Reconnect Alice to LND
	if err := s.deps.Alice.Connect(); err != nil {
		respondJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"status":  "partial",
			"message": "Alice reconnection failed. Queue will remain pending.",
			"error":   err.Error(),
		})
		return
	}

	// Resume sync engine
	s.deps.Reconciler.Resume()

	// Process pending events
	result, err := s.deps.Reconciler.Reconcile(ctx)
	if err != nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"status":  "reconnected",
			"message": "Connected, but reconciliation failed.",
			"error":   err.Error(),
		})
		return
	}

	// Attempt settlement for each reconciled event
	pending, _ := s.deps.Store.ListByStatus(ctx, engine.StatusReconciling)
	for _, evt := range pending {
		// Attempt LND payment
		_, payErr := s.deps.Alice.SendPayment(ctx, evt.Invoice, evt.AmountSats)
		if payErr != nil {
			evt.Transition(engine.StatusFailed)
		} else {
			evt.Transition(engine.StatusSettled)
		}
		s.deps.Store.Update(ctx, evt)

		s.hub.Broadcast(map[string]interface{}{
			"type":   "event_updated",
			"event":  evt,
			"status": evt.Status,
		})
	}

	// Broadcast reconnection
	s.hub.Broadcast(map[string]interface{}{
		"type":   "node_status",
		"node":   "Alice",
		"status": "online",
	})

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "online",
		"message":   "Alice reconnected. Sync engine resumed.",
		"reconcile": result,
	})
}

// respondJSON sends a JSON response with the given status code
func respondJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

// respondError sends a formatted error response
func respondError(w http.ResponseWriter, code int, format string, args ...interface{}) {
	respondJSON(w, code, map[string]interface{}{
		"error": fmt.Sprintf(format, args...),
	})
}
