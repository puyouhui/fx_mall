import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    name: 'Layout',
    component: () => import('../layout/Layout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('../views/Dashboard.vue'),
        meta: {
          title: '仪表盘'
        }
      },
      {
        path: 'products',
        name: 'Products',
        component: () => import('../views/Products.vue'),
        meta: {
          title: '商品管理'
        }
      },
      {
        path: 'orders',
        name: 'Orders',
        component: () => import('../views/Orders.vue'),
        meta: {
          title: '历史记录'
        }
      }
    ]
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: {
      title: '登录'
    }
  },
  {
    path: '/mobile/goods',
    name: 'MobileGoods',
    component: () => import('../views/MobileGoods.vue'),
    meta: {
      title: '待备货货物列表',
      requiresAuth: false // 不需要token验证
    }
  }
]

// 创建路由实例
const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL || '/'),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  // 设置页面标题
  document.title = to.meta.title || '供应商后台管理系统'
  
  // 检查是否需要登录（移动端页面不需要token）
  if (to.meta.requiresAuth === false) {
    // 不需要token验证的页面，直接访问
    next()
  } else {
    const token = localStorage.getItem('token')
    if (to.path !== '/login' && !token) {
      // 未登录，重定向到登录页
      next('/login')
    } else {
      // 已登录，继续访问
      next()
    }
  }
})

export default router

