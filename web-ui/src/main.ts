import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import i18n from './i18n'
import './styles/theme.css'
import { initWebVitals } from './utils/webVitals'

const isDev = import.meta.env.DEV
const perf = window.performance
const appStart = perf.now()

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.use(i18n)

router.isReady().then(() => {
  app.mount('#app')
  const mountTime = perf.now() - appStart
  if (isDev) console.log(`[perf] app mount: ${mountTime.toFixed(1)}ms`)

  ;(window as any).__APP_START_TIME = appStart
  ;(window as any).__APP_MOUNT_TIME = perf.now()

  window.addEventListener('load', () => {
    const loadTime = perf.now() - appStart
    if (isDev) console.log(`[perf] page load: ${loadTime.toFixed(1)}ms`)
    try {
      const nav = perf.getEntriesByType('navigation')[0] as PerformanceNavigationTiming | undefined
      if (nav && isDev) {
        console.log(`[perf] TTFB: ${nav.responseStart.toFixed(1)}ms, DOMContentLoaded: ${nav.domContentLoadedEventEnd.toFixed(1)}ms, DOMComplete: ${nav.domComplete.toFixed(1)}ms`)
      }
    } catch {}
    initWebVitals()
    if ('serviceWorker' in navigator) {
      navigator.serviceWorker.register('/sw.js').catch(() => {})
    }
  })
})
