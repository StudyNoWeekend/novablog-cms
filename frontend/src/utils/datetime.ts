import i18n, { DEFAULT_LOCALE, isSupportedLocale, LOCALE_META } from '@/locales'

// 当前语言的 Intl 区域标签，供日期/时间/数字格式化统一使用
function currentLocaleTag(): string {
  const value = i18n.global.locale.value
  return isSupportedLocale(value) ? LOCALE_META[value].htmlLang : DEFAULT_LOCALE
}

function toDate(value: string | number | Date | null | undefined): Date | null {
  if (value === null || value === undefined || value === '') return null
  const date = value instanceof Date ? value : new Date(value)
  return Number.isNaN(date.getTime()) ? null : date
}

// formatDate 日期（不含时间）：zh-CN → 2025/01/02，en → Jan 2, 2025
export function formatDate(value: string | number | Date | null | undefined): string {
  const date = toDate(value)
  if (!date) return '-'
  return new Intl.DateTimeFormat(currentLocaleTag(), {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).format(date)
}

// formatDateTime 日期 + 时间（到分钟）
export function formatDateTime(value: string | number | Date | null | undefined): string {
  const date = toDate(value)
  if (!date) return '-'
  return new Intl.DateTimeFormat(currentLocaleTag(), {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

// formatTime 仅时间（到分钟）
export function formatTime(value: string | number | Date | null | undefined): string {
  const date = toDate(value)
  if (!date) return '-'
  return new Intl.DateTimeFormat(currentLocaleTag(), {
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}
