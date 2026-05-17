<template>
  <el-row :gutter="16">
    <el-col :span="12"><el-card><template #header>我的邀请码</template><h2 style="color:#00D4FF;font-size:32px">{{ summary?.referral_code || '--' }}</h2></el-card></el-col>
    <el-col :span="12"><el-card><template #header>返佣收益</template><h2 style="color:#2ED573;font-size:32px">{{ summary?.total_earnings || 0 }}</h2></el-card></el-col>
    <el-col :span="12" style="margin-top:16px"><el-card><template #header>邀请好友</template><p>{{ summary?.total_referrals || 0 }} 人</p></el-card></el-col>
    <el-col :span="12" style="margin-top:16px"><el-card><template #header>邀请链接</template><el-input v-model="link" readonly><template #append><el-button @click="copyLink">复制</el-button></template></el-input></el-card></el-col>
  </el-row>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/store/auth'
import { getReferralSummary, generateReferralLink } from '@/api/referral'
import { ElMessage } from 'element-plus'

const auth = useAuthStore()
const summary = ref<any>(null)
const link = ref('')

onMounted(async () => {
  const uid = auth.userId; if (!uid) return
  try { const r = await getReferralSummary(uid); summary.value = r.data.data } catch {}
  try { const r = await generateReferralLink(); link.value = r.data.data?.referral_link || '' } catch {}
})

function copyLink() { navigator.clipboard.writeText(link.value); ElMessage.success('已复制') }
</script>
