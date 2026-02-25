<script setup lang="ts">
import { computed } from 'vue'
import { marked } from 'marked'
import type { Step } from '@/types/simulation'

const AGENT_COLORS = ['#7c6ef0', '#e06c75', '#98c379', '#e5c07b', '#61afef']

const props = defineProps<{
  step: Step
  totalRounds: number
  agentIndex: number
}>()

const renderedContent = computed(() => {
  return marked.parse(props.step.content) as string
})

const formattedTime = computed(() => {
  return new Date(props.step.timestamp).toLocaleTimeString()
})

const borderColor = computed(() => AGENT_COLORS[props.agentIndex % AGENT_COLORS.length])
</script>

<template>
  <div class="step-card" :style="{ borderLeftColor: borderColor }">
    <div class="step-header">
      <div class="step-header-left">
        <span class="agent-name" :style="{ color: borderColor }">{{ step.agent_name }}</span>
      </div>
      <span class="step-time">{{ formattedTime }}</span>
    </div>
    <div class="step-content" v-html="renderedContent" />
  </div>
</template>

<style scoped>
.step-card {
  background: #1e1e2e;
  border: 1px solid #333;
  border-left: 3px solid #7c6ef0;
  border-radius: 8px;
  padding: 1rem 1.25rem;
  margin-bottom: 0.5rem;
}

.step-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
}

.step-header-left {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.agent-name {
  font-size: 0.8rem;
  font-weight: 600;
}

.step-time {
  font-size: 0.75rem;
  color: #666;
}

.step-content {
  font-size: 0.9rem;
  line-height: 1.6;
  color: #ccc;
}

.step-content :deep(p) {
  margin: 0.5em 0;
}

.step-content :deep(ul),
.step-content :deep(ol) {
  padding-left: 1.5em;
}

.step-content :deep(code) {
  background: #2a2a3e;
  padding: 0.15em 0.4em;
  border-radius: 3px;
  font-size: 0.85em;
}

.step-content :deep(pre) {
  background: #2a2a3e;
  padding: 0.75rem;
  border-radius: 6px;
  overflow-x: auto;
}
</style>
