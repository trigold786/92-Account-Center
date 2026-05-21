import { ref, onMounted, onUnmounted } from 'vue'

interface CacheEntry {
  blob: Blob
  timestamp: number
  size: number
}

const MAX_DISK_SIZE = 50 * 1024 * 1024
const CACHE_NAME = 'neuro-img-cache-v1'

class LRUCache<K, V> {
  private cache = new Map<K, V>()
  private readonly maxSize: number

  constructor(maxSize: number = 100) {
    this.maxSize = maxSize
  }

  get(key: K): V | undefined {
    const value = this.cache.get(key)
    if (value !== undefined) {
      this.cache.delete(key)
      this.cache.set(key, value)
    }
    return value
  }

  set(key: K, value: V): void {
    if (this.cache.has(key)) {
      this.cache.delete(key)
    }
    if (this.cache.size >= this.maxSize) {
      const firstKey = this.cache.keys().next().value
      if (firstKey !== undefined) this.cache.delete(firstKey)
    }
    this.cache.set(key, value)
  }

  has(key: K): boolean {
    return this.cache.has(key)
  }

  clear(): void {
    this.cache.clear()
  }
}

const memoryCache = new LRUCache<string, string>(50)

let supportsWebP: boolean | null = null
let supportsAVIF: boolean | null = null

function detectFormats(): Promise<{ webp: boolean; avif: boolean }> {
  return new Promise((resolve) => {
    const webpImg = new Image()
    webpImg.onload = () => {
      supportsWebP = webpImg.width > 0
      checkAvif()
    }
    webpImg.onerror = () => {
      supportsWebP = false
      checkAvif()
    }
    webpImg.src = 'data:image/webp;base64,UklGRiQAAABXRUJQVlA4IBgAAAAwAQCdASoBAAEAAwA0JaQAA3AA/vuUAAA='

    function checkAvif() {
      const avifImg = new Image()
      avifImg.onload = () => {
        supportsAVIF = avifImg.width > 0
        resolve({ webp: supportsWebP!, avif: supportsAVIF! })
      }
      avifImg.onerror = () => {
        supportsAVIF = false
        resolve({ webp: supportsWebP!, avif: false })
      }
      avifImg.src = 'data:image/avif;base64,AAAAIGZ0eXBhdmlmAAAAAGF2aWZtaWYxbWlhZk1BMUIAAADybWV0YQAAAAAAAAAoaGRscgAAAAAAAAAAcGljdAAAAAAAAAAAAAAAAGxpYmF2aWYAAAAIWnSlpAC6h5j5v8aSKJCAB8JYc0Yhq4RZmQgA='
    }
  })
}

function getOptimizedUrl(originalUrl: string): string {
  if (supportsAVIF) {
    return originalUrl.replace(/\.(jpg|jpeg|png|webp)(\?.*)?$/i, '.avif$2')
  }
  if (supportsWebP) {
    return originalUrl.replace(/\.(jpg|jpeg|png)(\?.*)?$/i, '.webp$2')
  }
  return originalUrl
}

async function getCachedImage(url: string): Promise<string | null> {
  if (memoryCache.has(url)) {
    return memoryCache.get(url)!
  }

  try {
    const cache = await caches.open(CACHE_NAME)
    const response = await cache.match(url)
    if (response) {
      const blob = await response.blob()
      const objectUrl = URL.createObjectURL(blob)
      memoryCache.set(url, objectUrl)
      return objectUrl
    }
  } catch {}

  return null
}

async function cacheImage(url: string, blob: Blob): Promise<void> {
  try {
    const cache = await caches.open(CACHE_NAME)
    await cache.put(url, new Response(blob))
    memoryCache.set(url, URL.createObjectURL(blob))
    await pruneDiskCache()
  } catch {}
}

async function pruneDiskCache(): Promise<void> {
  try {
    const cache = await caches.open(CACHE_NAME)
    const keys = await cache.keys()
    if (keys.length === 0) return

    let totalSize = 0
    const entries: { url: string; response: Response; size: number }[] = []

    for (const request of keys) {
      const response = await cache.match(request)
      if (response) {
        const blob = await response.blob()
        totalSize += blob.size
        entries.push({ url: request.url, response, size: blob.size })
      }
    }

    if (totalSize > MAX_DISK_SIZE) {
      entries.sort((a, b) => 0 - 0)
      for (const entry of entries) {
        if (totalSize <= MAX_DISK_SIZE) break
        await cache.delete(entry.url)
        totalSize -= entry.size
      }
    }
  } catch {}
}

function useLazyLoad(elementSelector: string, threshold: number = 0.1) {
  const isVisible = ref(false)
  let observer: IntersectionObserver | null = null

  onMounted(() => {
    const el = document.querySelector(elementSelector)
    if (!el) return

    observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            isVisible.value = true
            observer?.disconnect()
          }
        })
      },
      { threshold }
    )
    observer.observe(el)
  })

  onUnmounted(() => {
    observer?.disconnect()
  })

  return { isVisible }
}

export function useImageOptimization() {
  const isLoading = ref(true)
  const hasError = ref(false)
  const imageSrc = ref<string>('')
  const placeholderSrc = ref<string>('')

  async function loadImage(originalUrl: string) {
    isLoading.value = true
    hasError.value = false

    if (supportsWebP === null) {
      await detectFormats()
    }

    const cached = await getCachedImage(originalUrl)
    if (cached) {
      imageSrc.value = cached
      isLoading.value = false
      return
    }

    const optimizedUrl = getOptimizedUrl(originalUrl)

    try {
      const response = await fetch(optimizedUrl)
      if (!response.ok) throw new Error('Failed to load')
      const blob = await response.blob()

      await cacheImage(originalUrl, blob)
      imageSrc.value = URL.createObjectURL(blob)
    } catch {
      try {
        const response = await fetch(originalUrl)
        if (!response.ok) throw new Error('Failed to load')
        const blob = await response.blob()
        await cacheImage(originalUrl, blob)
        imageSrc.value = URL.createObjectURL(blob)
      } catch {
        hasError.value = true
      }
    } finally {
      isLoading.value = false
    }
  }

  function generateBlurPlaceholder(width: number = 10, height: number = 10): string {
    const canvas = document.createElement('canvas')
    canvas.width = width
    canvas.height = height
    const ctx = canvas.getContext('2d')!
    ctx.fillStyle = '#1C2333'
    ctx.fillRect(0, 0, width, height)
    return canvas.toDataURL()
  }

  return {
    isLoading,
    hasError,
    imageSrc,
    placeholderSrc,
    loadImage,
    generateBlurPlaceholder,
    getOptimizedUrl,
  }
}

export { useLazyLoad, detectFormats, getOptimizedUrl, getCachedImage, cacheImage }
