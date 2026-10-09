import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import UnoCSS from 'unocss/vite'
import { resolve } from 'path'

// 临时本地验证配置：5174 → 后端 8112（仅验证用，验证后可删除）
export default defineConfig({
  base: '/',
  plugins: [vue(), UnoCSS()],
  resolve: { alias: { '@': resolve(__dirname, 'src') } },
  server: {
    port: 5174,
    proxy: { '/api': { target: 'http://localhost:8112', changeOrigin: true } },
  },
})
