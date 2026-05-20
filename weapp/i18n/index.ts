import zhCN from './zh-CN'
import enUS from './en-US'

type LocaleKey = string
type NestedMessages = { [key: string]: string | NestedMessages }

const messages: Record<string, NestedMessages> = {
  'zh-CN': zhCN,
  'en-US': enUS,
}

const STORAGE_KEY = 'app_locale'

function getSystemLocale(): string {
  try {
    const sysInfo = wx.getSystemInfoSync()
    const lang = sysInfo.language || 'zh_CN'
    return lang.replace('_', '-') === 'zh-CN' ? 'zh-CN' : 'en-US'
  } catch {
    return 'zh-CN'
  }
}

function getStoredLocale(): string | null {
  try {
    return wx.getStorageSync(STORAGE_KEY) || null
  } catch {
    return null
  }
}

let currentLocale: string = getStoredLocale() || getSystemLocale()

function getLocale(): string {
  return currentLocale
}

function setLocale(locale: string): void {
  currentLocale = locale
  try {
    wx.setStorageSync(STORAGE_KEY, locale)
  } catch {}
}

function resolve(obj: NestedMessages, path: string): string {
  const keys = path.split('.')
  let current: NestedMessages | string = obj
  for (const key of keys) {
    if (typeof current === 'string') return path
    current = (current as NestedMessages)[key]
    if (current === undefined) return path
  }
  return typeof current === 'string' ? current : path
}

function t(key: string): string {
  const msg = messages[currentLocale]
  if (!msg) return key
  return resolve(msg, key)
}

export { getLocale, setLocale, t }
