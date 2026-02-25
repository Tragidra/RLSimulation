export interface Step {
  round: number
  content: string
  timestamp: string
}

export interface Simulation {
  id: string
  description: string
  preconditions: string
  rounds: number
  show_only_result: boolean
  status: 'running' | 'completed' | 'failed'
  steps: Step[]
  final_result?: string
  created_at: string
}

export interface CreateSimulationRequest {
  description: string
  preconditions: string
  rounds: number
  show_only_result: boolean
}
