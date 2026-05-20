import http from 'k6/http'
import { check, sleep } from 'k6'

export const options = {
  vus: 10,
  duration: '30s',
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
    http_req_failed: ['rate<0.01'],
  },
}

const BASE_URL = __ENV.BASE_URL || 'http://localhost:30300'

export default function () {
  const res = http.get(`${BASE_URL}/health`)
  check(res, {
    'health check status is 200': (r) => r.status === 200,
  })
  sleep(1)
}
