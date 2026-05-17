import client from './client'

export function getRFMScore(userId: number) {
  return client.get(`/data/rfm/${userId}`)
}

export function getDashboardOverview() {
  return client.get('/data/dashboard/overview')
}

export function getSubscriptionFunnel() {
  return client.get('/data/funnel/subscription')
}
