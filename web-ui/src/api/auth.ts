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
  return client.post('/account/register', {
    phone_number: params.phone,
    password: params.password,
    code: params.code,
    account_id: `user_${Date.now()}`,
    agree_to_terms: true,
  })
}

export function sendSMSCode(phone: string) {
  return client.post('/sms/send', { phone_number: phone })
}

export function verifySMSCode(phone: string, code: string) {
  return client.post('/sms/verify', { phone_number: phone, code })
}
