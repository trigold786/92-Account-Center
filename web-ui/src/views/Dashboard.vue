<template>
  <div>
    <el-row :gutter="16">
      <el-col :span="6"><el-card><h3>RFM 评分</h3><p style="font-size:36px;color:#6C63FF">{{ rfm?.total_score || '--' }}</p></el-card></el-col>
      <el-col :span="6"><el-card><h3>积分余额</h3><p style="font-size:36px;color:#00D4FF">{{ credit?.balance || '--' }}</p></el-card></el-col>
      <el-col :span="6"><el-card><h3>订阅状态</h3><p style="font-size:24px;color:#2ED573">{{ subscription?.status || '无' }}</p></el-card></el-col>
      <el-col :span="6"><el-card><h3>用户等级</h3><p style="font-size:24px;color:#FFA502">{{ tier || '--' }}</p></el-card></el-col>
    </el-row>
    <el-card style="margin-top:16px">
      <template #header>快捷入口</template>
      <el-row :gutter="12">
        <el-col :span="6" v-for="item in quickLinks" :key="item.path">
          <el-card shadow="hover" style="cursor:pointer;text-align:center;margin-bottom:12px" @click="$router.push(item.path)">
            <p>{{ item.name }}</p>
          </el-card>
        </el-col>
      </el-row>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/store/auth'
import { getRFMScore } from '@/api/data'
import { getCreditAccount } from '@/api/credits'
import { getSubscriptions } from '@/api/subscriptions'
import { getTier } from '@/api/account'

const auth = useAuthStore()
const rfm = ref<any>(null)
const credit = ref<any>(null)
const subscription = ref<any>(null)
const tier = ref('')
const quickLinks = [
  { path: '/account', name: '账户设置' }, { path: '/credits', name: '积分' },
  { path: '/subscriptions', name: '订阅' }, { path: '/referral', name: '推荐' },
  { path: '/devices', name: '设备管理' }, { path: '/admin', name: '管理后台' },
]

onMounted(async () => {
  const uid = auth.userId
  if (!uid) return
  try { const r = await getRFMScore(uid); rfm.value = r.data.data } catch {}
  try { const r = await getCreditAccount(uid); credit.value = r.data.data } catch {}
  try { const r = await getSubscriptions(uid); subscription.value = r.data.data?.[0] } catch {}
  try { const r = await getTier(uid); tier.value = r.data.data?.tier } catch {}
})
</script>
