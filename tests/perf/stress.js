import http from 'k6/http'
import { check, sleep } from 'k6'

export const options = {
  stages: [
    { target: 500, duration: '3m' },
    { target: 1000, duration: '5m' },
    { target: 0, duration: '2m' },
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000', 'p(99)<2000'],
    http_req_failed: ['rate<0.05'],
  },
}

const BASE_URL = __ENV.BASE_URL || 'http://localhost:30300'

export default function () {
  const res = http.get(`${BASE_URL}/health`)
  check(res, { 'status is 200': (r) => r.status === 200 })
  sleep(0.1)
}
