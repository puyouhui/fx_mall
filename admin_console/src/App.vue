<script setup>
import { onMounted } from 'vue'
import { hiprint } from 'vue-plugin-hiprint'
import { getPrinterAddress, canConnectToPrinter, clearClientCache, isOnlineEnvironment } from './utils/printer'

// 检查连接状态
const checkConnectionStatus = () => {
  try {
    if (hiprint && hiprint.hiwebSocket) {
      const isConnected = hiprint.hiwebSocket.opened || false
      const printerAddress = getPrinterAddress()
      if (isConnected) {
        console.log('✅ 打印客户端已连接')
      } else {
        console.warn('⚠️ 打印客户端未连接，请检查:')
        console.warn('  1. 打印客户端是否正在运行')
        console.warn(`  2. 地址是否正确: ${printerAddress}`)
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
    // 检查是否可以在当前环境下连接打印机
    const connectionCheck = canConnectToPrinter()
    if (!connectionCheck.canConnect) {
      console.warn('⚠️ 打印机连接受限:', connectionCheck.reason)
      if (connectionCheck.suggestion) {
        console.warn('💡 建议:', connectionCheck.suggestion)
      }
      // 仍然尝试初始化，但用户应该知道可能无法连接
    }
    
    const printerAddress = getPrinterAddress()
    hiprint.init({
      host: printerAddress, // 从本地存储获取打印机地址
      token: "vue-plugin-hiprint", // 与打印客户端相同的 token
    })
    
    console.log('hiprint 初始化完成', hiprint)
    console.log('打印机地址:', printerAddress)
    
    // 监听连接状态
    if (hiprint.hiwebSocket) {
      // 监听连接打开事件
      hiprint.hiwebSocket.onopen = () => {
        console.log('✅ 打印客户端连接成功')
        
        // 如果是线上环境，连接成功后获取客户端列表
        if (isOnlineEnvironment() && hiprint.hiwebSocket.socket) {
          // 清除旧的缓存
          clearClientCache()
          
          // 延迟一下，确保连接完全建立
          setTimeout(() => {
            // 触发获取客户端列表（会在打印时自动获取）
            console.log('📡 准备获取客户端列表...')
          }, 1000)
        }
      }
      
      // 监听连接关闭事件
      hiprint.hiwebSocket.onclose = () => {
        console.warn('⚠️ 打印客户端连接已关闭')
        // 清除客户端缓存
        clearClientCache()
      }
      
      // 监听连接错误事件
      hiprint.hiwebSocket.onerror = (error) => {
        console.error('❌ 打印客户端连接错误:', error)
        // 如果是 HTTPS 页面且配置的是本地地址，提示使用中转服务
        if (window.location.protocol === 'https:') {
          const printerAddress = getPrinterAddress()
          if (printerAddress.startsWith('http://')) {
            console.error('💡 提示: HTTPS 页面无法连接到本地 HTTP 打印机。')
            console.error('   解决方案：使用中转服务（node-hiprint-transit），配置地址为 https://域名:端口')
            console.error('   参考：https://github.com/Xavier9896/node-hiprint-transit')
          }
        }
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
