<script setup lang="ts">
import { useSimulationStore } from '@/stores/simulation'
import HistorySidebar from '@/components/HistorySidebar.vue'
import SimulationForm from '@/components/SimulationForm.vue'
import SimulationViewer from '@/components/SimulationViewer.vue'

const store = useSimulationStore()
</script>

<template>
  <div class="main-layout">
    <aside class="sidebar-panel">
      <HistorySidebar />
    </aside>
    <main class="content-panel">
      <SimulationViewer v-if="store.currentSimulation" />
      <SimulationForm v-else />
    </main>
  </div>
</template>

<style scoped>
.main-layout {
  display: flex;
  height: 100vh;
  background: #12121e;
  color: #e0e0e0;
}

.sidebar-panel {
  width: 300px;
  min-width: 300px;
  border-right: 1px solid #2a2a3e;
  overflow: hidden;
}

.content-panel {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
