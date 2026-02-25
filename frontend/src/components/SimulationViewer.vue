<script setup lang="ts">
import { computed, watch, ref, nextTick } from 'vue'
import { marked } from 'marked'
import { useSimulationStore } from '@/stores/simulation'
import { useLocale } from '@/composables/useLocale'
import StepCard from './StepCard.vue'
import type { Step } from '@/types/simulation'

const store = useSimulationStore()
const { t } = useLocale()
const logContainer = ref<HTMLElement | null>(null)

const sim = computed(() => store.currentSimulation)

const statusClass = computed(() => {
  if (!sim.value) return ''
  return `status-${sim.value.status}`
})

const statusLabel = computed(() => {
  if (!sim.value) return ''
  const key = sim.value.status as keyof typeof t.value.status
  return t.value.status[key] || sim.value.status
})

// Group steps by round
interface RoundGroup {
  round: number
  steps: Step[]
}

const roundGroups = computed<RoundGroup[]>(() => {
  if (!sim.value) return []
  const groups: RoundGroup[] = []
  let currentGroup: RoundGroup | null = null
  for (const step of sim.value.steps) {
    if (!currentGroup || currentGroup.round !== step.round) {
      currentGroup = { round: step.round, steps: [] }
      groups.push(currentGroup)
    }
    currentGroup.steps.push(step)
  }
  return groups
})

// Count completed rounds (rounds where all agents have acted)
const completedRounds = computed(() => {
  if (!sim.value) return 0
  const agentCount = sim.value.agents?.length || 1
  return Math.floor(sim.value.steps.length / agentCount)
})

const roundProgress = computed(() => {
  if (!sim.value) return ''
  return `${completedRounds.value}/${sim.value.rounds}`
})

// Build agent index map for consistent coloring
const agentIndexMap = computed<Record<string, number>>(() => {
  if (!sim.value?.agents) return {}
  const map: Record<string, number> = {}
  sim.value.agents.forEach((a, i) => { map[a.id] = i })
  return map
})

const finalHtml = computed(() => {
  if (!sim.value?.final_result) return ''
  return marked.parse(sim.value.final_result) as string
})

const showSpinner = computed(() => {
  return sim.value?.status === 'running' && sim.value?.show_only_result
})

const showSteps = computed(() => {
  if (!sim.value) return false
  if (sim.value.show_only_result && sim.value.status === 'running') return false
  return true
})

// Auto-scroll when new steps arrive
watch(
  () => sim.value?.steps.length,
  async () => {
    await nextTick()
    if (logContainer.value) {
      logContainer.value.scrollTop = logContainer.value.scrollHeight
    }
  }
)
</script>

<template>
  <div class="viewer" v-if="sim">
    <div class="viewer-header">
      <div class="viewer-title">
        <h2>{{ sim.description }}</h2>
        <div class="viewer-meta">
          <span class="status-badge" :class="statusClass">{{ statusLabel }}</span>
          <span class="round-counter">{{ t.viewer.roundsLabel }}: {{ roundProgress }}</span>
        </div>
      </div>
    </div>

    <div class="viewer-body" ref="logContainer">
      <!-- Spinner mode -->
      <div v-if="showSpinner" class="spinner-container">
        <div class="spinner" />
        <p>{{ t.viewer.running }} ({{ roundProgress }} {{ t.viewer.roundsComplete }})</p>
      </div>

      <!-- Steps grouped by round -->
      <template v-if="showSteps">
        <div v-for="group in roundGroups" :key="group.round" class="round-group">
          <div class="round-label">{{ t.viewer.roundLabel }} {{ group.round }}/{{ sim.rounds }}</div>
          <StepCard
            v-for="(step, idx) in group.steps"
            :key="`${step.round}-${step.agent_id}`"
            :step="step"
            :total-rounds="sim.rounds"
            :agent-index="agentIndexMap[step.agent_id] ?? idx"
          />
        </div>
      </template>

      <!-- Running indicator -->
      <div v-if="sim.status === 'running' && !sim.show_only_result" class="running-indicator">
        <div class="dot-pulse" />
        <span>{{ t.viewer.thinking }}</span>
      </div>

      <!-- Final result -->
      <div v-if="sim.final_result" class="final-result">
        <div class="final-header">{{ t.viewer.finalSummary }}</div>
        <div class="final-content" v-html="finalHtml" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.viewer {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.viewer-header {
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid #333;
  flex-shrink: 0;
}

.viewer-title h2 {
  margin: 0 0 0.5rem;
  font-size: 1.15rem;
  color: #e0e0e0;
}

.viewer-meta {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.status-badge {
  font-size: 0.75rem;
  font-weight: 600;
  padding: 0.2rem 0.6rem;
  border-radius: 12px;
}

.status-running {
  background: #2d4a22;
  color: #8eff6a;
}

.status-completed {
  background: #1a3a5c;
  color: #6ac5ff;
}

.status-failed {
  background: #5c1a1a;
  color: #ff6a6a;
}

.round-counter {
  font-size: 0.8rem;
  color: #888;
}

.round-group {
  margin-bottom: 1rem;
}

.round-label {
  font-size: 0.75rem;
  font-weight: 600;
  color: #7c6ef0;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 0.4rem;
  padding: 0.2rem 0;
  border-bottom: 1px solid #2a2a3e;
}

.viewer-body {
  flex: 1;
  overflow-y: auto;
  padding: 1.25rem 1.5rem;
}

.spinner-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem 0;
  color: #888;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #333;
  border-top-color: #7c6ef0;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin-bottom: 1rem;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.running-indicator {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 0;
  color: #888;
  font-size: 0.85rem;
}

.dot-pulse {
  width: 8px;
  height: 8px;
  background: #7c6ef0;
  border-radius: 50%;
  animation: pulse 1s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 0.3; }
  50% { opacity: 1; }
}

.final-result {
  margin-top: 1rem;
  border: 1px solid #7c6ef0;
  border-radius: 8px;
  overflow: hidden;
}

.final-header {
  background: #7c6ef0;
  color: #fff;
  padding: 0.5rem 1rem;
  font-weight: 600;
  font-size: 0.85rem;
}

.final-content {
  padding: 1rem 1.25rem;
  font-size: 0.9rem;
  line-height: 1.6;
  color: #ccc;
}

.final-content :deep(p) {
  margin: 0.5em 0;
}
</style>
