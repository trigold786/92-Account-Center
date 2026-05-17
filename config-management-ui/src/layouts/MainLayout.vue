<template>
  <el-container style="height: 100vh">
    <el-aside width="220px" style="background: #1d1e1f; color: #fff">
      <div style="padding: 20px; font-size: 18px; font-weight: bold; border-bottom: 1px solid #333">
        ⚙️ 配置管理后台
      </div>
      <el-menu
        :default-active="route.path"
        router
        background-color="#1d1e1f"
        text-color="#bfcbd9"
        active-text-color="#409eff"
        style="border-right: none"
      >
        <el-menu-item index="/dashboard">
          <el-icon><DataBoard /></el-icon>
          <span>配置总览</span>
        </el-menu-item>
        <el-menu-item index="/config">
          <el-icon><Setting /></el-icon>
          <span>配置管理</span>
        </el-menu-item>
        <el-menu-item index="/release">
          <el-icon><Upload /></el-icon>
          <span>发布审批</span>
        </el-menu-item>
        <el-menu-item index="/audit">
          <el-icon><Document /></el-icon>
          <span>审计日志</span>
        </el-menu-item>
        <el-menu-item index="/permission">
          <el-icon><User /></el-icon>
          <span>权限管理</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header style="background: #fff; border-bottom: 1px solid #e4e7ed; display: flex; align-items: center; justify-content: space-between">
        <h2 style="margin: 0; font-size: 16px">{{ route.meta.title }}</h2>
        <div>
          <el-input
            v-model="operatorInput"
            placeholder="操作人"
            size="small"
            style="width: 150px; margin-right: 8px"
            @keyup.enter="updateOperator"
          />
          <el-button size="small" type="primary" @click="updateOperator">切换</el-button>
        </div>
      </el-header>
      <el-main style="background: #f5f7fa">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'

const route = useRoute()
const userStore = useUserStore()
const operatorInput = ref(userStore.operator)

function updateOperator() {
  if (operatorInput.value) {
    userStore.setOperator(operatorInput.value)
  }
}
</script>
