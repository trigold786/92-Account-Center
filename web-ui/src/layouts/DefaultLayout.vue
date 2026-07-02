<template>
  <SkipLink />
  <el-container :style="{ minHeight: '100vh', flexDirection: isMobile ? 'column' : 'row' }">
    <div v-if="isMobile" class="mobile-header">
      <el-button :icon="Menu" @click="drawerOpen = true" text aria-label="打开导航菜单" />
      <span class="logo-text">Account Center</span>
    </div>

    <el-aside v-if="!isMobile" width="220px" style="background: #161B22; border-right: 1px solid #30363D" aria-label="侧边导航栏">
      <SidebarMenu />
    </el-aside>

    <el-drawer v-if="isMobile" v-model="drawerOpen" direction="ltr" size="220px" :with-header="false" aria-label="导航菜单">
      <SidebarMenu @navigate="drawerOpen = false" />
    </el-drawer>

    <el-main id="main-content" :style="{ background: '#0D1117', padding: isMobile ? '12px' : '24px' }">
      <router-view />
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Menu } from '@element-plus/icons-vue'
import { useBreakpoint } from '@/composables/useBreakpoint'
import SidebarMenu from '@/components/SidebarMenu.vue'
import SkipLink from '@/components/a11y/SkipLink.vue'

const { isMobile } = useBreakpoint()
const drawerOpen = ref(false)
</script>

<style scoped>
.mobile-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 16px;
  background: #161B22;
  border-bottom: 1px solid #30363D;
  position: sticky;
  top: 0;
  z-index: 100;
}
.logo-text { font-size: 20px; font-weight: 700; background: linear-gradient(135deg, #6C63FF, #00D4FF); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }
</style>
