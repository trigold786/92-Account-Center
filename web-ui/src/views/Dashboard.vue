<template>
  <div>
    <el-row :gutter="16">
      <el-col :xs="12" :sm="12" :md="6"><el-card><h3>{{ $t('dashboard.rfm_score') }}</h3><p style="font-size:36px;color:#6C63FF">{{ rfm?.total_score || '--' }}</p></el-card></el-col>
      <el-col :xs="12" :sm="12" :md="6"><el-card><h3>{{ $t('dashboard.credit_balance') }}</h3><p style="font-size:36px;color:#00D4FF">{{ credit?.balance || '--' }}</p></el-card></el-col>
      <el-col :xs="12" :sm="12" :md="6"><el-card><h3>{{ $t('dashboard.subscription_status') }}</h3><p style="font-size:24px;color:#2ED573">{{ subscription?.status || $t('dashboard.no_subscription') }}</p></el-card></el-col>
      <el-col :xs="12" :sm="12" :md="6"><el-card><h3>{{ $t('dashboard.user_level') }}</h3><p style="font-size:24px;color:#FFA502">{{ tier || '--' }}</p></el-card></el-col>
    </el-row>
    <el-card style="margin-top:16px">
      <template #header>{{ $t('dashboard.quick_links') }}</template>
      <el-row :gutter="12">
        <el-col :xs="12" :sm="12" :md="6" v-for="item in quickLinks" :key="item.path">
          <el-card shadow="hover" style="cursor:pointer;text-align:center;margin-bottom:12px" @click="$router.push(item.path)">
            <p>{{ item.name }}</p>
          </el-card>
        </el-col>
      </el-row>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/store/auth'
import { usePermissionStore } from '@/store/permission'
import { getRFMScore } from '@/api/data'
import { getCreditAccount } from '@/api/credits'
import { getSubscriptions } from '@/api/subscriptions'
import { getTier } from '@/api/account'

const { t } = useI18n()
const auth = useAuthStore()
const perm = usePermissionStore()
interface RFMScore { total_score?: number }
const rfm = ref<RFMScore | null>(null)
interface CreditBalance { balance?: number }
const credit = ref<CreditBalance | null>(null)
interface SubData { status?: string }
const subscription = ref<SubData | null>(null)
const tier = ref('')
const quickLinks = computed(() => {
  const links = [
    { path: '/account', name: t('dashboard.account_settings') }, { path: '/credits', name: t('dashboard.credits') },
    { path: '/subscriptions', name: t('dashboard.subscription') }, { path: '/pricing', name: t('dashboard.view_pricing') }, { path: '/referral', name: t('dashboard.referral') },
    { path: '/devices', name: t('dashboard.device_management') },
  ]
  if (perm.hasAnyRole(['admin', 'system_owner', 'operator', 'finance'])) links.push({ path: '/finance', name: t('dashboard.finance_admin') })
  if (perm.hasAnyRole(['admin', 'system_owner', 'operator', 'finance', 'support'])) links.push({ path: '/admin', name: t('dashboard.admin_panel') })
  return links
})

onMounted(async () => {
  const uid = auth.userId
  if (!uid) return
  const [rfmRes, creditRes, subRes, tierRes] = await Promise.allSettled([
    getRFMScore(uid),
    getCreditAccount(uid),
    getSubscriptions(uid),
    getTier(uid),
  ])
  if (rfmRes.status === 'fulfilled') rfm.value = rfmRes.value.data.data
  if (creditRes.status === 'fulfilled') credit.value = creditRes.value.data.data
  if (subRes.status === 'fulfilled') subscription.value = subRes.value.data.data?.[0]
  if (tierRes.status === 'fulfilled') tier.value = tierRes.value.data.data?.tier
})
</script>
