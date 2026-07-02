<template>
  <div style="max-width: 1200px; margin: 0 auto; padding: 20px">
    <h1 style="text-align:center;color:var(--text-primary);margin-bottom:8px">选择适合您的方案</h1>
    <p style="text-align:center;color:var(--text-secondary);margin-bottom:32px">透明的价格，灵活的积分抵扣</p>

    <!-- Monthly/Yearly toggle -->
    <div style="text-align:center;margin-bottom:32px">
      <el-radio-group v-model="billingCycle">
        <el-radio-button label="monthly">月付</el-radio-button>
        <el-radio-button label="yearly">年付（省约17%）</el-radio-button>
      </el-radio-group>
    </div>

    <!-- Tier cards -->
    <el-row :gutter="20" style="margin-bottom:40px">
      <el-col :xs="24" :sm="12" :md="8" v-for="tier in pricing?.tiers" :key="tier.tier_level">
        <el-card :shadow="tier.tier_level === 3 ? 'always' : 'hover'" style="text-align:center;position:relative">
          <el-tag v-if="tier.tier_level === 3" type="warning" style="position:absolute;top:-10px;left:50%;transform:translateX(-50%)">推荐</el-tag>
          <h2 style="color:var(--text-primary);margin-bottom:8px">{{ tier.name }}</h2>
          <div style="margin-bottom:16px">
            <span style="font-size:36px;font-weight:700;color:var(--brand-secondary)">
              ¥{{ billingCycle === 'yearly' ? tier.yearly_price : tier.monthly_price }}
            </span>
            <span style="color:var(--text-secondary)">/{{ billingCycle === 'yearly' ? '年' : '月' }}</span>
            <div v-if="billingCycle === 'yearly'" style="font-size:12px;color:var(--text-secondary);margin-top:4px">
              约¥{{ (tier.yearly_price / 12).toFixed(1) }}/月
            </div>
          </div>

          <!-- Features -->
          <div style="text-align:left;margin-bottom:20px">
            <div v-for="feat in tier.features" :key="feat.name" style="margin-bottom:8px;display:flex;align-items:center;gap:8px">
              <el-icon v-if="feat.included" color="#67C23A"><Check /></el-icon>
              <el-icon v-else color="#909399"><Close /></el-icon>
              <span :style="{ color: feat.included ? 'var(--text-primary)' : 'var(--text-secondary)', fontSize: '14px' }">{{ feat.name }}</span>
              <span style="color:var(--text-secondary);font-size:12px;margin-left:auto">{{ feat.description }}</span>
            </div>
          </div>

          <!-- Entitlements -->
          <div style="margin-bottom:20px">
            <el-tag v-for="ent in tier.entitlements" :key="ent.feature_code" style="margin:2px">
              {{ ent.feature_code }}: {{ ent.quota }}{{ ent.unit }}
            </el-tag>
          </div>

          <el-button type="primary" style="width:100%" @click="handlePurchase(tier)">选择此方案</el-button>
        </el-card>
      </el-col>
    </el-row>

    <!-- Credit discount calculator -->
    <el-card shadow="hover">
      <h3 style="color:var(--text-primary);margin-bottom:16px">积分抵扣计算器</h3>
      <p style="color:var(--text-secondary);font-size:13px;margin-bottom:16px">
        积分兑换率: ¥{{ pricing?.credit_rate || 0.01 }}/积分，最高可抵扣 {{ ((pricing?.max_credit_discount || 0.5) * 100) }}%
      </p>
      <el-form label-width="80px" style="max-width:500px">
        <el-form-item label="原价">
          <el-input-number v-model="calcForm.price" :min="0.01" :precision="2" style="width:100%" />
        </el-form-item>
        <el-form-item label="使用积分">
          <el-input-number v-model="calcForm.credits" :min="0" :step="100" style="width:100%" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="calcLoading" @click="doCalculate">计算折扣</el-button>
        </el-form-item>
      </el-form>
      <div v-if="calcResult" style="margin-top:16px">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="原价">¥{{ calcResult.original_price }}</el-descriptions-item>
          <el-descriptions-item label="使用积分">{{ calcResult.credits_used }}</el-descriptions-item>
          <el-descriptions-item label="积分价值">¥{{ calcResult.credit_value }}</el-descriptions-item>
          <el-descriptions-item label="折扣率">{{ calcResult.discount_percent }}%</el-descriptions-item>
          <el-descriptions-item label="最终价格"><span style="font-size:20px;font-weight:700;color:var(--brand-secondary)">¥{{ calcResult.final_price }}</span></el-descriptions-item>
        </el-descriptions>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Check, Close } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { getPricing, calculateDiscount } from '@/api/pricing'
import type { PricingPageData, CreditDiscountResult, PricingTier } from '@/api/pricing'

const router = useRouter()
const pricing = ref<PricingPageData | null>(null)
const billingCycle = ref<'monthly' | 'yearly'>('monthly')
const calcForm = reactive({ price: 29.9, credits: 1000 })
const calcLoading = ref(false)
const calcResult = ref<CreditDiscountResult | null>(null)

onMounted(async () => {
  try {
    const res = await getPricing()
    pricing.value = res.data
  } catch (e: any) {
    ElMessage.error('获取定价信息失败')
  }
})

async function doCalculate() {
  calcLoading.value = true
  try {
    const res = await calculateDiscount(calcForm.price, calcForm.credits)
    calcResult.value = res.data
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '计算失败')
  }
  calcLoading.value = false
}

function handlePurchase(tier: PricingTier) {
  const token = localStorage.getItem('access_token')
  if (token) {
    router.push('/subscriptions')
  } else {
    router.push('/login?redirect=/pricing')
  }
}
</script>
