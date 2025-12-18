<template>
  <el-container class="app-container">
    <!-- 侧边栏 -->
    <el-aside width="200px" class="app-sidebar">
      <div class="logo-container">
        <h2 class="logo">管理后台</h2>
      </div>
      <el-menu 
        :default-active="activeMenu" 
        class="el-menu-vertical-demo" 
        router
        :default-openeds="defaultOpeneds"
        unique-opened
      >
        <!-- 仪表盘 -->
        <el-menu-item index="/dashboard">
          <el-icon>
            <HomeFilled />
          </el-icon>
          <span>仪表盘</span>
        </el-menu-item>

        <!-- 商品管理 -->
        <el-sub-menu index="product-management">
          <template #title>
            <el-icon>
              <ShoppingBag />
            </el-icon>
            <span>商品管理</span>
          </template>
          <el-menu-item index="/products">
            <el-icon>
              <ShoppingBag />
            </el-icon>
            <span>商品列表</span>
          </el-menu-item>
          <el-menu-item index="/categories">
            <el-icon>
              <Grid />
            </el-icon>
            <span>分类管理</span>
          </el-menu-item>
          <el-menu-item index="/hot-products">
            <el-icon>
              <ShoppingCart />
            </el-icon>
            <span>热销产品</span>
          </el-menu-item>
          <el-menu-item index="/carousel">
            <el-icon>
              <Picture />
            </el-icon>
            <span>轮播图管理</span>
          </el-menu-item>
          <el-menu-item index="/hot-search-keywords">
            <el-icon>
              <Search />
            </el-icon>
            <span>热门搜索</span>
          </el-menu-item>
        </el-sub-menu>

        <!-- 订单管理 -->
        <el-sub-menu index="order-management">
          <template #title>
            <el-icon>
              <Document />
            </el-icon>
            <span>订单管理</span>
          </template>
          <el-menu-item index="/orders">
            <el-icon>
              <Document />
            </el-icon>
            <span>订单列表</span>
          </el-menu-item>
          <el-menu-item index="/delivery-records">
            <el-icon>
              <Box />
            </el-icon>
            <span>配送记录</span>
          </el-menu-item>
          <el-menu-item index="/delivery-income">
            <el-icon>
              <Money />
            </el-icon>
            <span>配送费结算</span>
          </el-menu-item>
          <el-menu-item index="/sales-commission">
            <el-icon>
              <Money />
            </el-icon>
            <span>销售分成管理</span>
          </el-menu-item>
        </el-sub-menu>

        <!-- 用户管理 -->
        <el-sub-menu index="user-management">
          <template #title>
            <el-icon>
              <User />
            </el-icon>
            <span>用户管理</span>
          </template>
          <el-menu-item index="/mini-users">
            <el-icon>
              <User />
            </el-icon>
            <span>小程序用户</span>
          </el-menu-item>
          <el-menu-item index="/employees">
            <el-icon>
              <UserFilled />
            </el-icon>
            <span>员工管理</span>
          </el-menu-item>
          <el-menu-item index="/employee-locations">
            <el-icon>
              <Location />
            </el-icon>
            <span>员工位置</span>
          </el-menu-item>
        </el-sub-menu>

        <!-- 营销管理 -->
        <el-sub-menu index="marketing-management">
          <template #title>
            <el-icon>
              <Ticket />
            </el-icon>
            <span>营销管理</span>
          </template>
          <el-menu-item index="/coupons">
            <el-icon>
              <Ticket />
            </el-icon>
            <span>优惠券管理</span>
          </el-menu-item>
        </el-sub-menu>

        <!-- 供应商管理 -->
        <el-menu-item index="/suppliers">
          <el-icon>
            <Shop />
          </el-icon>
          <span>供应商管理</span>
        </el-menu-item>

        <!-- 系统设置 -->
        <el-sub-menu index="system-settings">
          <template #title>
            <el-icon>
              <Setting />
            </el-icon>
            <span>系统设置</span>
          </template>
          <el-menu-item index="/settings">
            <el-icon>
              <Setting />
            </el-icon>
            <span>基础设置</span>
          </el-menu-item>
          <el-menu-item index="/delivery-fee">
            <el-icon>
              <Money />
            </el-icon>
            <span>配送费设置</span>
          </el-menu-item>
        </el-sub-menu>
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
  UserFilled,
  SwitchFilled,
  Setting,
  Shop,
  Money,
  Ticket,
  Document,
  Box,
  Location,
  Search
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

// 默认展开的菜单（可以根据需要设置）
const defaultOpeneds = ref([])

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

.el-menu-item,
:deep(.el-sub-menu__title) {
  color: #657288;
  font-size: 14px;
}

.el-menu-item:hover,
.el-menu-item.is-active,
:deep(.el-sub-menu__title:hover) {
  background-color: #E0EBFF;
  color: #2E74FF;
}

:deep(.el-sub-menu.is-opened > .el-sub-menu__title) {
  color: #2E74FF;
}

:deep(.el-sub-menu .el-menu-item) {
  padding-left: 50px !important;
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