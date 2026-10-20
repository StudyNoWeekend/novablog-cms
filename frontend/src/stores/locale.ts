import { computed } from 'vue'
import { defineStore } from 'pinia'
import i18n, {
  DEFAULT_LOCALE,
  isSupportedLocale,
  resolveInitialLocale,
  applyDocumentLang,
  applyDayjsLocale,
} from '@/locales'
import type { AppLocale } from '@/locales'
import { localeStorage } from '@/utils/storage'

// 语言状态中枢：i18n.global 为唯一数据源，store 负责切换副作用（html lang / dayjs / 持久化）
export const useLocaleStore = defineStore('locale', () => {
  const locale = computed<AppLocale>({
    get: () => {
      const current = i18n.global.locale.value
      return isSupportedLocale(current) ? current : DEFAULT_LOCALE
    },
    set: (next) => setLocale(next),
  })

  function setLocale(next: AppLocale) {
    i18n.global.locale.value = next
    applyDocumentLang(next)
    applyDayjsLocale(next)
    localeStorage.set(next)
  }

  setLocale(resolveInitialLocale())

  return { locale, setLocale }
})
