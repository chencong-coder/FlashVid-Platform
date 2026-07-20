import { fileURLToPath, URL } from 'node:url'

import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vite'
import compression from 'vite-plugin-compression'

export default defineConfig(({ mode }) => ({
  plugins: [
    vue(),
    mode === 'production' &&
      compression({
        algorithm: 'gzip',
        ext: '.gz',
        threshold: 10 * 1024,
      }),
  ].filter(Boolean),
  resolve: {
    alias: [
      { find: '@', replacement: fileURLToPath(new URL('./src', import.meta.url)) },
      {
        find: /^vue3-video-play$/,
        replacement: fileURLToPath(
          new URL('./node_modules/vue3-video-play/dist/index.mjs', import.meta.url),
        ),
      },
    ],
  },
  server: { host: '0.0.0.0', port: 5173 },
  build: {
    target: 'es2018',
    cssCodeSplit: true,
    chunkSizeWarningLimit: 650,
    rollupOptions: {
      output: {
        manualChunks: {
          vue: ['vue', 'vue-router', 'pinia'],
          video: ['vue3-video-play', 'vue-virtual-scroller'],
          vant: ['vant'],
        },
      },
    },
  },
}))
