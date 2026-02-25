<script setup lang="ts">
import { ref, computed } from 'vue'
import { useSimulationStore } from '@/stores/simulation'
import { useLocale } from '@/composables/useLocale'
import type { AgentRequest } from '@/types/simulation'

const store = useSimulationStore()
const { locale, t } = useLocale()

const description = ref('')
const preconditions = ref('')
const roundMode = ref<'standard' | 'custom'>('standard')
const customRounds = ref(10)
const showOnlyResult = ref(false)
const depth = ref<'shallow' | 'medium' | 'deep'>('medium')
const agents = ref<AgentRequest[]>([{ name: '', role: '' }])

const rounds = computed(() => roundMode.value === 'standard' ? 5 : customRounds.value)

const allRolesEmpty = computed(() => agents.value.every(a => !a.role.trim()))

function addAgent() {
  if (agents.value.length < 5) {
    agents.value.push({ name: '', role: '' })
  }
}

function removeAgent(index: number) {
  if (agents.value.length > 1) {
    agents.value.splice(index, 1)
  }
}

async function submit() {
  if (!description.value.trim()) return
  await store.startSimulation({
    description: description.value.trim(),
    preconditions: preconditions.value.trim(),
    rounds: rounds.value,
    show_only_result: showOnlyResult.value,
    agents: agents.value.map(a => ({
      name: a.name.trim() || 'Agent',
      role: a.role.trim(),
    })),
    language: locale.value,
    depth: depth.value,
  })
  description.value = ''
  preconditions.value = ''
  roundMode.value = 'standard'
  customRounds.value = 10
  showOnlyResult.value = false
  depth.value = 'medium'
  agents.value = [{ name: '', role: '' }]
}
</script>

<template>
  <div class="sim-form">
    <h2>{{ t.form.title }}</h2>

    <div class="field">
      <label for="description">{{ t.form.description }}</label>
      <textarea
        id="description"
        v-model="description"
        rows="4"
        :placeholder="t.form.descriptionPlaceholder"
      />
    </div>

    <div class="field">
      <label for="preconditions">{{ t.form.preconditions }}</label>
      <textarea
        id="preconditions"
        v-model="preconditions"
        rows="3"
        :placeholder="t.form.preconditionsPlaceholder"
      />
    </div>

    <!-- Agents section -->
    <div class="field">
      <label>{{ t.form.agents }}</label>
      <div
        v-for="(agent, idx) in agents"
        :key="idx"
        class="agent-card"
      >
        <div class="agent-fields">
          <input
            type="text"
            v-model="agent.name"
            :placeholder="t.form.agentNamePlaceholder"
            class="agent-input"
          />
          <input
            type="text"
            v-model="agent.role"
            :placeholder="t.form.agentRolePlaceholder"
            class="agent-input"
          />
        </div>
        <button
          v-if="agents.length > 1"
          class="btn-remove-agent"
          @click="removeAgent(idx)"
        >
          &times;
        </button>
      </div>
      <button
        v-if="agents.length < 5"
        class="btn-add-agent"
        @click="addAgent"
      >
        {{ t.form.addAgent }}
      </button>
      <div v-if="agents.length > 1 && allRolesEmpty" class="hint">
        {{ t.form.agentHintIndependent }}
      </div>
    </div>

    <!-- Depth selector -->
    <div class="field">
      <label>{{ t.form.depth }}</label>
      <div class="radio-group">
        <label class="radio-label">
          <input type="radio" v-model="depth" value="shallow" />
          {{ t.form.depthShallow }}
        </label>
        <label class="radio-label">
          <input type="radio" v-model="depth" value="medium" />
          {{ t.form.depthMedium }}
        </label>
        <label class="radio-label">
          <input type="radio" v-model="depth" value="deep" />
          {{ t.form.depthDeep }}
        </label>
      </div>
    </div>

    <!-- Rounds -->
    <div class="field">
      <label>{{ t.form.rounds }}</label>
      <div class="radio-group">
        <label class="radio-label">
          <input type="radio" v-model="roundMode" value="standard" />
          {{ t.form.roundsStandard }}
        </label>
        <label class="radio-label">
          <input type="radio" v-model="roundMode" value="custom" />
          {{ t.form.roundsCustom }}
        </label>
      </div>
      <input
        v-if="roundMode === 'custom'"
        v-model.number="customRounds"
        type="number"
        min="1"
        max="50"
        class="custom-rounds-input"
      />
    </div>

    <div class="field checkbox-field">
      <label>
        <input type="checkbox" v-model="showOnlyResult" />
        {{ t.form.showOnlyResult }}
      </label>
    </div>

    <button
      class="btn-run"
      :disabled="!description.trim() || store.loading"
      @click="submit"
    >
      {{ store.loading ? t.form.starting : t.form.run }}
    </button>
  </div>
</template>

<style scoped>
.sim-form {
  padding: 1.5rem;
  max-width: 700px;
}

.sim-form h2 {
  margin: 0 0 1.25rem;
  font-size: 1.25rem;
  color: #e0e0e0;
}

.field {
  margin-bottom: 1rem;
}

.field > label {
  display: block;
  margin-bottom: 0.35rem;
  font-size: 0.85rem;
  color: #aaa;
  font-weight: 500;
}

.field textarea,
.field input[type='number'] {
  width: 100%;
  padding: 0.6rem 0.75rem;
  border: 1px solid #444;
  border-radius: 6px;
  background: #1e1e2e;
  color: #e0e0e0;
  font-size: 0.9rem;
  font-family: inherit;
  resize: vertical;
  box-sizing: border-box;
}

.field textarea:focus,
.field input:focus {
  outline: none;
  border-color: #7c6ef0;
}

.checkbox-field label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  font-size: 0.85rem;
  color: #aaa;
}

.checkbox-field input[type='checkbox'] {
  accent-color: #7c6ef0;
}

/* Agent cards */
.agent-card {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
  background: #1e1e2e;
  border: 1px solid #333;
  border-radius: 6px;
  padding: 0.5rem 0.6rem;
}

.agent-fields {
  display: flex;
  gap: 0.5rem;
  flex: 1;
}

.agent-input {
  flex: 1;
  padding: 0.4rem 0.6rem;
  border: 1px solid #444;
  border-radius: 4px;
  background: #12121e;
  color: #e0e0e0;
  font-size: 0.85rem;
  font-family: inherit;
}

.agent-input:focus {
  outline: none;
  border-color: #7c6ef0;
}

.btn-remove-agent {
  background: none;
  border: none;
  color: #666;
  font-size: 1.2rem;
  cursor: pointer;
  padding: 0.1rem 0.3rem;
  border-radius: 4px;
  line-height: 1;
}

.btn-remove-agent:hover {
  background: #5c1a1a;
  color: #ff6a6a;
}

.btn-add-agent {
  background: none;
  border: 1px dashed #444;
  color: #7c6ef0;
  border-radius: 6px;
  padding: 0.4rem 0.75rem;
  font-size: 0.8rem;
  cursor: pointer;
  transition: border-color 0.2s;
}

.btn-add-agent:hover {
  border-color: #7c6ef0;
}

.hint {
  margin-top: 0.4rem;
  font-size: 0.75rem;
  color: #888;
  font-style: italic;
}

/* Radio groups */
.radio-group {
  display: flex;
  gap: 1rem;
  margin-top: 0.25rem;
}

.radio-label {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.85rem;
  color: #ccc;
  cursor: pointer;
}

.radio-label input[type='radio'] {
  accent-color: #7c6ef0;
}

.custom-rounds-input {
  margin-top: 0.4rem;
  width: 120px;
  padding: 0.4rem 0.6rem;
  border: 1px solid #444;
  border-radius: 6px;
  background: #1e1e2e;
  color: #e0e0e0;
  font-size: 0.9rem;
  box-sizing: border-box;
}

.btn-run {
  margin-top: 0.5rem;
  width: 100%;
  padding: 0.7rem;
  border: none;
  border-radius: 6px;
  background: #7c6ef0;
  color: #fff;
  font-size: 0.95rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
}

.btn-run:hover:not(:disabled) {
  background: #6b5cd4;
}

.btn-run:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
