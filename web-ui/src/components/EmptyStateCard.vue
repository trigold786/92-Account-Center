<template>
  <div class="empty-state-card">
    <div class="empty-state-icon">{{ icon }}</div>
    <h3 class="empty-state-title">{{ title }}</h3>
    <p class="empty-state-description">{{ description }}</p>
    <router-link v-if="actionRoute" :to="actionRoute" class="empty-state-action">
      {{ actionText }}
    </router-link>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  icon: string
  title: string
  description: string
  actionText?: string
  actionRoute?: string
}>()
</script>

<style scoped>
.empty-state-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 24px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 16px;
  text-align: center;
}

.empty-state-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.empty-state-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.empty-state-description {
  font-size: 14px;
  color: var(--text-secondary);
  max-width: 320px;
  line-height: 1.5;
  margin-bottom: 24px;
}

.empty-state-action {
  display: inline-block;
  padding: 10px 32px;
  background: linear-gradient(135deg, var(--brand-primary), var(--brand-secondary));
  color: #fff;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  text-decoration: none;
  transition: opacity 0.2s ease;
}

.empty-state-action:hover {
  opacity: 0.9;
}
</style>
