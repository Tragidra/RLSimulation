import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Simulation, CreateSimulationRequest, Step } from '@/types/simulation'
import * as api from '@/services/api'
import { SimulationWebSocket } from '@/services/websocket'

export const useSimulationStore = defineStore('simulation', () => {
  const simulations = ref<Simulation[]>([])
  const currentSimulation = ref<Simulation | null>(null)
  const loading = ref(false)
  const wsConnection = ref<SimulationWebSocket | null>(null)

  async function fetchSimulations() {
    simulations.value = await api.listSimulations()
  }

  async function fetchSimulation(id: string) {
    currentSimulation.value = await api.getSimulation(id)
  }

  async function startSimulation(req: CreateSimulationRequest) {
    loading.value = true
    try {
      const sim = await api.createSimulation(req)
      currentSimulation.value = sim
      simulations.value.unshift(sim)
      connectWebSocket(sim.id)
      return sim
    } finally {
      loading.value = false
    }
  }

  function connectWebSocket(simId: string) {
    disconnectWebSocket()
    const ws = new SimulationWebSocket()

    ws.onStep((step: Step) => {
      if (!currentSimulation.value || currentSimulation.value.id !== simId) return

      if (step.round === -1) {
        // Completion sentinel
        currentSimulation.value.final_result = step.content
        currentSimulation.value.status = 'completed'
        disconnectWebSocket()
      } else {
        currentSimulation.value.steps.push(step)
      }
    })

    ws.onClose(() => {
      // Refresh to get final state
      if (currentSimulation.value && currentSimulation.value.id === simId) {
        api.getSimulation(simId).then((sim) => {
          currentSimulation.value = sim
          // Also update in list
          const idx = simulations.value.findIndex((s) => s.id === simId)
          if (idx !== -1) simulations.value[idx] = sim
        })
      }
    })

    ws.connect(simId)
    wsConnection.value = ws
  }

  function disconnectWebSocket() {
    if (wsConnection.value) {
      wsConnection.value.disconnect()
      wsConnection.value = null
    }
  }

  async function selectSimulation(id: string) {
    disconnectWebSocket()
    await fetchSimulation(id)
    if (currentSimulation.value?.status === 'running') {
      connectWebSocket(id)
    }
  }

  async function removeSimulation(id: string) {
    await api.deleteSimulation(id)
    simulations.value = simulations.value.filter((s) => s.id !== id)
    if (currentSimulation.value?.id === id) {
      currentSimulation.value = null
      disconnectWebSocket()
    }
  }

  function clearCurrent() {
    disconnectWebSocket()
    currentSimulation.value = null
  }

  return {
    simulations,
    currentSimulation,
    loading,
    fetchSimulations,
    startSimulation,
    selectSimulation,
    removeSimulation,
    clearCurrent,
  }
})
