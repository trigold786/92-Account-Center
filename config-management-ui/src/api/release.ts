import request from './request'
import type { ConfigRelease, ConfigReleaseItem, ApiResponse } from '@/types'

export function listReleases(params: {
  status?: string
  page?: number
  page_size?: number
}): Promise<ApiResponse<ConfigRelease[]> & { total: number }> {
  return request.get('/config/releases', { params })
}

export function getRelease(id: number): Promise<ApiResponse<ConfigRelease>> {
  return request.get(`/config/releases/${id}`)
}

export function createRelease(data: Partial<ConfigRelease>): Promise<ApiResponse<ConfigRelease>> {
  return request.post('/config/releases', data)
}

export function submitRelease(id: number): Promise<ApiResponse<void>> {
  return request.put(`/config/releases/${id}/submit`)
}

export function approveRelease(id: number): Promise<ApiResponse<void>> {
  return request.put(`/config/releases/${id}/approve`)
}

export function rejectRelease(id: number): Promise<ApiResponse<void>> {
  return request.put(`/config/releases/${id}/reject`)
}

export function executeRelease(id: number): Promise<ApiResponse<void>> {
  return request.post(`/config/releases/${id}/execute`)
}

export function listReleaseItems(releaseId: number): Promise<ApiResponse<ConfigReleaseItem[]>> {
  return request.get(`/config/releases/${releaseId}/items`)
}

export function addReleaseItem(releaseId: number, data: Partial<ConfigReleaseItem>): Promise<ApiResponse<ConfigReleaseItem>> {
  return request.post(`/config/releases/${releaseId}/items`, data)
}
