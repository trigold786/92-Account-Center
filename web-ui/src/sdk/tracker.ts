interface TrackEvent {
  event_type: string
  properties?: Record<string, unknown>
  timestamp?: number
}

class Tracker {
  private apiUrl: string
  private sessionId: string
  private queue: TrackEvent[] = []
  private maxBatchSize = 10
  private flushInterval = 5000
  private timer: ReturnType<typeof setInterval> | null = null
  private userId: number | null = null

  constructor(apiUrl: string) {
    this.apiUrl = apiUrl
    this.sessionId = this.generateId()
    this.restoreQueue()
    this.startAutoFlush()
    this.autoCapture()
  }

  private generateId(): string {
    return Math.random().toString(36).substring(2, 15)
  }

  setUserId(id: number | null) {
    this.userId = id
  }

  track(eventType: string, properties?: Record<string, unknown>) {
    if (!eventType) return
    const event: TrackEvent = {
      event_type: eventType,
      properties: properties ?? {},
      timestamp: Date.now(),
    }
    this.queue.push(event)
    if (this.queue.length >= this.maxBatchSize) {
      this.flush()
    }
  }

  private async flush() {
    if (this.queue.length === 0) return
    const batch = this.queue.splice(0, this.maxBatchSize)
    try {
      await fetch(`${this.apiUrl}/api/v1/events/batch`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(batch),
        keepalive: true,
      })
    } catch {
      this.queue.unshift(...batch)
      this.saveQueue()
    }
  }

  private startAutoFlush() {
    this.timer = setInterval(() => this.flush(), this.flushInterval)
    window.addEventListener('beforeunload', () => {
      this.flush()
      if (this.timer) clearInterval(this.timer)
    })
  }

  private autoCapture() {
    let lastUrl = location.href
    new MutationObserver(() => {
      const url = location.href
      if (url !== lastUrl) {
        lastUrl = url
        this.track('page_view', { url })
      }
    }).observe(document, { subtree: true, childList: true })

    document.addEventListener('click', (e) => {
      const target = e.target as HTMLElement
      const trackAttr = target.getAttribute('data-track')
      if (trackAttr) {
        this.track('click', { element: trackAttr, text: target.textContent?.trim() })
      }
    })
  }

  private saveQueue() {
    try {
      localStorage.setItem('tracker_queue', JSON.stringify(this.queue))
    } catch {}
  }

  private restoreQueue() {
    try {
      const saved = localStorage.getItem('tracker_queue')
      if (saved) {
        this.queue = JSON.parse(saved)
        localStorage.removeItem('tracker_queue')
      }
    } catch {}
  }
}

export default Tracker
