// File: apps/dashboard/src/components/BitcoinCore.tsx
// Purpose: Live blockchain widget (block height, mempool, network)
// Connects to: App.tsx (grid layout), hooks/useApi.ts (bitcoin status)

import React from 'react'
import { BlockchainStatus } from '../types'

interface Props {
  data: BlockchainStatus | null
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

const statusStyle = (online: boolean): React.CSSProperties => ({
  width: '8px',
  height: '8px',
  borderRadius: '50%',
  background: online ? '#059669' : '#dc2626',
})

const valueStyle: React.CSSProperties = {
  fontSize: '24px',
  fontWeight: 700,
  color: '#f8fafc',
  marginBottom: '4px',
}

const labelStyle: React.CSSProperties = {
  fontSize: '12px',
  color: '#64748b',
}

export default function BitcoinCore({ data }: Props) {
  const online = data?.status === 'online'

  return (
    <div style={cardStyle}>
      <div style={headerStyle}>
        <div style={titleStyle}>Bitcoin Core</div>
        <div style={statusStyle(online)} />
      </div>
      <div style={valueStyle}>{online ? data?.height ?? '—' : '—'}</div>
      <div style={labelStyle}>Block Height</div>
      <div style={{ marginTop: '12px', display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '8px' }}>
        <div>
          <div style={{ fontSize: '16px', fontWeight: 600, color: '#e2e8f0' }}>{online ? data?.mempool_size ?? 0 : '—'}</div>
          <div style={labelStyle}>Mempool</div>
        </div>
        <div>
          <div style={{ fontSize: '16px', fontWeight: 600, color: '#e2e8f0' }}>{online ? data?.network ?? '—' : '—'}</div>
          <div style={labelStyle}>Network</div>
        </div>
      </div>
      {!online && (
        <div style={{ marginTop: '8px', fontSize: '11px', color: '#dc2626' }}>
          {data?.message ?? 'Waiting for node...'}
        </div>
      )}
    </div>
  )
}
