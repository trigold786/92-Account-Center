import { ref, computed, onMounted, onUnmounted } from 'vue'

type Breakpoint = 'mobile' | 'tablet' | 'desktop'

const BREAKPOINTS = {
  mobile: 0,
  tablet: 768,
  desktop: 1024,
} as const

export function useBreakpoint() {
  const width = ref(typeof window !== 'undefined' ? window.innerWidth : 1024)

  const current = computed<Breakpoint>(() => {
    if (width.value < BREAKPOINTS.tablet) return 'mobile'
    if (width.value < BREAKPOINTS.desktop) return 'tablet'
    return 'desktop'
  })

  const isMobile = computed(() => current.value === 'mobile')
  const isTablet = computed(() => current.value === 'tablet')
  const isDesktop = computed(() => current.value === 'desktop')
  const isMobileOrTablet = computed(() => current.value === 'mobile' || current.value === 'tablet')

  let rafId: number | null = null

  function onResize() {
    if (rafId !== null) return
    rafId = requestAnimationFrame(() => {
      width.value = window.innerWidth
      rafId = null
    })
  }

  onMounted(() => {
    window.addEventListener('resize', onResize, { passive: true })
  })

  onUnmounted(() => {
    window.removeEventListener('resize', onResize)
    if (rafId !== null) {
      cancelAnimationFrame(rafId)
      rafId = null
    }
  })

  return {
    width,
    current,
    isMobile,
    isTablet,
    isDesktop,
    isMobileOrTablet,
  }
}
