import { ref, computed } from 'vue'
import i18n from '../i18n'

const availableLocales = [
  { value: 'zh-CN', label: '简体中文' },
  { value: 'en-US', label: 'English' },
]

const currentLocale = ref(i18n.global.locale.value)

function getCurrentLocale() {
  return currentLocale.value
}

function setLocale(locale: string) {
  currentLocale.value = locale
  i18n.global.locale.value = locale as any
  localStorage.setItem('locale', locale)
  document.documentElement.setAttribute('lang', locale)
}

export function useLocale() {
  return {
    currentLocale: computed(() => currentLocale.value),
    availableLocales,
    getCurrentLocale,
    setLocale,
  }
}
