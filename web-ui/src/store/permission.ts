import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getUserPermissions } from '@/api/permission'
import { useAuthStore } from './auth'

export const usePermissionStore = defineStore('permission', () => {
  const permissions = ref<string[]>([])

  async function loadPermissions(userId: string | number) {
    try {
      const res = await getUserPermissions(userId)
      permissions.value = res.data.data || []
    } catch (e: any) {
      console.warn('[permission] failed to load permissions for', userId, e?.response?.data || e)
      permissions.value = []
    }
  }

  function hasPermission(perm: string): boolean {
    return permissions.value.includes(perm)
  }

  function hasRole(role: string): boolean {
    const auth = useAuthStore()
    return auth.roles.includes(role)
  }

  function hasAnyRole(roles: string[]): boolean {
    const auth = useAuthStore()
    return roles.some(r => auth.roles.includes(r))
  }

  return { permissions, loadPermissions, hasPermission, hasRole, hasAnyRole }
})
