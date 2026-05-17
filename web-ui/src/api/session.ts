import client from './client'

export function getUserSessions(userId: number) {
  return client.get(`/session/user/${userId}`)
}

export function invalidateSession(sessionId: string) {
  return client.post('/session/invalidate', { session_id: sessionId })
}

export function invalidateAllSessions() {
  return client.post('/session/invalidate-all')
}
