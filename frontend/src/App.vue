<template>
  <a-config-provider :locale="antdLocale" :theme="antTheme">
    <router-view />
  </a-config-provider>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { theme as antdTheme } from 'ant-design-vue'
import zhCN from 'ant-design-vue/es/locale/zh_CN'
import zhTW from 'ant-design-vue/es/locale/zh_TW'
import enUS from 'ant-design-vue/es/locale/en_US'
import jaJP from 'ant-design-vue/es/locale/ja_JP'
import koKR from 'ant-design-vue/es/locale/ko_KR'
import i18n from '@/locales'
import { useLocaleStore } from '@/stores/locale'

const localeStore = useLocaleStore()
const route = useRoute()

// antd 内置文案（分页/日期选择/表格空态等）跟随当前语言
const ANTD_LOCALES = { 'zh-CN': zhCN, 'zh-TW': zhTW, en: enUS, ja: jaJP, ko: koKR } as const
const antdLocale = computed(() => ANTD_LOCALES[localeStore.locale])

// 页面标题跟随当前语言与路由 meta.title（值为 menu.* 的 i18n key）
const APP_TITLE = import.meta.env.VITE_APP_TITLE || 'NovaBlog CMS'
watch(
  [() => localeStore.locale, () => route.meta.title],
  () => {
    const key = route.meta.title as string | undefined
    document.title = key ? `${i18n.global.t(key)} · ${APP_TITLE}` : APP_TITLE
  },
  { immediate: true },
)

// AntD 全局设计令牌：与 variables.css / uno.config.ts 保持同一套色板
const antTheme = {
  algorithm: antdTheme.defaultAlgorithm,
  token: {
    colorPrimary: '#526FE8',
    colorSuccess: '#16805D',
    colorWarning: '#B7791F',
    colorError: '#C2414B',
    colorInfo: '#3569C8',
    colorTextBase: '#29365C',
    colorBgLayout: '#F5F7FC',
    colorBorder: '#E5E9F2',
    colorBorderSecondary: '#EEF1F7',
    colorTextSecondary: '#667085',
    colorTextTertiary: '#8A93A8',
    borderRadius: 8,
    fontSize: 14,
  },
}
</script>
