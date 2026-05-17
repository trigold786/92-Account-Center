import { API_BASE } from './constants'
import { getToken, getRefreshToken, setToken, setRefreshToken, clearAuth } from './storage'

interface ApiResponse<T = any> {
  code: number
  message: string
  data?: T
  total?: number
}

function getHeaders(): Record<string, string> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  const token = getToken()
  if (token) headers['Authorization'] = `Bearer ${token}`
  return headers
}

async function refreshToken(): Promise<boolean> {
  const rToken = getRefreshToken()
  if (!rToken) return false
  try {
    const res = await wx.request({
      url: `${API_BASE}/auth/refresh`,
      method: 'POST',
      data: { refresh_token: rToken },
      header: { 'Content-Type': 'application/json' },
    })
    const data = res.data as ApiResponse<{ access_token: string; refresh_token: string }>
    if (data.code === 0 && data.data) {
      setToken(data.data.access_token)
      setRefreshToken(data.data.refresh_token)
      return true
    }
    return false
  } catch {
    return false
  }
}

export async function request<T = any>(
  path: string,
  method: 'GET' | 'POST' | 'PUT' | 'DELETE' = 'GET',
  data?: any,
  retry = true
): Promise<ApiResponse<T>> {
  try {
    const res = await wx.request({
      url: `${API_BASE}${path}`,
      method,
      data,
      header: getHeaders(),
      timeout: 15000,
    })
    const result = res.data as ApiResponse<T>
    if (result.code === 401 && retry) {
      const refreshed = await refreshToken()
      if (refreshed) return request<T>(path, method, data, false)
      clearAuth()
      wx.redirectTo({ url: '/pages/login/login' })
      return result
    }
    return result
  } catch (err: any) {
    return { code: -1, message: err.errMsg || '网络错误' }
  }
}

export const api = {
  get: <T = any>(path: string) => request<T>(path, 'GET'),
  post: <T = any>(path: string, data?: any) => request<T>(path, 'POST', data),
  put: <T = any>(path: string, data?: any) => request<T>(path, 'PUT', data),
  del: <T = any>(path: string) => request<T>(path, 'DELETE'),
}
