import client from './client'

export function getSubscriptions(userId: number) {
  return client.get(`/subscriptions/${userId}`)
}

export function purchaseSubscription(data: { user_id: number; plan_id: string }) {
  return client.post('/subscriptions/purchase', data)
}

export function upgradeSubscription(data: { user_id: number; new_plan_id: string }) {
  return client.post('/subscriptions/upgrade', data)
}

export function renewSubscription(data: { user_id: number; subscription_id: number }) {
  return client.post('/subscriptions/renew', data)
}
