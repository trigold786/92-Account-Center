<template>
  <el-card style="width: 420px; max-width: 100%">
    <template #header><h2 style="text-align:center;color:var(--text-primary)">{{ $t('register.title') }}</h2></template>
    <el-form @submit.prevent="doRegister">
      <el-input v-model="form.phone" :placeholder="$t('register.phone_placeholder')" :aria-label="$t('register.phone_placeholder')" style="margin-bottom:16px" @keyup.enter="doRegister" />
      <el-input v-model="form.password" type="password" :placeholder="$t('register.password_placeholder')" show-password :aria-label="$t('register.password_placeholder')" style="margin-bottom:16px" @keyup.enter="doRegister" />
      <el-input v-model="form.confirm" type="password" :placeholder="$t('register.confirm_placeholder')" show-password :aria-label="$t('register.confirm_placeholder')" style="margin-bottom:16px" @keyup.enter="doRegister" />
      <div style="display:flex;gap:8px;margin-bottom:16px">
        <el-input v-model="form.code" :placeholder="$t('register.code_placeholder')" :aria-label="$t('register.code_placeholder')" @keyup.enter="doRegister" />
        <el-button :disabled="codeSending" @click="sendCode">{{ codeBtnText }}</el-button>
      </div>
      <el-checkbox v-model="agree" style="margin-bottom:16px">{{ $t('register.agree_terms') }}</el-checkbox>
      <el-button type="primary" style="width:100%" :loading="loading" @click="doRegister">{{ $t('register.register_button') }}</el-button>
    </el-form>
    <div style="text-align:center;margin-top:16px">
      <router-link to="/login" style="color:var(--brand-secondary)">{{ $t('register.has_account') }}</router-link>
    </div>
  </el-card>

  <el-dialog v-model="captchaVisible" :title="$t('register.captcha_title')" width="min(380px, 90vw)" :close-on-click-modal="false" :show-close="false" align-center>
    <div style="text-align:center">
      <div style="margin-bottom:12px;font-size:14px;color:var(--text-secondary)">{{ captcha.captchaText.value }}</div>
      <el-button @click="captcha.refresh()" size="small" text style="margin-bottom:12px">{{ $t('register.captcha_change') }}</el-button>

      <template v-if="captcha.challenge.value.type === 'math'">
        <el-input v-model="captcha.userAnswer.value" :placeholder="$t('register.captcha_input_placeholder')" @keyup.enter="submitCaptcha" style="width:200px" />
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
        <el-button type="primary" @click="submitCaptcha" :disabled="!captcha.userAnswer.value">{{ $t('common.confirm') }}</el-button>
        <el-button @click="captcha.refresh()">{{ $t('register.captcha_change') }}</el-button>
      </div>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/store/auth'
import { register, sendSMSCode, verifySMSCode } from '@/api/auth'
import { useCaptcha } from '@/composables/useCaptcha'
import { ElMessage } from 'element-plus'

const { t } = useI18n()
const router = useRouter()
const auth = useAuthStore()
const loading = ref(false)
const codeSending = ref(false)
const codeBtnText = ref(t('register.get_code'))
const agree = ref(false)
const form = reactive({ phone: '', password: '', confirm: '', code: '' })
const captcha = useCaptcha()
const captchaVisible = ref(false)

async function sendCode() {
  if (!form.phone) { ElMessage.warning(t('messages.enter_phone')); return }
  captcha.refresh()
  captchaVisible.value = true
}

async function submitCaptcha() {
  if (!captcha.validate()) { ElMessage.warning(t('messages.verify_failed')); captcha.refresh(); return }
  captchaVisible.value = false
  codeSending.value = true
  try {
    await sendSMSCode(form.phone)
    const isDev = import.meta.env.DEV
    ElMessage.success(isDev ? t('messages.code_sent_dev') : t('messages.code_sent'))
    let sec = 60; codeBtnText.value = `${sec}s`
    const timer = setInterval(() => { sec--; codeBtnText.value = `${sec}s`; if (sec <= 0) { clearInterval(timer); codeBtnText.value = t('register.resend_code'); codeSending.value = false } }, 1000)
  } catch (e: any) {
    const msg = e?.response?.data?.error || e?.response?.data?.message || t('messages.code_send_failed')
    if (msg.includes('rate limit') || msg.includes('频繁')) {
      ElMessage.warning(t('messages.code_rate_limit'))
    } else {
      ElMessage.error(msg)
    }
    codeSending.value = false
  }
}

async function doRegister() {
  if (!form.phone || !form.password || !form.code) { ElMessage.warning(t('messages.fill_complete')); return }
  if (form.password.length < 6) { ElMessage.warning(t('messages.password_min_length')); return }
  if (form.password !== form.confirm) { ElMessage.warning(t('messages.password_mismatch')); return }
  if (!agree.value) { ElMessage.warning(t('messages.agree_terms')); return }
  loading.value = true
  try {
    await verifySMSCode(form.phone, form.code)
    await register({ phone: form.phone, password: form.password, code: form.code })
    await auth.doLogin(form.phone, form.password)
    ElMessage.success(t('messages.register_success'))
    router.push('/')
  } catch (e: any) { ElMessage.error(e.message || t('messages.register_failed')) }
  loading.value = false
}
</script>
