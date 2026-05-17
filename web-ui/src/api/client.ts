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
          if (res.data.code === 0 && res.data.data) {
            localStorage.setItem('access_token', res.data.data.access_token)
            localStorage.setItem('refresh_token', res.data.data.refresh_token)
            error.config.headers.Authorization = `Bearer ${res.data.data.access_token}`
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
