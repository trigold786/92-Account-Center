import client from './client'

export function getRiskHistory(userId: number) {
  return client.get(`/risk/history/${userId}`)
}

export function getAuditLogs(params: { page?: number; page_size?: number }) {
  return client.get('/audit/logs', { params })
}

export function verifyAuditLog(logId: number) {
  return client.get(`/audit/logs/${logId}/verify`)
}

export function cleanupAuditLogs(beforeDays: number) {
  return client.post('/audit/logs/cleanup', { before_days: beforeDays })
}

export function listBlacklist(params: { type?: string; page?: number; page_size?: number }) {
  return client.get('/blacklist/', { params })
}

export function addBlacklistEntry(data: { type: string; value: string; reason: string }) {
  return client.post('/blacklist/', data)
}

export function removeBlacklistEntry(type: string, value: string) {
  return client.delete(`/blacklist/${type}/${value}`)
}

export function getSMSProviderStatus() {
  return client.get('/sms/providers/status')
}
