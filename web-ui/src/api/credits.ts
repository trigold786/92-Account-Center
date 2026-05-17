import client from './client'

export function getCreditAccount(userId: number) {
  return client.get(`/credits/${userId}/account`)
}

export function getTransactions(userId: number) {
  return client.get(`/credits/${userId}/transactions`)
}

export function calculateDiscount(data: { amount: number; userId: number }) {
  return client.post('/credits/calculate-discount', data)
}
