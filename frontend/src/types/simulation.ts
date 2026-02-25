export interface Agent {
  id: string
  name: string
  role: string
}

export interface Step {
  round: number
  agent_id: string
  agent_name: string
  content: string
  timestamp: string
}

export interface Simulation {
  id: string
  description: string
  preconditions: string
  rounds: number
  show_only_result: boolean
  agents: Agent[]
  language: 'en' | 'ru'
  depth: 'shallow' | 'medium' | 'deep'
  status: 'running' | 'completed' | 'failed'
  steps: Step[]
  final_result?: string
  created_at: string
}

export interface AgentRequest {
  name: string
  role: string
}

export interface CreateSimulationRequest {
  description: string
  preconditions: string
  rounds: number
  show_only_result: boolean
  agents: AgentRequest[]
  language: 'en' | 'ru'
  depth: 'shallow' | 'medium' | 'deep'
}
