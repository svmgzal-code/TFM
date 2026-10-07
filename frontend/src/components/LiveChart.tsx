import { useEffect, useMemo, useRef } from 'react'
import uPlot from 'uplot'
import 'uplot/dist/uPlot.min.css'

interface LiveChartProps {
  mode: 'fmcw' | 'cw'
}

// Fake FD magnitude with a couple of plausible target peaks over range.
// TODO: replace with real /ws frames (decimated for the UI).
function fakeFdData() {
  const zeroPad = 2
  const rangeBinUm = 599600
  const freqPoints = 256
  const x: number[] = []
  const y: number[] = []

  for (let n = 1; n <= freqPoints; n++) {
    const rangeM = ((n - 1) * rangeBinUm) / (zeroPad * 1e6)
    const level =
      -92 +
      48 * Math.exp(-(((rangeM - 15) / 2.5) ** 2)) +
      36 * Math.exp(-(((rangeM - 42) / 1.4) ** 2))
    x.push(Number(rangeM.toFixed(2)))
    y.push(Number(level.toFixed(1)))
  }
  return [x, y] as uPlot.AlignedData
}

export function LiveChart({ mode }: LiveChartProps) {
  const targetRef = useRef<HTMLDivElement>(null)
  const data = useMemo(fakeFdData, [])

  useEffect(() => {
    const target = targetRef.current
    if (!target) return

    const opts: uPlot.Options = {
      width: target.clientWidth,
      height: 360,
      title: 'Frequency-domain magnitude',
      scales: {
        x: { time: false },
        y: { range: [-100, -20] },
      },
      series: [
        {},
        { label: 'FD magnitude', stroke: '#4cc2ff', width: 1, fill: 'rgba(76, 194, 255, 0.08)' },
      ],
      axes: [
        { label: mode === 'cw' ? 'Velocity (m/s)' : 'Range (m)', grid: { stroke: 'rgba(255,255,255,0.06)' } },
        { label: 'Level (dB)', grid: { stroke: 'rgba(255,255,255,0.06)' } },
      ],
      cursor: { show: true, points: { size: 6 } },
      legend: { show: true },
    }

    const plot = new uPlot(opts, data, target)
    return () => plot.destroy()
  }, [data, mode])

  return (
    <section className="card chart-card">
      <div className="chart-header">
        <h2>Live spectrum</h2>
        <span className="chart-tag">{mode === 'cw' ? 'CW · velocity' : 'FMCW · range'}</span>
      </div>
      <div ref={targetRef} className="chart-canvas" />
    </section>
  )
}
