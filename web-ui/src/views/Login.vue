<template>
  <el-card style="width: 420px; max-width: 100%">
    <template #header><h2 style="text-align:center;color:var(--text-primary)">{{ $t('login.title') }}</h2></template>
    <el-tabs v-model="loginMode" stretch>
      <el-tab-pane :label="$t('login.password_tab')" name="password">
        <el-form @submit.prevent="doLogin">
          <el-input v-model="form.credential" :placeholder="$t('login.credential_placeholder')" :aria-label="$t('login.credential_placeholder')" style="margin-bottom:16px" @keyup.enter="doLogin" />
          <el-input v-model="form.password" type="password" :placeholder="$t('login.password_placeholder')" show-password :aria-label="$t('login.password_placeholder')" style="margin-bottom:16px" @keyup.enter="doLogin" />
          <el-button type="primary" style="width:100%" :loading="loading" @click="doLogin">{{ $t('login.login_button') }}</el-button>
        </el-form>
      </el-tab-pane>
      <el-tab-pane :label="$t('login.code_tab')" name="code">
        <el-form @submit.prevent="doLoginWithCode">
          <el-input v-model="form.credential" :placeholder="$t('login.phone_placeholder')" :aria-label="$t('login.phone_placeholder')" style="margin-bottom:16px" @keyup.enter="doLoginWithCode" />
          <div style="display:flex;gap:8px;margin-bottom:16px">
            <el-input v-model="form.code" :placeholder="$t('login.code_placeholder')" :aria-label="$t('login.code_placeholder')" @keyup.enter="doLoginWithCode" />
            <el-button :disabled="codeSending" @click="sendCode">{{ codeBtnText }}</el-button>
          </div>
          <el-button type="primary" style="width:100%" :loading="loading" @click="doLoginWithCode">{{ $t('login.login_button') }}</el-button>
        </el-form>
      </el-tab-pane>
    </el-tabs>
    <div style="text-align:center;margin-top:12px">
      <el-divider content-position="center" style="margin:12px 0"><span style="color:var(--text-secondary);font-size:12px">{{ $t('login.third_party_divider') }}</span></el-divider>
      <div style="display:flex;gap:12px;justify-content:center">
        <el-button @click="oauthLogin('google')" :loading="oauthLoading" style="padding:8px 16px">
          <svg viewBox="0 0 24 24" width="18" height="18" style="margin-right:6px"><path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z" fill="#4285F4"/><path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/><path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" fill="#FBBC05"/><path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/></svg>
          Google
        </el-button>
        <el-button @click="oauthLogin('wechat')" :loading="oauthLoading" style="padding:8px 16px">
          <svg viewBox="0 0 24 24" width="18" height="18" style="margin-right:6px"><path d="M8.691 2.188C3.891 2.188 0 5.476 0 9.53c0 2.212 1.17 4.203 3.002 5.55a.59.59 0 0 1 .213.665l-.39 1.48c-.019.07-.048.141-.048.213 0 .163.13.295.29.295a.326.326 0 0 0 .167-.054l1.903-1.114a.864.864 0 0 1 .717-.098 10.16 10.16 0 0 0 2.837.403c.276 0 .543-.027.811-.05a6.329 6.329 0 0 1-.256-1.786c0-3.64 3.393-6.594 7.579-6.594.26 0 .514.015.764.04C16.871 4.578 13.122 2.188 8.691 2.188z" fill="#51C332"/><path d="M22.766 14.846c0-3.115-3.016-5.64-6.735-5.64s-6.735 2.525-6.735 5.64c0 3.116 3.016 5.64 6.735 5.64.71 0 1.394-.098 2.036-.269a.67.67 0 0 1 .559.076l1.483.867a.254.254 0 0 0 .13.043.228.228 0 0 0 .226-.229c0-.056-.023-.111-.038-.165l-.305-1.153a.46.46 0 0 1 .166-.519c1.426-1.046 2.353-2.61 2.353-4.267h.126z" fill="#51C332"/></svg>
          {{ $t('login.wechat') }}
        </el-button>
        <el-button @click="oauthLogin('apple')" :loading="oauthLoading" style="padding:8px 16px">
          <svg viewBox="0 0 24 24" width="18" height="18" style="margin-right:6px"><path d="M17.05 20.28c-.98.95-2.05.88-3.08.4-1.09-.5-2.08-.48-3.24 0-1.44.62-2.2.44-3.06-.4C2.79 15.25 3.51 7.59 9.05 7.31c1.35.07 2.29.74 3.08.8 1.18-.24 2.31-.93 3.57-.84 1.51.12 2.65.72 3.4 1.8-3.12 1.87-2.38 5.98.48 7.13-.57 1.5-1.31 2.99-2.54 4.09zM12.03 7.25c-.15-2.23 1.66-4.07 3.74-4.25.29 2.58-2.34 4.5-3.74 4.25z" fill="#000"/></svg>
          Apple
        </el-button>
      </div>
    </div>
    <div style="text-align:center;margin-top:16px">
      <router-link to="/register" style="color:var(--brand-secondary)">{{ $t('login.no_account') }}</router-link>
    </div>
  </el-card>

  <el-dialog v-model="captchaVisible" :title="$t('login.captcha_title')" width="min(380px, 90vw)" :close-on-click-modal="false" :show-close="false" align-center>
    <div style="text-align:center">
      <div style="margin-bottom:12px;font-size:14px;color:var(--text-secondary)">{{ captcha.captchaText.value }}</div>
      <el-button @click="captcha.refresh()" size="small" text style="margin-bottom:12px">{{ $t('login.captcha_change') }}</el-button>

      <template v-if="captcha.challenge.value.type === 'math'">
        <el-input v-model="captcha.userAnswer.value" :placeholder="$t('login.captcha_input_placeholder')" @keyup.enter="submitCaptcha" style="width:200px" />
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
        <el-button @click="captcha.refresh()">{{ $t('login.captcha_change') }}</el-button>
      </div>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/store/auth'
import { sendSMSCode, loginWithCode, oauthAuthorize } from '@/api/auth'
import { useCaptcha } from '@/composables/useCaptcha'
import { ElMessage } from 'element-plus'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const loginMode = ref('password')
const loading = ref(false)
const codeSending = ref(false)
const codeBtnText = ref(t('login.get_code'))
const form = reactive({ credential: '', password: '', code: '' })
const captcha = useCaptcha()
const captchaVisible = ref(false)
const oauthLoading = ref(false)

async function doLogin() {
  if (!form.credential || !form.password) { ElMessage.warning(t('messages.fill_complete')); return }
  loading.value = true
  try {
    await auth.doLogin(form.credential, form.password)
    ElMessage.success(t('messages.login_success'))
    const redirect = (route.query.redirect as string) || '/'
    const safeRedirect = redirect.startsWith('/') && !redirect.startsWith('//') ? redirect : '/'
    router.push(safeRedirect)
  } catch (e: any) { ElMessage.error(e.message || t('messages.login_failed')) }
  loading.value = false
}

async function doLoginWithCode() {
  if (!form.credential || !form.code) { ElMessage.warning(t('messages.fill_complete')); return }
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
      ElMessage.success(t('messages.login_success'))
      const redirect = (route.query.redirect as string) || '/'
      const safeRedirect = redirect.startsWith('/') && !redirect.startsWith('//') ? redirect : '/'
      router.push(safeRedirect)
    } else {
      ElMessage.error(d.message || d.error || t('messages.login_failed'))
    }
  } catch (e: any) { ElMessage.error(e.message || t('messages.login_failed')) }
  loading.value = false
}

async function sendCode() {
  if (!form.credential) { ElMessage.warning(t('messages.enter_phone')); return }
  captcha.refresh()
  captchaVisible.value = true
}

async function oauthLogin(provider: string) {
  oauthLoading.value = true
  try {
    const res = await oauthAuthorize(provider)
    const d = res.data.data ?? res.data
    if (d.auth_url) {
      window.location.href = d.auth_url
    } else {
      ElMessage.error(d.error || t('messages.oauth_get_url_failed'))
      oauthLoading.value = false
    }
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || t('messages.oauth_auth_failed'))
    oauthLoading.value = false
  }
}

async function submitCaptcha() {
  if (!captcha.validate()) { ElMessage.warning(t('messages.verify_failed')); captcha.refresh(); return }
  captchaVisible.value = false
  codeSending.value = true
  try {
    await sendSMSCode(form.credential)
    const isDev = import.meta.env.DEV
    ElMessage.success(isDev ? t('messages.code_sent_dev') : t('messages.code_sent'))
    let sec = 60; codeBtnText.value = `${sec}s`
    const timer = setInterval(() => { sec--; codeBtnText.value = `${sec}s`; if (sec <= 0) { clearInterval(timer); codeBtnText.value = t('login.resend_code'); codeSending.value = false } }, 1000)
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
</script>
