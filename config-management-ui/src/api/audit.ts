import request from './request'
import type { AuditLog, ApiResponse } from '@/types'

export function listAuditLogs(params: {
  operation_type?: string
  operator?: string
  start_time?: number
  end_time?: number
  page?: number
  page_size?: number
}): Promise<ApiResponse<AuditLog[]> & { total: number }> {
  return request.get('/config/audit-logs', { params })
}

export function getAuditLog(id: number): Promise<ApiResponse<AuditLog>> {
  return request.get(`/config/audit-logs/${id}`)
}
