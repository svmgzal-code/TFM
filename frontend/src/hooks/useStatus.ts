import { useCallback, useEffect, useState } from 'react'
import { getStatus } from '../api/client'
import type { RadarStatus } from '../types'

const POLL_MS = 2000

const INITIAL_STATUS: RadarStatus = {
  connection: 'disconnected',
  recording: false,
  host: '',
  commandPort: 0,
  streamPort: 0,
}

export function useStatus() {
  const [status, setStatus] = useState<RadarStatus>(INITIAL_STATUS)
  const [error, setError] = useState<string | null>(null)

  const refresh = useCallback(async () => {
    try {
      const next = await getStatus()
      setStatus(next)
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    }
  }, [])

  useEffect(() => {
    void refresh()
    const id = window.setInterval(() => void refresh(), POLL_MS)
    return () => window.clearInterval(id)
  }, [refresh])

  return { status, error, refresh }
}
