<script setup lang="ts">
import { ref } from 'vue'
import { useSimulationStore } from '@/stores/simulation'

const store = useSimulationStore()

const description = ref('')
const preconditions = ref('')
const rounds = ref(5)
const showOnlyResult = ref(false)

async function submit() {
  if (!description.value.trim()) return
  await store.startSimulation({
    description: description.value.trim(),
    preconditions: preconditions.value.trim(),
    rounds: rounds.value,
    show_only_result: showOnlyResult.value,
  })
  description.value = ''
  preconditions.value = ''
  rounds.value = 5
  showOnlyResult.value = false
}
</script>

<template>
  <div class="sim-form">
    <h2>New Simulation</h2>

    <div class="field">
      <label for="description">Scenario Description</label>
      <textarea
        id="description"
        v-model="description"
        rows="4"
        placeholder="Describe the simulation scenario..."
      />
    </div>

    <div class="field">
      <label for="preconditions">Preconditions</label>
      <textarea
        id="preconditions"
        v-model="preconditions"
        rows="3"
        placeholder="Initial conditions, options, constraints..."
      />
    </div>

    <div class="field-row">
      <div class="field">
        <label for="rounds">Rounds (1–20)</label>
        <input
          id="rounds"
          v-model.number="rounds"
          type="number"
          min="1"
          max="20"
        />
      </div>

      <div class="field checkbox-field">
        <label>
          <input type="checkbox" v-model="showOnlyResult" />
          Show only final result
        </label>
      </div>
    </div>

    <button
      class="btn-run"
      :disabled="!description.trim() || store.loading"
      @click="submit"
    >
      {{ store.loading ? 'Starting...' : 'Run Simulation' }}
    </button>
  </div>
</template>

<style scoped>
.sim-form {
  padding: 1.5rem;
}

.sim-form h2 {
  margin: 0 0 1.25rem;
  font-size: 1.25rem;
  color: #e0e0e0;
}

.field {
  margin-bottom: 1rem;
}

.field label {
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

.field-row {
  display: flex;
  gap: 1rem;
  align-items: flex-end;
}

.field-row .field {
  flex: 1;
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
