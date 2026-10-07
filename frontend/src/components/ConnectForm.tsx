import { useState } from 'react'
import type { ConnectionState } from '../types'

interface ConnectFormProps {
  connection: ConnectionState
  onConnectionChange: (next: ConnectionState) => void
}

const STATUS_LABEL: Record<ConnectionState, string> = {
  disconnected: 'Disconnected',
  connected: 'Connected',
  streaming: 'Streaming',
}

export function ConnectForm({ connection, onConnectionChange }: ConnectFormProps) {
  const [host, setHost] = useState('192.168.1.2')
  const [commandPort, setCommandPort] = useState('1024')
  const [streamPort, setStreamPort] = useState('47475')

  const isConnected = connection !== 'disconnected'

  const handleToggle = () => {
    // TODO: replace with real POST /api/connect and POST /api/disconnect.
    onConnectionChange(isConnected ? 'disconnected' : 'connected')
  }

  return (
    <section className="card connect-card">
      <h2>Connection</h2>
      <div className="field-row">
        <label htmlFor="host">Radar IP</label>
        <input
          id="host"
          value={host}
          onChange={(e) => setHost(e.target.value)}
          placeholder="192.168.1.2"
          disabled={isConnected}
        />
      </div>
      <div className="field-row two-col">
        <label htmlFor="command-port">Command port</label>
        <input
          id="command-port"
          value={commandPort}
          onChange={(e) => setCommandPort(e.target.value)}
          disabled={isConnected}
        />
      </div>
      <div className="field-row two-col">
        <label htmlFor="stream-port">Stream port</label>
        <input
          id="stream-port"
          value={streamPort}
          onChange={(e) => setStreamPort(e.target.value)}
          disabled={isConnected}
        />
      </div>
      <div className="connect-actions">
        <button className="btn" onClick={handleToggle}>
          {isConnected ? 'Disconnect' : 'Connect'}
        </button>
        <span className={`badge badge-${connection}`}>{STATUS_LABEL[connection]}</span>
      </div>
    </section>
  )
}
