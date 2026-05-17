import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUserStore = defineStore('user', () => {
  const operator = ref(localStorage.getItem('operator') || 'admin')
  const permissions = ref<string[]>(['*'])

  function setOperator(name: string) {
    operator.value = name
    localStorage.setItem('operator', name)
  }

  function hasPermission(permission: string): boolean {
    if (permissions.value.includes('*')) return true
    return permissions.value.includes(permission)
  }

  return { operator, permissions, setOperator, hasPermission }
})
