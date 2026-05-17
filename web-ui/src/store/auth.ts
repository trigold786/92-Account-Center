import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login, logout as apiLogout } from '@/api/auth'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('access_token') || '')
  const refreshToken = ref(localStorage.getItem('refresh_token') || '')
  const userId = ref(Number(localStorage.getItem('user_id') || 0))
  const accountId = ref(localStorage.getItem('account_id') || '')

  const isLoggedIn = computed(() => !!token.value)

  async function doLogin(credential: string, password: string) {
    const res = await login(credential, password)
    if (res.data.code === 0 && res.data.data) {
      token.value = res.data.data.access_token
      refreshToken.value = res.data.data.refresh_token
      userId.value = res.data.data.user_id
      accountId.value = res.data.data.account_id
      localStorage.setItem('access_token', token.value)
      localStorage.setItem('refresh_token', refreshToken.value)
      localStorage.setItem('user_id', String(userId.value))
      localStorage.setItem('account_id', accountId.value)
      return true
    }
    throw new Error(res.data.message || '登录失败')
  }

  async function doLogout() {
    try { await apiLogout() } catch {}
    token.value = ''
    refreshToken.value = ''
    userId.value = 0
    accountId.value = ''
    localStorage.clear()
  }

  return { token, refreshToken, userId, accountId, isLoggedIn, doLogin, doLogout }
})
