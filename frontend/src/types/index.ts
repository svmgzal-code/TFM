export type ConnectionState = 'disconnected' | 'connected' | 'streaming'

export type MeasurementMode = 'fmcw_up' | 'fmcw_down' | 'cw'

export interface RadarStatus {
  connection: ConnectionState
  recording: boolean
  host: string
  commandPort: number
  streamPort: number
}

export interface RadarInfo {
  firmwareVersion: string
  firmwareRevision: number
  sntBoardVersion: string
  basebandBoardVersion: string
  frontendBoardVersion: string
  availableChannels: number
  availableAlgorithms: number
  radarHardwareId: number
  radarNumber: number
  flashingDate: string
  phaseOffsetDeg: number
}

export interface SystemParams {
  band: number
  tRampMs: number
  zeroPad: number
  fftDataType: number
  frontendEn: boolean
  powerSaveEn: boolean
  norm: boolean
  actChannels: number
  freqPoints: number
  minFreqMHz: number
  manualBWMHz: number
  attenuation: number
  rangeBinUm: number
  dopplerBinUmS: number
  frequencyBinHz: number
}

export interface OpParams {
  autoDelayMs: number
}

export interface DataSelection {
  temperatures: boolean
  timeDomain: boolean
  frequencyDomain: boolean
}

export interface Recording {
  id: string
  startedAt: string
  durationSec: number
  frameCount: number
  sizeBytes: number
}
