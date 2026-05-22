import client from './client'

export function getTier(userId: number) {
  return client.get(`/account/${userId}/tier`)
}

export function sendPasswordCode(credential: string) {
  const contactType = credential.includes('@') ? 'email' : 'phone'
  return client.post('/account/password/send-verification-code', {
    contact_type: contactType,
    contact_value: credential,
  })
}

export function changePassword(data: {
  current_password: string
  new_password: string
  verification_code: string
  verification_type: string
}) {
  return client.post('/account/password/change', {
    current_password: data.current_password,
    new_password: data.new_password,
    confirm_password: data.new_password,
    verification_code: data.verification_code,
    verification_type: data.verification_type,
  })
}

export function requestDeletion(data: { verification_code: string; verification_type: string }) {
  return client.post('/account/deletion/request', data)
}

export function cancelDeletion() {
  return client.post('/account/deletion/cancel')
}

export function getDeletionStatus() {
  return client.get('/account/deletion/status')
}

export function getEntitlements(userId: number) {
  return client.get(`/entitlements/${userId}`)
}
