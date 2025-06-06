import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      // всё, что начинается на /api, проксируем на Go-бекенд
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        ws: true             // ← обязательно, иначе WebSocket-ы не пройдут
      }
    }
  }
});