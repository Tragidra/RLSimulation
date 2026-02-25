import type { Step } from '@/types/simulation'

export class SimulationWebSocket {
  private ws: WebSocket | null = null
  private onStepCallback: ((step: Step) => void) | null = null
  private onCloseCallback: (() => void) | null = null

  connect(simulationId: string): void {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const url = `${protocol}//${host}/api/simulations/${simulationId}/ws`

    this.ws = new WebSocket(url)

    this.ws.onmessage = (event: MessageEvent) => {
      try {
        const step: Step = JSON.parse(event.data)
        if (this.onStepCallback) {
          this.onStepCallback(step)
        }
      } catch (e) {
        console.error('Failed to parse WebSocket message:', e)
      }
    }

    this.ws.onclose = () => {
      if (this.onCloseCallback) {
        this.onCloseCallback()
      }
    }

    this.ws.onerror = (err) => {
      console.error('WebSocket error:', err)
    }
  }

  onStep(callback: (step: Step) => void): void {
    this.onStepCallback = callback
  }

  onClose(callback: () => void): void {
    this.onCloseCallback = callback
  }

  disconnect(): void {
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }
}
