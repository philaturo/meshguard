# MeshGuard

**Resilient Transaction Coordination Layer for Bitcoin, Lightning & Nostr**

&gt; Built for Unreliable Networks

---

## Overview

MeshGuard is a resilient transaction coordination system that ensures payments survive network partitions. When connectivity drops, transactions queue locally. When the network returns, the reconciliation engine replays and settles queued events atomically.

**Core Philosophy:** When networks fail, transactions shouldn't.

---

## Architecture
```bash meshguard/
├── apps/
│   ├── api/                    # HTTP server + WebSocket hub
│   │   ├── main.go             # Server entry point, DI container
│   │   ├── handlers.go         # REST endpoints (GET/POST)
│   │   └── websocket.go        # Real-time push to dashboard
│   └── dashboard/              # Vite + React frontend
│       ├── src/
│       │   ├── components/     # UI components (BitcoinCore, NodeCard, etc.)
│       │   ├── hooks/          # useApi.ts, useWebSocket.ts
│       │   └── types/          # TypeScript interfaces
│       ├── vite.config.ts      # Proxy config (ports 8082, 5173)
│       └── index.html          # Dark theme entry
├── drivers/
│   ├── bitcoin/
│   │   └── rpc_client.go       # Bitcoin Core JSON-RPC adapter
│   └── lightning/
│       ├── interfaces.go       # LightningDriver abstraction
│       └── lnd_client.go       # LND REST API + lncli fallback
├── sdk/
│   ├── types/
│   │   ├── event.go            # MeshGuardEvent, EventStatus, state machine
│   │   ├── clock.go            # Atomic sequence counter
│   │   └── reconcile.go        # Reconciliation summary
│   ├── queue/
│   │   ├── store.go            # EventStore interface
│   │   └── sqlite_store.go     # SQLite WAL implementation
│   └── engine/
│       └── reconciler.go       # Pause/Resume/Reconcile logic
├── bin/                        # Compiled binaries
├── data/                       # SQLite database + LND data
├── scripts/                    # Automation scripts
├── go.mod                      # Go module definition
├── go.sum                      # Dependency checksums
├── Makefile                    # Build automation
```


---

## Technology Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| **Blockchain** | Bitcoin Core (regtest) | Base layer, block production |
| **Lightning** | LND 0.21.99-beta | Payment channels, HTLC routing |
| **Backend** | Go 1.21+ | REST API, WebSocket, reconciliation |
| **Frontend** | React 18 + Vite | Real-time dashboard |
| **Database** | SQLite (WAL mode) | Event queue persistence |
| **Transport** | HTTP/JSON + WebSocket | Client-server communication |

---

## Prerequisites

- Go 1.21+
- Node.js 18+ (for dashboard)
- Bitcoin Core (regtest mode)
- LND 0.21.99-beta (Alice + Bob nodes)
- `lncli` in `$PATH` (at `/home/aturo/go/bin/lncli`)

---

## Quick Start

### 1. Start Bitcoin Core

```bash
bitcoind -regtest -daemon


### Startland nodes
```bash
lnd --lnddir=$HOME/bootcamp-code/day3/alice
```

Unlock Alice wallet
```bash
lncli \
  --network=regtest \
  --rpcserver=127.0.0.1:10009 \
  --tlscertpath=$HOME/bootcamp-code/day3/alice/tls.cert \
  unlock
  ```

### Bob 

```bash
lnd --lnddir=$HOME/bootcamp-code/day3/bob
```

Unlock Bob's wallet

```bash
lncli \
  --network=regtest \
  --rpcserver=127.0.0.1:10010 \
  --tlscertpath=$HOME/bootcamp-code/day3/bob/tls.cert \
  unlock
```

### Verify Channel
```bash
lncli --network=regtest --rpcserver=127.0.0.1:10009 \
  --tlscertpath=$HOME/bootcamp-code/day3/alice/tls.cert \
  --macaroonpath=$HOME/bootcamp-code/day3/alice/data/chain/bitcoin/regtest/admin.macaroon \
  listchannels
  ```


## Start MeshGuard
```bash
cd ~/bootcamp-code/meshguard
go build -o bin/meshguard-api ./apps/api
./bin/meshguard-api &
```

## Start DAshboard
```bash
cd ~/bootcamp-code/meshguard/apps/dashboard
npm run dev
```

