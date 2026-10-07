import type { Recording } from '../types'

interface RecordControlsProps {
  recording: boolean
  recordings: Recording[]
  onRecordingChange: (recording: boolean) => void
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export function RecordControls({
  recording,
  recordings,
  onRecordingChange,
}: RecordControlsProps) {
  return (
    <section className="card">
      <div className="record-header">
        <h2>Recording</h2>
        {recording && <span className="recording-indicator">● REC</span>}
      </div>

      <button
        className={`btn ${recording ? 'btn-stop' : 'btn-record'}`}
        onClick={() => onRecordingChange(!recording)}
      >
        {recording ? 'Stop recording' : 'Start recording'}
      </button>

      <div className="recordings-list">
        <h3>Past recordings</h3>
        {recordings.length === 0 ? (
          <p className="empty-note">No recordings yet.</p>
        ) : (
          <ul>
            {recordings.map((rec) => (
              <li key={rec.id} className="recording-item">
                <div className="recording-meta">
                  <span className="recording-name">{rec.id}</span>
                  <span className="recording-date">{rec.startedAt}</span>
                </div>
                <span className="recording-stats">
                  {rec.durationSec}s · {rec.frameCount.toLocaleString()} frames ·{' '}
                  {formatBytes(rec.sizeBytes)}
                </span>
              </li>
            ))}
          </ul>
        )}
      </div>
    </section>
  )
}
