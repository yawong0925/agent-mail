// sysmgr-web/vite.config.js
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [
    vue(),
    tailwindcss(),
  ],
  server: {
    host: true, // Allows you to access the dev server via your local network IP
    port: 5174  // Running on 5174 so it doesn't collide with the port 8080 frontend
  }
})