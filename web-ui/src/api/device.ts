import client from './client'

export function getUserDevices(userId: number) {
  return client.get(`/device/user/${userId}`)
}

export function registerDevice(data: { device_name: string; device_type: string; device_fingerprint: string }) {
  return client.post('/device/register', data)
}

export function trustDevice(fingerprintId: string) {
  return client.post('/device/trust', { fingerprint_id: fingerprintId })
}

export function removeDevice(deviceId: string) {
  return client.delete(`/device/${deviceId}`)
}
