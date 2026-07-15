import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { VitePWA } from 'vite-plugin-pwa'

export default defineConfig({
  plugins: [
    vue(),
    VitePWA({
      registerType: 'autoUpdate',
      includeAssets: ['icon.svg'],
      manifest: {
        name: '棋境 · Xiangqi Lab',
        short_name: '棋境',
        description: '中国象棋人机对战、棋谱学习与复盘',
        theme_color: '#f3eee4',
        background_color: '#f3eee4',
        display: 'standalone',
        start_url: '/',
        icons: [{ src: '/icon.svg', sizes: 'any', type: 'image/svg+xml', purpose: 'any maskable' }],
      },
      workbox: {
        navigateFallbackDenylist: [/^\/api\//],
        cleanupOutdatedCaches: true,
      },
    }),
  ],
  resolve: { alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) } },
  server: { port: 5666, strictPort: true },
  preview: { port: 4173 },
  test: { environment: 'jsdom', globals: true, setupFiles: ['./src/tests/setup.ts'] },
})
