import { useState } from 'react'
import { ConnectForm } from './components/ConnectForm'
import { StatusBar } from './components/StatusBar'
import { InfoPanel } from './components/InfoPanel'
import { ParamsForm } from './components/ParamsForm'
import { LiveChart } from './components/LiveChart'
import { RecordControls } from './components/RecordControls'
import type {
  ConnectionState,
  DataSelection,
  MeasurementMode,
  RadarInfo,
  Recording,
  SystemParams,
} from './types'

// Hardcoded mock data — replaced by real backend data in a later slice.
// TODO: replace with real GET /api/info, GET /api/params, GET /api/recordings.
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

const MOCK_PARAMS: SystemParams = {
  band: 4,
  tRampMs: 2,
  zeroPad: 2,
  fftDataType: 0,
  frontendEn: true,
  powerSaveEn: false,
  norm: true,
  actChannels: 0x000f,
  freqPoints: 256,
  minFreqMHz: 0,
  manualBWMHz: 0,
  attenuation: 0,
  rangeBinUm: 599600,
  dopplerBinUmS: 0,
  frequencyBinHz: 0,
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
  const [connection, setConnection] = useState<ConnectionState>('disconnected')
  const [recording, setRecording] = useState(false)
  const [mode, setMode] = useState<MeasurementMode>('fmcw_up')
  const [selection, setSelection] = useState<DataSelection>({
    temperatures: true,
    timeDomain: false,
    frequencyDomain: true,
  })

  const chartMode = mode === 'cw' ? 'cw' : 'fmcw'

  return (
    <div className="app">
      <StatusBar
        connection={connection}
        recording={recording}
        host="192.168.1.2"
        commandPort={1024}
        streamPort={47475}
      />
      <main className="layout">
        <div className="left-column">
          <ConnectForm connection={connection} onConnectionChange={setConnection} />
          <ParamsForm
            params={MOCK_PARAMS}
            mode={mode}
            selection={selection}
            onModeChange={setMode}
            onSelectionChange={setSelection}
          />
        </div>
        <div className="right-column">
          <LiveChart mode={chartMode} />
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
