import axios from 'axios'

const client = axios.create({ baseURL: '/api/v1', timeout: 15000 })

client.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

client.interceptors.response.use(
  (res) => res,
  async (error) => {
    if (error.response?.status === 401) {
      const refreshToken = localStorage.getItem('refresh_token')
      if (refreshToken) {
        try {
          const res = await axios.post('/api/v1/auth/refresh', { refresh_token: refreshToken })
          const d = res.data.data ?? res.data
          if (d.access_token) {
            localStorage.setItem('access_token', d.access_token)
            localStorage.setItem('refresh_token', d.refresh_token)
            error.config.headers.Authorization = `Bearer ${d.access_token}`
            return axios(error.config)
          }
        } catch {}
      }
      localStorage.clear()
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default client
