// File: apps/api/main.go
// Purpose: HTTP server entry point — initializes all drivers, SDK, and routes
// Connects to: handlers.go (REST endpoints), websocket.go (real-time push)
// Drivers: bitcoin/rpc_client.go, lightning/lnd_client.go
// SDK: queue/sqlite_store.go, engine/reconciler.go
// Usage: go run apps/api/main.go or ./bin/meshguard-api

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"meshguard/drivers/bitcoin"
	"meshguard/drivers/lightning"
	"meshguard/sdk/engine"
	"meshguard/sdk/queue"
)

const (
	dataDir         = "./data"
	dbPath          = dataDir + "/meshguard.db"
	bitcoinRPCHost  = "localhost:18443"
	bitcoinRPCUser  = "bootcamp"
	bitcoinRPCPass  = "bootcamp123"
	aliceRPCAddr    = "localhost:10009"
	bobRPCAddr      = "localhost:10010"
)

func main() {
	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("create data dir: %v", err)
	}

	// Initialize SQLite store
	store, err := queue.NewSQLiteStore(dbPath)
	if err != nil {
		log.Fatalf("init store: %v", err)
	}
	defer store.Close()

	// Initialize sequence clock from highest existing sequence
	events, err := store.ListAll(context.Background(), 1)
	var startSeq uint64
	if len(events) > 0 {
		startSeq = events[0].Sequence
	}
	clock := engine.NewSequenceClock(startSeq)

	// Initialize reconciler
	reconciler := engine.NewReconciler(store, clock)

	// Initialize Bitcoin Core RPC (live connection)
	btcClient := bitcoin.NewRPCClient(bitcoinRPCHost, bitcoinRPCUser, bitcoinRPCPass)

	// Initialize LND clients (deferred connection until nodes are ready)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("get home dir: %v", err)
	}

	aliceClient := lightning.NewLNDClient(lightning.NodeConfig{
		Name:         "Alice",
		RPCAddr:      aliceRPCAddr,
		TLSCertPath:  homeDir + "/bootcamp-code/day3/alice/tls.cert",
		MacaroonPath: homeDir + "/bootcamp-code/day3/alice/data/chain/bitcoin/regtest/admin.macaroon",
	})

	bobClient := lightning.NewLNDClient(lightning.NodeConfig{
		Name:         "Bob",
		RPCAddr:      bobRPCAddr,
		TLSCertPath:  homeDir + "/bootcamp-code/day3/bob/tls.cert",
		MacaroonPath: homeDir + "/bootcamp-code/day3/bob/data/chain/bitcoin/regtest/admin.macaroon",
	})

	// Attempt initial connections (non-fatal if nodes not ready)
	if err := aliceClient.Connect(); err != nil {
		log.Printf("Alice not ready: %v", err)
	}
	if err := bobClient.Connect(); err != nil {
		log.Printf("Bob not ready: %v", err)
	}

	// Build server with all dependencies
	server := NewServer(ServerDeps{
		Store:       store,
		Clock:       clock,
		Reconciler:  reconciler,
		Bitcoin:     btcClient,
		Alice:       aliceClient,
		Bob:         bobClient,
	})

	// HTTP server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: server.Router(),
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		log.Println("Shutting down server...")
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()

	log.Println("MeshGuard API listening on http://localhost:8080")
	log.Println("Dashboard available at http://localhost:5173")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

// ServerDeps holds all injected dependencies for the HTTP server
type ServerDeps struct {
	Store      queue.EventStore
	Clock      *engine.SequenceClock
	Reconciler *engine.Reconciler
	Bitcoin    *bitcoin.RPCClient
	Alice      lightning.LightningDriver
	Bob        lightning.LightningDriver
}
