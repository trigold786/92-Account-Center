import { api } from './api'
import { setToken, setRefreshToken, setUserInfo, setUserId, clearAuth } from './storage'

interface LoginParams {
  credential: string
  password?: string
  code?: string
}

interface LoginResponse {
  access_token: string
  refresh_token: string
  expires_in: number
  user_id: number
  account_id: string
}

export async function loginWithPassword(credential: string, password: string): Promise<boolean> {
  const res = await api.post<LoginResponse>('/auth/login', { credential, password })
  if (res.code === 0 && res.data) {
    setToken(res.data.access_token)
    setRefreshToken(res.data.refresh_token)
    setUserId(res.data.user_id)
    return true
  }
  throw new Error(res.message || '登录失败')
}

export async function loginWithCode(phone: string, code: string): Promise<boolean> {
  const res = await api.post<LoginResponse>('/auth/login', { credential: phone, code })
  if (res.code === 0 && res.data) {
    setToken(res.data.access_token)
    setRefreshToken(res.data.refresh_token)
    setUserId(res.data.user_id)
    return true
  }
  throw new Error(res.message || '登录失败')
}

export async function sendSMSCode(phone: string): Promise<void> {
  const res = await api.post('/sms/send', { phone, scene: 'login' })
  if (res.code !== 0) throw new Error(res.message || '发送失败')
}

export async function register(params: { phone: string; password: string; code: string }): Promise<void> {
  const res = await api.post('/account/register', params)
  if (res.code !== 0) throw new Error(res.message || '注册失败')
}

export async function logout(): Promise<void> {
  await api.post('/auth/logout')
  clearAuth()
}

export function sendVerificationCode(phone: string, scene: 'register' | 'password_change') {
  return api.post('/sms/send', { phone, scene })
}
