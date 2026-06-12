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
	"meshguard/sdk/types"
)

// Server holds all HTTP handlers with injected dependencies
type Server struct {
	deps ServerDeps
	hub  *Hub
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

	api := r.PathPrefix("/api").Subrouter()

	api.HandleFunc("/bitcoin/status", s.handleBitcoinStatus).Methods("GET")
	api.HandleFunc("/nodes/status", s.handleNodesStatus).Methods("GET")
	api.HandleFunc("/channels", s.handleChannels).Methods("GET")
	api.HandleFunc("/events", s.handleEvents).Methods("GET")
	api.HandleFunc("/sync/status", s.handleSyncStatus).Methods("GET")

	api.HandleFunc("/offline", s.handleGoOffline).Methods("POST")
	api.HandleFunc("/payment", s.handleCreatePayment).Methods("POST")
	api.HandleFunc("/reconnect", s.handleReconnect).Methods("POST")

	r.HandleFunc("/ws", s.handleWebSocket)

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
		respondError(w, http.StatusInternalServerError, "count events: "+err.Error())
		return
	}

	events, err := s.deps.Store.ListAll(ctx, 50)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "list events: "+err.Error())
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
			"pending": counts[types.StatusQueued] + counts[types.StatusOffline],
			"settled": counts[types.StatusSettled],
			"failed":  counts[types.StatusFailed],
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

// handleCreatePayment creates a payment event with invoice generation
func (s *Server) handleCreatePayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		FromNode string `json:"from_node"`
		ToNode   string `json:"to_node"`
		Amount   int64  `json:"amount_sats"`
		Invoice  string `json:"invoice,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "decode request: "+err.Error())
		return
	}

	event := &types.MeshGuardEvent{
		ID:         fmt.Sprintf("evt-%d", s.deps.Clock.Next()),
		Type:       types.EventTypePayment,
		Status:     types.StatusCreated,
		FromNode:   req.FromNode,
		ToNode:     req.ToNode,
		AmountSats: req.Amount,
		Sequence:   s.deps.Clock.Next(),
		Timestamp:  time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Create invoice from Bob if not provided — ensures reconciliation can execute payment
	if req.Invoice == "" && req.ToNode == "Bob" {
		invoiceResult, err := s.deps.Bob.AddInvoice(ctx, req.Amount, fmt.Sprintf("Payment from %s", req.FromNode))
		if err != nil {
			log.Printf("invoice creation failed: %v", err)
		} else {
			event.Invoice = invoiceResult.PaymentRequest
			log.Printf("Created invoice for event %s: %s", event.ID, event.Invoice)
		}
	} else {
		event.Invoice = req.Invoice
	}

	// Determine status based on network state
	if !s.deps.Reconciler.IsActive() || !s.deps.Alice.IsConnected() {
		event.Status = types.StatusOffline
	} else {
		event.Status = types.StatusQueued
	}

	if err := s.deps.Store.Create(ctx, event); err != nil {
		respondError(w, http.StatusInternalServerError, "create event: "+err.Error())
		return
	}

	s.hub.Broadcast(map[string]interface{}{
		"type":    "new_event",
		"event":   event,
		"message": fmt.Sprintf("Payment %s: %d sats %s -> %s", event.ID, event.AmountSats, event.FromNode, event.ToNode),
	})

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"event":   event,
		"status":  event.Status,
		"message": fmt.Sprintf("Payment %s. Status: %s", map[bool]string{true: "queued", false: "stored"}[event.Status == types.StatusQueued], event.Status),
	})
}

// handleReconnect restores connectivity and processes queue
func (s *Server) handleReconnect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.deps.Alice.Connect(); err != nil {
		respondJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"status":  "partial",
			"message": "Alice reconnection failed. Queue will remain pending.",
			"error":   err.Error(),
		})
		return
	}

	s.deps.Reconciler.Resume()

	result, err := s.deps.Reconciler.Reconcile(ctx)
	if err != nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"status":  "reconnected",
			"message": "Connected, but reconciliation failed.",
			"error":   err.Error(),
		})
		return
	}

	// Process reconciling events — attempt actual Lightning payment
	pending, _ := s.deps.Store.ListByStatus(ctx, types.StatusReconciling)
	for _, evt := range pending {
		// If no invoice, try to create one now via Bob
		if evt.Invoice == "" && evt.ToNode == "Bob" {
			invoiceResult, err := s.deps.Bob.AddInvoice(ctx, evt.AmountSats, fmt.Sprintf("Payment from %s", evt.FromNode))
			if err != nil {
				log.Printf("Failed to create invoice for %s: %v", evt.ID, err)
				evt.Transition(types.StatusFailed)
				s.deps.Store.Update(ctx, evt)
				continue
			}
			evt.Invoice = invoiceResult.PaymentRequest
			log.Printf("Created invoice for %s: %s", evt.ID, evt.Invoice)
		}

		// Attempt payment via Alice
		if evt.Invoice != "" {
			payResult, payErr := s.deps.Alice.SendPayment(ctx, evt.Invoice, evt.AmountSats)
			if payErr != nil {
				log.Printf("Payment failed for %s: %v", evt.ID, payErr)
				evt.Transition(types.StatusFailed)
			} else {
				log.Printf("Payment succeeded for %s: preimage=%s", evt.ID, payResult.Preimage)
				evt.Transition(types.StatusSettled)
			}
		} else {
			log.Printf("No invoice for %s, marking failed", evt.ID)
			evt.Transition(types.StatusFailed)
		}

		s.deps.Store.Update(ctx, evt)

		s.hub.Broadcast(map[string]interface{}{
			"type":   "event_updated",
			"event":  evt,
			"status": evt.Status,
		})
	}

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

// === RESPONSE HELPERS ===

// respondJSON writes a JSON response with status code
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("[ERROR] encoding JSON response: %v", err)
	}
}

// respondError writes an error JSON response
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
