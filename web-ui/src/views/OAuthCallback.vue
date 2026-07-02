<template>
  <div style="display:flex;justify-content:center;align-items:center;min-height:100vh">
    <el-card style="width:400px;max-width:90%;text-align:center">
      <el-icon v-if="loading" class="is-loading" :size="32" style="margin-bottom:16px"><Loading /></el-icon>
      <el-icon v-else-if="error" :size="32" style="margin-bottom:16px;color:#F56C6C"><CircleCloseFilled /></el-icon>
      <p v-if="loading">正在完成登录...</p>
      <p v-else-if="error" style="color:#F56C6C">{{ error }}</p>
      <el-button v-if="error" type="primary" @click="router.push('/login')" style="margin-top:12px">返回登录</el-button>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { oauthCallback } from '@/api/auth'
import { ElMessage } from 'element-plus'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  const code = route.query.code as string
  const state = route.query.state as string
  const provider = (route.query.provider as string) || 'google'

  if (!code) {
    error.value = '授权失败：未收到授权码'
    loading.value = false
    return
  }

  try {
    const res = await oauthCallback(provider, code, state || '')
    const d = res.data.data ?? res.data
    if (d.access_token) {
      auth.token = d.access_token
      auth.refreshToken = d.refresh_token
      auth.userId = d.user_id
      auth.accountId = d.account_id
      localStorage.setItem('access_token', d.access_token)
      localStorage.setItem('refresh_token', d.refresh_token)
      localStorage.setItem('user_id', String(d.user_id))
      localStorage.setItem('account_id', d.account_id)
      ElMessage.success(d.is_new_user ? '注册并登录成功' : '登录成功')
      router.push('/')
    } else {
      error.value = d.error || '登录失败'
      loading.value = false
    }
  } catch (e: any) {
    error.value = e?.response?.data?.error || 'OAuth 登录失败'
    loading.value = false
  }
})
</script>
