import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    // DEV: frontend :5173, backend Go :8080 → proxy chuyển tiếp
    // → code chỉ gọi axios.get('/api/products'), không hardcode domain
    // PROD: Nginx proxy y hệt → cùng 1 code chạy cả 2 nơi
    proxy: {
      '/api': 'http://localhost:8080',
      '/uploads': 'http://localhost:8080',
    },
  },
})
