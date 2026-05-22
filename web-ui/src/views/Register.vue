<template>
  <el-card style="width: 420px; max-width: 100%">
    <template #header><h2 style="text-align:center;color:var(--text-primary)">注册</h2></template>
    <el-form>
      <el-input v-model="form.phone" placeholder="手机号" style="margin-bottom:16px" />
      <el-input v-model="form.password" type="password" placeholder="密码（至少6位）" show-password style="margin-bottom:16px" />
      <el-input v-model="form.confirm" type="password" placeholder="确认密码" show-password style="margin-bottom:16px" />
      <div style="display:flex;gap:8px;margin-bottom:16px">
        <el-input v-model="form.code" placeholder="验证码" />
        <el-button :disabled="codeSending" @click="sendCode">{{ codeBtnText }}</el-button>
      </div>
      <el-checkbox v-model="agree" style="margin-bottom:16px">我已阅读并同意服务条款</el-checkbox>
      <el-button type="primary" style="width:100%" :loading="loading" @click="doRegister">注册</el-button>
    </el-form>
    <div style="text-align:center;margin-top:16px">
      <router-link to="/login" style="color:var(--brand-secondary)">已有账号？去登录</router-link>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { register, sendSMSCode } from '@/api/auth'
import { ElMessage } from 'element-plus'

const router = useRouter()
const auth = useAuthStore()
const loading = ref(false)
const codeSending = ref(false)
const codeBtnText = ref('获取验证码')
const agree = ref(false)
const form = reactive({ phone: '', password: '', confirm: '', code: '' })

async function sendCode() {
  if (!form.phone) { ElMessage.warning('请输入手机号'); return }
  codeSending.value = true
  try {
    await sendSMSCode(form.phone)
    ElMessage.success('验证码已发送')
    let sec = 60; codeBtnText.value = `${sec}s`
    const timer = setInterval(() => { sec--; codeBtnText.value = `${sec}s`; if (sec <= 0) { clearInterval(timer); codeBtnText.value = '重新获取'; codeSending.value = false } }, 1000)
  } catch { ElMessage.error('发送失败'); codeSending.value = false }
}

async function doRegister() {
  if (!form.phone || !form.password || !form.code) { ElMessage.warning('请填写完整'); return }
  if (form.password.length < 6) { ElMessage.warning('密码至少6位'); return }
  if (form.password !== form.confirm) { ElMessage.warning('两次密码不一致'); return }
  if (!agree.value) { ElMessage.warning('请同意服务条款'); return }
  loading.value = true
  try {
    await register({ phone: form.phone, password: form.password, code: form.code })
    await auth.doLogin(form.phone, form.password)
    ElMessage.success('注册成功')
    router.push('/')
  } catch (e: any) { ElMessage.error(e.message || '注册失败') }
  loading.value = false
}
</script>
