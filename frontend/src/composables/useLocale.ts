import { ref, computed } from 'vue'
import en from '@/locales/en'
import ru from '@/locales/ru'

type Locale = 'en' | 'ru'
type Messages = typeof en

const locales: Record<Locale, Messages> = { en, ru }

const stored = localStorage.getItem('simarena-locale') as Locale | null
const currentLocale = ref<Locale>(stored === 'ru' ? 'ru' : 'en')

export function useLocale() {
  const t = computed<Messages>(() => locales[currentLocale.value])

  function setLocale(locale: Locale) {
    currentLocale.value = locale
    localStorage.setItem('simarena-locale', locale)
  }

  return {
    locale: currentLocale,
    t,
    setLocale,
  }
}
