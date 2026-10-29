import { useCallback, useState } from 'react'
import { getParams, updateParams } from '../api/client'
import type { ParamsUpdate, SystemParams } from '../types'

export function useParams() {
  const [params, setParams] = useState<SystemParams | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [applying, setApplying] = useState(false)

  const refresh = useCallback(async () => {
    try {
      const next = await getParams()
      setParams(next)
      setError(null)
      return true
    } catch (err) {
      setParams(null)
      setError(err instanceof Error ? err.message : String(err))
      return false
    }
  }, [])

  const apply = useCallback(
    async (update: ParamsUpdate) => {
      setApplying(true)
      try {
        const next = await updateParams(update)
        setParams(next)
        setError(null)
      } catch (err) {
        setError(err instanceof Error ? err.message : String(err))
        throw err
      } finally {
        setApplying(false)
      }
    },
    [],
  )

  return { params, error, applying, refresh, apply }
}
