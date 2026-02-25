import type { Simulation, CreateSimulationRequest } from '@/types/simulation'

const BASE_URL = '/api'

export async function createSimulation(req: CreateSimulationRequest): Promise<Simulation> {
  const res = await fetch(`${BASE_URL}/simulations`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: 'Unknown error' }))
    throw new Error(err.error || `HTTP ${res.status}`)
  }
  return res.json()
}

export async function listSimulations(): Promise<Simulation[]> {
  const res = await fetch(`${BASE_URL}/simulations`)
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  return res.json()
}

export async function getSimulation(id: string): Promise<Simulation> {
  const res = await fetch(`${BASE_URL}/simulations/${id}`)
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  return res.json()
}

export async function deleteSimulation(id: string): Promise<void> {
  const res = await fetch(`${BASE_URL}/simulations/${id}`, { method: 'DELETE' })
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
}
