import { useEffect, useState } from 'react'
import type { MeasurementMode, ParamsUpdate, SystemParams } from '../types'

const MODES: { value: MeasurementMode; label: string }[] = [
  { value: 'fmcw_up', label: 'FMCW up-ramp' },
  { value: 'fmcw_down', label: 'FMCW down-ramp' },
  { value: 'cw', label: 'CW' },
]

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

const CHANNELS = [1, 2, 3, 4]

// Band table from docs/protocol.md section 5: [start MHz, stop MHz].
const BAND_RANGE_MHZ: [number, number][] = [
  [24000, 24050],
  [24000, 24075],
  [24000, 24125],
  [24000, 24230],
  [24000, 24250],
  [24000, 24500],
  [24000, 24800],
  [24000, 25000],
  [24000, 25200],
  [24000, 25500],
  [24000, 25800],
  [23500, 25500],
  [23500, 25700],
  [23200, 25700],
]

const MIN_FREQ_MHZ = 23000
const MAX_FREQ_MHZ = 26000
const EU_STOP_MHZ = 24250

interface ParamsFormProps {
  params: SystemParams
  mode: MeasurementMode
  onModeChange: (mode: MeasurementMode) => void
  onApply: (update: ParamsUpdate) => Promise<void>
}

interface FrequencyRange {
  start: number
  stop: number
}

function freqDefaults(params: SystemParams): FrequencyRange {
  if (params.minFreqMHz > 0 && params.manualBWMHz > 0) {
    return { start: params.minFreqMHz, stop: params.minFreqMHz + params.manualBWMHz }
  }
  const [start, stop] = BAND_RANGE_MHZ[params.band] ?? BAND_RANGE_MHZ[0]
  return { start, stop }
}

export function ParamsForm({ params, mode, onModeChange, onApply }: ParamsFormProps) {
  const [rampTimeMs, setRampTimeMs] = useState(String(params.tRampMs))
  const [zeroPad, setZeroPad] = useState(params.zeroPad)
  const [attenuation, setAttenuation] = useState(params.attenuation)
  const [frontendEn, setFrontendEn] = useState(params.frontendEn)
  const [powerSaveEn, setPowerSaveEn] = useState(params.powerSaveEn)
  const [norm, setNorm] = useState(params.norm)
  const [channels, setChannels] = useState(params.actChannels)

  const defaults = freqDefaults(params)
  const [startFreq, setStartFreq] = useState(String(defaults.start))
  const [stopFreq, setStopFreq] = useState(String(defaults.stop))
  const [staged, setStaged] = useState<FrequencyRange | null>(null)
  const [freqError, setFreqError] = useState<string | null>(null)
  const [applyError, setApplyError] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    const d = freqDefaults(params)
    setRampTimeMs(String(params.tRampMs))
    setZeroPad(params.zeroPad)
    setAttenuation(params.attenuation)
    setFrontendEn(params.frontendEn)
    setPowerSaveEn(params.powerSaveEn)
    setNorm(params.norm)
    setChannels(params.actChannels)
    setStartFreq(String(d.start))
    setStopFreq(String(d.stop))
    setStaged(null)
    setFreqError(null)
  }, [params])

  const applied = staged ?? defaults
  const outOfEuBand = applied.start < 24000 || applied.stop > EU_STOP_MHZ

  const toggleChannel = (index: number) => setChannels((c) => c ^ (1 << index))

  const handleSetFrequency = () => {
    const start = Number(startFreq)
    const stop = Number(stopFreq)
    if (!Number.isFinite(start) || !Number.isFinite(stop) || start <= 0 || stop <= 0) {
      setFreqError('Enter numeric start and stop frequencies in MHz.')
      return
    }
    if (start >= stop) {
      setFreqError('Start frequency must be below stop frequency.')
      return
    }
    if (start < MIN_FREQ_MHZ || stop > MAX_FREQ_MHZ) {
      setFreqError('Frequencies are limited to 23-26 GHz.')
      return
    }
    setStaged({ start, stop })
    setFreqError(null)
  }

  const handleApply = async () => {
    setApplyError(null)
    setBusy(true)
    try {
      await onApply({
        tRampMs: Number(rampTimeMs),
        zeroPad,
        frontendEn,
        powerSaveEn,
        norm,
        actChannels: channels,
        minFreqMHz: applied.start,
        manualBWMHz: applied.stop - applied.start,
        attenuation,
        mode,
      })
    } catch (err) {
      setApplyError(err instanceof Error ? err.message : String(err))
    } finally {
      setBusy(false)
    }
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

      <fieldset className="group">
        <legend>Receiver channels</legend>
        <div className="channel-grid">
          {CHANNELS.map((n) => {
            const bit = n - 1
            return (
              <label key={n} className="checkbox-row">
                <input
                  type="checkbox"
                  checked={(channels & (1 << bit)) !== 0}
                  onChange={() => toggleChannel(bit)}
                />
                Rx{n}
              </label>
            )
          })}
        </div>
      </fieldset>

      <div className="field-row">
        <label htmlFor="ramp-time">Ramp time (ms)</label>
        <input
          id="ramp-time"
          type="number"
          min={1}
          value={rampTimeMs}
          onChange={(e) => setRampTimeMs(e.target.value)}
        />
      </div>

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
        <legend>Signal processing</legend>
        <label className="checkbox-row">
          <input
            type="checkbox"
            checked={frontendEn}
            onChange={(e) => setFrontendEn(e.target.checked)}
          />
          Frontend power
        </label>
        <label className="checkbox-row">
          <input
            type="checkbox"
            checked={powerSaveEn}
            onChange={(e) => setPowerSaveEn(e.target.checked)}
          />
          Power saving
        </label>
        <label className="checkbox-row">
          <input
            type="checkbox"
            checked={norm}
            onChange={(e) => setNorm(e.target.checked)}
          />
          Normalization
        </label>
      </fieldset>

      <fieldset className="group">
        <legend>Frequency</legend>
        <div className="field-row">
          <label htmlFor="freq-start">Start frequency (MHz)</label>
          <input
            id="freq-start"
            type="number"
            value={startFreq}
            onChange={(e) => setStartFreq(e.target.value)}
          />
        </div>
        <div className="field-row">
          <label htmlFor="freq-stop">Stop frequency (MHz)</label>
          <input
            id="freq-stop"
            type="number"
            value={stopFreq}
            onChange={(e) => setStopFreq(e.target.value)}
          />
        </div>
        {freqError && <div className="warning-banner">{freqError}</div>}
        {outOfEuBand && (
          <div className="warning-banner">
            {applied.start}-{applied.stop} MHz exceeds the 24.000-24.250 GHz license-free EU range.
            Confirm before applying.
          </div>
        )}
        <button type="button" className="btn" onClick={handleSetFrequency}>
          Set frequency
        </button>
      </fieldset>

      <div className="derived-row">
        <span className="info-label">Range bin</span>
        <span className="info-value">{params.rangeBinUm} µm</span>
        <span className="info-label">Doppler bin</span>
        <span className="info-value">{params.dopplerBinUmS} µm/s</span>
      </div>

      {applyError && <div className="warning-banner">{applyError}</div>}

      <button className="btn btn-primary" onClick={handleApply} disabled={busy}>
        {busy ? 'Applying…' : 'Apply settings'}
      </button>
    </section>
  )
}
