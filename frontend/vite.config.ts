import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import UnoCSS from 'unocss/vite'
import { resolve } from 'path'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  // 生产可经 VITE_BASE_PATH=/admin/ 部署到子路径（nginx 统一入口），默认 /
  const basePath = env.VITE_BASE_PATH || '/'

  return {
    base: basePath,
    plugins: [vue(), UnoCSS()],
    resolve: {
      alias: {
        '@': resolve(__dirname, 'src'),
      },
    },
    server: {
      port: 5173,
      proxy: {
        '/api': {
          target: 'http://localhost:8111',
          changeOrigin: true,
        },
      },
    },
  }
})