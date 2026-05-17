import request from './request'
import type { ConfigGroup, ConfigItem, ConfigVersion, ApiResponse } from '@/types'

export function listGroups(): Promise<ApiResponse<ConfigGroup[]>> {
  return request.get('/config/groups')
}

export function getGroup(id: number): Promise<ApiResponse<ConfigGroup>> {
  return request.get(`/config/groups/${id}`)
}

export function createGroup(data: Partial<ConfigGroup>): Promise<ApiResponse<ConfigGroup>> {
  return request.post('/config/groups', data)
}

export function updateGroup(id: number, data: Partial<ConfigGroup>): Promise<ApiResponse<ConfigGroup>> {
  return request.put(`/config/groups/${id}`, data)
}

export function deleteGroup(id: number): Promise<ApiResponse<void>> {
  return request.delete(`/config/groups/${id}`)
}

export function listItems(params: {
  group_id?: number
  code?: string
  name?: string
  data_type?: string
  page?: number
  page_size?: number
}): Promise<ApiResponse<ConfigItem[]> & { total: number }> {
  return request.get('/config/items', { params })
}

export function getItem(id: number): Promise<ApiResponse<ConfigItem>> {
  return request.get(`/config/items/${id}`)
}

export function createItem(data: Partial<ConfigItem>): Promise<ApiResponse<ConfigItem>> {
  return request.post('/config/items', data)
}

export function updateItem(id: number, data: Partial<ConfigItem>, changeReason?: string): Promise<ApiResponse<ConfigItem>> {
  return request.put(`/config/items/${id}`, data, { params: { change_reason: changeReason } })
}

export function deleteItem(id: number): Promise<ApiResponse<void>> {
  return request.delete(`/config/items/${id}`)
}

export function resetItemToDefault(id: number): Promise<ApiResponse<void>> {
  return request.post(`/config/items/${id}/reset-default`)
}

export function listVersions(itemId: number): Promise<ApiResponse<ConfigVersion[]>> {
  return request.get(`/config/items/${itemId}/versions`)
}
