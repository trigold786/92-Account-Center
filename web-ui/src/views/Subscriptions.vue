<template>
  <div>
    <el-row :gutter="16">
      <el-col :span="24">
        <el-card>
          <template #header>当前订阅</template>
          <div v-if="subscription">
            <h3>{{ subscription.plan_name }}</h3>
            <p>状态: <el-tag :type="statusType">{{ subscription.status }}</el-tag></p>
            <p>开始: {{ subscription.current_period_start }}</p>
            <p>到期: {{ subscription.current_period_end }}</p>
          </div>
          <p v-else style="color:var(--text-secondary)">暂无订阅</p>
        </el-card>
      </el-col>
    </el-row>
    <el-card style="margin-top:16px"><template #header>用户等级</template><h2 style="color:#6C63FF">{{ tier || '--' }}</h2></el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAuthStore } from '@/store/auth'
import { getSubscriptions } from '@/api/subscriptions'
import { getTier } from '@/api/account'
import type { Subscription } from '@/types/api'

const auth = useAuthStore()
interface SubscriptionData { plan_name: string; status: string; current_period_start: string; current_period_end: string }
const subscription = ref<SubscriptionData | null>(null)
const tier = ref('')
const statusType = computed(() => {
  const statusMap: Record<string, string> = { active: 'success', expired: 'danger', cancelled: 'info', trialing: 'warning' }
  return statusMap[subscription.value?.status ?? ''] || 'info'
})

onMounted(async () => {
  const uid = auth.userId; if (!uid) return
  try { const r = await getSubscriptions(uid); subscription.value = r.data.data?.[0] } catch {}
  try { const r = await getTier(uid); tier.value = r.data.data?.tier } catch {}
})
</script>
