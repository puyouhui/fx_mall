<template>
  <div class="dashboard-container">
    <div class="dashboard-header">
      <h1 class="page-title">运营数据中心</h1>
      <div class="time-range-selector">
        <el-radio-group v-model="timeRange" @change="handleTimeRangeChange">
          <el-radio-button label="today">今日</el-radio-button>
          <el-radio-button label="week">本周</el-radio-button>
          <el-radio-button label="month">本月</el-radio-button>
        </el-radio-group>
      </div>
    </div>

    <!-- 顶部核心指标卡片 -->
    <div class="stats-cards">
      <el-card class="stat-card stat-card-orders" shadow="hover">
        <div class="stat-content">
          <div class="stat-icon-wrapper orders">
            <el-icon class="stat-icon"><ShoppingCart /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-label">今日订单数</div>
            <div class="stat-value">{{ formatNumber(orderStats.total_orders) }}</div>
            <div class="stat-growth" :class="orderStats.growth >= 0 ? 'positive' : 'negative'">
              <el-icon v-if="orderStats.growth >= 0"><ArrowUp /></el-icon>
              <el-icon v-else><ArrowDown /></el-icon>
              <span>{{ formatPercent(Math.abs(orderStats.growth)) }}%</span>
              <span class="growth-label">环比</span>
            </div>
          </div>
        </div>
      </el-card>

      <el-card class="stat-card stat-card-revenue" shadow="hover">
        <div class="stat-content">
          <div class="stat-icon-wrapper revenue">
            <el-icon class="stat-icon"><Money /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-label">今日收入</div>
            <div class="stat-value">¥{{ formatMoney(revenueStats.total_revenue) }}</div>
            <div class="stat-growth" :class="revenueStats.growth >= 0 ? 'positive' : 'negative'">
              <el-icon v-if="revenueStats.growth >= 0"><ArrowUp /></el-icon>
              <el-icon v-else><ArrowDown /></el-icon>
              <span>{{ formatPercent(Math.abs(revenueStats.growth)) }}%</span>
              <span class="growth-label">环比</span>
            </div>
          </div>
        </div>
      </el-card>

      <el-card class="stat-card stat-card-profit" shadow="hover">
        <div class="stat-content">
          <div class="stat-icon-wrapper profit">
            <el-icon class="stat-icon"><TrendCharts /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-label">今日净利润</div>
            <div class="stat-value">¥{{ formatMoney(revenueStats.net_profit) }}</div>
            <div class="stat-extra">
              <span class="profit-rate">利润率: {{ formatPercent(profitRate) }}%</span>
            </div>
          </div>
        </div>
      </el-card>

      <el-card class="stat-card stat-card-users" shadow="hover">
        <div class="stat-content">
          <div class="stat-icon-wrapper users">
            <el-icon class="stat-icon"><User /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-label">今日新增用户</div>
            <div class="stat-value">{{ formatNumber(userStats.new_users) }}</div>
            <div class="stat-extra">
              <span class="total-users">总用户: {{ formatNumber(userStats.total_users) }}</span>
            </div>
          </div>
        </div>
      </el-card>
    </div>

    <!-- 订单状态分布 -->
    <el-row :gutter="20" class="dashboard-row">
      <el-col :span="12">
        <el-card shadow="hover" class="info-card">
          <template #header>
            <div class="card-header">
              <el-icon class="header-icon"><Document /></el-icon>
              <span>订单状态分布</span>
            </div>
          </template>
          <div class="order-status-list">
            <div class="status-item status-pending">
              <div class="status-info">
                <span class="status-dot"></span>
                <span class="status-label">待配送</span>
              </div>
              <span class="status-value">{{ formatNumber(orderStats.pending_delivery) }}</span>
            </div>
            <div class="status-item status-delivering">
              <div class="status-info">
                <span class="status-dot"></span>
                <span class="status-label">配送中</span>
              </div>
              <span class="status-value">{{ formatNumber(orderStats.delivering) }}</span>
            </div>
            <div class="status-item status-delivered">
              <div class="status-info">
                <span class="status-dot"></span>
                <span class="status-label">已送达</span>
              </div>
              <span class="status-value">{{ formatNumber(orderStats.delivered) }}</span>
            </div>
            <div class="status-item status-paid">
              <div class="status-info">
                <span class="status-dot"></span>
                <span class="status-label">已收款</span>
              </div>
              <span class="status-value">{{ formatNumber(orderStats.paid) }}</span>
            </div>
            <div class="status-item status-cancelled">
              <div class="status-info">
                <span class="status-dot"></span>
                <span class="status-label">已取消</span>
              </div>
              <span class="status-value">{{ formatNumber(orderStats.cancelled) }}</span>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="12">
        <el-card shadow="hover" class="info-card">
          <template #header>
            <div class="card-header">
              <el-icon class="header-icon"><Money /></el-icon>
              <span>收入成本分析</span>
            </div>
          </template>
          <div class="revenue-cost-list">
            <div class="revenue-item revenue-income">
              <div class="revenue-info">
                <span class="revenue-icon">📈</span>
                <span class="revenue-label">总收入</span>
              </div>
              <span class="revenue-value positive">¥{{ formatMoney(revenueStats.total_revenue) }}</span>
            </div>
            <div class="revenue-item revenue-cost">
              <div class="revenue-info">
                <span class="revenue-icon">📦</span>
                <span class="revenue-label">商品成本</span>
              </div>
              <span class="revenue-value">¥{{ formatMoney(revenueStats.goods_cost) }}</span>
            </div>
            <div class="revenue-item revenue-cost">
              <div class="revenue-info">
                <span class="revenue-icon">🚚</span>
                <span class="revenue-label">配送成本</span>
              </div>
              <span class="revenue-value">¥{{ formatMoney(revenueStats.delivery_cost) }}</span>
            </div>
            <div class="revenue-item revenue-cost">
              <div class="revenue-info">
                <span class="revenue-icon">💰</span>
                <span class="revenue-label">销售分成</span>
              </div>
              <span class="revenue-value">¥{{ formatMoney(revenueStats.sales_commission) }}</span>
            </div>
            <div class="revenue-item revenue-total">
              <div class="revenue-info">
                <span class="revenue-icon">✨</span>
                <span class="revenue-label">净利润</span>
              </div>
              <span class="revenue-value positive">¥{{ formatMoney(revenueStats.net_profit) }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 趋势图表 -->
    <el-row :gutter="20" class="dashboard-row">
      <el-col :span="24">
        <el-card shadow="hover" class="chart-card">
          <template #header>
            <div class="card-header">
              <el-icon class="header-icon"><TrendCharts /></el-icon>
              <span>订单趋势</span>
            </div>
          </template>
          <div class="chart-container">
            <canvas ref="orderTrendChart"></canvas>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="dashboard-row">
      <el-col :span="24">
        <el-card shadow="hover" class="chart-card">
          <template #header>
            <div class="card-header">
              <el-icon class="header-icon"><TrendCharts /></el-icon>
              <span>收入利润趋势</span>
            </div>
          </template>
          <div class="chart-container">
            <canvas ref="revenueTrendChart"></canvas>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 热销商品和绩效排名 -->
    <el-row :gutter="20" class="dashboard-row">
      <el-col :span="12">
        <el-card shadow="hover" class="ranking-card">
          <template #header>
            <div class="card-header">
              <el-icon class="header-icon"><Trophy /></el-icon>
              <span>热销商品 Top 10</span>
            </div>
          </template>
          <el-table :data="hotProducts" size="small" stripe class="ranking-table">
            <el-table-column type="index" label="排名" width="60" />
            <el-table-column prop="product_name" label="商品名称" />
            <el-table-column prop="total_quantity" label="销量" width="80" align="right" />
            <el-table-column prop="total_amount" label="销售额" width="120" align="right">
              <template #default="scope">
                ¥{{ formatMoney(scope.row.total_amount) }}
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>

      <el-col :span="12">
        <el-card shadow="hover" class="ranking-card">
          <template #header>
            <div class="card-header">
              <el-icon class="header-icon"><Box /></el-icon>
              <span>配送员绩效排名</span>
            </div>
          </template>
          <el-table :data="deliveryRanking" size="small" stripe class="ranking-table">
            <el-table-column type="index" label="排名" width="60" />
            <el-table-column prop="employee_name" label="配送员" />
            <el-table-column prop="order_count" label="订单数" width="80" align="right" />
            <el-table-column prop="total_fee" label="配送费" width="120" align="right">
              <template #default="scope">
                ¥{{ formatMoney(scope.row.total_fee) }}
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="dashboard-row">
      <el-col :span="24">
        <el-card shadow="hover" class="ranking-card">
          <template #header>
            <div class="card-header">
              <el-icon class="header-icon"><Trophy /></el-icon>
              <span>销售员绩效排名</span>
            </div>
          </template>
          <el-table :data="salesRanking" size="small" stripe class="ranking-table">
            <el-table-column type="index" label="排名" width="60" />
            <el-table-column prop="employee_name" label="销售员" />
            <el-table-column prop="order_count" label="订单数" width="100" align="right" />
            <el-table-column prop="total_sales" label="销售额" width="120" align="right">
              <template #default="scope">
                ¥{{ formatMoney(scope.row.total_sales) }}
              </template>
            </el-table-column>
            <el-table-column prop="total_commission" label="分成" width="120" align="right">
              <template #default="scope">
                ¥{{ formatMoney(scope.row.total_commission) }}
              </template>
            </el-table-column>
            <el-table-column prop="new_customer_count" label="新客数" width="100" align="right" />
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed, nextTick } from 'vue'
import { 
  ArrowUp, 
  ArrowDown, 
  ShoppingCart, 
  Money, 
  User, 
  TrendCharts,
  Document,
  Box,
  Trophy
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import Chart from 'chart.js/auto'
import { getDashboardStats } from '../api/dashboard'

// 数据
const timeRange = ref('today')
const orderStats = ref({
  total_orders: 0,
  pending_delivery: 0,
  delivering: 0,
  delivered: 0,
  paid: 0,
  cancelled: 0,
  growth: 0
})
const revenueStats = ref({
  total_revenue: 0,
  goods_cost: 0,
  delivery_cost: 0,
  sales_commission: 0,
  net_profit: 0,
  growth: 0
})
const userStats = ref({
  total_users: 0,
  new_users: 0,
  active_users: 0
})
const hotProducts = ref([])
const deliveryRanking = ref([])
const salesRanking = ref([])
const orderTrend = ref([])
const revenueTrend = ref([])

// 图表引用
const orderTrendChart = ref(null)
const revenueTrendChart = ref(null)
let orderChartInstance = null
let revenueChartInstance = null

// 计算属性
const profitRate = computed(() => {
  if (revenueStats.value.total_revenue === 0) return 0
  return (revenueStats.value.net_profit / revenueStats.value.total_revenue) * 100
})

// 加载数据
const loadDashboardData = async () => {
  try {
    const response = await getDashboardStats({ time_range: timeRange.value })
    if (response.code === 200) {
      const data = response.data
      
      orderStats.value = {
        ...data.order_stats,
        growth: data.order_stats.growth || 0
      }
      revenueStats.value = {
        ...data.revenue_stats,
        growth: data.revenue_stats.growth || 0
      }
      userStats.value = data.user_stats
      hotProducts.value = data.hot_products || []
      deliveryRanking.value = data.delivery_ranking || []
      salesRanking.value = data.sales_ranking || []
      orderTrend.value = data.order_trend || []
      revenueTrend.value = data.revenue_trend || []

      // 更新图表
      await nextTick()
      updateCharts()
    } else {
      ElMessage.error(response.message || '获取数据失败')
    }
  } catch (error) {
    console.error('获取仪表盘数据失败:', error)
    ElMessage.error('获取数据失败，请稍后再试')
  }
}

// 更新图表
const updateCharts = () => {
  updateOrderTrendChart()
  updateRevenueTrendChart()
}

// 更新订单趋势图
const updateOrderTrendChart = () => {
  if (!orderTrendChart.value) return

  if (orderChartInstance) {
    orderChartInstance.destroy()
  }

  const labels = orderTrend.value.map(item => item.date)
  const orderCounts = orderTrend.value.map(item => item.order_count)
  const amounts = orderTrend.value.map(item => item.total_amount)

  orderChartInstance = new Chart(orderTrendChart.value, {
    type: 'line',
    data: {
      labels: labels,
      datasets: [
        {
          label: '订单数',
          data: orderCounts,
          borderColor: '#409EFF',
          backgroundColor: 'rgba(64, 158, 255, 0.15)',
          yAxisID: 'y',
          tension: 0.4,
          fill: true,
          borderWidth: 3,
          pointRadius: 5,
          pointHoverRadius: 7,
          pointBackgroundColor: '#fff',
          pointBorderColor: '#409EFF',
          pointBorderWidth: 2,
          pointHoverBackgroundColor: '#409EFF',
          pointHoverBorderColor: '#fff',
          pointHoverBorderWidth: 3
        },
        {
          label: '订单金额',
          data: amounts,
          borderColor: '#67C23A',
          backgroundColor: 'rgba(103, 194, 58, 0.15)',
          yAxisID: 'y1',
          tension: 0.4,
          fill: true,
          borderWidth: 3,
          pointRadius: 5,
          pointHoverRadius: 7,
          pointBackgroundColor: '#fff',
          pointBorderColor: '#67C23A',
          pointBorderWidth: 2,
          pointHoverBackgroundColor: '#67C23A',
          pointHoverBorderColor: '#fff',
          pointHoverBorderWidth: 3
        }
      ]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          display: true,
          position: 'top',
          labels: {
            usePointStyle: true,
            padding: 20,
            font: {
              size: 13,
              weight: '600'
            }
          }
        },
        tooltip: {
          backgroundColor: 'rgba(0, 0, 0, 0.8)',
          padding: 12,
          titleFont: {
            size: 14,
            weight: '600'
          },
          bodyFont: {
            size: 13
          },
          borderColor: 'rgba(255, 255, 255, 0.1)',
          borderWidth: 1,
          cornerRadius: 8,
          displayColors: true,
          callbacks: {
            label: function(context) {
              if (context.datasetIndex === 0) {
                return `订单数: ${context.parsed.y}`
              } else {
                return `订单金额: ¥${context.parsed.y.toLocaleString('zh-CN')}`
              }
            }
          }
        }
      },
      interaction: {
        mode: 'index',
        intersect: false
      },
      scales: {
        x: {
          grid: {
            display: false
          },
          ticks: {
            font: {
              size: 12
            },
            color: '#909399'
          }
        },
        y: {
          type: 'linear',
          display: true,
          position: 'left',
          title: {
            display: true,
            text: '订单数',
            font: {
              size: 13,
              weight: '600'
            },
            color: '#606266'
          },
          grid: {
            color: 'rgba(0, 0, 0, 0.05)'
          },
          ticks: {
            font: {
              size: 12
            },
            color: '#909399'
          }
        },
        y1: {
          type: 'linear',
          display: true,
          position: 'right',
          title: {
            display: true,
            text: '订单金额（元）',
            font: {
              size: 13,
              weight: '600'
            },
            color: '#606266'
          },
          grid: {
            drawOnChartArea: false
          },
          ticks: {
            font: {
              size: 12
            },
            color: '#909399',
            callback: function(value) {
              return '¥' + value.toLocaleString('zh-CN')
            }
          }
        }
      }
    }
  })
}

// 更新收入利润趋势图
const updateRevenueTrendChart = () => {
  if (!revenueTrendChart.value) return

  if (revenueChartInstance) {
    revenueChartInstance.destroy()
  }

  const labels = revenueTrend.value.map(item => item.date)
  const revenues = revenueTrend.value.map(item => item.revenue)
  const profits = revenueTrend.value.map(item => item.profit)
  const netProfits = revenueTrend.value.map(item => item.net_profit)

  revenueChartInstance = new Chart(revenueTrendChart.value, {
    type: 'line',
    data: {
      labels: labels,
      datasets: [
        {
          label: '收入',
          data: revenues,
          borderColor: '#409EFF',
          backgroundColor: 'rgba(64, 158, 255, 0.15)',
          tension: 0.4,
          fill: true,
          borderWidth: 3,
          pointRadius: 5,
          pointHoverRadius: 7,
          pointBackgroundColor: '#fff',
          pointBorderColor: '#409EFF',
          pointBorderWidth: 2,
          pointHoverBackgroundColor: '#409EFF',
          pointHoverBorderColor: '#fff',
          pointHoverBorderWidth: 3
        },
        {
          label: '利润',
          data: profits,
          borderColor: '#67C23A',
          backgroundColor: 'rgba(103, 194, 58, 0.15)',
          tension: 0.4,
          fill: true,
          borderWidth: 3,
          pointRadius: 5,
          pointHoverRadius: 7,
          pointBackgroundColor: '#fff',
          pointBorderColor: '#67C23A',
          pointBorderWidth: 2,
          pointHoverBackgroundColor: '#67C23A',
          pointHoverBorderColor: '#fff',
          pointHoverBorderWidth: 3
        },
        {
          label: '净利润',
          data: netProfits,
          borderColor: '#E6A23C',
          backgroundColor: 'rgba(230, 162, 60, 0.15)',
          tension: 0.4,
          fill: true,
          borderWidth: 3,
          pointRadius: 5,
          pointHoverRadius: 7,
          pointBackgroundColor: '#fff',
          pointBorderColor: '#E6A23C',
          pointBorderWidth: 2,
          pointHoverBackgroundColor: '#E6A23C',
          pointHoverBorderColor: '#fff',
          pointHoverBorderWidth: 3
        }
      ]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          display: true,
          position: 'top',
          labels: {
            usePointStyle: true,
            padding: 20,
            font: {
              size: 13,
              weight: '600'
            }
          }
        },
        tooltip: {
          backgroundColor: 'rgba(0, 0, 0, 0.8)',
          padding: 12,
          titleFont: {
            size: 14,
            weight: '600'
          },
          bodyFont: {
            size: 13
          },
          borderColor: 'rgba(255, 255, 255, 0.1)',
          borderWidth: 1,
          cornerRadius: 8,
          displayColors: true,
          callbacks: {
            label: function(context) {
              return `${context.dataset.label}: ¥${context.parsed.y.toLocaleString('zh-CN')}`
            }
          }
        }
      },
      interaction: {
        mode: 'index',
        intersect: false
      },
      scales: {
        x: {
          grid: {
            display: false
          },
          ticks: {
            font: {
              size: 12
            },
            color: '#909399'
          }
        },
        y: {
          beginAtZero: true,
          grid: {
            color: 'rgba(0, 0, 0, 0.05)'
          },
          ticks: {
            font: {
              size: 12
            },
            color: '#909399',
            callback: function(value) {
              return '¥' + value.toLocaleString('zh-CN')
            }
          }
        }
      }
    }
  })
}

// 时间范围改变
const handleTimeRangeChange = () => {
  loadDashboardData()
}

// 格式化函数
const formatNumber = (num) => {
  if (num === null || num === undefined) return '0'
  return Number(num).toLocaleString('zh-CN')
}

const formatMoney = (num) => {
  if (num === null || num === undefined) return '0.00'
  return Number(num).toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
}

const formatPercent = (num) => {
  if (num === null || num === undefined) return '0.00'
  return Number(num).toFixed(2)
}

// 生命周期
onMounted(() => {
  loadDashboardData()
})

onUnmounted(() => {
  if (orderChartInstance) {
    orderChartInstance.destroy()
  }
  if (revenueChartInstance) {
    revenueChartInstance.destroy()
  }
})
</script>

<style scoped>
.dashboard-container {
  padding: 24px;
  background: linear-gradient(135deg, #f5f7fa 0%, #e9ecef 100%);
  min-height: calc(100vh - 60px);
}

.dashboard-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding: 20px 24px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.04);
}

.page-title {
  font-size: 28px;
  font-weight: 700;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  margin: 0;
}

.time-range-selector :deep(.el-radio-group) {
  display: flex;
  gap: 8px;
}

.time-range-selector :deep(.el-radio-button__inner) {
  border-radius: 8px;
  padding: 8px 20px;
  font-weight: 500;
  transition: all 0.3s ease;
}

.time-range-selector :deep(.el-radio-button__orig-radio:checked + .el-radio-button__inner) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-color: transparent;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

/* 统计卡片样式 */
.stats-cards {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  margin-bottom: 24px;
}

.stat-card {
  border-radius: 16px;
  border: none;
  overflow: hidden;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
}

.stat-card:hover {
  transform: translateY(-4px);
}

.stat-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.8), transparent);
}

.stat-card-orders::before {
  background: linear-gradient(90deg, #409EFF, #66b1ff);
}

.stat-card-revenue::before {
  background: linear-gradient(90deg, #67C23A, #85ce61);
}

.stat-card-profit::before {
  background: linear-gradient(90deg, #E6A23C, #ebb563);
}

.stat-card-users::before {
  background: linear-gradient(90deg, #F56C6C, #f78989);
}

.stat-content {
  display: flex;
  align-items: center;
  padding: 20px;
  gap: 16px;
}

.stat-icon-wrapper {
  width: 64px;
  height: 64px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  position: relative;
  overflow: hidden;
}

.stat-icon-wrapper::before {
  content: '';
  position: absolute;
  inset: 0;
  opacity: 0.1;
  background: inherit;
}

.stat-icon-wrapper.orders {
  background: linear-gradient(135deg, #409EFF 0%, #66b1ff 100%);
}

.stat-icon-wrapper.revenue {
  background: linear-gradient(135deg, #67C23A 0%, #85ce61 100%);
}

.stat-icon-wrapper.profit {
  background: linear-gradient(135deg, #E6A23C 0%, #ebb563 100%);
}

.stat-icon-wrapper.users {
  background: linear-gradient(135deg, #F56C6C 0%, #f78989 100%);
}

.stat-icon {
  font-size: 32px;
  color: #fff;
  z-index: 1;
}

.stat-info {
  flex: 1;
  min-width: 0;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-bottom: 8px;
  font-weight: 500;
}

.stat-value {
  font-size: 32px;
  font-weight: 700;
  color: #303133;
  margin-bottom: 8px;
  line-height: 1.2;
  word-break: break-all;
}

.stat-growth {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  font-weight: 600;
}

.stat-growth.positive {
  color: #67C23A;
}

.stat-growth.negative {
  color: #F56C6C;
}

.growth-label {
  font-size: 12px;
  color: #909399;
  font-weight: 400;
  margin-left: 4px;
}

.stat-extra {
  font-size: 13px;
  color: #909399;
}

.profit-rate,
.total-users {
  font-weight: 500;
}

.dashboard-row {
  margin-bottom: 24px;
}

/* 卡片通用样式 */
.info-card,
.chart-card,
.ranking-card {
  border-radius: 16px;
  border: none;
  transition: all 0.3s ease;
}

.info-card:hover,
.chart-card:hover,
.ranking-card:hover {
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.header-icon {
  font-size: 20px;
  color: #409EFF;
}

/* 订单状态列表 */
.order-status-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.status-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  background: #f8f9fa;
  border-radius: 12px;
  transition: all 0.3s ease;
  border-left: 4px solid transparent;
}

.status-item:hover {
  background: #f0f2f5;
  transform: translateX(4px);
}

.status-item.status-pending {
  border-left-color: #E6A23C;
}

.status-item.status-delivering {
  border-left-color: #409EFF;
}

.status-item.status-delivered {
  border-left-color: #67C23A;
}

.status-item.status-paid {
  border-left-color: #909399;
}

.status-item.status-cancelled {
  border-left-color: #F56C6C;
}

.status-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.status-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-pending .status-dot {
  background: #E6A23C;
}

.status-delivering .status-dot {
  background: #409EFF;
}

.status-delivered .status-dot {
  background: #67C23A;
}

.status-paid .status-dot {
  background: #909399;
}

.status-cancelled .status-dot {
  background: #F56C6C;
}

.status-label {
  font-size: 15px;
  color: #606266;
  font-weight: 500;
}

.status-value {
  font-size: 20px;
  font-weight: 700;
  color: #303133;
}

/* 收入成本列表 */
.revenue-cost-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.revenue-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  background: #f8f9fa;
  border-radius: 12px;
  transition: all 0.3s ease;
}

.revenue-item:hover {
  background: #f0f2f5;
}

.revenue-item.revenue-income {
  background: linear-gradient(135deg, rgba(103, 194, 58, 0.1) 0%, rgba(103, 194, 58, 0.05) 100%);
}

.revenue-item.revenue-total {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
  border: 2px solid rgba(102, 126, 234, 0.2);
  padding: 20px 16px;
  margin-top: 8px;
}

.revenue-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.revenue-icon {
  font-size: 20px;
}

.revenue-label {
  font-size: 15px;
  color: #606266;
  font-weight: 500;
}

.revenue-value {
  font-size: 18px;
  font-weight: 700;
  color: #303133;
}

.revenue-value.positive {
  color: #67C23A;
  font-size: 20px;
}

.revenue-total .revenue-value.positive {
  font-size: 24px;
}

/* 图表容器 */
.chart-container {
  height: 350px;
  position: relative;
  padding: 16px;
}

/* 排名表格 */
.ranking-table :deep(.el-table__header) {
  background: #f8f9fa;
}

.ranking-table :deep(.el-table__header th) {
  background: transparent;
  color: #606266;
  font-weight: 600;
  border-bottom: 2px solid #e4e7ed;
}

.ranking-table :deep(.el-table__row:hover) {
  background: #f0f7ff;
}

.ranking-table :deep(.el-table__row:nth-child(1)) {
  background: linear-gradient(135deg, rgba(255, 215, 0, 0.1) 0%, rgba(255, 215, 0, 0.05) 100%);
}

.ranking-table :deep(.el-table__row:nth-child(2)) {
  background: linear-gradient(135deg, rgba(192, 192, 192, 0.1) 0%, rgba(192, 192, 192, 0.05) 100%);
}

.ranking-table :deep(.el-table__row:nth-child(3)) {
  background: linear-gradient(135deg, rgba(205, 127, 50, 0.1) 0%, rgba(205, 127, 50, 0.05) 100%);
}

/* 响应式设计 */
@media (max-width: 1400px) {
  .stats-cards {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .dashboard-container {
    padding: 16px;
  }

  .stats-cards {
    grid-template-columns: 1fr;
  }
  
  .dashboard-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
    padding: 16px;
  }

  .page-title {
    font-size: 24px;
  }

  .stat-content {
    padding: 16px;
  }

  .stat-icon-wrapper {
    width: 56px;
    height: 56px;
  }

  .stat-icon {
    font-size: 28px;
  }

  .stat-value {
    font-size: 28px;
  }

  .chart-container {
    height: 280px;
  }
}

/* 动画效果 */
@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.stat-card {
  animation: fadeInUp 0.6s ease-out;
}

.stat-card:nth-child(1) {
  animation-delay: 0.1s;
}

.stat-card:nth-child(2) {
  animation-delay: 0.2s;
}

.stat-card:nth-child(3) {
  animation-delay: 0.3s;
}

.stat-card:nth-child(4) {
  animation-delay: 0.4s;
}
</style>
