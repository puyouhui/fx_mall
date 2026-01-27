<template>
  <el-container class="app-container">
    <!-- 侧边栏 -->
    <el-aside width="200px" class="app-sidebar">
      <div class="logo-container">
        <h2 class="logo">供应链管理后台</h2>
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
          <el-menu-item index="/image-library">
            <el-icon>
              <Picture />
            </el-icon>
            <span>图库管理</span>
          </el-menu-item>
          <el-menu-item index="/hot-search-keywords">
            <el-icon>
              <Search />
            </el-icon>
            <span>热门搜索</span>
          </el-menu-item>
          <el-menu-item index="/sort-management">
            <el-icon>
              <Sort />
            </el-icon>
            <span>排序管理</span>
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
          <el-menu-item index="/delivery-management">
            <el-icon>
              <Box />
            </el-icon>
            <span>配送管理</span>
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
          <el-menu-item index="/payment-verification">
            <el-icon>
              <Money />
            </el-icon>
            <span>收款审核</span>
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
          <el-menu-item index="/rich-content">
            <el-icon>
              <Document />
            </el-icon>
            <span>富文本内容</span>
          </el-menu-item>
          <el-menu-item index="/product-requests">
            <el-icon>
              <Document />
            </el-icon>
            <span>新品需求管理</span>
          </el-menu-item>
          <el-menu-item index="/supplier-applications">
            <el-icon>
              <Document />
            </el-icon>
            <span>供应商合作申请</span>
          </el-menu-item>
          <el-menu-item index="/price-feedback" @click="handleMenuClick('/price-feedback')">
            <el-icon><Document /></el-icon>
            <span>价格反馈管理</span>
          </el-menu-item>
          <el-menu-item index="/reward-activities" @click="handleMenuClick('/reward-activities')">
            <el-icon><Ticket /></el-icon>
            <span>奖励活动管理</span>
          </el-menu-item>
        </el-sub-menu>

        <!-- 供应商管理 -->
        <el-sub-menu index="supplier-management">
          <template #title>
            <el-icon>
              <Shop />
            </el-icon>
            <span>供应商</span>
          </template>
          <el-menu-item index="/suppliers">
            <el-icon>
              <Shop />
            </el-icon>
            <span>供应商管理</span>
          </el-menu-item>
          <el-menu-item index="/suppliers/payments">
            <el-icon>
              <Money />
            </el-icon>
            <span>付款统计</span>
          </el-menu-item>
        </el-sub-menu>

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
          <div class="history-tabs-container">
            <div class="history-tabs-wrapper" ref="tabsWrapperRef">
              <div
                v-for="(item, index) in historyList"
                :key="item.path"
                :class="['history-tab', { 'is-active': item.path === route.path }]"
                @click="handleTabClick(item.path)"
              >
                <span class="tab-title">{{ item.title }}</span>
                <el-icon
                  v-if="historyList.length > 1"
                  class="tab-close"
                  @click.stop="handleCloseTab(item.path, index)"
                >
                  <Close />
                </el-icon>
              </div>
            </div>
            <el-button
              v-if="historyList.length > 0"
              type="text"
              class="clear-history-btn"
              @click="handleClearHistory"
              title="清空所有标签"
            >
              <el-icon>
                <Delete />
              </el-icon>
            </el-button>
          </div>
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
  Search,
  Clock,
  Delete,
  Close,
  Sort
} from '@element-plus/icons-vue'
import { logout, getAdminInfo } from '../api/auth'
import { ElMessage } from 'element-plus'

const router = useRouter()
const route = useRoute()
const username = ref('')
const historyList = ref([])
const tabsWrapperRef = ref(null)

// 历史记录存储键名
const HISTORY_STORAGE_KEY = 'admin_history_list'
const MAX_HISTORY_COUNT = 15

// 当前激活的菜单
const activeMenu = computed(() => {
  return route.path
})

// 默认展开的菜单（可以根据需要设置）
const defaultOpeneds = ref([])

// 加载历史记录
const loadHistory = () => {
  try {
    const stored = localStorage.getItem(HISTORY_STORAGE_KEY)
    if (stored) {
      historyList.value = JSON.parse(stored)
    }
  } catch (error) {
    console.error('加载历史记录失败:', error)
    historyList.value = []
  }
}

// 保存历史记录
const saveHistory = () => {
  try {
    localStorage.setItem(HISTORY_STORAGE_KEY, JSON.stringify(historyList.value))
  } catch (error) {
    console.error('保存历史记录失败:', error)
  }
}

// 添加历史记录
const addHistory = (path, title) => {
  // 排除登录页和首页
  if (path === '/login' || path === '/') {
    return
  }

  // 检查是否已存在
  const existingIndex = historyList.value.findIndex(item => item.path === path)
  
  if (existingIndex !== -1) {
    // 如果已存在，只更新标题和时间戳，不改变位置
    historyList.value[existingIndex].title = title
    historyList.value[existingIndex].timestamp = Date.now()
  } else {
    // 如果不存在，添加到开头
    historyList.value.unshift({
      path,
      title,
      timestamp: Date.now()
    })

    // 限制最大数量
    if (historyList.value.length > MAX_HISTORY_COUNT) {
      historyList.value = historyList.value.slice(0, MAX_HISTORY_COUNT)
    }
  }

  // 保存到localStorage
  saveHistory()
}

// 处理标签点击
const handleTabClick = (path) => {
  router.push(path)
}

// 处理关闭标签
const handleCloseTab = (path, index) => {
  // 如果关闭的是当前页面，需要跳转到其他标签或首页
  if (path === route.path) {
    // 如果还有其他标签，跳转到下一个或上一个
    if (historyList.value.length > 1) {
      const nextIndex = index < historyList.value.length - 1 ? index : index - 1
      const nextPath = historyList.value[nextIndex]?.path || '/dashboard'
      router.push(nextPath)
    } else {
      // 没有其他标签了，跳转到首页
      router.push('/dashboard')
    }
  }
  
  // 从历史记录中移除
  historyList.value = historyList.value.filter(item => item.path !== path)
  saveHistory()
}

// 清空所有历史记录
const handleClearHistory = () => {
  historyList.value = []
  saveHistory()
  router.push('/dashboard')
  ElMessage.success('已清空所有标签')
}

// 退出登录
const handleLogout = async () => {
  try {
    await logout()
    localStorage.removeItem('token')
    localStorage.removeItem('username')
    // 使用相对路径，Vue Router会自动处理base URL
    router.push('/login')
  } catch (error) {
    console.error('退出登录失败:', error)
  }
}

// 监听路由变化，添加历史记录
router.afterEach((to) => {
  if (to.meta && to.meta.title) {
    addHistory(to.path, to.meta.title)
  }
})

// 初始化用户信息
onMounted(async () => {
  // 加载历史记录
  loadHistory()

  // 添加当前页面到历史记录
  if (route.meta && route.meta.title) {
    addHistory(route.path, route.meta.title)
  }

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
  gap: 10px;
  flex: 1;
  overflow: hidden;
}

.history-tabs-container {
  display: flex;
  align-items: center;
  flex: 1;
  overflow: hidden;
  margin-left: 10px;
}

.history-tabs-wrapper {
  display: flex;
  align-items: center;
  gap: 4px;
  overflow-x: auto;
  overflow-y: hidden;
  flex: 1;
  scrollbar-width: thin;
  scrollbar-color: #c0c4cc transparent;
}

.history-tabs-wrapper::-webkit-scrollbar {
  height: 4px;
}

.history-tabs-wrapper::-webkit-scrollbar-track {
  background: transparent;
}

.history-tabs-wrapper::-webkit-scrollbar-thumb {
  background: #c0c4cc;
  border-radius: 2px;
}

.history-tabs-wrapper::-webkit-scrollbar-thumb:hover {
  background: #a0a4a8;
}

.history-tab {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background-color: #f5f7fa;
  border: 1px solid #e4e7ed;
  border-radius: 4px 4px 0 0;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.2s;
  position: relative;
  min-width: 80px;
  max-width: 200px;
}

.history-tab:hover {
  background-color: #ecf5ff;
  border-color: #b3d8ff;
}

.history-tab.is-active {
  background-color: #409eff;
  border-color: #409eff;
  color: #fff;
  z-index: 1;
}

.history-tab.is-active .tab-title {
  color: #fff;
}

.tab-title {
  font-size: 13px;
  color: #606266;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 1;
}

.tab-close {
  font-size: 12px;
  color: #909399;
  cursor: pointer;
  transition: color 0.2s;
  flex-shrink: 0;
}

.tab-close:hover {
  color: #f56c6c;
}

.history-tab.is-active .tab-close {
  color: rgba(255, 255, 255, 0.8);
}

.history-tab.is-active .tab-close:hover {
  color: #fff;
}

.clear-history-btn {
  margin-left: 8px;
  flex-shrink: 0;
  padding: 4px 8px;
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