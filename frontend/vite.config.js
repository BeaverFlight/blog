import { defineConfig } from 'vite'  // Добавьте этот импорт
import react from '@vitejs/plugin-react'
import path from 'path'  // Для работы с путями

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src')
    }
  },
  server: {
    host: true,
    proxy: {
      '/api': {
        target: 'http://backend:8080',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, '')
      }
    }
  }
})
