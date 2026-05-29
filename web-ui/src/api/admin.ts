import client from './client'

export function getRiskHistory(userId: number) {
  return client.get(`/risk/history/${userId}`)
}

export function getAuditLogs(params: { start_time: string; end_time: string; limit?: number; offset?: number }) {
  return client.get('/audit/logs', { params })
}

export function verifyAuditLog(logId: string) {
  return client.get(`/audit/logs/${logId}/verify`)
}

export function cleanupAuditLogs(beforeDays: number) {
  return client.post('/audit/logs/cleanup', { retention_days: beforeDays })
}

export function listBlacklist(params: { type?: string; limit?: number; offset?: number }) {
  return client.get('/blacklist/', { params })
}

export function addBlacklistEntry(data: { type: string; value: string; reason: string }) {
  return client.post('/blacklist/', {
    entry_type: data.type.toUpperCase(),
    entry_value: data.value,
    reason: data.reason,
  })
}

export function removeBlacklistEntry(type: string, value: string) {
  return client.delete(`/blacklist/${type}/${value}`)
}

export function getSMSProviderStatus() {
  return client.get('/sms/providers/status')
}

export function listUsers(params: { q?: string; page?: number; limit?: number }) {
  return client.get('/admin/users', { params })
}

export function updateUserStatus(userId: number, action: string) {
  return client.put(`/admin/users/${userId}/status`, { action })
}

export function updateUserTier(userId: number, tier: string) {
  return client.put(`/admin/users/${userId}/tier`, { identity_tier: tier })
}
