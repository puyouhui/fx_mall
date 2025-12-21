<script setup>
import { onMounted } from 'vue'
import { hiprint } from 'vue-plugin-hiprint'

// 检查连接状态
const checkConnectionStatus = () => {
  try {
    if (hiprint && hiprint.hiwebSocket) {
      const isConnected = hiprint.hiwebSocket.opened || false
      if (isConnected) {
        console.log('✅ 打印客户端已连接')
      } else {
        console.warn('⚠️ 打印客户端未连接，请检查:')
        console.warn('  1. 打印客户端是否正在运行')
        console.warn('  2. 地址是否正确: http://198.18.0.1:17521')
        console.warn('  3. 防火墙是否阻止了连接')
      }
    } else {
      console.warn('⚠️ hiprint.hiwebSocket 未初始化')
    }
  } catch (error) {
    console.error('检查连接状态失败:', error)
  }
}

// 初始化 hiprint 打印客户端
onMounted(() => {
  try {
    hiprint.init({
      host: "http://198.18.0.1:17521", // 打印客户端的地址
      token: "vue-plugin-hiprint", // 与打印客户端相同的 token
    })
    
    console.log('hiprint 初始化完成', hiprint)
    
    // 监听连接状态
    if (hiprint.hiwebSocket) {
      // 监听连接打开事件
      hiprint.hiwebSocket.onopen = () => {
        console.log('✅ 打印客户端连接成功')
      }
      
      // 监听连接关闭事件
      hiprint.hiwebSocket.onclose = () => {
        console.warn('⚠️ 打印客户端连接已关闭')
      }
      
      // 监听连接错误事件
      hiprint.hiwebSocket.onerror = (error) => {
        console.error('❌ 打印客户端连接错误:', error)
      }
      
      // 检查当前连接状态（延迟检查，给连接一些时间建立）
      setTimeout(() => {
        checkConnectionStatus()
      }, 1000)
      
      // 每5秒检查一次连接状态
      setInterval(() => {
        checkConnectionStatus()
      }, 5000)
    } else {
      console.warn('⚠️ hiprint.hiwebSocket 未初始化')
    }
  } catch (error) {
    console.error('hiprint 初始化失败:', error)
  }
})
</script>

<template>
  <div id="app">
    <!-- 路由视图 -->
    <router-view />
  </div>
</template>

<style>
/* 全局样式重置 */
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

#app {
  width: 100%;
  height: 100vh;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  padding: 0 0;
  margin: 0 auto;
}

/* 自定义滚动条 */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 4px;
}

::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}
</style>
