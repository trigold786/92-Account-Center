import { ref, computed } from 'vue'
import i18n from '../i18n'

const availableLocales = [
  { value: 'zh-CN', label: '简体中文' },
  { value: 'en-US', label: 'English' },
]

type SupportedLocale = 'zh-CN' | 'en-US'
const currentLocale = ref<SupportedLocale>((i18n.global.locale.value as SupportedLocale) || 'zh-CN')

function getCurrentLocale() {
  return currentLocale.value
}

function setLocale(locale: string) {
  const next = (locale === 'en-US' ? 'en-US' : 'zh-CN') as SupportedLocale
  currentLocale.value = next
  i18n.global.locale.value = next
  localStorage.setItem('locale', next)
  document.documentElement.setAttribute('lang', next)
}

export function useLocale() {
  return {
    currentLocale: computed(() => currentLocale.value),
    availableLocales,
    getCurrentLocale,
    setLocale,
  }
}
