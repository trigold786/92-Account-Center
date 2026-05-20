import { ref, onUnmounted } from 'vue'

export function announceToScreenReader(message: string) {
  const el = document.createElement('div')
  el.setAttribute('role', 'status')
  el.setAttribute('aria-live', 'polite')
  el.setAttribute('aria-atomic', 'true')
  Object.assign(el.style, {
    position: 'absolute',
    width: '1px',
    height: '1px',
    padding: '0',
    margin: '-1px',
    overflow: 'hidden',
    clip: 'rect(0, 0, 0, 0)',
    whiteSpace: 'nowrap',
    border: '0',
  })
  document.body.appendChild(el)
  el.textContent = message
  setTimeout(() => {
    document.body.removeChild(el)
  }, 1000)
}

export function checkContrast(fg: string, bg: string): { ratio: number; passesAA: boolean } {
  function parseColor(color: string): [number, number, number] {
    const hex = color.replace('#', '')
    return [
      parseInt(hex.substring(0, 2), 16),
      parseInt(hex.substring(2, 4), 16),
      parseInt(hex.substring(4, 6), 16),
    ]
  }

  function relativeLuminance(r: number, g: number, b: number): number {
    const [rs, gs, bs] = [r, g, b].map((c) => {
      const s = c / 255
      return s <= 0.03928 ? s / 12.92 : Math.pow((s + 0.055) / 1.055, 2.4)
    })
    return 0.2126 * rs + 0.7152 * gs + 0.0722 * bs
  }

  const [fgR, fgG, fgB] = parseColor(fg)
  const [bgR, bgG, bgB] = parseColor(bg)
  const l1 = relativeLuminance(fgR, fgG, fgB)
  const l2 = relativeLuminance(bgR, bgG, bgB)
  const lighter = Math.max(l1, l2)
  const darker = Math.min(l1, l2)
  const ratio = (lighter + 0.05) / (darker + 0.05)

  return {
    ratio: Math.round(ratio * 100) / 100,
    passesAA: ratio >= 4.5,
  }
}

export function trapFocus(element: HTMLElement) {
  const FOCUSABLE =
    'a[href], button:not([disabled]), textarea:not([disabled]), input:not([disabled]), select:not([disabled]), [tabindex]:not([tabindex="-1"])'

  let previouslyFocused: HTMLElement | null = null

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key !== 'Tab') return

    const focusableEls = element.querySelectorAll<HTMLElement>(FOCUSABLE)
    if (focusableEls.length === 0) return

    const firstFocusable = focusableEls[0]
    const lastFocusable = focusableEls[focusableEls.length - 1]

    if (e.shiftKey) {
      if (document.activeElement === firstFocusable) {
        e.preventDefault()
        lastFocusable.focus()
      }
    } else {
      if (document.activeElement === lastFocusable) {
        e.preventDefault()
        firstFocusable.focus()
      }
    }
  }

  previouslyFocused = document.activeElement as HTMLElement

  const focusableEls = element.querySelectorAll<HTMLElement>(FOCUSABLE)
  if (focusableEls.length > 0) {
    focusableEls[0].focus()
  }

  element.addEventListener('keydown', handleKeyDown)

  return () => {
    element.removeEventListener('keydown', handleKeyDown)
    if (previouslyFocused) {
      previouslyFocused.focus()
    }
  }
}

export function prefersReducedMotion(): boolean {
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches
}
