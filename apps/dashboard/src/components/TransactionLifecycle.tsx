// File: apps/dashboard/src/components/TransactionLifecycle.tsx
// Purpose: 5-state timeline (CREATED -> OFFLINE -> QUEUED -> RECONCILING -> SETTLED)
// Connects to: App.tsx (left column), hooks/useApi.ts (events, sync)

import React from 'react'
import { MeshEvent, SyncStatus } from '../types'

interface Props {
  events: MeshEvent[]
  sync: SyncStatus | null
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
  marginBottom: '16px',
}

const stepContainer: React.CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'space-between',
  marginBottom: '12px',
}

const step = (active: boolean, completed: boolean): React.CSSProperties => ({
  display: 'flex',
  flexDirection: 'column',
  alignItems: 'center',
  gap: '4px',
})

const circle = (active: boolean, completed: boolean): React.CSSProperties => ({
  width: '32px',
  height: '32px',
  borderRadius: '50%',
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
  fontSize: '12px',
  fontWeight: 700,
  background: completed ? '#059669' : active ? '#3b82f6' : '#1e293b',
  color: '#fff',
  border: active ? '2px solid #60a5fa' : '2px solid transparent',
})

const line = (completed: boolean): React.CSSProperties => ({
  flex: 1,
  height: '2px',
  background: completed ? '#059669' : '#1e293b',
  margin: '0 8px',
  marginBottom: '16px',
})

const labelStyle: React.CSSProperties = {
  fontSize: '10px',
  color: '#64748b',
  textTransform: 'uppercase',
}

export default function TransactionLifecycle({ events, sync }: Props) {

const activeEvent = events.find(e => e.status !== 'SETTLED' && e.status !== 'FAILED')


  const currentStatus = activeEvent?.status ?? (sync?.status === 'offline' ? 'OFFLINE' : 'CREATED')

  const steps = ['CREATED', 'OFFLINE', 'QUEUED', 'RECONCILING', 'SETTLED']
  const stepIndex = steps.indexOf(currentStatus)

  return (
    <div style={cardStyle}>
      <div style={titleStyle}>Transaction Lifecycle</div>
      <div style={stepContainer}>
        {steps.map((s, i) => (
          <React.Fragment key={s}>
            <div style={step(i === stepIndex, i < stepIndex)}>
              <div style={circle(i === stepIndex, i < stepIndex)}>{i + 1}</div>
              <div style={labelStyle}>{s}</div>
            </div>
            {i < steps.length - 1 && <div style={line(i < stepIndex)} />}
          </React.Fragment>
        ))}
      </div>
      {activeEvent && (
        <div style={{ marginTop: '8px', padding: '8px', background: '#1e293b', borderRadius: '6px', fontSize: '12px' }}>
          <div style={{ color: '#e2e8f0', fontWeight: 600 }}>Event {activeEvent.id}</div>
          <div style={{ color: '#64748b' }}>{activeEvent.amount} sats {activeEvent.from} {'->'} {activeEvent.to}</div>
          <div style={{ color: activeEvent.status === 'FAILED' ? '#dc2626' : '#3b82f6', fontWeight: 600 }}>
            Status: {activeEvent.status}
          </div>
        </div>
      )}
      {!activeEvent && (
        <div style={{ textAlign: 'center', color: '#64748b', fontSize: '12px', padding: '12px' }}>
          No active transactions. Click "Create Payment" to begin.
        </div>
      )}
    </div>
  )
}
