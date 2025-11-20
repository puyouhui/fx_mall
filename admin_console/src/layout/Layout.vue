<template>
  <el-container class="app-container">
    <!-- 侧边栏 -->
    <el-aside width="200px" class="app-sidebar">
      <div class="logo-container">
        <h2 class="logo">云鹿进货管理后台</h2>
      </div>
      <el-menu :default-active="activeMenu" class="el-menu-vertical-demo" router>
        <el-menu-item index="/dashboard">
          <el-icon>
            <HomeFilled />
          </el-icon>
          <span>仪表盘</span>
        </el-menu-item>
        <el-menu-item index="/products">
          <el-icon>
            <ShoppingBag />
          </el-icon>
          <span>商品管理</span>
        </el-menu-item>
        <el-menu-item index="/categories">
          <el-icon>
            <Grid />
          </el-icon>
          <span>分类管理</span>
        </el-menu-item>
        <el-menu-item index="/carousel">
          <el-icon>
            <Picture />
          </el-icon>
          <span>轮播图管理</span>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <!-- 主内容区 -->
    <el-container>
      <!-- 顶部导航栏 -->
      <el-header class="app-header">
        <div class="header-left">
          <el-button type="text" @click="toggleSidebar">
            <el-icon>
              <menu />
            </el-icon>
          </el-button>
        </div>
        <div class="header-right">
          <el-dropdown>
            <span class="user-info">
              <el-icon>
                <user />
              </el-icon>
              <span>{{ username || '管理员' }}</span>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="handleLogout">
                  <el-icon>
                    <SwitchFilled />
                  </el-icon>
                  <span>退出登录</span>
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- 页面内容 -->
      <el-main class="app-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  HomeFilled,
  ShoppingBag,
  Grid,
  Picture,
  ShoppingCart,
  Menu,
  User,
  SwitchFilled
} from '@element-plus/icons-vue'
import { logout, getAdminInfo } from '../api/auth'
import { ElMessage } from 'element-plus'

const router = useRouter()
const route = useRoute()
const username = ref('')
const sidebarOpened = ref(true)

// 当前激活的菜单
const activeMenu = computed(() => {
  return route.path
})

// 切换侧边栏
const toggleSidebar = () => {
  sidebarOpened.value = !sidebarOpened.value
}

// 退出登录
const handleLogout = async () => {
  try {
    await logout()
    localStorage.removeItem('token')
    localStorage.removeItem('username')
    router.push('/login')
  } catch (error) {
    console.error('退出登录失败:', error)
  }
}

// 初始化用户信息
onMounted(async () => {
  try {
    // 从localStorage获取用户名作为备用
    username.value = localStorage.getItem('username') || ''

    // 尝试调用API获取管理员信息
    const response = await getAdminInfo()
    if (response.data && response.data.admin && response.data.admin.username) {
      username.value = response.data.admin.username
      localStorage.setItem('username', response.data.admin.username)
    }
  } catch (error) {
    console.error('获取管理员信息失败:', error)
    // 失败时继续使用localStorage中的用户名
  }
})
</script>

<style scoped>
.app-container {
  height: 100vh;
}

.app-sidebar {
  background-color: #EFF5FF;
  color: #fff;
  overflow-y: auto;
}

.logo-container {
  /* padding: 20px; */
  text-align: center;
  height: 60px;
  line-height: 60px;
  /* border-bottom: 1px solid #657288; */
}

.logo {
  color: #2E74FF;
  font-size: 18px;
  margin: 0;
}

.el-menu-vertical-demo {
  background-color: transparent;
  border-right: none;
}

.el-menu-item {
  color: #657288;
  font-size: 14px;
}

.el-menu-item:hover,
.el-menu-item.is-active {
  background-color: #E0EBFF;
  color: #2E74FF;
}

.app-header {
  background-color: #fff;
  border-bottom: 1px solid #e6e6e6;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px;
}

.header-left {
  display: flex;
  align-items: center;
}

.header-right {
  display: flex;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  cursor: pointer;
  padding: 0 10px;
  height: 40px;
  border-radius: 4px;
}

.user-info:hover {
  background-color: #FFFFFF;
}

.app-main {
  padding: 20px;
  overflow-y: auto;
  background-color: #FFFFFF;
}
</style>