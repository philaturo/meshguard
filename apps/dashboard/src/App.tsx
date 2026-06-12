// File: apps/dashboard/src/App.tsx
// Purpose: Root layout — single-screen MVP dashboard
// Connects to: All components in src/components/, hooks/useWebSocket.ts
// Layout: Top bar -> Node cards -> Channel state -> Transaction lifecycle + Event queue -> Controls
// No navigation, no sidebars, no settings. One screen tells the entire story.

import React from 'react'
import BitcoinCore from './components/BitcoinCore'
import NodeCard from './components/NodeCard'
import ChannelState from './components/ChannelState'
import TransactionLifecycle from './components/TransactionLifecycle'
import EventQueue from './components/EventQueue'
import DemoControls from './components/DemoControls'
import SystemStatus from './components/SystemStatus'
import { useWebSocket } from './hooks/useWebSocket'
import { useBitcoinStatus, useNodesStatus, useChannels, useEvents, useSyncStatus } from './hooks/useApi'

const styles: Record<string, React.CSSProperties> = {
  container: {
    minHeight: '100vh',
    background: '#0a0e1a',
    color: '#e0e6ed',
    fontFamily: "'Segoe UI', system-ui, sans-serif",
    padding: '20px',
    boxSizing: 'border-box',
  },
  header: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginBottom: '24px',
    paddingBottom: '16px',
    borderBottom: '1px solid #1e293b',
  },
  title: {
    fontSize: '28px',
    fontWeight: 700,
    color: '#f8fafc',
    margin: 0,
    letterSpacing: '-0.5px',
  },
  subtitle: {
    fontSize: '13px',
    color: '#64748b',
    marginTop: '4px',
    fontWeight: 400,
  },
  badge: {
    background: '#059669',
    color: '#fff',
    padding: '4px 12px',
    borderRadius: '12px',
    fontSize: '11px',
    fontWeight: 600,
    textTransform: 'uppercase',
    letterSpacing: '0.5px',
  },
  grid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(3, 1fr)',
    gap: '16px',
    marginBottom: '20px',
  },
  row: {
    display: 'grid',
    gridTemplateColumns: '1fr 1fr',
    gap: '16px',
    marginBottom: '20px',
  },
  fullWidth: {
    marginBottom: '20px',
  },
}

export default function App() {
  const { connected: wsConnected } = useWebSocket()
  const { data: bitcoin } = useBitcoinStatus()
  const { data: nodes } = useNodesStatus()
  const { data: channels } = useChannels()
  const { data: events } = useEvents()
  const { data: sync } = useSyncStatus()

  return (
    <div style={styles.container}>
      <div style={styles.header}>
        <div>
          <h1 style={styles.title}>MeshGuard</h1>
          <div style={styles.subtitle}>Built for Unreliable Networks</div>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
          <SystemStatus wsConnected={wsConnected} sync={sync} />
        </div>
      </div>

      <div style={styles.grid}>
        <BitcoinCore data={bitcoin} />
        <NodeCard title="Alice" data={nodes?.alice} />
        <NodeCard title="Bob" data={nodes?.bob} />
      </div>

      <div style={styles.fullWidth}>
        <ChannelState data={channels} />
      </div>

      <div style={styles.row}>
        <TransactionLifecycle events={events?.events || []} sync={sync} />
        <EventQueue data={events} />
      </div>

      <div style={styles.fullWidth}>
        <DemoControls />
      </div>
    </div>
  )
}
