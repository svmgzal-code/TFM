import type { RadarInfo } from '../types'

interface InfoPanelProps {
  info: RadarInfo
}

function Row({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="info-row">
      <span className="info-label">{label}</span>
      <span className="info-value">{value}</span>
    </div>
  )
}

export function InfoPanel({ info }: InfoPanelProps) {
  // TODO: replace with real GET /api/info data.
  return (
    <section className="card">
      <h2>Device info</h2>
      <div className="info-table">
        <Row label="Firmware" value={info.firmwareVersion} />
        <Row label="Revision" value={info.firmwareRevision} />
        <Row label="Radar number" value={info.radarNumber} />
        <Row label="Flashing date" value={info.flashingDate} />
        <Row label="SNT board" value={info.sntBoardVersion} />
        <Row label="Baseband board" value={info.basebandBoardVersion} />
        <Row label="Frontend board" value={info.frontendBoardVersion} />
        <Row label="Channels (mask)" value={`0x${info.availableChannels.toString(16).toUpperCase().padStart(4, '0')}`} />
        <Row label="Phase offset" value={`${info.phaseOffsetDeg}°`} />
      </div>
    </section>
  )
}
