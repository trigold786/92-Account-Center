<template>
  <el-card style="width: 420px; max-width: 100%">
    <template #header><h2 style="text-align:center;color:var(--text-primary)">注册</h2></template>
    <el-form @submit.prevent="doRegister">
      <el-input v-model="form.phone" placeholder="手机号" style="margin-bottom:16px" @keyup.enter="doRegister" />
      <el-input v-model="form.password" type="password" placeholder="密码（至少6位）" show-password style="margin-bottom:16px" @keyup.enter="doRegister" />
      <el-input v-model="form.confirm" type="password" placeholder="确认密码" show-password style="margin-bottom:16px" @keyup.enter="doRegister" />
      <div style="display:flex;gap:8px;margin-bottom:16px">
        <el-input v-model="form.code" placeholder="验证码" @keyup.enter="doRegister" />
        <el-button :disabled="codeSending" @click="sendCode">{{ codeBtnText }}</el-button>
      </div>
      <el-checkbox v-model="agree" style="margin-bottom:16px">我已阅读并同意服务条款</el-checkbox>
      <el-button type="primary" style="width:100%" :loading="loading" @click="doRegister">注册</el-button>
    </el-form>
    <div style="text-align:center;margin-top:16px">
      <router-link to="/login" style="color:var(--brand-secondary)">已有账号？去登录</router-link>
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
import { register, sendSMSCode, verifySMSCode } from '@/api/auth'
import { useCaptcha } from '@/composables/useCaptcha'
import { ElMessage } from 'element-plus'

const router = useRouter()
const auth = useAuthStore()
const loading = ref(false)
const codeSending = ref(false)
const codeBtnText = ref('获取验证码')
const agree = ref(false)
const form = reactive({ phone: '', password: '', confirm: '', code: '' })
const captcha = useCaptcha()
const captchaVisible = ref(false)

async function sendCode() {
  if (!form.phone) { ElMessage.warning('请输入手机号'); return }
  captcha.refresh()
  captchaVisible.value = true
}

async function submitCaptcha() {
  if (!captcha.validate()) { ElMessage.warning('验证失败，请重试'); captcha.refresh(); return }
  captchaVisible.value = false
  codeSending.value = true
  try {
    await sendSMSCode(form.phone)
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

async function doRegister() {
  if (!form.phone || !form.password || !form.code) { ElMessage.warning('请填写完整'); return }
  if (form.password.length < 6) { ElMessage.warning('密码至少6位'); return }
  if (form.password !== form.confirm) { ElMessage.warning('两次密码不一致'); return }
  if (!agree.value) { ElMessage.warning('请同意服务条款'); return }
  loading.value = true
  try {
    await verifySMSCode(form.phone, form.code)
    await register({ phone: form.phone, password: form.password, code: form.code })
    await auth.doLogin(form.phone, form.password)
    ElMessage.success('注册成功')
    router.push('/')
  } catch (e: any) { ElMessage.error(e.message || '注册失败') }
  loading.value = false
}
</script>
