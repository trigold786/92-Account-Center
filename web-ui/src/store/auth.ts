import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login, logout as apiLogout } from '@/api/auth'
import { usePermissionStore } from './permission'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('access_token') || '')
  const refreshToken = ref(localStorage.getItem('refresh_token') || '')
  const userId = ref(Number(localStorage.getItem('user_id') || 0))
  const accountId = ref(localStorage.getItem('account_id') || '')
  const roles = ref<string[]>(JSON.parse(localStorage.getItem('roles') || '[]'))

  const isLoggedIn = computed(() => !!token.value)

  async function doLogin(credential: string, password: string) {
    const res = await login(credential, password)
    const d = res.data.data ?? res.data
    if (d.access_token) {
      token.value = d.access_token
      refreshToken.value = d.refresh_token
      userId.value = d.user_id
      accountId.value = d.account_id
      roles.value = d.roles || []
      localStorage.setItem('access_token', token.value)
      localStorage.setItem('refresh_token', refreshToken.value)
      localStorage.setItem('user_id', String(userId.value))
      localStorage.setItem('account_id', accountId.value)
      localStorage.setItem('roles', JSON.stringify(roles.value))
      const permStore = usePermissionStore()
      await permStore.loadPermissions(d.account_id || accountId.value)
      return true
    }
    throw new Error(d.message || d.error || '登录失败')
  }

  async function doLogout() {
    try { await apiLogout() } catch {}
    token.value = ''
    refreshToken.value = ''
    userId.value = 0
    accountId.value = ''
    roles.value = []
    localStorage.clear()
  }

  return { token, refreshToken, userId, accountId, roles, isLoggedIn, doLogin, doLogout }
})
