import type { RadarStatus } from '../types'

const BASE = '/api'

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...init,
  })
  if (!res.ok) {
    let message = `${res.status} ${res.statusText}`
    try {
      const body = (await res.json()) as { error?: string }
      if (body.error) message = body.error
    } catch {
      // non-JSON error body; keep the status text
    }
    throw new Error(message)
  }
  return (await res.json()) as T
}

export interface ConnectRequest {
  host: string
  commandPort: number
  streamPort: number
}

export function connect(req: ConnectRequest): Promise<RadarStatus> {
  return request<RadarStatus>('/connect', {
    method: 'POST',
    body: JSON.stringify(req),
  })
}

export function disconnect(): Promise<RadarStatus> {
  return request<RadarStatus>('/disconnect', { method: 'POST' })
}

export function getStatus(): Promise<RadarStatus> {
  return request<RadarStatus>('/status')
}
