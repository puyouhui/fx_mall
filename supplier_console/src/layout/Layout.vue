<template>
  <el-container class="app-container">
    <!-- 全屏水印 -->
    <Watermark :text="watermarkText" />
    <!-- 侧边栏 -->
    <el-aside width="200px" class="app-sidebar">
      <div class="logo-container">
        <h2 class="logo">供应商后台</h2>
      </div>
      <el-menu 
        :default-active="activeMenu" 
        class="el-menu-vertical-demo" 
        router
        unique-opened
      >
        <!-- 仪表盘 -->
        <el-menu-item index="/dashboard">
          <el-icon><HomeFilled /></el-icon>
          <span>仪表盘</span>
        </el-menu-item>

        <!-- 商品管理 -->
        <el-menu-item index="/products">
          <el-icon><ShoppingBag /></el-icon>
          <span>商品管理</span>
        </el-menu-item>

        <!-- 订单管理 -->
        <el-menu-item index="/orders">
          <el-icon><Document /></el-icon>
          <span>订单管理</span>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <!-- 主内容区 -->
    <el-container>
      <!-- 顶部导航栏 -->
      <el-header class="app-header">
        <div class="header-left">
          <span class="header-title">{{ currentPageTitle }}</span>
        </div>
        <div class="header-right">
          <span class="username">{{ username }}</span>
          <el-button type="text" @click="handleLogout">退出登录</el-button>
        </div>
      </el-header>

      <!-- 内容区域 -->
      <el-main class="app-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessageBox, ElMessage } from 'element-plus'
import { HomeFilled, ShoppingBag, Document } from '@element-plus/icons-vue'
import { logout } from '../api/auth'
import Watermark from '../components/Watermark.vue'

const router = useRouter()
const route = useRoute()

const activeMenu = computed(() => route.path)
const username = ref(localStorage.getItem('username') || '供应商')
const currentPageTitle = computed(() => route.meta.title || '供应商后台')

// 水印文本：显示供应商名称
const watermarkText = computed(() => {
  const supplierName = localStorage.getItem('supplierName') || '供应商'
  return `${supplierName} - 供应商后台`
})

// 退出登录
const handleLogout = async () => {
  try {
    await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    // 调用退出登录API
    try {
      await logout()
    } catch (error) {
      console.error('退出登录API调用失败:', error)
    }
    
    // 清除本地存储
    localStorage.removeItem('token')
    localStorage.removeItem('username')
    
    ElMessage.success('退出登录成功')
    router.push('/login')
  } catch (error) {
    // 用户取消操作
  }
}

onMounted(() => {
  // 检查登录状态
  const token = localStorage.getItem('token')
  if (!token) {
    router.push('/login')
  }
})
</script>

<style scoped>
.app-container {
  height: 100vh;
}

.app-sidebar {
  background-color: #304156;
  overflow-y: auto;
}

.logo-container {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #2b3a4a;
}

.logo {
  color: #fff;
  font-size: 18px;
  font-weight: bold;
  margin: 0;
}

.el-menu {
  border-right: none;
  background-color: #304156;
}

.el-menu-item {
  color: #bfcbd9;
}

.el-menu-item:hover {
  background-color: #263445;
  color: #409eff;
}

.el-menu-item.is-active {
  background-color: #409eff;
  color: #fff;
}

.app-header {
  background-color: #fff;
  border-bottom: 1px solid #e4e7ed;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
}

.header-left {
  flex: 1;
}

.header-title {
  font-size: 18px;
  font-weight: 500;
  color: #303133;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 15px;
}

.username {
  color: #606266;
  font-size: 14px;
}

.app-main {
  background-color: #f0f2f5;
  padding: 20px;
  overflow-y: auto;
}
</style>

