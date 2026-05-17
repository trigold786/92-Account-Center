import client from './client'

export function login(credential: string, password: string) {
  return client.post('/auth/login', { credential, password })
}

export function loginWithCode(credential: string, code: string) {
  return client.post('/auth/login', { credential, code })
}

export function refreshToken(refresh_token: string) {
  return client.post('/auth/refresh', { refresh_token })
}

export function logout() {
  return client.post('/auth/logout')
}

export function register(params: { phone: string; password: string; code: string }) {
  return client.post('/account/register', params)
}

export function sendSMSCode(phone: string, scene: string = 'login') {
  return client.post('/sms/send', { phone, scene })
}

export function verifySMSCode(phone: string, code: string) {
  return client.post('/sms/verify', { phone, code })
}
