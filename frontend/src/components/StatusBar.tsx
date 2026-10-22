import type { ConnectionState } from '../types'
import type { Theme } from '../hooks/useTheme'
import { ThemeToggle } from './ThemeToggle'

interface StatusBarProps {
  connection: ConnectionState
  recording: boolean
  host: string
  commandPort: number
  streamPort: number
  theme: Theme
  onToggleTheme: () => void
}

const CONNECTION_LABEL: Record<ConnectionState, string> = {
  disconnected: 'Disconnected',
  connected: 'Connected',
  streaming: 'Streaming',
}

export function StatusBar({
  connection,
  recording,
  host,
  commandPort,
  streamPort,
  theme,
  onToggleTheme,
}: StatusBarProps) {
  return (
    <div className="status-bar">
      <div className="status-bar-left">
        <span className="app-title">sR-1200 Radar Console</span>
        <span className="status-meta">
          {host}:{commandPort} · stream {streamPort}
        </span>
      </div>
      <div className="status-bar-right">
        <span className={`badge badge-${connection}`}>{CONNECTION_LABEL[connection]}</span>
        {recording && <span className="badge badge-recording">● Recording</span>}
        <ThemeToggle theme={theme} onToggle={onToggleTheme} />
      </div>
    </div>
  )
}
