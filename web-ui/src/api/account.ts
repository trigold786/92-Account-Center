import client from './client'

export function getTier(userId: number) {
  return client.get(`/account/${userId}/tier`)
}

export function changePassword(data: { current_password: string; new_password: string; code: string }) {
  return client.post('/account/password/change', data)
}

export function sendPasswordCode(credential: string) {
  return client.post('/account/password/send-verification-code', { credential })
}

export function requestDeletion() {
  return client.post('/account/deletion/request')
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
