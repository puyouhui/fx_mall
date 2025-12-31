import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  base: '/supplier/', // 使用 /supplier/ 作为基础路径，匹配 Nginx 配置 location /supplier/
  server: {
    port: 21321,
    open: false
  },
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
    sourcemap: false,
    // 优化构建
    rollupOptions: {
      output: {
        // 手动分包
        manualChunks: {
          'vue-vendor': ['vue', 'vue-router'],
          'element-plus': ['element-plus'],
          'chart': ['echarts']
        }
      }
    },
    // 构建大小限制警告
    chunkSizeWarningLimit: 1000
  }
})
 