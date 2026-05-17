import request from './request'
import type { Role, RolePermission, UserRole, ApiResponse } from '@/types'

export function listRoles(): Promise<ApiResponse<Role[]>> {
  return request.get('/config/roles')
}

export function createRole(data: Partial<Role>): Promise<ApiResponse<Role>> {
  return request.post('/config/roles', data)
}

export function getRolePermissions(roleId: number): Promise<ApiResponse<RolePermission[]>> {
  return request.get(`/config/roles/${roleId}/permissions`)
}

export function addRolePermission(roleId: number, data: Partial<RolePermission>): Promise<ApiResponse<RolePermission>> {
  return request.post(`/config/roles/${roleId}/permissions`, data)
}

export function getUserRoles(userId: string): Promise<ApiResponse<UserRole[]>> {
  return request.get(`/config/users/${userId}/roles`)
}

export function setUserRole(userId: string, data: Partial<UserRole>): Promise<ApiResponse<UserRole>> {
  return request.post(`/config/users/${userId}/roles`, data)
}
