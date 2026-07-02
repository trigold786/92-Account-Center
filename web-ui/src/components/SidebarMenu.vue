<template>
  <div class="sidebar-menu">
    <div class="logo">
      <span class="logo-text">Account Center</span>
    </div>
    <el-menu :router="true" background-color="#161B22" text-color="#8B949E" active-text-color="#6C63FF" style="border: none" @select="onSelect">
      <el-menu-item index="/" v-if="hasPermission('nav.dashboard')"><el-icon><Odometer /></el-icon>仪表盘</el-menu-item>
      <el-menu-item index="/account" v-if="hasPermission('nav.account')"><el-icon><User /></el-icon>我的账户</el-menu-item>
      <el-menu-item index="/credits" v-if="hasPermission('nav.credits')"><el-icon><Coin /></el-icon>积分</el-menu-item>
      <el-menu-item index="/subscriptions" v-if="hasPermission('nav.subscriptions')"><el-icon><Ticket /></el-icon>订阅</el-menu-item>
      <el-menu-item index="/pricing" v-if="hasPermission('nav.subscriptions')"><el-icon><PriceTag /></el-icon>定价</el-menu-item>
      <el-menu-item index="/referral" v-if="hasPermission('nav.referral')"><el-icon><Share /></el-icon>推荐</el-menu-item>
      <el-menu-item index="/devices" v-if="hasPermission('nav.devices')"><el-icon><Monitor /></el-icon>设备</el-menu-item>
      <el-menu-item index="/finance" v-if="hasPermission('nav.finance') || hasAnyRole(['admin', 'system_owner', 'operator', 'finance'])"><el-icon><Money /></el-icon>财务后台</el-menu-item>
      <el-menu-item index="/admin" v-if="hasPermission('nav.admin')"><el-icon><Setting /></el-icon>管理后台</el-menu-item>
    </el-menu>
    <div class="logout-area">
      <el-button type="danger" plain style="width: 90%; margin: 0 5%" @click="handleLogout">退出登录</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { usePermissionStore } from '@/store/permission'
import { ElMessage } from 'element-plus'
import { Odometer, User, Coin, Ticket, PriceTag, Share, Monitor, Money, Setting } from '@element-plus/icons-vue'

const emit = defineEmits<{ (e: 'navigate'): void }>()

const router = useRouter()
const auth = useAuthStore()
const perm = usePermissionStore()
const hasPermission = (p: string) => perm.hasPermission(p)
const hasAnyRole = (roles: string[]) => perm.hasAnyRole(roles)

function onSelect() {
  emit('navigate')
}

async function handleLogout() {
  await auth.doLogout()
  ElMessage.success('已退出登录')
  emit('navigate')
  router.push('/login')
}
</script>

<style scoped>
.sidebar-menu { position: relative; height: 100%; }
.logo { padding: 20px; text-align: center; border-bottom: 1px solid #30363D; margin-bottom: 8px; }
.logo-text { font-size: 20px; font-weight: 700; background: linear-gradient(135deg, #6C63FF, #00D4FF); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }
.logout-area { position: absolute; bottom: 20px; width: 100%; left: 0; }
</style>
