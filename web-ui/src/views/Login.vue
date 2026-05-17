<template>
  <el-card style="width: 420px; max-width: 100%">
    <template #header><h2 style="text-align:center;color:var(--text-primary)">登录</h2></template>
    <el-tabs v-model="loginMode" stretch>
      <el-tab-pane label="密码登录" name="password">
        <el-form @submit.prevent="doLogin">
          <el-input v-model="form.credential" placeholder="手机号 / 邮箱 / 账号" style="margin-bottom:16px" />
          <el-input v-model="form.password" type="password" placeholder="密码" show-password style="margin-bottom:16px" />
          <el-button type="primary" style="width:100%" :loading="loading" @click="doLogin">登录</el-button>
        </el-form>
      </el-tab-pane>
      <el-tab-pane label="验证码登录" name="code">
        <el-form @submit.prevent="doLoginWithCode">
          <el-input v-model="form.credential" placeholder="手机号" style="margin-bottom:16px" />
          <div style="display:flex;gap:8px;margin-bottom:16px">
            <el-input v-model="form.code" placeholder="验证码" />
            <el-button :disabled="codeSending" @click="sendCode">{{ codeBtnText }}</el-button>
          </div>
          <el-button type="primary" style="width:100%" :loading="loading" @click="doLoginWithCode">登录</el-button>
        </el-form>
      </el-tab-pane>
    </el-tabs>
    <div style="text-align:center;margin-top:16px">
      <router-link to="/register" style="color:var(--brand-secondary)">还没有账号？立即注册</router-link>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { sendSMSCode } from '@/api/auth'
import { ElMessage } from 'element-plus'

const router = useRouter()
const auth = useAuthStore()
const loginMode = ref('password')
const loading = ref(false)
const codeSending = ref(false)
const codeBtnText = ref('获取验证码')
const form = reactive({ credential: '', password: '', code: '' })

async function doLogin() {
  if (!form.credential || !form.password) { ElMessage.warning('请填写完整'); return }
  loading.value = true
  try {
    await auth.doLogin(form.credential, form.password)
    ElMessage.success('登录成功')
    router.push('/')
  } catch (e: any) { ElMessage.error(e.message || '登录失败') }
  loading.value = false
}

async function doLoginWithCode() {
  if (!form.credential || !form.code) { ElMessage.warning('请填写完整'); return }
  loading.value = true
  try {
    await auth.doLogin(form.credential, form.password)
    ElMessage.success('登录成功')
    router.push('/')
  } catch (e: any) { ElMessage.error(e.message || '登录失败') }
  loading.value = false
}

async function sendCode() {
  if (!form.credential) { ElMessage.warning('请输入手机号'); return }
  codeSending.value = true
  try {
    await sendSMSCode(form.credential)
    ElMessage.success('验证码已发送')
    let sec = 60; codeBtnText.value = `${sec}s`
    const timer = setInterval(() => { sec--; codeBtnText.value = `${sec}s`; if (sec <= 0) { clearInterval(timer); codeBtnText.value = '重新获取'; codeSending.value = false } }, 1000)
  } catch { ElMessage.error('发送失败'); codeSending.value = false }
}
</script>
