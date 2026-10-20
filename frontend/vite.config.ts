import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import UnoCSS from 'unocss/vite'
import VueI18nPlugin from '@intlify/unplugin-vue-i18n/vite'
import { resolve } from 'path'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  // 生产可经 VITE_BASE_PATH=/admin/ 部署到子路径（nginx 统一入口），默认 /
  const basePath = env.VITE_BASE_PATH || '/'

  return {
    base: basePath,
    plugins: [
      vue(),
      UnoCSS(),
      // 语言包预编译：JSON 在构建期编译为消息函数，运行时不携带消息编译器
      VueI18nPlugin({
        include: [resolve(__dirname, 'src/locales/**/*.json')],
      }),
    ],
    resolve: {
      alias: {
        '@': resolve(__dirname, 'src'),
        // 配合语言包预编译，使用 vue-i18n 纯运行时构建
        'vue-i18n': 'vue-i18n/dist/vue-i18n.runtime.esm-bundler.js',
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