<template>
  <a-config-provider :theme="antTheme">
    <router-view />
  </a-config-provider>
</template>

<script setup lang="ts">
import { theme as antdTheme } from 'ant-design-vue'

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
