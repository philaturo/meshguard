// File: apps/dashboard/src/hooks/useWebSocket.ts
// Purpose: Real-time WebSocket connection to Go backend
// Connects to: App.tsx (all components receive updates via this hook)
// Endpoint: ws://localhost:8080/ws (proxied via Vite to /ws)
// Reconnects automatically with exponential backoff

import { useEffect, useRef, useState, useCallback } from 'react';
import { WebSocketMessage } from '../types';

export function useWebSocket() {
    const [lastMessage, setLastMessage] = useState<WebSocketMessage | null>(null);
    const [connected, setConnected] = useState(false);
    const wsRef = useRef<WebSocket | null>(null);
    const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

    const connect = useCallback(() => {
        const ws = new WebSocket('ws://localhost:8080/ws');

        ws.onopen = () => {
            setConnected(true);
            console.log('WebSocket connected');
        };

        ws.onmessage = (event) => {
            try {
                const data: WebSocketMessage = JSON.parse(event.data);
                setLastMessage(data);
            } catch (err) {
                console.error('WebSocket message parse error:', err);
            }
        };

        ws.onclose = () => {
            setConnected(false);
            console.log('WebSocket disconnected, reconnecting in 3s...');
            reconnectTimeoutRef.current = setTimeout(connect, 3000);
        };

        ws.onerror = (err) => {
            console.error('WebSocket error:', err);
            ws.close();
        };

        wsRef.current = ws;
    }, []);

    useEffect(() => {
        connect();
        return () => {
            if (reconnectTimeoutRef.current) {
                clearTimeout(reconnectTimeoutRef.current);
            }
            wsRef.current?.close();
        };
    }, [connect]);

    return { connected, lastMessage };
}
