import client from './client'

export function getUserPermissions(userId: string | number) {
  return client.get(`/config/users/${userId}/permissions`)
}
