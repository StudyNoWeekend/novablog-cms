import { createI18n } from 'vue-i18n'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import 'dayjs/locale/zh-tw'
import 'dayjs/locale/ja'
import 'dayjs/locale/ko'

import { localeStorage } from '@/utils/storage'

export const SUPPORTED_LOCALES = ['zh-CN', 'zh-TW', 'en', 'ja', 'ko'] as const
export type AppLocale = (typeof SUPPORTED_LOCALES)[number]
export const DEFAULT_LOCALE: AppLocale = 'zh-CN'

// 各语言的展示元信息：name 为该语言的原生写法（语言切换器内保持原生名，不做翻译）
export const LOCALE_META: Record<AppLocale, { name: string; htmlLang: string; dayjsLocale: string }> = {
  'zh-CN': { name: '简体中文', htmlLang: 'zh-CN', dayjsLocale: 'zh-cn' },
  'zh-TW': { name: '繁體中文', htmlLang: 'zh-TW', dayjsLocale: 'zh-tw' },
  en: { name: 'English', htmlLang: 'en', dayjsLocale: 'en' },
  ja: { name: '日本語', htmlLang: 'ja', dayjsLocale: 'ja' },
  ko: { name: '한국어', htmlLang: 'ko', dayjsLocale: 'ko' },
}

// 语言包自动装配：<语言目录>/<命名空间>.json → messages[语言][命名空间]。
// 新增命名空间时只需在各语言目录新建同名 JSON，无需修改本文件。
const localeFiles = import.meta.glob<{ default: Record<string, unknown> }>('./*/*.json', { eager: true })

const messages = Object.fromEntries(
  SUPPORTED_LOCALES.map((locale) => [
    locale,
    Object.fromEntries(
      Object.entries(localeFiles)
        .filter(([path]) => path.startsWith(`./${locale}/`))
        .map(([path, mod]) => [path.slice(`./${locale}/`.length, -'.json'.length), mod.default]),
    ),
  ]),
)

export function isSupportedLocale(value: unknown): value is AppLocale {
  return typeof value === 'string' && (SUPPORTED_LOCALES as readonly string[]).includes(value)
}

// resolveInitialLocale 解析初始语言：本地偏好 → 浏览器语言（精确匹配 → 前缀匹配）→ 简体中文
export function resolveInitialLocale(): AppLocale {
  const stored = localeStorage.get()
  if (isSupportedLocale(stored)) return stored

  const candidates = typeof navigator !== 'undefined' ? [...(navigator.languages ?? []), navigator.language] : []
  for (const candidate of candidates) {
    if (!candidate) continue
    const exact = SUPPORTED_LOCALES.find((locale) => locale.toLowerCase() === candidate.toLowerCase())
    if (exact) return exact

    const prefix = candidate.toLowerCase().split('-')[0]
    if (prefix === 'zh') {
      // 繁体地区（台湾/香港/澳门）优先映射到繁体中文
      return /-(hant|tw|hk|mo)/i.test(candidate) ? 'zh-TW' : 'zh-CN'
    }
    const partial = SUPPORTED_LOCALES.find((locale) => locale.toLowerCase().split('-')[0] === prefix)
    if (partial) return partial
  }
  return DEFAULT_LOCALE
}

export function applyDocumentLang(locale: AppLocale): void {
  document.documentElement.lang = LOCALE_META[locale].htmlLang
}

export function applyDayjsLocale(locale: AppLocale): void {
  dayjs.locale(LOCALE_META[locale].dayjsLocale)
}

const i18n = createI18n({
  legacy: false,
  globalInjection: true,
  locale: resolveInitialLocale(),
  fallbackLocale: DEFAULT_LOCALE,
  messages,
  missingWarn: import.meta.env.DEV,
  fallbackWarn: import.meta.env.DEV,
})

export default i18n
