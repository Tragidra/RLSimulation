<script setup lang="ts">
import { onMounted } from 'vue'
import { useSimulationStore } from '@/stores/simulation'
import { useLocale } from '@/composables/useLocale'

const store = useSimulationStore()
const { locale, t, setLocale } = useLocale()

onMounted(() => {
  store.fetchSimulations()
})

function truncate(text: string, max: number): string {
  if (text.length <= max) return text
  return text.slice(0, max) + '...'
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}
</script>

<template>
  <div class="sidebar">
    <div class="sidebar-header">
      <h3>{{ t.app.title }}</h3>
      <div class="header-actions">
        <div class="lang-toggle">
          <button
            class="lang-btn"
            :class="{ active: locale === 'en' }"
            @click="setLocale('en')"
          >{{ t.lang.en }}</button>
          <button
            class="lang-btn"
            :class="{ active: locale === 'ru' }"
            @click="setLocale('ru')"
          >{{ t.lang.ru }}</button>
        </div>
        <button class="btn-new" @click="store.clearCurrent()">{{ t.sidebar.newButton }}</button>
      </div>
    </div>

    <div class="sim-list">
      <div
        v-for="sim in store.simulations"
        :key="sim.id"
        class="sim-item"
        :class="{ active: store.currentSimulation?.id === sim.id }"
        @click="store.selectSimulation(sim.id)"
      >
        <div class="sim-item-title">{{ truncate(sim.description, 50) }}</div>
        <div class="sim-item-meta">
          <span class="sim-status" :class="`st-${sim.status}`">{{ sim.status }}</span>
          <span class="sim-date">{{ formatDate(sim.created_at) }}</span>
        </div>
        <button
          class="btn-delete"
          @click.stop="store.removeSimulation(sim.id)"
          title="Delete"
        >
          &times;
        </button>
      </div>

      <div v-if="store.simulations.length === 0" class="empty">
        {{ t.sidebar.empty }}
      </div>
    </div>
  </div>
</template>

<style scoped>
.sidebar {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #161622;
}

.sidebar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid #2a2a3e;
}

.sidebar-header h3 {
  margin: 0;
  font-size: 1.1rem;
  color: #e0e0e0;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.lang-toggle {
  display: flex;
  border: 1px solid #444;
  border-radius: 4px;
  overflow: hidden;
}

.lang-btn {
  background: transparent;
  border: none;
  color: #888;
  font-size: 0.7rem;
  font-weight: 600;
  padding: 0.2rem 0.45rem;
  cursor: pointer;
  transition: all 0.15s;
}

.lang-btn.active {
  background: #7c6ef0;
  color: #fff;
}

.lang-btn:hover:not(.active) {
  color: #ccc;
}

.btn-new {
  background: #7c6ef0;
  color: #fff;
  border: none;
  border-radius: 6px;
  padding: 0.35rem 0.75rem;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
}

.btn-new:hover {
  background: #6b5cd4;
}

.sim-list {
  flex: 1;
  overflow-y: auto;
  padding: 0.5rem;
}

.sim-item {
  position: relative;
  padding: 0.75rem 1rem;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.15s;
  margin-bottom: 2px;
}

.sim-item:hover {
  background: #1e1e2e;
}

.sim-item.active {
  background: #252540;
  border-left: 3px solid #7c6ef0;
}

.sim-item-title {
  font-size: 0.85rem;
  color: #d0d0d0;
  margin-bottom: 0.3rem;
  padding-right: 1.5rem;
}

.sim-item-meta {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.sim-status {
  font-size: 0.7rem;
  font-weight: 600;
  padding: 0.1rem 0.4rem;
  border-radius: 8px;
}

.st-running { background: #2d4a22; color: #8eff6a; }
.st-completed { background: #1a3a5c; color: #6ac5ff; }
.st-failed { background: #5c1a1a; color: #ff6a6a; }

.sim-date {
  font-size: 0.7rem;
  color: #666;
}

.btn-delete {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  background: none;
  border: none;
  color: #666;
  font-size: 1.1rem;
  cursor: pointer;
  padding: 0.1rem 0.3rem;
  border-radius: 4px;
  line-height: 1;
}

.btn-delete:hover {
  background: #5c1a1a;
  color: #ff6a6a;
}

.empty {
  padding: 2rem 1rem;
  text-align: center;
  color: #555;
  font-size: 0.85rem;
}
</style>
