import { useState } from 'react'
import type { DataSelection, MeasurementMode, SystemParams } from '../types'

const MODES: { value: MeasurementMode; label: string }[] = [
  { value: 'fmcw_up', label: 'FMCW up-ramp' },
  { value: 'fmcw_down', label: 'FMCW down-ramp' },
  { value: 'cw', label: 'CW' },
]

const BANDS = Array.from({ length: 14 }, (_, i) => i)
const ZERO_PADS = [1, 2, 4, 8]
const ATTENUATIONS = [
  { value: 0, label: '0 dB' },
  { value: 1, label: '0.4 dB' },
  { value: 2, label: '0.8 dB' },
  { value: 3, label: '1.4 dB' },
  { value: 4, label: '2.5 dB' },
  { value: 5, label: '4 dB' },
  { value: 6, label: '6 dB' },
  { value: 7, label: '9 dB' },
]

interface ParamsFormProps {
  params: SystemParams
  mode: MeasurementMode
  selection: DataSelection
  onModeChange: (mode: MeasurementMode) => void
  onSelectionChange: (selection: DataSelection) => void
}

export function ParamsForm({
  params,
  mode,
  selection,
  onModeChange,
  onSelectionChange,
}: ParamsFormProps) {
  const [band, setBand] = useState(params.band)
  const [zeroPad, setZeroPad] = useState(params.zeroPad)
  const [freqPoints, setFreqPoints] = useState(String(params.freqPoints))
  const [attenuation, setAttenuation] = useState(params.attenuation)

  const outOfEuBand = band > 4

  const handleApply = () => {
    // TODO: replace with real PUT /api/params (read-modify-write: 0x0029 -> 0x0028 -> 0x0020).
    void 0
  }

  return (
    <section className="card">
      <h2>Parameters</h2>

      <fieldset className="group">
        <legend>Measurement mode</legend>
        <div className="segmented">
          {MODES.map((m) => (
            <button
              key={m.value}
              type="button"
              className={`segment ${mode === m.value ? 'segment-active' : ''}`}
              onClick={() => onModeChange(m.value)}
            >
              {m.label}
            </button>
          ))}
        </div>
      </fieldset>

      <div className="field-row">
        <label htmlFor="band">Band</label>
        <select id="band" value={band} onChange={(e) => setBand(Number(e.target.value))}>
          {BANDS.map((b) => (
            <option key={b} value={b}>
              {b}
              {b <= 4 ? ' (EU)' : ''}
            </option>
          ))}
        </select>
      </div>

      {outOfEuBand && (
        <div className="warning-banner">
          Band {band} exceeds the 24.000–24.250 GHz license-free EU range. Confirm before
          applying.
        </div>
      )}

      <div className="field-row">
        <label htmlFor="zero-pad">Zero padding</label>
        <select id="zero-pad" value={zeroPad} onChange={(e) => setZeroPad(Number(e.target.value))}>
          {ZERO_PADS.map((z) => (
            <option key={z} value={z}>
              {z}
            </option>
          ))}
        </select>
      </div>

      <div className="field-row">
        <label htmlFor="freq-points">Frequency points</label>
        <input
          id="freq-points"
          type="number"
          min={5}
          max={513}
          value={freqPoints}
          onChange={(e) => setFreqPoints(e.target.value)}
        />
      </div>

      <div className="field-row">
        <label htmlFor="attenuation">Attenuation</label>
        <select
          id="attenuation"
          value={attenuation}
          onChange={(e) => setAttenuation(Number(e.target.value))}
        >
          {ATTENUATIONS.map((a) => (
            <option key={a.value} value={a.value}>
              {a.label}
            </option>
          ))}
        </select>
      </div>

      <fieldset className="group">
        <legend>Stream data</legend>
        <label className="checkbox-row">
          <input
            type="checkbox"
            checked={selection.temperatures}
            onChange={(e) => onSelectionChange({ ...selection, temperatures: e.target.checked })}
          />
          Temperatures &amp; Tx power
        </label>
        <label className="checkbox-row">
          <input
            type="checkbox"
            checked={selection.timeDomain}
            onChange={(e) => onSelectionChange({ ...selection, timeDomain: e.target.checked })}
          />
          Time-domain data
        </label>
        <label className="checkbox-row">
          <input
            type="checkbox"
            checked={selection.frequencyDomain}
            onChange={(e) => onSelectionChange({ ...selection, frequencyDomain: e.target.checked })}
          />
          Frequency-domain data
        </label>
      </fieldset>

      <div className="derived-row">
        <span className="info-label">Range bin</span>
        <span className="info-value">{params.rangeBinUm} µm</span>
        <span className="info-label">Doppler bin</span>
        <span className="info-value">{params.dopplerBinUmS} µm/s</span>
      </div>

      <button className="btn btn-primary" onClick={handleApply}>
        Apply settings
      </button>
    </section>
  )
}
