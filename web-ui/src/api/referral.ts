import client from './client'

export function getReferralSummary(userId: number) {
  return client.get(`/referral/${userId}/summary`)
}

export function generateReferralLink(userId: string) {
  return client.post('/referral/generate-link', { user_id: userId })
}

export function bindReferral(code: string) {
  return client.post('/referral/bind', { referrer_code: code })
}
