import { useState } from 'react'
import type { ConnectionState } from '../types'

interface ConnectFormProps {
  connection: ConnectionState
  onConnect: (host: string, commandPort: number) => Promise<void>
  onDisconnect: () => Promise<void>
}

const STATUS_LABEL: Record<ConnectionState, string> = {
  disconnected: 'Disconnected',
  connected: 'Connected',
  streaming: 'Streaming',
}

export function ConnectForm({ connection, onConnect, onDisconnect }: ConnectFormProps) {
  const [host, setHost] = useState('192.168.1.2')
  const [commandPort, setCommandPort] = useState('1024')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const isConnected = connection !== 'disconnected'

  const handleToggle = async () => {
    setError(null)
    setBusy(true)
    try {
      if (isConnected) {
        await onDisconnect()
      } else {
        await onConnect(host, Number(commandPort))
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setBusy(false)
    }
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
          disabled={isConnected || busy}
        />
      </div>
      <div className="field-row two-col">
        <label htmlFor="command-port">Command port</label>
        <input
          id="command-port"
          value={commandPort}
          onChange={(e) => setCommandPort(e.target.value)}
          disabled={isConnected || busy}
        />
      </div>
      {error && <div className="warning-banner">{error}</div>}
      <div className="connect-actions">
        <button className="btn" onClick={handleToggle} disabled={busy}>
          {busy ? 'Working…' : isConnected ? 'Disconnect' : 'Connect'}
        </button>
        <span className={`badge badge-${connection}`}>{STATUS_LABEL[connection]}</span>
      </div>
    </section>
  )
}
