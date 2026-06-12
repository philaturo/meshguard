// File: apps/dashboard/src/components/DemoControls.tsx
// Purpose: Three buttons: Go Offline / Create Payment / Reconnect
// Connects to: App.tsx (bottom row), hooks/useApi.ts (demo actions)

import React, { useState } from 'react'
import { goOffline, createPayment, reconnect } from '../hooks/useApi'

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

const buttonRow: React.CSSProperties = {
  display: 'flex',
  gap: '12px',
  flexWrap: 'wrap',
}

const button = (variant: 'danger' | 'primary' | 'success'): React.CSSProperties => ({
  padding: '10px 20px',
  borderRadius: '8px',
  border: 'none',
  fontSize: '13px',
  fontWeight: 600,
  cursor: 'pointer',
  color: '#fff',
  background: variant === 'danger' ? '#dc2626' : variant === 'primary' ? '#3b82f6' : '#059669',
  transition: 'opacity 0.2s',
})

const statusStyle: React.CSSProperties = {
  marginTop: '12px',
  padding: '8px',
  borderRadius: '6px',
  background: '#1e293b',
  fontSize: '12px',
  color: '#e2e8f0',
  minHeight: '20px',
}

export default function DemoControls() {
  const [status, setStatus] = useState<string>('')

  const handleOffline = async () => {
    setStatus('Disconnecting Alice...')
    try {
      const res = await goOffline()
      setStatus(res.message)
    } catch (err) {
      setStatus(`Error: ${err instanceof Error ? err.message : 'Unknown'}`)
    }
  }

  const handlePayment = async () => {
    setStatus('Creating payment...')
    try {
      const res = await createPayment('Alice', 'Bob', 10000)
      setStatus(res.message)
    } catch (err) {
      setStatus(`Error: ${err instanceof Error ? err.message : 'Unknown'}`)
    }
  }

  const handleReconnect = async () => {
    setStatus('Reconnecting Alice...')
    try {
      const res = await reconnect()
      setStatus(res.message)
    } catch (err) {
      setStatus(`Error: ${err instanceof Error ? err.message : 'Unknown'}`)
    }
  }

  return (
    <div style={cardStyle}>
      <div style={titleStyle}>Demo Controls</div>
      <div style={buttonRow}>
        <button style={button('danger')} onClick={handleOffline}>Go Offline</button>
        <button style={button('primary')} onClick={handlePayment}>Create Payment</button>
        <button style={button('success')} onClick={handleReconnect}>Reconnect</button>
      </div>
      <div style={statusStyle}>{status || 'Ready'}</div>
    </div>
  )
}
