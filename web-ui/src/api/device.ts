import client from './client'

export function getUserDevices(userId: number) {
  return client.get(`/device/user/${userId}`)
}

export function registerDevice(data: { device_name: string; device_type: string; device_fingerprint: string }) {
  return client.post('/device/register', data)
}

export function trustDevice(deviceId: string) {
  return client.post('/device/trust', { device_id: deviceId })
}

export function removeDevice(deviceId: string) {
  return client.delete(`/device/${deviceId}`)
}
