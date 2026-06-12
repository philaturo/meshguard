// File: apps/dashboard/src/components/EventQueue.tsx
// Purpose: Pending/Settled/Failed events table with counts
// Connects to: App.tsx (right column), hooks/useApi.ts (events)

import React from 'react'
import { EventsResponse } from '../types'

interface Props {
  data: EventsResponse | null
}

const cardStyle: React.CSSProperties = {
  background: '#111827',
  borderRadius: '12px',
  padding: '16px',
  border: '1px solid #1e293b',
}

const titleStyle: React.CSSProperties = {
  fontSize: '14px',
  fontWeight: 600,
  color: '#94a3b8',
  textTransform: 'uppercase',
  letterSpacing: '0.5px',
  marginBottom: '12px',
}

const countsRow: React.CSSProperties = {
  display: 'grid',
  gridTemplateColumns: 'repeat(3, 1fr)',
  gap: '8px',
  marginBottom: '12px',
}

const countCard = (color: string): React.CSSProperties => ({
  background: '#1e293b',
  borderRadius: '8px',
  padding: '8px',
  textAlign: 'center',
  borderLeft: `3px solid ${color}`,
})

const countValue: React.CSSProperties = {
  fontSize: '20px',
  fontWeight: 700,
  color: '#f8fafc',
}

const countLabel: React.CSSProperties = {
  fontSize: '10px',
  color: '#64748b',
  textTransform: 'uppercase',
}

const tableStyle: React.CSSProperties = {
  width: '100%',
  fontSize: '12px',
  borderCollapse: 'collapse',
}

const thStyle: React.CSSProperties = {
  textAlign: 'left',
  padding: '6px 4px',
  color: '#64748b',
  borderBottom: '1px solid #1e293b',
  fontWeight: 600,
}

const tdStyle: React.CSSProperties = {
  padding: '6px 4px',
  color: '#e2e8f0',
  borderBottom: '1px solid #1e293b',
}

const statusColor: Record<string, string> = {
  CREATED: '#6b7280',
  OFFLINE: '#f59e0b',
  QUEUED: '#3b82f6',
  RECONCILING: '#8b5cf6',
  SETTLED: '#059669',
  FAILED: '#dc2626',
}

export default function EventQueue({ data }: Props) {
  const counts = data?.counts ?? { pending: 0, settled: 0, failed: 0 }
  const events = data?.events?.slice(0, 5) ?? []

  return (
    <div style={cardStyle}>
      <div style={titleStyle}>Event Queue</div>
      <div style={countsRow}>
        <div style={countCard('#f59e0b')}>
          <div style={countValue}>{counts.pending}</div>
          <div style={countLabel}>Pending</div>
        </div>
        <div style={countCard('#059669')}>
          <div style={countValue}>{counts.settled}</div>
          <div style={countLabel}>Settled</div>
        </div>
        <div style={countCard('#dc2626')}>
          <div style={countValue}>{counts.failed}</div>
          <div style={countLabel}>Failed</div>
        </div>
      </div>
      <table style={tableStyle}>
        <thead>
          <tr>
            <th style={thStyle}>ID</th>
            <th style={thStyle}>Type</th>
            <th style={thStyle}>Status</th>
            <th style={thStyle}>Amount</th>
          </tr>
        </thead>
        <tbody>
          {events.map(e => (
            <tr key={e.id}>
              <td style={tdStyle}>{e.id.slice(0, 8)}...</td>
              <td style={tdStyle}>{e.type}</td>
              <td style={tdStyle}>
                <span style={{ color: statusColor[e.status] ?? '#64748b', fontWeight: 600 }}>{e.status}</span>
              </td>
              <td style={tdStyle}>{e.amount} sats</td>
            </tr>
          ))}
          {events.length === 0 && (
            <tr>
              <td colSpan={4} style={{ ...tdStyle, textAlign: 'center', color: '#64748b' }}>
                No events recorded
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  )
}
