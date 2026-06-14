import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { VitePWA } from 'vite-plugin-pwa'

declare const process: { env: Record<string, string | undefined> }

export default defineConfig(() => {
  const base = process.env.VITE_BASE || '/'
  return {
    base,
    plugins: [
      react(),
      VitePWA({
        registerType: 'autoUpdate',
        manifest: {
          name: 'VOUCH — Verified Outgoing-call Universal Compliance Hub',
          short_name: 'VOUCH',
          description: 'VOUCH — 통화 연결 전 발신 신원 인증 브로커 프로토콜',
          theme_color: '#0a0a0f',
          background_color: '#0a0a0f',
          display: 'standalone',
          orientation: 'portrait',
          scope: base,
          start_url: base
        },
        workbox: {
          globPatterns: ['**/*.{js,css,html,svg,png,ico}']
        }
      })
    ],
    server: {
      port: 3000,
      host: true,
      strictPort: false
    }
  }
})
