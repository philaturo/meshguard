// File: apps/dashboard/src/components/SystemStatus.tsx
// Purpose: Top-right health indicators (WebSocket, Engine, Queue)
// Connects to: App.tsx (header), hooks/useWebSocket.ts and useApi.ts

import React from 'react'
import { SyncStatus } from '../types'

interface Props {
  wsConnected: boolean
  sync: SyncStatus | null
}

const badge = (online: boolean): React.CSSProperties => ({
  display: 'flex',
  alignItems: 'center',
  gap: '6px',
  padding: '4px 10px',
  borderRadius: '6px',
  background: online ? '#05966920' : '#dc262620',
  fontSize: '11px',
  fontWeight: 600,
  color: online ? '#059669' : '#dc2626',
  border: `1px solid ${online ? '#05966940' : '#dc262640'}`,
})

const dot = (online: boolean): React.CSSProperties => ({
  width: '6px',
  height: '6px',
  borderRadius: '50%',
  background: online ? '#059669' : '#dc2626',
})

export default function SystemStatus({ wsConnected, sync }: Props) {
  return (
    <div style={{ display: 'flex', gap: '8px' }}>
      <div style={badge(wsConnected)}>
        <div style={dot(wsConnected)} />
        WebSocket
      </div>
      <div style={badge(sync?.status === 'online')}>
        <div style={dot(sync?.status === 'online')} />
        Engine
      </div>
    </div>
  )
}
