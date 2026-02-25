<script setup lang="ts">
import { computed, watch, ref, nextTick } from 'vue'
import { marked } from 'marked'
import { useSimulationStore } from '@/stores/simulation'
import StepCard from './StepCard.vue'

const store = useSimulationStore()
const logContainer = ref<HTMLElement | null>(null)

const sim = computed(() => store.currentSimulation)

const statusClass = computed(() => {
  if (!sim.value) return ''
  return `status-${sim.value.status}`
})

const statusLabel = computed(() => {
  if (!sim.value) return ''
  return sim.value.status.charAt(0).toUpperCase() + sim.value.status.slice(1)
})

const roundProgress = computed(() => {
  if (!sim.value) return ''
  return `${sim.value.steps.length}/${sim.value.rounds}`
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
          <span class="round-counter">Rounds: {{ roundProgress }}</span>
        </div>
      </div>
    </div>

    <div class="viewer-body" ref="logContainer">
      <!-- Spinner mode -->
      <div v-if="showSpinner" class="spinner-container">
        <div class="spinner" />
        <p>Simulation running... ({{ roundProgress }} rounds complete)</p>
      </div>

      <!-- Steps display -->
      <template v-if="showSteps">
        <StepCard
          v-for="step in sim.steps"
          :key="step.round"
          :step="step"
          :total-rounds="sim.rounds"
        />
      </template>

      <!-- Running indicator -->
      <div v-if="sim.status === 'running' && !sim.show_only_result" class="running-indicator">
        <div class="dot-pulse" />
        <span>Agent is thinking...</span>
      </div>

      <!-- Final result -->
      <div v-if="sim.final_result" class="final-result">
        <div class="final-header">Final Summary</div>
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
