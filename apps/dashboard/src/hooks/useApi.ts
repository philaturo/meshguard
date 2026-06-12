// File: apps/dashboard/src/hooks/useApi.ts
// Purpose: HTTP API fetch hooks for all dashboard data sources
// Connects to: All components that need polling data (Bitcoin, Nodes, Channels, Events)
// Base URL: /api (proxied to localhost:8080 by Vite)

import { useState, useEffect, useCallback } from 'react';
import {
    BlockchainStatus,
    NodesStatus,
    ChannelsResponse,
    EventsResponse,
    SyncStatus,
} from '../types';

const API_BASE = '/api';

async function fetchJson<T>(path: string): Promise<T> {
    const res = await fetch(`${API_BASE}${path}`);
    if (!res.ok) {
        throw new Error(`HTTP ${res.status}: ${res.statusText}`);
    }
    return res.json();
}

export function useBitcoinStatus(pollInterval = 5000) {
    const [data, setData] = useState<BlockchainStatus | null>(null);
    const [error, setError] = useState<string | null>(null);

    const fetch = useCallback(async () => {
        try {
            const status = await fetchJson<BlockchainStatus>('/bitcoin/status');
            setData(status);
            setError(null);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Unknown error');
        }
    }, []);

    useEffect(() => {
        fetch();
        const interval = setInterval(fetch, pollInterval);
        return () => clearInterval(interval);
    }, [fetch, pollInterval]);

    return { data, error, refresh: fetch };
}

export function useNodesStatus(pollInterval = 3000) {
    const [data, setData] = useState<NodesStatus | null>(null);
    const [error, setError] = useState<string | null>(null);

    const fetch = useCallback(async () => {
        try {
            const status = await fetchJson<NodesStatus>('/nodes/status');
            setData(status);
            setError(null);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Unknown error');
        }
    }, []);

    useEffect(() => {
        fetch();
        const interval = setInterval(fetch, pollInterval);
        return () => clearInterval(interval);
    }, [fetch, pollInterval]);

    return { data, error, refresh: fetch };
}

export function useChannels(pollInterval = 5000) {
    const [data, setData] = useState<ChannelsResponse | null>(null);
    const [error, setError] = useState<string | null>(null);

    const fetch = useCallback(async () => {
        try {
            const status = await fetchJson<ChannelsResponse>('/channels');
            setData(status);
            setError(null);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Unknown error');
        }
    }, []);

    useEffect(() => {
        fetch();
        const interval = setInterval(fetch, pollInterval);
        return () => clearInterval(interval);
    }, [fetch, pollInterval]);

    return { data, error, refresh: fetch };
}

export function useEvents(pollInterval = 2000) {
    const [data, setData] = useState<EventsResponse | null>(null);
    const [error, setError] = useState<string | null>(null);

    const fetch = useCallback(async () => {
        try {
            const status = await fetchJson<EventsResponse>('/events');
            setData(status);
            setError(null);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Unknown error');
        }
    }, []);

    useEffect(() => {
        fetch();
        const interval = setInterval(fetch, pollInterval);
        return () => clearInterval(interval);
    }, [fetch, pollInterval]);

    return { data, error, refresh: fetch };
}

export function useSyncStatus(pollInterval = 2000) {
    const [data, setData] = useState<SyncStatus | null>(null);
    const [error, setError] = useState<string | null>(null);

    const fetch = useCallback(async () => {
        try {
            const status = await fetchJson<SyncStatus>('/sync/status');
            setData(status);
            setError(null);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Unknown error');
        }
    }, []);

    useEffect(() => {
        fetch();
        const interval = setInterval(fetch, pollInterval);
        return () => clearInterval(interval);
    }, [fetch, pollInterval]);

    return { data, error, refresh: fetch };
}

// Demo control actions
export async function goOffline(): Promise<{ status: string; message: string }> {
    const res = await fetch(`${API_BASE}/offline`, { method: 'POST' });
    return res.json();
}

export async function createPayment(from: string, to: string, amount: number): Promise<{ event: unknown; status: string; message: string }> {
    const res = await fetch(`${API_BASE}/payment`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ from_node: from, to_node: to, amount_sats: amount }),
    });
    return res.json();
}

export async function reconnect(): Promise<{ status: string; message: string; reconcile: unknown }> {
    const res = await fetch(`${API_BASE}/reconnect`, { method: 'POST' });
    return res.json();
}
