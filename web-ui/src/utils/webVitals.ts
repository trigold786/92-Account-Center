type Metric = {
  name: string
  value: number
  rating: 'good' | 'needs-improvement' | 'poor'
  delta: number
}

function getRating(name: string, value: number): 'good' | 'needs-improvement' | 'poor' {
  const thresholds: Record<string, [number, number]> = {
    LCP: [2500, 4000],
    FID: [100, 300],
    CLS: [0.1, 0.25],
    FCP: [1800, 3000],
    TTFB: [800, 1800],
    INP: [200, 500],
  }
  const t = thresholds[name]
  if (!t) return 'good'
  if (value <= t[0]) return 'good'
  if (value <= t[1]) return 'needs-improvement'
  return 'poor'
}

function reportMetric(metric: Metric) {
  if (import.meta.env.DEV) console.log(`[web-vitals] ${metric.name}: ${metric.value.toFixed(1)}ms (${metric.rating})`)
  if (typeof window !== 'undefined' && 'navigator' in window && navigator.sendBeacon) {
    try {
      navigator.sendBeacon('/api/v1/metrics/web-vitals', JSON.stringify({
        name: metric.name,
        value: metric.value,
        rating: metric.rating,
        delta: metric.delta,
        url: window.location.pathname,
        ua: navigator.userAgent,
        ts: Date.now(),
      }))
    } catch {}
  }
}

function observeMetric(entryName: string, callback: (entry: PerformanceEntry) => void) {
  try {
    const observer = new PerformanceObserver((list) => {
      for (const entry of list.getEntries()) {
        callback(entry)
      }
    })
    observer.observe({ type: entryName, buffered: true })
  } catch {}
}

export function initWebVitals() {
  observeMetric('largest-contentful-paint', (entry) => {
    const lcp = entry.startTime
    reportMetric({ name: 'LCP', value: lcp, rating: getRating('LCP', lcp), delta: lcp })
  })

  observeMetric('first-input', (entry) => {
    const fid = (entry as PerformanceEventTiming).processingStart - entry.startTime
    reportMetric({ name: 'FID', value: fid, rating: getRating('FID', fid), delta: fid })
  })

  observeMetric('layout-shift', (entry) => {
    const cls = (entry as any).value || 0
    reportMetric({ name: 'CLS', value: cls, rating: getRating('CLS', cls), delta: cls })
  })

  observeMetric('paint', (entry) => {
    if (entry.name === 'first-contentful-paint') {
      reportMetric({ name: 'FCP', value: entry.startTime, rating: getRating('FCP', entry.startTime), delta: entry.startTime })
    }
  })

  observeMetric('resource', (entry) => {
    if (entry.name.includes('/api/')) {
      const duration = entry.duration
      if (duration > 1000 && import.meta.env.DEV) {
        console.log(`[perf] slow API: ${entry.name} ${duration.toFixed(0)}ms`)
      }
    }
  })
}
