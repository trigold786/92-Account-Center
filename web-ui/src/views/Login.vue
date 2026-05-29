<template>
  <el-card style="width: 420px; max-width: 100%">
    <template #header><h2 style="text-align:center;color:var(--text-primary)">登录</h2></template>
    <el-tabs v-model="loginMode" stretch>
      <el-tab-pane label="密码登录" name="password">
        <el-form @submit.prevent="doLogin">
          <el-input v-model="form.credential" placeholder="手机号 / 邮箱 / 账号" style="margin-bottom:16px" @keyup.enter="doLogin" />
          <el-input v-model="form.password" type="password" placeholder="密码" show-password style="margin-bottom:16px" @keyup.enter="doLogin" />
          <el-button type="primary" style="width:100%" :loading="loading" @click="doLogin">登录</el-button>
        </el-form>
      </el-tab-pane>
      <el-tab-pane label="验证码登录" name="code">
        <el-form @submit.prevent="doLoginWithCode">
          <el-input v-model="form.credential" placeholder="手机号" style="margin-bottom:16px" @keyup.enter="doLoginWithCode" />
          <div style="display:flex;gap:8px;margin-bottom:16px">
            <el-input v-model="form.code" placeholder="验证码" @keyup.enter="doLoginWithCode" />
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

  <el-dialog v-model="captchaVisible" title="真人验证" width="380px" :close-on-click-modal="false" :show-close="false" align-center>
    <div style="text-align:center">
      <div style="margin-bottom:12px;font-size:14px;color:var(--text-secondary)">{{ captcha.captchaText.value }}</div>
      <el-button @click="captcha.refresh()" size="small" text style="margin-bottom:12px">换一个</el-button>

      <template v-if="captcha.challenge.value.type === 'math'">
        <el-input v-model="captcha.userAnswer.value" placeholder="输入答案" @keyup.enter="submitCaptcha" style="width:200px" />
      </template>

      <template v-else-if="captcha.challenge.value.type === 'category'">
        <div style="display:flex;gap:8px;justify-content:center;flex-wrap:wrap;margin-top:8px">
          <el-button v-for="opt in captcha.challenge.value.options" :key="opt.id" :type="captcha.userAnswer.value === opt.id ? 'primary' : 'default'" size="large" style="min-width:80px" @click="captcha.handleClick(opt.id)">{{ opt.label }}</el-button>
        </div>
      </template>

      <template v-else-if="captcha.challenge.value.type === 'shape'">
        <div style="display:flex;gap:8px;justify-content:center;flex-wrap:wrap;margin-top:8px">
          <el-button v-for="opt in captcha.challenge.value.options" :key="opt.id" :type="captcha.userAnswer.value === opt.id ? 'primary' : 'default'" size="large" style="font-size:28px;width:64px;height:64px" @click="captcha.handleClick(opt.id)">{{ opt.label }}</el-button>
        </div>
      </template>

      <template v-else-if="captcha.challenge.value.type === 'chinese' || captcha.challenge.value.type === 'letter'">
        <div style="display:flex;gap:6px;justify-content:center;flex-wrap:wrap;margin-top:8px;max-width:320px;margin-left:auto;margin-right:auto">
          <el-button v-for="opt in captcha.challenge.value.options" :key="opt.id" :type="captcha.isClicked(opt.id) ? 'success' : 'default'" :disabled="captcha.isClicked(opt.id)" size="default" style="font-size:16px;min-width:40px" @click="captcha.handleClick(opt.id)">{{ opt.label }}</el-button>
        </div>
      </template>

      <div style="margin-top:16px">
        <el-button type="primary" @click="submitCaptcha" :disabled="!captcha.userAnswer.value">确认</el-button>
        <el-button @click="captcha.refresh()">换一个</el-button>
      </div>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { sendSMSCode, loginWithCode } from '@/api/auth'
import { useCaptcha } from '@/composables/useCaptcha'
import { ElMessage } from 'element-plus'

const router = useRouter()
const auth = useAuthStore()
const loginMode = ref('password')
const loading = ref(false)
const codeSending = ref(false)
const codeBtnText = ref('获取验证码')
const form = reactive({ credential: '', password: '', code: '' })
const captcha = useCaptcha()
const captchaVisible = ref(false)

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
    const res = await loginWithCode(form.credential, form.code)
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
      ElMessage.success('登录成功')
      router.push('/')
    } else {
      ElMessage.error(d.message || d.error || '登录失败')
    }
  } catch (e: any) { ElMessage.error(e.message || '登录失败') }
  loading.value = false
}

async function sendCode() {
  if (!form.credential) { ElMessage.warning('请输入手机号'); return }
  captcha.refresh()
  captchaVisible.value = true
}

async function submitCaptcha() {
  if (!captcha.validate()) { ElMessage.warning('验证失败，请重试'); captcha.refresh(); return }
  captchaVisible.value = false
  codeSending.value = true
  try {
    await sendSMSCode(form.credential)
    const isDev = import.meta.env.DEV
    ElMessage.success(isDev ? '验证码已发送（开发模式验证码: 012345）' : '验证码已发送')
    let sec = 60; codeBtnText.value = `${sec}s`
    const timer = setInterval(() => { sec--; codeBtnText.value = `${sec}s`; if (sec <= 0) { clearInterval(timer); codeBtnText.value = '重新获取'; codeSending.value = false } }, 1000)
  } catch (e: any) {
    const msg = e?.response?.data?.error || e?.response?.data?.message || '发送失败'
    if (msg.includes('rate limit') || msg.includes('频繁')) {
      ElMessage.warning('发送过于频繁，请稍后再试')
    } else {
      ElMessage.error(msg)
    }
    codeSending.value = false
  }
}
</script>
