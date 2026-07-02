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
    sms_code: params.code,
    account_id: `user_${Date.now()}`,
    agree_terms: true,
  }).then(async (res) => {
    return res
  })
}

export function sendSMSCode(phone: string) {
  return client.post('/sms/send', { phone_number: phone })
}

export function verifySMSCode(phone: string, code: string) {
  return client.post('/sms/verify', { phone_number: phone, code })
}

export function oauthAuthorize(provider: string) {
  return client.get('/auth/oauth/authorize', { params: { provider } })
}

export function oauthCallback(provider: string, code: string, state: string) {
  const params = new URLSearchParams({ provider, code, state })
  return client.post('/auth/oauth/callback?' + params.toString())
}
