import client from './client'

export function listRoles() {
  return client.get('/config/roles')
}

export function createRole(data: { name: string; description: string }) {
  return client.post('/config/roles', data)
}

export function deleteRole(roleId: number) {
  return client.delete(`/config/roles/${roleId}`)
}

export function getRolePermissions(roleId: number) {
  return client.get(`/config/roles/${roleId}/permissions`)
}

export function addRolePermission(roleId: number, permission: string) {
  return client.post(`/config/roles/${roleId}/permissions`, { permission })
}

export function removeRolePermission(roleId: number, permId: number) {
  return client.delete(`/config/roles/${roleId}/permissions/${permId}`)
}

export function getUserRoles(userId: string) {
  return client.get(`/config/users/${userId}/roles`)
}

export function setUserRole(userId: string, roleId: number) {
  return client.post(`/config/users/${userId}/roles`, { role_id: roleId })
}

export function removeUserRole(userId: string, roleId: number) {
  return client.delete(`/config/users/${userId}/roles/${roleId}`)
}
