import client from './client'

export function getReferralSummary(userId: number) {
  return client.get(`/referral/${userId}/summary`)
}

export function generateReferralLink() {
  return client.post('/referral/generate-link')
}

export function bindReferral(code: string) {
  return client.post('/referral/bind', { code })
}
