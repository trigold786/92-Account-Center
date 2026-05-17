export function getToken(): string {
  return wx.getStorageSync('access_token') || ''
}

export function setToken(token: string): void {
  wx.setStorageSync('access_token', token)
}

export function getRefreshToken(): string {
  return wx.getStorageSync('refresh_token') || ''
}

export function setRefreshToken(token: string): void {
  wx.setStorageSync('refresh_token', token)
}

export function setUserInfo(info: any): void {
  wx.setStorageSync('user_info', info)
}

export function getUserInfo(): any {
  return wx.getStorageSync('user_info') || null
}

export function setUserId(id: number): void {
  wx.setStorageSync('user_id', id)
}

export function getUserId(): number {
  return wx.getStorageSync('user_id') || 0
}

export function clearAuth(): void {
  wx.removeStorageSync('access_token')
  wx.removeStorageSync('refresh_token')
  wx.removeStorageSync('user_info')
  wx.removeStorageSync('user_id')
  wx.removeStorageSync('account_id')
}

export function isLoggedIn(): boolean {
  return !!getToken()
}
