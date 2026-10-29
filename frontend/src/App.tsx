import { useCallback, useState } from 'react'
import { ConnectForm } from './components/ConnectForm'
import { StatusBar } from './components/StatusBar'
import { InfoPanel } from './components/InfoPanel'
import { ParamsForm } from './components/ParamsForm'
import { LiveChart } from './components/LiveChart'
import { RecordControls } from './components/RecordControls'
import { useTheme } from './hooks/useTheme'
import { useStatus } from './hooks/useStatus'
import { useParams } from './hooks/useParams'
import { connect, disconnect } from './api/client'
import type { MeasurementMode, RadarInfo, Recording } from './types'

// Hardcoded mock data — replaced by real backend data in a later slice.
// TODO: replace with real GET /api/info, GET /api/recordings.
const MOCK_INFO: RadarInfo = {
  firmwareVersion: '4.1.2',
  firmwareRevision: 104,
  sntBoardVersion: '0x00020100',
  basebandBoardVersion: '0x00030001',
  frontendBoardVersion: '0x00010002',
  availableChannels: 0x000f,
  availableAlgorithms: 0x0003,
  radarHardwareId: 0x0001,
  radarNumber: 0x0a1b2c3d,
  flashingDate: '20150301',
  phaseOffsetDeg: 0,
}

const MOCK_RECORDINGS: Recording[] = [
  {
    id: '2026-10-06T14-32-10Z',
    startedAt: '2026-10-06 14:32:10',
    durationSec: 62,
    frameCount: 6150,
    sizeBytes: 24_500_000,
  },
  {
    id: '2026-10-05T09-11-47Z',
    startedAt: '2026-10-05 09:11:47',
    durationSec: 180,
    frameCount: 17920,
    sizeBytes: 71_800_000,
  },
]

export default function App() {
  const [recording, setRecording] = useState(false)
  const [mode, setMode] = useState<MeasurementMode>('fmcw_up')
  const { theme, toggle } = useTheme()
  const { status, refresh } = useStatus()
  const { params, apply, refresh: refreshParams } = useParams()

  const handleConnect = useCallback(
    async (host: string, commandPort: number) => {
      await connect({ host, commandPort, streamPort: 47475 })
      await refresh()
      await refreshParams()
    },
    [refresh, refreshParams],
  )

  const handleDisconnect = useCallback(async () => {
    await disconnect()
    await refresh()
    await refreshParams()
  }, [refresh, refreshParams])

  const chartMode = mode === 'cw' ? 'cw' : 'fmcw'

  return (
    <div className="app">
      <StatusBar
        connection={status.connection}
        recording={recording}
        host={status.host || '—'}
        commandPort={status.commandPort}
        streamPort={status.streamPort}
        theme={theme}
        onToggleTheme={toggle}
      />
      <main className="layout">
        <div className="left-column">
          <ConnectForm
            connection={status.connection}
            onConnect={handleConnect}
            onDisconnect={handleDisconnect}
          />
          {params ? (
            <ParamsForm params={params} mode={mode} onModeChange={setMode} onApply={apply} />
          ) : (
            <section className="card">
              <h2>Parameters</h2>
              <p className="empty-note">Connect to a radar to edit parameters.</p>
            </section>
          )}
        </div>
        <div className="right-column">
          <LiveChart mode={chartMode} theme={theme} />
          <div className="bottom-row">
            <InfoPanel info={MOCK_INFO} />
            <RecordControls
              recording={recording}
              recordings={MOCK_RECORDINGS}
              onRecordingChange={setRecording}
            />
          </div>
        </div>
      </main>
    </div>
  )
}
