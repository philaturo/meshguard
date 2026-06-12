// File: apps/dashboard/src/components/ChannelState.tsx
// Purpose: Real channel visualization (capacity, local/remote balance, active status)
// Connects to: App.tsx (full width row), hooks/useApi.ts (channels)

import React from 'react'
import { ChannelsResponse } from '../types'

interface Props {
  data: ChannelsResponse | null
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
  marginBottom: '16px',
}

const titleStyle: React.CSSProperties = {
  fontSize: '14px',
  fontWeight: 600,
  color: '#94a3b8',
  textTransform: 'uppercase',
  letterSpacing: '0.5px',
}

const statusBadge = (active: boolean): React.CSSProperties => ({
  fontSize: '11px',
  fontWeight: 600,
  padding: '2px 8px',
  borderRadius: '6px',
  background: active ? '#059669' : '#dc2626',
  color: '#fff',
})

const barContainer: React.CSSProperties = {
  height: '24px',
  background: '#1e293b',
  borderRadius: '6px',
  overflow: 'hidden',
  display: 'flex',
  marginBottom: '8px',
}

const barSegment = (color: string, width: string): React.CSSProperties => ({
  height: '100%',
  width,
  background: color,
  transition: 'width 0.3s ease',
})

const labelRow: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  fontSize: '12px',
  color: '#64748b',
}

const valueStyle: React.CSSProperties = {
  fontSize: '18px',
  fontWeight: 700,
  color: '#f8fafc',
}

export default function ChannelState({ data }: Props) {
  const active = data?.status === 'active'
  const channel = data?.channels?.[0]

  const localSats = channel?.local_balance ?? 0
  const remoteSats = channel?.remote_balance ?? 0
  const capacity = channel?.capacity ?? 1
  const localPct = capacity > 0 ? (localSats / capacity) * 100 : 0
  const remotePct = capacity > 0 ? (remoteSats / capacity) * 100 : 0

  return (
    <div style={cardStyle}>
      <div style={headerStyle}>
        <div style={titleStyle}>Channel State</div>
        <div style={statusBadge(active)}>{active ? 'ACTIVE' : 'WAITING'}</div>
      </div>
      {channel ? (
        <>
          <div style={barContainer}>
            <div style={barSegment('#3b82f6', `${localPct}%`)} />
            <div style={barSegment('#8b5cf6', `${remotePct}%`)} />
          </div>
          <div style={labelRow}>
            <div>Alice: {(localSats / 1e8).toFixed(4)} BTC ({localPct.toFixed(1)}%)</div>
            <div>Bob: {(remoteSats / 1e8).toFixed(4)} BTC ({remotePct.toFixed(1)}%)</div>
          </div>
          <div style={{ marginTop: '12px', display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '12px' }}>
            <div>
              <div style={valueStyle}>{(capacity / 1e8).toFixed(4)} BTC</div>
              <div style={{ fontSize: '12px', color: '#64748b' }}>Capacity</div>
            </div>
            <div>
              <div style={valueStyle}>{channel.channel_id}</div>
              <div style={{ fontSize: '12px', color: '#64748b' }}>Channel ID</div>
            </div>
            <div>
              <div style={valueStyle}>{truncate(channel.remote_pubkey, 16)}</div>
              <div style={{ fontSize: '12px', color: '#64748b' }}>Remote Pubkey</div>
            </div>
          </div>
        </>
      ) : (
        <div style={{ fontSize: '14px', color: '#64748b', textAlign: 'center', padding: '20px' }}>
          {data?.message ?? 'No active channels. Complete bootcamp Day 3 to open Alice-Bob channel.'}
        </div>
      )}
    </div>
  )
}

function truncate(s: string, len: number): string {
  return s.length > len ? s.slice(0, len) + '...' : s
}
