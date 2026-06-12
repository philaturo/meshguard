// File: apps/dashboard/src/types/index.ts
// Purpose: Shared TypeScript interfaces for all dashboard components
// Connects to: All components in src/components/, hooks/useWebSocket.ts
// Mirrors: Go structs in sdk/types/ and api response shapes

export interface BlockchainStatus {
    status: 'online' | 'waiting' | 'error';
    height: number;
    headers: number;
    best_block_hash: string;
    network: string;
    mempool_size: number;
    mempool_bytes: number;
    error?: string;
    message?: string;
}

export interface NodeStatus {
    status: 'online' | 'waiting' | 'error';
    alias: string;
    pubkey?: string;
    channels?: number;
    balance_total?: number;
    balance_confirmed?: number;
    balance_unconfirmed?: number;
    error?: string;
    message?: string;
}

export interface NodesStatus {
    alice: NodeStatus;
    bob: NodeStatus;
}

export interface Channel {
    channel_id: string;
    capacity: number;
    local_balance: number;
    remote_balance: number;
    active: boolean;
    remote_pubkey: string;
}

export interface ChannelsResponse {
    status: 'active' | 'waiting';
    channels: Channel[];
    count: number;
    message?: string;
}

export interface MeshEvent {
    id: string;
    type: 'PAYMENT' | 'INVOICE';
    status: 'CREATED' | 'OFFLINE' | 'QUEUED' | 'RECONCILING' | 'SETTLED' | 'FAILED';
    from: string;
    to: string;
    amount: number;
    sequence: number;
    timestamp: string;
    created_at: string;
    updated_at: string;
}

export interface EventCounts {
    pending: number;
    settled: number;
    failed: number;
}

export interface EventsResponse {
    counts: EventCounts;
    events: MeshEvent[];
}

export interface SyncStatus {
    reconciler_active: boolean;
    status: 'online' | 'offline';
}

export interface WebSocketMessage {
    type: 'node_status' | 'new_event' | 'event_updated' | 'sync_status';
    node?: string;
    status?: string;
    event?: MeshEvent;
    message?: string;
}
