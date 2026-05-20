<template>
  <div class="skeleton-screen" :class="{ 'skeleton-fade-in': !loading }">
    <template v-if="loading">
      <div v-if="layout === 'dashboard'" class="skeleton-layout">
        <div class="skeleton-line skeleton-pulse" style="width: 40%; height: 24px; margin-bottom: 24px;" />
        <div class="skeleton-grid">
          <div v-for="i in 4" :key="i" class="skeleton-card skeleton-pulse">
            <div class="skeleton-circle skeleton-pulse" style="width: 40px; height: 40px;" />
            <div class="skeleton-line skeleton-pulse" style="width: 60%; height: 16px; margin-top: 12px;" />
            <div class="skeleton-line skeleton-pulse" style="width: 40%; height: 24px; margin-top: 8px;" />
          </div>
        </div>
        <div class="skeleton-card skeleton-pulse" style="margin-top: 16px; height: 200px;" />
      </div>

      <div v-else-if="layout === 'credits'" class="skeleton-layout">
        <div class="skeleton-line skeleton-pulse" style="width: 30%; height: 24px; margin-bottom: 24px;" />
        <div class="skeleton-card skeleton-pulse" style="height: 120px; margin-bottom: 16px;">
          <div class="skeleton-line skeleton-pulse" style="width: 50%; height: 32px;" />
          <div class="skeleton-line skeleton-pulse" style="width: 30%; height: 16px; margin-top: 8px;" />
        </div>
        <div v-for="i in 5" :key="i" class="skeleton-card skeleton-pulse" style="height: 64px; margin-bottom: 8px;" />
      </div>

      <div v-else-if="layout === 'subscription'" class="skeleton-layout">
        <div class="skeleton-line skeleton-pulse" style="width: 35%; height: 24px; margin-bottom: 24px;" />
        <div class="skeleton-grid" style="grid-template-columns: repeat(3, 1fr);">
          <div v-for="i in 3" :key="i" class="skeleton-card skeleton-pulse" style="height: 280px;" />
        </div>
      </div>

      <div v-else-if="layout === 'referral'" class="skeleton-layout">
        <div class="skeleton-line skeleton-pulse" style="width: 25%; height: 24px; margin-bottom: 24px;" />
        <div class="skeleton-card skeleton-pulse" style="height: 160px; margin-bottom: 16px;" />
        <div class="skeleton-card skeleton-pulse" style="height: 100px;" />
      </div>

      <div v-else-if="layout === 'orders'" class="skeleton-layout">
        <div class="skeleton-line skeleton-pulse" style="width: 30%; height: 24px; margin-bottom: 24px;" />
        <div v-for="i in 6" :key="i" class="skeleton-card skeleton-pulse" style="height: 72px; margin-bottom: 8px;">
          <div class="skeleton-line skeleton-pulse" style="width: 50%; height: 16px; margin-bottom: 8px;" />
          <div class="skeleton-line skeleton-pulse" style="width: 30%; height: 12px;" />
        </div>
      </div>

      <div v-else class="skeleton-layout">
        <slot />
      </div>
    </template>

    <div v-else class="skeleton-content">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  loading: boolean
  layout?: 'dashboard' | 'credits' | 'subscription' | 'referral' | 'orders' | 'custom'
}>()
</script>

<style scoped>
.skeleton-screen {
  width: 100%;
}

.skeleton-layout {
  padding: 16px;
}

.skeleton-card {
  background: var(--bg-card);
  border-radius: 12px;
  padding: 16px;
}

.skeleton-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 16px;
}

.skeleton-line {
  border-radius: 4px;
  background: linear-gradient(90deg, #1C2333 25%, #2D3548 50%, #1C2333 75%);
  background-size: 200% 100%;
}

.skeleton-circle {
  border-radius: 50%;
  background: linear-gradient(90deg, #1C2333 25%, #2D3548 50%, #1C2333 75%);
  background-size: 200% 100%;
}

.skeleton-pulse {
  animation: skeleton-pulse 1.5s ease-in-out infinite;
}

@keyframes skeleton-pulse {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

.skeleton-fade-in {
  animation: skeleton-fade-in 300ms ease-out;
}

@keyframes skeleton-fade-in {
  from { opacity: 0; }
  to { opacity: 1; }
}

.skeleton-content {
  animation: skeleton-fade-in 300ms ease-out;
}
</style>
