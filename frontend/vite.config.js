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
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, '')
      }
    }
  }
})
