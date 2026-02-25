# SimArena — Iteration 2 Spec

## Important

Before making any changes, read the entire current codebase — both `backend/` and `frontend/`. Understand the existing structure, data models, API endpoints, WebSocket logic, and how the simulation engine works. All changes in this spec must be implemented on top of the existing working code, not from scratch.

---

## Changes Overview

Iteration 2 adds: multi-agent support (1–5 agents with roles), language toggle (EN/RU), simulation depth setting, round count options, and a context window strategy for long simulations.

---

## 1. Multi-Agent Support (1–5 agents)

### Frontend: Agent Configuration

In the simulation creation form, replace the single-agent setup with a dynamic agent list:

- "Number of agents" selector: 1 to 5
- For each agent, show a card/row with:
  - **Agent name** (text input, required) — e.g. "Warrior", "Diplomat", "Necromancer"
  - **Role** (text input, optional) — custom role description for this agent within the simulation
- Default behavior hint (show as a tooltip or small note):
  - If ALL agents have empty roles → all agents act as independent "main characters" solving the scenario in parallel, unaware of each other. Show hint: "All agents work independently as main characters. Add custom roles to enable interaction between agents."
  - If ANY agent has a custom role → all agents are aware of each other, react to each other's actions, and the simulation becomes interactive/collaborative/competitive

### Backend: Data Model Changes

```go
type Agent struct {
    ID   string `json:"id"`   // uuid
    Name string `json:"name"` // display name
    Role string `json:"role"` // custom role, empty = "main character"
}

type Simulation struct {
    // ... existing fields ...
    Agents []Agent `json:"agents"` // replaces implicit single agent
    // ... rest of fields ...
}

type Step struct {
    Round     int       `json:"round"`
    AgentID   string    `json:"agent_id"`   // which agent produced this step
    AgentName string    `json:"agent_name"` // for display
    Content   string    `json:"content"`
    Timestamp time.Time `json:"timestamp"`
}
```

### Backend: Simulation Engine Changes

Each round now processes ALL agents sequentially (agent 1, then agent 2, ... then agent N), then moves to the next round.

**Independent mode** (all roles empty):
- Each agent gets its own prompt with only its own previous steps
- Agents don't see each other's actions
- Essentially N parallel simulations in one view

**Interactive mode** (any role is set):
- Each agent's prompt includes:
  - The scenario and preconditions
  - Its own role
  - Names and roles of all other agents
  - Actions of other agents from previous rounds (and current round if already acted this round)
  - Its own previous actions
- Agent prompt should make it clear: "You are {name}, your role: {role}. Other participants: {list}. React to their actions and the evolving situation."

### Frontend: Simulation Viewer Changes

- Steps are grouped by round
- Within each round, show each agent's step as a separate card
- Agent cards should be visually distinguishable (different colored left border or avatar/icon per agent)
- Agent name shown on each step card

---

## 2. Language Toggle (EN / RU)

### Frontend

- Language toggle in the top navigation bar (simple switch/dropdown: EN | RU)
- Default: EN
- Store selection in localStorage
- All UI text must come from a locale file (i18n approach)
- Create two locale files:
  - `locales/en.ts` — all UI strings in English
  - `locales/ru.ts` — all UI strings in Russian
- Use Vue's `provide/inject` or a simple composable for the current locale

### Backend

- Simulation creation request gets a new field: `language: "en" | "ru"`
- All LLM prompts must be in the selected language
- Maintain two sets of prompt templates (English and Russian)
- The LLM system prompt must include an instruction like:
  - EN: "Respond in English."
  - RU: "Отвечай на русском языке."

---

## 3. Simulation Depth

### Frontend

Add a "Simulation depth" selector to the creation form with 3 options:

| Label (EN) | Label (RU)  | max_tokens |
|------------|-------------|------------|
| Shallow    | Маленькая   | 400        |
| Medium     | Средняя     | 1200       |
| Deep       | Глубокая    | 0 (no limit / use model default) |

Default: Medium

### Backend

```go
type Simulation struct {
    // ... existing fields ...
    Depth string `json:"depth"` // "shallow", "medium", "deep"
}
```

Map depth to `max_tokens` in the LLM client when making requests:
- `"shallow"` → `max_tokens: 400`
- `"medium"` → `max_tokens: 1200`
- `"deep"` → omit `max_tokens` or set to model maximum

---

## 4. Round Count Options

### Frontend

Replace the current round number input with two options:

- Radio/toggle: **Standard (5 rounds)** | **Custom**
- If "Custom" selected: show number input (min: 1, max: 50)
- Default: Standard (5)

### Backend

No structural change needed — `rounds` field already exists as int. Just ensure validation accepts 1–50.

---

## 5. Context Window Strategy

This controls how much history each agent receives in its prompt.

### Rules

- **If total rounds ≤ 5:** Each agent receives FULL history — all steps from all previous rounds (own actions + other agents' actions in interactive mode)
- **If total rounds > 5:** Each agent receives only the last 2 rounds of history (own actions + other agents' actions for rounds N-1 and N)

### Implementation

In the prompt builder:

```go
func buildAgentContext(sim *Simulation, agentID string, currentRound int) []Step {
    if sim.Rounds <= 5 {
        // return ALL steps before current round
        return getAllPreviousSteps(sim, agentID, currentRound)
    }
    // return only steps from rounds (currentRound-2) and (currentRound-1)
    minRound := max(1, currentRound-2)
    return getStepsInRange(sim, agentID, minRound, currentRound)
}
```

The prompt should clearly indicate when context is truncated:
- EN: "Note: For efficiency, only the last 2 rounds are shown. Earlier history is omitted."
- RU: "Примечание: для эффективности показаны только последние 2 раунда. Более ранняя история опущена."

---

## Prompt Templates

### Interactive Mode (example, English)

```
SYSTEM:
You are an agent in a simulation.
Your name: {agent_name}
Your role: {agent_role}

Scenario: {description}
Preconditions: {preconditions}

Other participants:
{for each other agent}
- {name}: {role}
{end}

This simulation has {total_rounds} rounds. Current round: {current_round}/{total_rounds}.

{if history}
=== History ===
{for each historical round}
--- Round {N} ---
{for each agent step in round}
[{agent_name}]: {content}
{end}
{end}
{if truncated}
(Only last 2 rounds shown for context efficiency)
{end}
=== End History ===
{end}

USER:
Execute round {current_round}. Consider other agents' actions and the evolving situation. Describe your analysis, decisions, and actions. Stay in character as {agent_name} ({agent_role}).
Respond in English.
```

### Independent Mode

Same as above but:
- No "Other participants" section
- History only includes this agent's own steps
- No instruction to consider other agents

---

## WebSocket Changes

Step messages now include `agent_id` and `agent_name`:

```json
{
  "round": 3,
  "agent_id": "uuid-here",
  "agent_name": "Necromancer",
  "content": "I raise the fallen...",
  "timestamp": "2025-02-26T12:00:00Z"
}
```

Frontend groups incoming steps by round and displays them under the correct agent.

---

## Summary of API Changes

### POST /api/simulations — request body changes

```json
{
  "description": "scenario text",
  "preconditions": "conditions text",
  "rounds": 5,
  "show_only_result": false,
  "agents": [
    { "name": "Warrior", "role": "A brave fighter seeking glory" },
    { "name": "Mage", "role": "" }
  ],
  "language": "en",
  "depth": "medium"
}
```

### Backwards compatibility

If `agents` is empty or missing, default to a single agent with name "Agent" and empty role (preserves iteration 1 behavior).
