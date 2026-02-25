# SimArena — LLM Simulation Platform

## Overview

SimArena is a platform for running LLM-powered simulations. Users describe a scenario and conditions, configure parameters, and observe how an LLM agent reasons through the situation step by step.

**Example use case:** "Develop a theory for becoming the most powerful being in the simulation using one of 5 provided concepts" — the agent picks a concept, then over N rounds develops and refines its strategy.

---

## Architecture

```
┌─────────────────────────────────────────────┐
│                  Frontend                    │
│            Vue 3 + Vite + TypeScript         │
│                                              │
│  ┌──────────┐ ┌──────────┐ ┌──────────────┐ │
│  │ Sim Form │ │ Live Log │ │  History     │ │
│  │          │ │ (WS)     │ │  (past runs) │ │
│  └──────────┘ └──────────┘ └──────────────┘ │
└──────────────────┬──────────────────────────┘
                   │ HTTP + WebSocket
┌──────────────────▼──────────────────────────┐
│                  Backend (Go)                │
│                                              │
│  ┌───────────┐ ┌────────────┐ ┌───────────┐ │
│  │ HTTP API  │ │ WS Hub     │ │ Simulator │ │
│  │ (chi/std) │ │            │ │ Engine    │ │
│  └───────────┘ └────────────┘ └───────────┘ │
│                                              │
│  ┌───────────┐ ┌────────────┐               │
│  │ LLM       │ │ Storage    │               │
│  │ Client    │ │ (JSON)     │               │
│  └───────────┘ └────────────┘               │
└──────────────────┬──────────────────────────┘
                   │ OpenAI-compatible API
┌──────────────────▼──────────────────────────┐
│          LLM Studio (local)                  │
│    http://localhost:1234/v1/chat/completions  │
└──────────────────────────────────────────────┘
```

---

## Iteration 1 — Single Agent

### What the user sees

1. **Setup form:**
   - `description` (textarea) — scenario description
   - `preconditions` (textarea) — initial conditions, options, constraints
   - `rounds` (number, 1–20) — how many sequential reasoning steps the agent performs
   - `show_only_result` (checkbox) — if checked, only final result is shown; otherwise steps stream in real-time

2. **Simulation view:**
   - If `show_only_result = false`: steps appear one by one via WebSocket as the agent completes each round
   - If `show_only_result = true`: spinner/progress bar, then final result
   - Each step shows: round number, agent's reasoning/action text, timestamp

3. **History sidebar:**
   - List of past simulations (title auto-generated from description)
   - Click to view full log of any past simulation

### Backend

#### API Endpoints

```
POST   /api/simulations          — create and start a simulation
GET    /api/simulations          — list all simulations
GET    /api/simulations/{id}     — get simulation details + all steps
DELETE /api/simulations/{id}     — delete a simulation
WS     /api/simulations/{id}/ws  — stream simulation steps in real-time
```

#### Simulation Engine Flow

1. Receive simulation config (description, preconditions, rounds)
2. Generate UUID, save simulation record with status `running`
3. For each round `i` from 1 to N:
   a. Build prompt:
      - System prompt: simulation context + rules
      - Round 1: scenario description + preconditions
      - Round 2+: previous round's output as context (chain of thought)
   b. Call LLM API (streaming)
   c. Save step result
   d. Broadcast step via WebSocket to connected clients
4. After all rounds: compose final summary prompt → save → set status `completed`

#### Prompt Structure

```
SYSTEM:
You are a simulation agent. You are participating in a simulation
with the following scenario and conditions.

Scenario: {description}
Preconditions: {preconditions}

This simulation has {rounds} rounds. You are currently on round {i}/{rounds}.

Your task: analyze the situation, make decisions, and take actions
within the simulation. Think strategically. Each round builds on
the previous one.

{if round > 1}
Here is what happened in previous rounds:
{previous_rounds_summary}
{/if}

USER:
Execute round {i}. Describe your analysis, decisions, and actions
for this round. Be detailed and strategic.
```

#### Data Model

```go
type Simulation struct {
    ID             string    `json:"id"`
    Description    string    `json:"description"`
    Preconditions  string    `json:"preconditions"`
    Rounds         int       `json:"rounds"`
    ShowOnlyResult bool      `json:"show_only_result"`
    Status         string    `json:"status"` // "running", "completed", "failed"
    Steps          []Step    `json:"steps"`
    FinalResult    string    `json:"final_result,omitempty"`
    CreatedAt      time.Time `json:"created_at"`
}

type Step struct {
    Round     int       `json:"round"`
    Content   string    `json:"content"`
    Timestamp time.Time `json:"timestamp"`
}
```

#### Storage

JSON file-based storage in `./data/` directory:
- `./data/simulations.json` — array of all simulations
- Simple read/write with file lock (sync.Mutex)
- Good enough for local single-user usage

#### LLM Client

OpenAI-compatible client with configurable base URL:

```go
type LLMConfig struct {
    BaseURL string // default: http://localhost:1234/v1
    Model   string // default: openai/gpt-oss-20b
    APIKey  string // default: "not-needed" (LM Studio)
}
```

- Support streaming responses (SSE) for real-time step display
- Timeout per request: 120s (local models can be slow)
- Retry on failure: 1 retry with 5s delay

### Frontend

#### Tech Stack
- Vue 3 (Composition API + `<script setup>`)
- Vite
- TypeScript
- PrimeVue or Naive UI (component library)
- Pinia (state management)

#### Pages / Views

1. **Main view** (`/`) — split layout:
   - Left sidebar: history list + "New Simulation" button
   - Right panel: either setup form or simulation viewer

2. **Setup form** — fields as described above, "Run" button

3. **Simulation viewer:**
   - Header: simulation description, status badge, round counter (3/5)
   - Body: scrollable log of steps (markdown rendered)
   - Each step: card with round number, content, timestamp
   - Auto-scroll to latest step during streaming
   - Final result highlighted at the bottom

#### WebSocket Integration

```typescript
// Connect when viewing a running simulation
const ws = new WebSocket(`ws://localhost:8080/api/simulations/${id}/ws`)

ws.onmessage = (event) => {
  const step: Step = JSON.parse(event.data)
  simulation.steps.push(step)
}
```

---

## Iteration 2 — Multi-Agent (future)

> Not implemented in iteration 1. Documented here for architectural awareness.

### Changes

- `agents` field in simulation config: array of agent definitions (name, role, personality)
- Each agent runs in its own goroutine
- Agents communicate via message broker (start with Go channels, migrate to NATS/Kafka later)
- Shared state visible to all agents (world state)
- Turn-based or concurrent execution modes
- Frontend shows multiple agent streams side by side

### Message Broker Interface

```go
type MessageBroker interface {
    Publish(topic string, msg Message) error
    Subscribe(topic string) (<-chan Message, error)
}
```

Start with `ChanBroker` (in-memory Go channels), swap to `NATSBroker` or `KafkaBroker` later.

---

## Project Structure

```
simarena/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go              — entry point
│   ├── internal/
│   │   ├── api/
│   │   │   ├── handler.go           — HTTP handlers
│   │   │   ├── router.go            — route setup
│   │   │   └── websocket.go         — WS hub + connections
│   │   ├── simulation/
│   │   │   ├── engine.go            — simulation orchestration
│   │   │   └── prompt.go            — prompt building
│   │   ├── llm/
│   │   │   ├── client.go            — OpenAI-compatible client
│   │   │   └── types.go             — request/response types
│   │   ├── storage/
│   │   │   └── json.go              — JSON file storage
│   │   └── models/
│   │       └── simulation.go        — data models
│   ├── go.mod
│   └── go.sum
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── SimulationForm.vue
│   │   │   ├── SimulationViewer.vue
│   │   │   ├── StepCard.vue
│   │   │   └── HistorySidebar.vue
│   │   ├── stores/
│   │   │   └── simulation.ts
│   │   ├── services/
│   │   │   ├── api.ts               — HTTP client
│   │   │   └── websocket.ts         — WS connection manager
│   │   ├── types/
│   │   │   └── simulation.ts
│   │   ├── views/
│   │   │   └── MainView.vue
│   │   ├── App.vue
│   │   └── main.ts
│   ├── index.html
│   ├── vite.config.ts
│   ├── tsconfig.json
│   └── package.json
├── data/                             — JSON storage (gitignored)
├── SPEC.md
├── .gitignore
└── README.md
```

---

## Running Locally

```bash
# Terminal 1: Backend
cd backend
go run ./cmd/server

# Terminal 2: Frontend
cd frontend
npm install
npm run dev
```

Backend runs on `:8080`, frontend on `:5173` with Vite proxy to backend.

---

## Configuration

Environment variables or `config.yaml`:

```yaml
server:
  port: 8080
  cors_origin: "http://localhost:5173"

llm:
  base_url: "http://localhost:1234/v1"
  model: "openai/gpt-oss-20b"
  api_key: "not-needed"
  timeout: 120s
  max_tokens: 4096

storage:
  path: "./data"
```

---

## Claude Code Prompt

After placing this SPEC.md in the repo root, use this prompt:

```
Read SPEC.md and implement Iteration 1 fully. Start with the backend:
set up the Go project structure, implement the LLM client, simulation
engine, JSON storage, HTTP API and WebSocket hub. Then implement the
frontend: Vue 3 + Vite + TypeScript project with PrimeVue, simulation
form, real-time step viewer, and history sidebar. Make sure WebSocket
streaming works end-to-end. Use the project structure from the spec.
```
