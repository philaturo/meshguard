// File: apps/dashboard/src/components/NodeCard.tsx
// Purpose: Alice/Bob status cards (alias, balance, channels, pubkey)
// Connects to: App.tsx (grid layout), hooks/useApi.ts (nodes status)

import React from 'react'
import { NodeStatus } from '../types'

interface Props {
  title: string
  data: NodeStatus | undefined
}

const cardStyle: React.CSSProperties = {
  background: '#111827',
  borderRadius: '12px',
  padding: '16px',
  border: '1px solid #1e293b',
}

const headerStyle: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'space-between',
  marginBottom: '12px',
}

const titleStyle: React.CSSProperties = {
  fontSize: '14px',
  fontWeight: 600,
  color: '#94a3b8',
  textTransform: 'uppercase',
  letterSpacing: '0.5px',
}

const statusBadge = (status: string): React.CSSProperties => ({
  fontSize: '11px',
  fontWeight: 600,
  padding: '2px 8px',
  borderRadius: '6px',
  background: status === 'online' ? '#059669' : status === 'error' ? '#dc2626' : '#6b7280',
  color: '#fff',
})

const valueStyle: React.CSSProperties = {
  fontSize: '20px',
  fontWeight: 700,
  color: '#f8fafc',
  marginBottom: '4px',
}

const labelStyle: React.CSSProperties = {
  fontSize: '12px',
  color: '#64748b',
}

const truncate = (s: string | undefined, len = 16): string => {
  if (!s) return '—'
  return s.length > len ? s.slice(0, len) + '...' : s
}

export default function NodeCard({ title, data }: Props) {
  const status = data?.status ?? 'waiting'

  return (
    <div style={cardStyle}>
      <div style={headerStyle}>
        <div style={titleStyle}>{title}</div>
        <div style={statusBadge(status)}>{status}</div>
      </div>
      <div style={valueStyle}>{truncate(data?.alias, 12)}</div>
      <div style={labelStyle}>Alias</div>
      <div style={{ marginTop: '12px', display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '8px' }}>
        <div>
          <div style={{ fontSize: '14px', fontWeight: 600, color: '#e2e8f0' }}>
            {data?.balance_total !== undefined ? `${(data.balance_total / 1e8).toFixed(4)} BTC` : '—'}
          </div>
          <div style={labelStyle}>Balance</div>
        </div>
        <div>
          <div style={{ fontSize: '14px', fontWeight: 600, color: '#e2e8f0' }}>{data?.channels ?? '—'}</div>
          <div style={labelStyle}>Channels</div>
        </div>
      </div>
      {data?.pubkey && (
        <div style={{ marginTop: '8px', fontSize: '10px', color: '#475569', fontFamily: 'monospace' }}>
          {truncate(data.pubkey, 24)}
        </div>
      )}
      {status !== 'online' && (
        <div style={{ marginTop: '8px', fontSize: '11px', color: '#dc2626' }}>
          {data?.message ?? 'Waiting for node...'}
        </div>
      )}
    </div>
  )
}
