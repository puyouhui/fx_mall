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
        path: 'categories',
        name: 'Categories',
        component: () => import('../views/Categories.vue'),
        meta: {
          title: '分类管理'
        }
      },
      {
        path: 'carousel',
        name: 'Carousel',
        component: () => import('../views/Carousel.vue'),
        meta: {
          title: '轮播图管理'
        }
      },
      {
        path: 'settings',
        name: 'Settings',
        component: () => import('../views/Settings.vue'),
        meta: {
          title: '系统设置'
        }
      },
      {
        path: 'delivery-fee',
        name: 'DeliveryFee',
        component: () => import('../views/DeliveryFee.vue'),
        meta: {
          title: '配送费设置'
        }
      },
      {
        path: 'suppliers',
        name: 'Suppliers',
        component: () => import('../views/Suppliers.vue'),
        meta: {
          title: '供应商管理'
        }
      },
      {
        path: 'suppliers/payments',
        name: 'SupplierPayments',
        component: () => import('../views/SupplierPayments.vue'),
        meta: {
          title: '供应商付款统计'
        }
      },
      {
        path: 'hot-products',
        name: 'HotProducts',
        component: () => import('../views/HotProducts.vue'),
        meta: {
          title: '热销产品管理'
        }
      },
      {
        path: 'hot-search-keywords',
        name: 'HotSearchKeywords',
        component: () => import('../views/HotSearchKeywords.vue'),
        meta: {
          title: '热门搜索关键词'
        }
      },
      {
        path: 'mini-users',
        name: 'MiniUsers',
        component: () => import('../views/MiniUsers.vue'),
        meta: {
          title: '小程序用户'
        }
      },
      {
        path: 'employees',
        name: 'Employees',
        component: () => import('../views/Employees.vue'),
        meta: {
          title: '员工管理'
        }
      },
      {
        path: 'coupons',
        name: 'Coupons',
        component: () => import('../views/Coupons.vue'),
        meta: {
          title: '优惠券管理'
        }
      },
      {
        path: 'orders',
        name: 'Orders',
        component: () => import('../views/Orders.vue'),
        meta: {
          title: '订单管理'
        }
      },
      {
        path: 'delivery-records',
        name: 'DeliveryRecords',
        component: () => import('../views/DeliveryRecords.vue'),
        meta: {
          title: '配送记录'
        }
      },
      {
        path: 'delivery-management',
        name: 'DeliveryManagement',
        component: () => import('../views/DeliveryManagement.vue'),
        meta: {
          title: '配送管理'
        }
      },
      {
        path: 'delivery-income',
        name: 'DeliveryIncome',
        component: () => import('../views/DeliveryIncome.vue'),
        meta: {
          title: '配送费结算管理'
        }
      },
      {
        path: 'sales-commission',
        name: 'SalesCommission',
        component: () => import('../views/SalesCommission.vue'),
        meta: {
          title: '销售分成管理'
        }
      },
      {
        path: 'product-requests',
        name: 'ProductRequests',
        component: () => import('../views/ProductRequests.vue'),
        meta: {
          title: '新品需求管理'
        }
      },
      {
        path: 'supplier-applications',
        name: 'SupplierApplications',
        component: () => import('../views/SupplierApplications.vue'),
        meta: {
          title: '供应商合作申请'
        }
      },
      {
        path: 'price-feedback',
        name: 'PriceFeedback',
        component: () => import('../views/PriceFeedback.vue'),
        meta: {
          title: '价格反馈管理'
        }
      },
      {
        path: 'payment-verification',
        name: 'PaymentVerification',
        component: () => import('../views/PaymentVerification.vue'),
        meta: {
          title: '收款审核管理'
        }
      },
      {
        path: 'employee-locations',
        name: 'EmployeeLocations',
        component: () => import('../views/EmployeeLocations.vue'),
        meta: {
          title: '员工位置'
        }
      },
      {
        path: 'rich-content',
        name: 'RichContent',
        component: () => import('../views/RichContent.vue'),
        meta: {
          title: '富文本内容管理'
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
  document.title = to.meta.title || '后台管理系统'
  
  // 检查是否需要登录
  const token = localStorage.getItem('token')
  if (to.path !== '/login' && !token) {
    // 未登录，重定向到登录页
    next('/login')
  } else {
    // 已登录，继续访问
    next()
  }
})

export default router