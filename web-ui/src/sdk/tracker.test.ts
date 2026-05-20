import Tracker from './tracker'

describe('Tracker', () => {
  let tracker: Tracker

  beforeEach(() => {
    tracker = new Tracker('http://localhost:30300')
  })

  test('should track event type', () => {
    tracker.track('page_view', { url: '/home' })
    expect((tracker as any).queue.length).toBe(1)
  })

  test('should not track empty event type', () => {
    tracker.track('')
    expect((tracker as any).queue.length).toBe(0)
  })

  test('should flush when queue reaches max size', () => {
    for (let i = 0; i < 10; i++) {
      tracker.track('page_view', { url: `/page/${i}` })
    }
    expect((tracker as any).queue.length).toBeLessThanOrEqual(1)
  })
})
