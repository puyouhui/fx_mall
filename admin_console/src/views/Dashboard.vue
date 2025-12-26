<template>
  <div class="dashboard-container">
    <div class="dashboard-header">
      <div class="header-left">
        <h1 class="page-title">销售额数据</h1>
        <p class="page-subtitle">销售汇总</p>
      </div>
      <div class="header-right">
        <!-- <el-button type="primary" class="export-btn">
          <el-icon><Download /></el-icon>
          <span>导出</span>
        </el-button> -->
        <div class="time-range-selector">
          <el-radio-group v-model="timeRange" @change="handleTimeRangeChange">
            <el-radio-button label="today">今日</el-radio-button>
            <el-radio-button label="week">本周</el-radio-button>
            <el-radio-button label="month">本月</el-radio-button>
          </el-radio-group>
        </div>
      </div>
    </div>

    <!-- 顶部核心指标卡片 -->
    <div class="stats-cards">
      <div class="stat-card stat-card-orders">
        <div class="stat-content">
          <div class="stat-icon-wrapper orders">
            <el-icon class="stat-icon"><ShoppingCart /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-label">总销量</div>
            <div class="stat-value">¥{{ formatMoney(orderStats.total_amount) }}</div>
            <div class="stat-growth-wrapper">
              <div class="stat-growth" :class="revenueStats.growth >= 0 ? 'positive' : 'negative'">
                <el-icon v-if="revenueStats.growth >= 0"><ArrowUp /></el-icon>
                <el-icon v-else><ArrowDown /></el-icon>
                <span class="growth-percent">{{ formatPercent(Math.abs(revenueStats.growth)) }}%</span>
              </div>
              <span class="growth-label">{{ growthLabel }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="stat-card stat-card-revenue">
        <div class="stat-content">
          <div class="stat-icon-wrapper revenue">
            <el-icon class="stat-icon"><Document /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-label">总订单</div>
            <div class="stat-value">{{ formatNumber(orderStats.total_orders) }}</div>
            <div class="stat-growth-wrapper">
              <div class="stat-growth" :class="orderStats.growth >= 0 ? 'positive' : 'negative'">
                <el-icon v-if="orderStats.growth >= 0"><ArrowUp /></el-icon>
                <el-icon v-else><ArrowDown /></el-icon>
                <span class="growth-percent">{{ formatPercent(Math.abs(orderStats.growth)) }}%</span>
              </div>
              <span class="growth-label">{{ growthLabel }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="stat-card stat-card-profit">
        <div class="stat-content">
          <div class="stat-icon-wrapper profit">
            <el-icon class="stat-icon"><Box /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-label">销售的产品</div>
            <div class="stat-value">{{ formatNumber(hotProducts.length) }}</div>
            <div class="stat-growth-wrapper" v-if="productGrowth !== null">
              <div class="stat-growth" :class="productGrowth >= 0 ? 'positive' : 'negative'">
                <el-icon v-if="productGrowth >= 0"><ArrowUp /></el-icon>
                <el-icon v-else><ArrowDown /></el-icon>
                <span class="growth-percent">{{ formatPercent(Math.abs(productGrowth)) }}%</span>
              </div>
              <span class="growth-label">{{ growthLabel }}</span>
            </div>
            <div class="stat-growth-wrapper" v-else>
              <span class="no-data-text">暂无对比数据</span>
            </div>
          </div>
        </div>
      </div>

      <div class="stat-card stat-card-users">
        <div class="stat-content">
          <div class="stat-icon-wrapper users">
            <el-icon class="stat-icon"><User /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-label">新客户</div>
            <div class="stat-value">{{ formatNumber(userStats.new_users) }}</div>
            <div class="stat-growth-wrapper" v-if="userGrowth !== null">
              <div class="stat-growth" :class="userGrowth >= 0 ? 'positive' : 'negative'">
                <el-icon v-if="userGrowth >= 0"><ArrowUp /></el-icon>
                <el-icon v-else><ArrowDown /></el-icon>
                <span class="growth-percent">{{ formatPercent(Math.abs(userGrowth)) }}%</span>
              </div>
              <span class="growth-label">{{ growthLabel }}</span>
            </div>
            <div class="stat-growth-wrapper" v-else>
              <span class="no-data-text">暂无对比数据</span>
            </div>
          </div>
        </div>
      </div>
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
                <div class="status-icon-wrapper pending">
                  <el-icon class="status-icon"><Clock /></el-icon>
                </div>
                <div class="status-text">
                  <span class="status-label">待配送</span>
                  <span class="status-value">{{ formatNumber(orderStats.pending_delivery) }}</span>
                </div>
              </div>
              <div class="status-percent">{{ getStatusPercent('pending') }}%</div>
            </div>
            <div class="status-item status-delivering">
              <div class="status-info">
                <div class="status-icon-wrapper delivering">
                  <el-icon class="status-icon"><Van /></el-icon>
                </div>
                <div class="status-text">
                  <span class="status-label">配送中</span>
                  <span class="status-value">{{ formatNumber(orderStats.delivering) }}</span>
                </div>
              </div>
              <div class="status-percent">{{ getStatusPercent('delivering') }}%</div>
            </div>
            <div class="status-item status-delivered">
              <div class="status-info">
                <div class="status-icon-wrapper delivered">
                  <el-icon class="status-icon"><CircleCheck /></el-icon>
                </div>
                <div class="status-text">
                  <span class="status-label">已送达</span>
                  <span class="status-value">{{ formatNumber(orderStats.delivered) }}</span>
                </div>
              </div>
              <div class="status-percent">{{ getStatusPercent('delivered') }}%</div>
            </div>
            <div class="status-item status-paid">
              <div class="status-info">
                <div class="status-icon-wrapper paid">
                  <el-icon class="status-icon"><Money /></el-icon>
                </div>
                <div class="status-text">
                  <span class="status-label">已收款</span>
                  <span class="status-value">{{ formatNumber(orderStats.paid) }}</span>
                </div>
              </div>
              <div class="status-percent">{{ getStatusPercent('paid') }}%</div>
            </div>
            <div class="status-item status-cancelled">
              <div class="status-info">
                <div class="status-icon-wrapper cancelled">
                  <el-icon class="status-icon"><Close /></el-icon>
                </div>
                <div class="status-text">
                  <span class="status-label">已取消</span>
                  <span class="status-value">{{ formatNumber(orderStats.cancelled) }}</span>
                </div>
              </div>
              <div class="status-percent">{{ getStatusPercent('cancelled') }}%</div>
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
                <div class="revenue-icon-wrapper income">
                  <el-icon class="revenue-icon"><TrendCharts /></el-icon>
                </div>
                <div class="revenue-text">
                  <span class="revenue-label">总收入</span>
                  <span class="revenue-value positive">¥{{ formatMoney(revenueStats.total_revenue) }}</span>
                </div>
              </div>
              <div class="revenue-trend positive">
                <el-icon><ArrowUp /></el-icon>
              </div>
            </div>
            <div class="revenue-item revenue-cost">
              <div class="revenue-info">
                <div class="revenue-icon-wrapper cost">
                  <el-icon class="revenue-icon"><Box /></el-icon>
                </div>
                <div class="revenue-text">
                  <span class="revenue-label">商品成本</span>
                  <span class="revenue-value">¥{{ formatMoney(revenueStats.goods_cost) }}</span>
                </div>
              </div>
              <div class="revenue-percent">{{ getCostPercent('goods') }}%</div>
            </div>
            <div class="revenue-item revenue-cost">
              <div class="revenue-info">
                <div class="revenue-icon-wrapper cost">
                  <el-icon class="revenue-icon"><Van /></el-icon>
                </div>
                <div class="revenue-text">
                  <span class="revenue-label">配送成本</span>
                  <span class="revenue-value">¥{{ formatMoney(revenueStats.delivery_cost) }}</span>
                </div>
              </div>
              <div class="revenue-percent">{{ getCostPercent('delivery') }}%</div>
            </div>
            <div class="revenue-item revenue-cost">
              <div class="revenue-info">
                <div class="revenue-icon-wrapper cost">
                  <el-icon class="revenue-icon"><Money /></el-icon>
                </div>
                <div class="revenue-text">
                  <span class="revenue-label">销售分成</span>
                  <span class="revenue-value">¥{{ formatMoney(revenueStats.sales_commission) }}</span>
                </div>
              </div>
              <div class="revenue-percent">{{ getCostPercent('commission') }}%</div>
            </div>
            <div class="revenue-item revenue-total">
              <div class="revenue-info">
                <div class="revenue-icon-wrapper total">
                  <el-icon class="revenue-icon"><Star /></el-icon>
                </div>
                <div class="revenue-text">
                  <span class="revenue-label">净利润</span>
                  <span class="revenue-value positive">¥{{ formatMoney(revenueStats.net_profit) }}</span>
                </div>
              </div>
              <div class="revenue-trend positive">
                <el-icon><ArrowUp /></el-icon>
                <span class="profit-rate-text">{{ formatPercent(profitRate) }}%</span>
              </div>
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
  Trophy,
  Clock,
  Van,
  CircleCheck,
  Close,
  Star
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

// 时间范围标签
const timeRangeLabel = computed(() => {
  const labelMap = {
    'today': '今日',
    'week': '本周',
    'month': '本月'
  }
  return labelMap[timeRange.value] || '今日'
})

// 增长提示标签
const growthLabel = computed(() => {
  const labelMap = {
    'today': '比昨天',
    'week': '比上周',
    'month': '比上月'
  }
  return labelMap[timeRange.value] || '比昨天'
})

// 产品增长（暂时设为null，因为后端没有提供）
const productGrowth = ref(null)

// 用户增长（暂时设为null，因为后端没有提供）
const userGrowth = ref(null)

// 加载数据
const loadDashboardData = async () => {
  try {
    const response = await getDashboardStats({ time_range: timeRange.value })
    if (response.code === 200) {
      const data = response.data
      
      // 订单统计
      orderStats.value = {
        total_orders: data.order_stats?.total_orders || 0,
        pending_delivery: data.order_stats?.pending_delivery || 0,
        delivering: data.order_stats?.delivering || 0,
        delivered: data.order_stats?.delivered || 0,
        paid: data.order_stats?.paid || 0,
        cancelled: data.order_stats?.cancelled || 0,
        growth: data.order_stats?.growth || 0,
        total_amount: data.order_stats?.total_amount || 0
      }
      
      // 收入统计
      revenueStats.value = {
        total_revenue: data.revenue_stats?.total_revenue || 0,
        goods_cost: data.revenue_stats?.goods_cost || 0,
        delivery_cost: data.revenue_stats?.delivery_cost || 0,
        sales_commission: data.revenue_stats?.sales_commission || 0,
        net_profit: data.revenue_stats?.net_profit || 0,
        growth: data.revenue_stats?.growth || 0
      }
      
      // 用户统计
      userStats.value = {
        total_users: data.user_stats?.total_users || 0,
        new_users: data.user_stats?.new_users || 0,
        active_users: data.user_stats?.active_users || 0
      }
      
      // 计算用户增长（如果有对比数据）
      if (data.user_stats?.growth !== undefined && data.user_stats?.growth !== null) {
        userGrowth.value = data.user_stats.growth
      } else {
        userGrowth.value = null
      }
      
      // 产品增长（暂时设为null，因为后端没有提供产品增长数据）
      productGrowth.value = null
      
      // 热销商品
      hotProducts.value = (data.hot_products || []).map(item => ({
        product_id: item.product_id,
        product_name: item.product_name || '未知商品',
        image: item.image || '',
        total_quantity: item.total_quantity || 0,
        total_amount: item.total_amount || 0
      }))
      
      // 配送员排名
      deliveryRanking.value = (data.delivery_ranking || []).map(item => ({
        employee_code: item.employee_code || '',
        employee_name: item.employee_name || '未知配送员',
        order_count: item.order_count || 0,
        total_fee: item.total_fee || 0
      }))
      
      // 销售员排名
      salesRanking.value = (data.sales_ranking || []).map(item => ({
        employee_code: item.employee_code || '',
        employee_name: item.employee_name || '未知销售员',
        order_count: item.order_count || 0,
        total_sales: item.total_sales || 0,
        total_commission: item.total_commission || 0,
        new_customer_count: item.new_customer_count || 0
      }))
      
      // 订单趋势
      orderTrend.value = (data.order_trend || []).map(item => ({
        date: item.date || '',
        order_count: item.order_count || 0,
        total_amount: item.total_amount || 0
      }))
      
      // 收入趋势
      revenueTrend.value = (data.revenue_trend || []).map(item => ({
        date: item.date || '',
        revenue: item.revenue || 0,
        profit: item.profit || 0,
        net_profit: item.net_profit || 0
      }))

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

// 生成完整日期范围
const generateDateRange = () => {
  const dates = []
  const now = new Date()
  let startDate, days
  
  if (timeRange.value === 'today') {
    // 今天的24小时
    for (let hour = 0; hour < 24; hour++) {
      dates.push(`${hour.toString().padStart(2, '0')}:00`)
    }
  } else if (timeRange.value === 'week') {
    // 最近7天
    days = 7
    for (let i = days - 1; i >= 0; i--) {
      const date = new Date(now)
      date.setDate(date.getDate() - i)
      dates.push(date.toISOString().split('T')[0])
    }
  } else if (timeRange.value === 'month') {
    // 最近30天
    days = 30
    for (let i = days - 1; i >= 0; i--) {
      const date = new Date(now)
      date.setDate(date.getDate() - i)
      dates.push(date.toISOString().split('T')[0])
    }
  }
  
  return dates
}

// 填充缺失日期的数据
const fillMissingDates = (data, dateField, valueFields) => {
  const allDates = generateDateRange()
  const dataMap = new Map(data.map(item => [item[dateField], item]))
  
  return allDates.map(date => {
    if (dataMap.has(date)) {
      return dataMap.get(date)
    } else {
      const emptyData = { [dateField]: date }
      valueFields.forEach(field => {
        emptyData[field] = 0
      })
      return emptyData
    }
  })
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

  // 填充缺失日期的数据
  const filledData = fillMissingDates(orderTrend.value, 'date', ['order_count', 'total_amount'])
  
  const labels = filledData.map(item => item.date)
  const orderCounts = filledData.map(item => item.order_count)
  const amounts = filledData.map(item => item.total_amount)

  orderChartInstance = new Chart(orderTrendChart.value, {
    type: 'line',
    data: {
      labels: labels,
      datasets: [
        {
          label: '订单数',
          data: orderCounts,
          borderColor: '#667eea',
          backgroundColor: function(context) {
            const chart = context.chart;
            const {ctx, chartArea} = chart;
            if (!chartArea) return 'rgba(102, 126, 234, 0.1)';
            const gradient = ctx.createLinearGradient(0, chartArea.top, 0, chartArea.bottom);
            gradient.addColorStop(0, 'rgba(102, 126, 234, 0.5)');
            gradient.addColorStop(1, 'rgba(102, 126, 234, 0.05)');
            return gradient;
          },
          yAxisID: 'y',
          tension: 0.4,
          fill: true,
          borderWidth: 3,
          pointRadius: 6,
          pointHoverRadius: 8,
          pointBackgroundColor: '#fff',
          pointBorderColor: '#667eea',
          pointBorderWidth: 3,
          pointHoverBackgroundColor: '#667eea',
          pointHoverBorderColor: '#fff',
          pointHoverBorderWidth: 3
        },
        {
          label: '订单金额',
          data: amounts,
          borderColor: '#10b981',
          backgroundColor: function(context) {
            const chart = context.chart;
            const {ctx, chartArea} = chart;
            if (!chartArea) return 'rgba(16, 185, 129, 0.1)';
            const gradient = ctx.createLinearGradient(0, chartArea.top, 0, chartArea.bottom);
            gradient.addColorStop(0, 'rgba(16, 185, 129, 0.5)');
            gradient.addColorStop(1, 'rgba(16, 185, 129, 0.05)');
            return gradient;
          },
          yAxisID: 'y1',
          tension: 0.4,
          fill: true,
          borderWidth: 3,
          pointRadius: 6,
          pointHoverRadius: 8,
          pointBackgroundColor: '#fff',
          pointBorderColor: '#10b981',
          pointBorderWidth: 3,
          pointHoverBackgroundColor: '#10b981',
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
          align: 'end',
          labels: {
            usePointStyle: true,
            pointStyle: 'circle',
            padding: 16,
            font: {
              size: 14,
              weight: '700',
              family: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif'
            },
            color: '#374151',
            boxWidth: 10,
            boxHeight: 10
          }
        },
        tooltip: {
          backgroundColor: 'rgba(17, 24, 39, 0.95)',
          padding: 16,
          titleFont: {
            size: 15,
            weight: '700',
            family: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif'
          },
          bodyFont: {
            size: 14,
            weight: '500'
          },
          borderColor: 'rgba(255, 255, 255, 0.15)',
          borderWidth: 1,
          cornerRadius: 12,
          displayColors: true,
          boxPadding: 8,
          usePointStyle: true,
          callbacks: {
            label: function(context) {
              if (context.datasetIndex === 0) {
                return ` 订单数: ${context.parsed.y}`
              } else {
                return ` 订单金额: ¥${context.parsed.y.toLocaleString('zh-CN')}`
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
            display: false,
            drawBorder: false
          },
          ticks: {
            font: {
              size: 13,
              weight: '600'
            },
            color: '#6B7280',
            padding: 8
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
              size: 14,
              weight: '700'
            },
            color: '#667eea',
            padding: { bottom: 8 }
          },
          grid: {
            color: 'rgba(0, 0, 0, 0.06)',
            drawBorder: false,
            lineWidth: 1
          },
          ticks: {
            font: {
              size: 13,
              weight: '600'
            },
            color: '#6B7280',
            padding: 8
          }
        },
        y1: {
          type: 'linear',
          display: true,
          position: 'right',
          title: {
            display: true,
            text: '订单金额 (¥)',
            font: {
              size: 14,
              weight: '700'
            },
            color: '#10b981',
            padding: { bottom: 8 }
          },
          grid: {
            drawOnChartArea: false,
            drawBorder: false
          },
          ticks: {
            font: {
              size: 13,
              weight: '600'
            },
            color: '#6B7280',
            padding: 8,
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

  // 填充缺失日期的数据
  const filledData = fillMissingDates(revenueTrend.value, 'date', ['revenue', 'profit', 'net_profit'])
  
  const labels = filledData.map(item => item.date)
  const revenues = filledData.map(item => item.revenue)
  const profits = filledData.map(item => item.profit)
  const netProfits = filledData.map(item => item.net_profit)

  revenueChartInstance = new Chart(revenueTrendChart.value, {
    type: 'line',
    data: {
      labels: labels,
      datasets: [
        {
          label: '收入',
          data: revenues,
          borderColor: '#3b82f6',
          backgroundColor: function(context) {
            const chart = context.chart;
            const {ctx, chartArea} = chart;
            if (!chartArea) return 'rgba(59, 130, 246, 0.1)';
            const gradient = ctx.createLinearGradient(0, chartArea.top, 0, chartArea.bottom);
            gradient.addColorStop(0, 'rgba(59, 130, 246, 0.5)');
            gradient.addColorStop(1, 'rgba(59, 130, 246, 0.05)');
            return gradient;
          },
          tension: 0.4,
          fill: true,
          borderWidth: 3,
          pointRadius: 6,
          pointHoverRadius: 8,
          pointBackgroundColor: '#fff',
          pointBorderColor: '#3b82f6',
          pointBorderWidth: 3,
          pointHoverBackgroundColor: '#3b82f6',
          pointHoverBorderColor: '#fff',
          pointHoverBorderWidth: 3
        },
        {
          label: '利润',
          data: profits,
          borderColor: '#10b981',
          backgroundColor: function(context) {
            const chart = context.chart;
            const {ctx, chartArea} = chart;
            if (!chartArea) return 'rgba(16, 185, 129, 0.1)';
            const gradient = ctx.createLinearGradient(0, chartArea.top, 0, chartArea.bottom);
            gradient.addColorStop(0, 'rgba(16, 185, 129, 0.5)');
            gradient.addColorStop(1, 'rgba(16, 185, 129, 0.05)');
            return gradient;
          },
          tension: 0.4,
          fill: true,
          borderWidth: 3,
          pointRadius: 6,
          pointHoverRadius: 8,
          pointBackgroundColor: '#fff',
          pointBorderColor: '#10b981',
          pointBorderWidth: 3,
          pointHoverBackgroundColor: '#10b981',
          pointHoverBorderColor: '#fff',
          pointHoverBorderWidth: 3
        },
        {
          label: '净利润',
          data: netProfits,
          borderColor: '#f59e0b',
          backgroundColor: function(context) {
            const chart = context.chart;
            const {ctx, chartArea} = chart;
            if (!chartArea) return 'rgba(245, 158, 11, 0.1)';
            const gradient = ctx.createLinearGradient(0, chartArea.top, 0, chartArea.bottom);
            gradient.addColorStop(0, 'rgba(245, 158, 11, 0.5)');
            gradient.addColorStop(1, 'rgba(245, 158, 11, 0.05)');
            return gradient;
          },
          tension: 0.4,
          fill: true,
          borderWidth: 3,
          pointRadius: 6,
          pointHoverRadius: 8,
          pointBackgroundColor: '#fff',
          pointBorderColor: '#f59e0b',
          pointBorderWidth: 3,
          pointHoverBackgroundColor: '#f59e0b',
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
          align: 'end',
          labels: {
            usePointStyle: true,
            pointStyle: 'circle',
            padding: 16,
            font: {
              size: 14,
              weight: '700',
              family: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif'
            },
            color: '#374151',
            boxWidth: 10,
            boxHeight: 10
          }
        },
        tooltip: {
          backgroundColor: 'rgba(17, 24, 39, 0.95)',
          padding: 16,
          titleFont: {
            size: 15,
            weight: '700',
            family: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif'
          },
          bodyFont: {
            size: 14,
            weight: '500'
          },
          borderColor: 'rgba(255, 255, 255, 0.15)',
          borderWidth: 1,
          cornerRadius: 12,
          displayColors: true,
          boxPadding: 8,
          usePointStyle: true,
          callbacks: {
            label: function(context) {
              return ` ${context.dataset.label}: ¥${context.parsed.y.toLocaleString('zh-CN')}`
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
            display: false,
            drawBorder: false
          },
          ticks: {
            font: {
              size: 13,
              weight: '600'
            },
            color: '#6B7280',
            padding: 8
          }
        },
        y: {
          beginAtZero: true,
          grid: {
            color: 'rgba(0, 0, 0, 0.06)',
            drawBorder: false,
            lineWidth: 1
          },
          ticks: {
            font: {
              size: 13,
              weight: '600'
            },
            color: '#6B7280',
            padding: 8,
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

// 计算订单状态百分比
const getStatusPercent = (status) => {
  const total = orderStats.value.total_orders || 1
  const statusMap = {
    'pending': orderStats.value.pending_delivery || 0,
    'delivering': orderStats.value.delivering || 0,
    'delivered': orderStats.value.delivered || 0,
    'paid': orderStats.value.paid || 0,
    'cancelled': orderStats.value.cancelled || 0
  }
  return ((statusMap[status] / total) * 100).toFixed(1)
}

// 计算成本占比
const getCostPercent = (type) => {
  const total = revenueStats.value.total_revenue || 1
  const costMap = {
    'goods': revenueStats.value.goods_cost || 0,
    'delivery': revenueStats.value.delivery_cost || 0,
    'commission': revenueStats.value.sales_commission || 0
  }
  return ((costMap[type] / total) * 100).toFixed(1)
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
  align-items: flex-start;
  margin-bottom: 24px;
  padding: 24px 28px;
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.04);
}

.header-left {
  flex: 1;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
  margin: 0 0 8px 0;
}

.page-subtitle {
  font-size: 14px;
  color: #909399;
  margin: 0;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.export-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 20px;
  border-radius: 8px;
  font-weight: 500;
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
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  cursor: pointer;
  background: #fff;
}

.stat-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  border-radius: 16px;
  padding: 2px;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.5), rgba(255, 255, 255, 0.1));
  -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
  -webkit-mask-composite: xor;
  mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
  mask-composite: exclude;
  pointer-events: none;
}

.stat-card:hover {
  transform: translateY(-6px);
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.12);
}

.stat-card-orders {
  background: linear-gradient(135deg, #10b981 0%, #34d399 50%, #6ee7b7 100%);
}

.stat-card-revenue {
  background: linear-gradient(135deg, #3b82f6 0%, #60a5fa 50%, #93c5fd 100%);
}

.stat-card-profit {
  background: linear-gradient(135deg, #14b8a6 0%, #2dd4bf 50%, #5eead4 100%);
}

.stat-card-users {
  background: linear-gradient(135deg, #8b5cf6 0%, #a78bfa 50%, #c4b5fd 100%);
}

.stat-content {
  display: flex;
  align-items: flex-start;
  padding: 24px;
  gap: 16px;
}

.stat-icon-wrapper {
  width: 56px;
  height: 56px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  position: relative;
  overflow: hidden;
  background: rgba(255, 255, 255, 0.25);
  backdrop-filter: blur(10px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1), inset 0 1px 0 rgba(255, 255, 255, 0.5);
}

.stat-card-orders .stat-icon-wrapper {
  background: rgba(255, 255, 255, 0.25);
}

.stat-card-revenue .stat-icon-wrapper {
  background: rgba(255, 255, 255, 0.25);
}

.stat-card-profit .stat-icon-wrapper {
  background: rgba(255, 255, 255, 0.25);
}

.stat-card-users .stat-icon-wrapper {
  background: rgba(255, 255, 255, 0.25);
}

.stat-icon {
  font-size: 28px;
  color: #ffffff;
  z-index: 1;
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.15));
}

.stat-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.stat-label {
  font-size: 14px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.95);
  letter-spacing: 0.3px;
  line-height: 1.4;
  text-transform: uppercase;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}

.stat-value {
  font-size: 32px;
  font-weight: 800;
  line-height: 1.2;
  word-break: break-all;
  color: #ffffff;
  letter-spacing: -0.5px;
  margin: 2px 0;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.15);
}

.stat-growth-wrapper {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 8px;
  margin-top: 4px;
}

.stat-growth {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  font-weight: 700;
  padding: 6px 12px;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(10px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.12);
}

.stat-growth.positive {
  color: #059669;
}

.stat-growth.negative {
  color: #DC2626;
}

.stat-growth .el-icon {
  font-size: 16px;
  font-weight: bold;
}

.growth-percent {
  font-weight: 800;
  letter-spacing: -0.3px;
}

.growth-label {
  font-size: 12px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.85);
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}

.no-data-text {
  font-size: 12px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.75);
  padding: 6px 12px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 8px;
  width: fit-content;
  backdrop-filter: blur(10px);
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
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
.info-card {
  border-radius: 16px;
  border: none;
  transition: all 0.3s ease;
  background: #fff;
}

.chart-card {
  border-radius: 20px;
  border: none;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  box-shadow: 0 8px 20px rgba(102, 126, 234, 0.3);
}

.chart-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 32px rgba(102, 126, 234, 0.4);
}

.chart-card :deep(.el-card__header) {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  border: none;
  padding: 20px 24px;
}

.chart-card :deep(.el-card__body) {
  background: rgba(255, 255, 255, 0.95);
  padding: 24px;
}

.ranking-card {
  border-radius: 16px;
  border: none;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  background: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.info-card:hover,
.ranking-card:hover {
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
  transform: translateY(-2px);
}

.chart-card .card-header {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 20px;
  font-weight: 700;
  color: #ffffff;
  letter-spacing: 0.5px;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.info-card .card-header,
.ranking-card .card-header {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 18px;
  font-weight: 700;
  color: #1f2937;
  letter-spacing: 0.3px;
}

.chart-card .header-icon {
  font-size: 24px;
  color: #ffffff;
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.2));
  background: rgba(255, 255, 255, 0.25);
  padding: 8px;
  border-radius: 10px;
  backdrop-filter: blur(10px);
}

.info-card .header-icon {
  font-size: 22px;
  color: #409EFF;
}

.ranking-card .header-icon {
  font-size: 22px;
  padding: 10px;
  border-radius: 10px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #ffffff;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

/* 订单状态列表 */
.order-status-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 4px;
}

.status-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  background: #fff;
  border-radius: 14px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border: 2px solid transparent;
  position: relative;
  overflow: hidden;
}

.status-item::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 4px;
  transition: width 0.3s ease;
}

.status-item:hover {
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.1);
  transform: translateX(4px);
}

.status-item:hover::before {
  width: 6px;
}

.status-item.status-pending {
  background: linear-gradient(135deg, rgba(255, 165, 0, 0.05) 0%, rgba(255, 165, 0, 0.02) 100%);
  border-color: rgba(255, 165, 0, 0.15);
}

.status-item.status-pending::before {
  background: linear-gradient(180deg, #FFA500 0%, #FFB732 100%);
}

.status-item.status-delivering {
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.05) 0%, rgba(59, 130, 246, 0.02) 100%);
  border-color: rgba(59, 130, 246, 0.15);
}

.status-item.status-delivering::before {
  background: linear-gradient(180deg, #3b82f6 0%, #60a5fa 100%);
}

.status-item.status-delivered {
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.05) 0%, rgba(16, 185, 129, 0.02) 100%);
  border-color: rgba(16, 185, 129, 0.15);
}

.status-item.status-delivered::before {
  background: linear-gradient(180deg, #10b981 0%, #34d399 100%);
}

.status-item.status-paid {
  background: linear-gradient(135deg, rgba(139, 92, 246, 0.05) 0%, rgba(139, 92, 246, 0.02) 100%);
  border-color: rgba(139, 92, 246, 0.15);
}

.status-item.status-paid::before {
  background: linear-gradient(180deg, #8b5cf6 0%, #a78bfa 100%);
}

.status-item.status-cancelled {
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.05) 0%, rgba(239, 68, 68, 0.02) 100%);
  border-color: rgba(239, 68, 68, 0.15);
}

.status-item.status-cancelled::before {
  background: linear-gradient(180deg, #ef4444 0%, #f87171 100%);
}

.status-info {
  display: flex;
  align-items: center;
  gap: 16px;
  flex: 1;
}

.status-icon-wrapper {
  width: 52px;
  height: 52px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.status-icon-wrapper.pending {
  background: linear-gradient(135deg, #FFA500 0%, #FFB732 100%);
}

.status-icon-wrapper.delivering {
  background: linear-gradient(135deg, #3b82f6 0%, #60a5fa 100%);
}

.status-icon-wrapper.delivered {
  background: linear-gradient(135deg, #10b981 0%, #34d399 100%);
}

.status-icon-wrapper.paid {
  background: linear-gradient(135deg, #8b5cf6 0%, #a78bfa 100%);
}

.status-icon-wrapper.cancelled {
  background: linear-gradient(135deg, #ef4444 0%, #f87171 100%);
}

.status-icon {
  font-size: 24px;
  color: #ffffff;
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.2));
}

.status-text {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.status-label {
  font-size: 13px;
  color: #6b7280;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.status-value {
  font-size: 24px;
  font-weight: 800;
  color: #1f2937;
  letter-spacing: -0.5px;
}

.status-percent {
  font-size: 18px;
  font-weight: 700;
  padding: 10px 16px;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
  border-radius: 10px;
  color: #667eea;
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.15);
  min-width: 70px;
  text-align: center;
}

/* 收入成本列表 */
.revenue-cost-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 4px;
}

.revenue-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  background: #fff;
  border-radius: 14px;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border: 2px solid transparent;
  position: relative;
  overflow: hidden;
}

.revenue-item::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 4px;
  transition: width 0.3s ease;
}

.revenue-item:hover {
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.1);
  transform: translateX(4px);
}

.revenue-item:hover::before {
  width: 6px;
}

.revenue-item.revenue-income {
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.05) 0%, rgba(16, 185, 129, 0.02) 100%);
  border-color: rgba(16, 185, 129, 0.15);
}

.revenue-item.revenue-income::before {
  background: linear-gradient(180deg, #10b981 0%, #34d399 100%);
}

.revenue-item.revenue-cost {
  background: linear-gradient(135deg, rgba(107, 114, 128, 0.05) 0%, rgba(107, 114, 128, 0.02) 100%);
  border-color: rgba(107, 114, 128, 0.15);
}

.revenue-item.revenue-cost::before {
  background: linear-gradient(180deg, #6b7280 0%, #9ca3af 100%);
}

.revenue-item.revenue-total {
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.08) 0%, rgba(59, 130, 246, 0.04) 100%);
  border: 3px solid rgba(59, 130, 246, 0.25);
  padding: 22px;
}

.revenue-item.revenue-total::before {
  background: linear-gradient(180deg, #3b82f6 0%, #60a5fa 100%);
  width: 5px;
}

.revenue-item.revenue-total:hover::before {
  width: 7px;
}

.revenue-info {
  display: flex;
  align-items: center;
  gap: 16px;
  flex: 1;
}

.revenue-icon-wrapper {
  width: 52px;
  height: 52px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.revenue-icon-wrapper.income {
  background: linear-gradient(135deg, #10b981 0%, #34d399 100%);
}

.revenue-icon-wrapper.cost {
  background: linear-gradient(135deg, #6b7280 0%, #9ca3af 100%);
}

.revenue-icon-wrapper.total {
  background: linear-gradient(135deg, #3b82f6 0%, #60a5fa 100%);
  width: 56px;
  height: 56px;
}

.revenue-icon {
  font-size: 24px;
  color: #ffffff;
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.2));
}

.revenue-icon-wrapper.total .revenue-icon {
  font-size: 26px;
}

.revenue-text {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.revenue-label {
  font-size: 13px;
  color: #6b7280;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.revenue-value {
  font-size: 24px;
  font-weight: 800;
  color: #1f2937;
  letter-spacing: -0.5px;
}

.revenue-value.positive {
  color: #10b981;
  font-size: 24px;
}

.revenue-total .revenue-value.positive {
  font-size: 26px;
}

.revenue-trend {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 16px;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 700;
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.2);
}

.revenue-trend.positive {
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.15) 0%, rgba(52, 211, 153, 0.1) 100%);
  color: #10b981;
}

.revenue-trend .el-icon {
  font-size: 20px;
}

.revenue-percent {
  font-size: 16px;
  font-weight: 700;
  padding: 8px 14px;
  background: linear-gradient(135deg, rgba(107, 114, 128, 0.1) 0%, rgba(156, 163, 175, 0.1) 100%);
  border-radius: 10px;
  color: #6b7280;
  box-shadow: 0 2px 8px rgba(107, 114, 128, 0.15);
  min-width: 65px;
  text-align: center;
}

.profit-rate-text {
  font-size: 14px;
  font-weight: 800;
}

/* 图表容器 */
.chart-container {
  height: 400px;
  position: relative;
  padding: 20px;
}

/* 排名表格 */
.ranking-table {
  border-radius: 8px;
  overflow: hidden;
}

.ranking-table :deep(.el-table) {
  border-radius: 8px;
}

.ranking-table :deep(.el-table__header) {
  background: linear-gradient(135deg, #667eea15 0%, #764ba215 100%);
}

.ranking-table :deep(.el-table__header th) {
  background: transparent;
  color: #374151;
  font-weight: 700;
  font-size: 14px;
  border-bottom: 2px solid #e5e7eb;
  padding: 16px 0;
  letter-spacing: 0.3px;
}

.ranking-table :deep(.el-table__body) {
  font-size: 14px;
}

.ranking-table :deep(.el-table__row) {
  transition: all 0.3s ease;
}

.ranking-table :deep(.el-table__row:hover) {
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.08) 0%, rgba(118, 75, 162, 0.05) 100%);
  transform: scale(1.01);
}

.ranking-table :deep(.el-table__row td) {
  padding: 14px 0;
  border-bottom: 1px solid #f3f4f6;
}

/* 前三名特殊样式 */
.ranking-table :deep(.el-table__row:nth-child(1)) {
  background: linear-gradient(135deg, rgba(255, 215, 0, 0.15) 0%, rgba(255, 215, 0, 0.08) 100%);
  font-weight: 600;
}

.ranking-table :deep(.el-table__row:nth-child(1):hover) {
  background: linear-gradient(135deg, rgba(255, 215, 0, 0.25) 0%, rgba(255, 215, 0, 0.15) 100%);
}

.ranking-table :deep(.el-table__row:nth-child(2)) {
  background: linear-gradient(135deg, rgba(192, 192, 192, 0.15) 0%, rgba(192, 192, 192, 0.08) 100%);
  font-weight: 600;
}

.ranking-table :deep(.el-table__row:nth-child(2):hover) {
  background: linear-gradient(135deg, rgba(192, 192, 192, 0.25) 0%, rgba(192, 192, 192, 0.15) 100%);
}

.ranking-table :deep(.el-table__row:nth-child(3)) {
  background: linear-gradient(135deg, rgba(205, 127, 50, 0.15) 0%, rgba(205, 127, 50, 0.08) 100%);
  font-weight: 600;
}

.ranking-table :deep(.el-table__row:nth-child(3):hover) {
  background: linear-gradient(135deg, rgba(205, 127, 50, 0.25) 0%, rgba(205, 127, 50, 0.15) 100%);
}

/* 排名列样式 */
.ranking-table :deep(.el-table__row:nth-child(1) .el-table-column--selection .cell) {
  color: #d4af37;
  font-weight: 800;
  font-size: 16px;
}

.ranking-table :deep(.el-table__row:nth-child(2) .el-table-column--selection .cell) {
  color: #a8a8a8;
  font-weight: 800;
  font-size: 16px;
}

.ranking-table :deep(.el-table__row:nth-child(3) .el-table-column--selection .cell) {
  color: #cd7f32;
  font-weight: 800;
  font-size: 16px;
}

/* 数值列样式 */
.ranking-table :deep(.el-table__body td) {
  color: #374151;
}

.ranking-table :deep(.el-table__row:nth-child(1) td),
.ranking-table :deep(.el-table__row:nth-child(2) td),
.ranking-table :deep(.el-table__row:nth-child(3) td) {
  font-weight: 700;
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
