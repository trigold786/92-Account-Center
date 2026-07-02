<template>
  <div>
    <el-row :gutter="16">
      <el-col :xs="24" :sm="12"><el-card><template #header>积分余额</template><p style="font-size:48px;color:#6C63FF">{{ account?.balance || '--' }}</p></el-card></el-col>
      <el-col :xs="24" :sm="12"><el-card><template #header>累计</template><p>获得: {{ account?.total_earned || 0 }}</p><p>消耗: {{ account?.total_consumed || 0 }}</p></el-card></el-col>
    </el-row>
    <el-card style="margin-top:16px"><template #header>交易记录</template>
      <div style="overflow-x: auto; width: 100%">
      <el-table :data="transactions" style="width:100%" stripe>
        <el-table-column prop="created_at" label="时间" width="180" />
        <el-table-column prop="reason" label="原因" />
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }"><el-tag :type="row.type === 'earn' ? 'success' : 'danger'">{{ row.type }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="amount" label="金额" width="120">
          <template #default="{ row }"><span :style="{color: row.type === 'earn' ? '#2ED573' : '#FF4757', fontWeight:600}">{{ row.type === 'earn' ? '+' : '-' }}{{ row.amount }}</span></template>
        </el-table-column>
      </el-table>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/store/auth'
import { getCreditAccount, getTransactions } from '@/api/credits'
import type { Transaction } from '@/types/api'

const auth = useAuthStore()
interface CreditAccount { balance?: number; total_earned?: number; total_consumed?: number }
const account = ref<CreditAccount | null>(null)
const transactions = ref<Transaction[]>([])

onMounted(async () => {
  const uid = auth.userId; if (!uid) return
  try { const r = await getCreditAccount(uid); account.value = r.data.data } catch {}
  try { const r = await getTransactions(uid); transactions.value = r.data.data || [] } catch {}
})
</script>
