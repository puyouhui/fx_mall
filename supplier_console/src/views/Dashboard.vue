<template>
  <div class="dashboard">
    <div class="page-header">
      <h2 class="page-title">数据总览</h2>
      <div class="period-selector">
        <el-radio-group v-model="selectedPeriod" @change="handlePeriodChange">
          <el-radio-button label="today">今日</el-radio-button>
          <el-radio-button label="7days">7日</el-radio-button>
          <el-radio-button label="month">本月</el-radio-button>
          <el-radio-button label="year">今年</el-radio-button>
        </el-radio-group>
      </div>
    </div>
    
    <!-- 第一行：核心指标 -->
    <el-row :gutter="20" style="margin-bottom: 20px;">
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background-color: #409eff;">
              <el-icon><ShoppingBag /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.total_products || 0 }}</div>
              <div class="stat-label">我供应的商品</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background-color: #67c23a;">
              <el-icon><Document /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.order_count || 0 }}</div>
              <div class="stat-label">
                {{ periodLabel }}订单数量
                <span v-if="stats.order_growth_rate !== undefined && (selectedPeriod === 'today' || selectedPeriod === 'month')" 
                      :class="['growth-rate', stats.order_growth_rate >= 0 ? 'positive' : 'negative']">
                  {{ stats.order_growth_rate >= 0 ? '↑' : '↓' }}{{ formatGrowthRate(stats.order_growth_rate) }}%
                </span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background-color: #e6a23c;">
              <el-icon><Box /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.item_count || 0 }}</div>
              <div class="stat-label">{{ periodLabel }}货物件数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background-color: #f56c6c;">
              <el-icon><Money /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">¥{{ formatAmount(stats.total_sales_amount) }}</div>
              <div class="stat-label">
                已完成总额
                <span v-if="stats.amount_growth_rate !== undefined && (selectedPeriod === 'today' || selectedPeriod === 'month')" 
                      :class="['growth-rate', stats.amount_growth_rate >= 0 ? 'positive' : 'negative']">
                  {{ stats.amount_growth_rate >= 0 ? '↑' : '↓' }}{{ formatGrowthRate(stats.amount_growth_rate) }}%
                </span>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 第二行：待备货信息和平均订单金额 -->
    <el-row :gutter="20" style="margin-bottom: 20px;">
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background-color: #909399;">
              <el-icon><Clock /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.pending_order_count || 0 }}</div>
              <div class="stat-label">待备货订单</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background-color: #909399;">
              <el-icon><Box /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.pending_item_count || 0 }}</div>
              <div class="stat-label">待备货商品数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background-color: #909399;">
              <el-icon><Money /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">¥{{ formatAmount(stats.pending_amount) }}</div>
              <div class="stat-label">待备货金额</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background-color: #9c27b0;">
              <el-icon><TrendCharts /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">¥{{ formatAmount(stats.avg_order_amount) }}</div>
              <div class="stat-label">平均订单金额</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 第三行：热销商品和最近订单并排 -->
    <el-row :gutter="20">
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>热销商品TOP 5</span>
            </div>
          </template>
          <div class="top-products">
            <div v-if="!stats.top_products || stats.top_products.length === 0" class="empty-data">暂无数据</div>
            <div v-else class="product-list">
              <div v-for="(product, index) in stats.top_products" :key="product.product_id" class="product-item">
                <div class="product-rank">{{ index + 1 }}</div>
                <div class="product-info">
                  <div class="product-name">{{ product.product_name }}</div>
                  <div class="product-stats">
                    <span>销量: {{ product.total_qty }}</span>
                    <span>金额: ¥{{ formatAmount(product.total_amount) }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>最近订单</span>
            </div>
          </template>
          <div class="recent-orders">
            <div v-if="!stats.recent_orders || stats.recent_orders.length === 0" class="empty-data">暂无数据</div>
            <el-table v-else :data="stats.recent_orders" style="width: 100%" stripe>
              <el-table-column label="客户编号" width="120" align="center">
                <template #default="scope">
                  <span>{{ scope.row.user_code || '-' }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="order_number" label="订单号" width="180" align="center" />
              <el-table-column label="状态" width="120" align="center">
                <template #default="scope">
                  <el-tag :type="scope.row.status === '已取货' ? 'success' : 'warning'" size="small">
                    {{ scope.row.status }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="商品数量" width="100" align="center">
                <template #default="scope">
                  {{ scope.row.item_count || 0 }} 件
                </template>
              </el-table-column>
              <el-table-column label="成本总额" width="150" align="center">
                <template #default="scope">
                  <span class="cost-price">¥{{ formatAmount(scope.row.total_cost) }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="created_at" label="创建时间" width="180" align="center">
                <template #default="scope">
                  {{ formatDateTime(scope.row.created_at) }}
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 第四行：最近15日销售情况 -->
    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="24">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>最近15日销售情况</span>
            </div>
          </template>
          <div class="daily-sales">
            <div v-if="!stats.daily_sales || stats.daily_sales.length === 0" class="empty-data">暂无数据</div>
            <div v-else>
              <!-- ECharts 图表 -->
              <div ref="chartRef" class="chart-container"></div>
              <!-- 数据表格 -->
              <el-table :data="stats.daily_sales" style="width: 100%; margin-top: 20px;" stripe>
                <el-table-column prop="date" label="日期" width="150">
                  <template #default="scope">
                    {{ formatDate(scope.row.date) }}
                  </template>
                </el-table-column>
                <el-table-column prop="order_count" label="订单数量" width="120" align="center" />
                <el-table-column prop="item_count" label="货物件数" width="120" align="center" />
                <el-table-column prop="sales_amount" label="销售金额" width="150" align="right">
                  <template #default="scope">
                    ¥{{ formatAmount(scope.row.sales_amount) }}
                  </template>
                </el-table-column>
                <el-table-column label="平均订单金额" width="150" align="right">
                  <template #default="scope">
                    <span v-if="scope.row.order_count > 0">
                      ¥{{ formatAmount(scope.row.sales_amount / scope.row.order_count) }}
                    </span>
                    <span v-else>¥0.00</span>
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { ShoppingBag, Document, Money, Clock, Box, TrendCharts } from '@element-plus/icons-vue'
import * as echarts from 'echarts'
import { getDashboard } from '../api/dashboard'

const selectedPeriod = ref('today')
const chartRef = ref(null)
let chartInstance = null

const stats = ref({
  total_products: 0,
  order_count: 0,
  item_count: 0,
  total_sales_amount: 0,
  avg_order_amount: 0,
  pending_order_count: 0,
  pending_item_count: 0,
  pending_amount: 0,
  top_products: [],
  recent_orders: [],
  daily_sales: [],
  order_growth_rate: 0,
  amount_growth_rate: 0,
  period: 'today'
})

// 根据选择的时间范围显示标签
const periodLabel = computed(() => {
  const labelMap = {
    'today': '今日',
    '7days': '7日',
    'month': '本月',
    'year': '今年'
  }
  return labelMap[selectedPeriod.value] || '今日'
})

// 格式化金额
const formatAmount = (amount) => {
  if (!amount) return '0.00'
  return Number(amount).toFixed(2)
}

// 格式化增长率
const formatGrowthRate = (rate) => {
  if (rate === undefined || rate === null) return '0.0'
  return Math.abs(rate).toFixed(1)
}

// 格式化日期时间
const formatDateTime = (dateTime) => {
  if (!dateTime) return ''
  const date = new Date(dateTime)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}`
}

// 格式化日期（用于表格显示，显示完整日期）
const formatDate = (dateStr) => {
  if (!dateStr) return ''
  // 如果已经是格式化的日期字符串，直接返回
  if (dateStr.includes('-')) {
    return dateStr
  }
  // 如果是日期对象，格式化
  const date = new Date(dateStr)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

// 格式化日期为月-日格式（用于图表显示）
const formatDateForChart = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${month}-${day}`
}


// 响应式调整图表大小
const handleResize = () => {
  if (chartInstance) {
    chartInstance.resize()
  }
}

// 初始化图表
const initChart = () => {
  if (!chartRef.value) return
  
  // 如果图表已存在，先销毁
  if (chartInstance) {
    chartInstance.dispose()
  }
  
  // 创建图表实例
  chartInstance = echarts.init(chartRef.value)
  
  // 更新图表数据
  updateChart()
  
  // 添加窗口大小调整监听
  window.addEventListener('resize', handleResize)
}

// 更新图表数据
const updateChart = () => {
  if (!chartInstance || !stats.value.daily_sales || stats.value.daily_sales.length === 0) {
    return
  }

  // 准备数据（数据已经是按日期正序排列的，直接使用）
  const dates = stats.value.daily_sales.map(item => formatDateForChart(item.date))
  const orderCounts = stats.value.daily_sales.map(item => item.order_count)
  const itemCounts = stats.value.daily_sales.map(item => item.item_count)
  const salesAmounts = stats.value.daily_sales.map(item => item.sales_amount)

  const option = {
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'cross'
      },
      formatter: function(params) {
        let result = params[0].axisValue + '<br/>'
        params.forEach(param => {
          if (param.seriesName === '销售金额') {
            result += param.marker + param.seriesName + ': ¥' + formatAmount(param.value) + '<br/>'
          } else {
            result += param.marker + param.seriesName + ': ' + param.value + '<br/>'
          }
        })
        return result
      }
    },
    legend: {
      data: ['订单数量', '货物件数', '销售金额'],
      top: 10
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: dates,
      axisLabel: {
        rotate: 45,
        formatter: function(value) {
          return value
        }
      }
    },
    yAxis: [
      {
        type: 'value',
        name: '数量',
        position: 'left',
        axisLabel: {
          formatter: '{value}'
        }
      },
      {
        type: 'value',
        name: '金额(元)',
        position: 'right',
        axisLabel: {
          formatter: '¥{value}'
        }
      }
    ],
    series: [
      {
        name: '订单数量',
        type: 'line',
        data: orderCounts,
        smooth: true,
        itemStyle: {
          color: '#409eff'
        },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(64, 158, 255, 0.3)' },
              { offset: 1, color: 'rgba(64, 158, 255, 0.1)' }
            ]
          }
        }
      },
      {
        name: '货物件数',
        type: 'line',
        data: itemCounts,
        smooth: true,
        itemStyle: {
          color: '#67c23a'
        },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(103, 194, 58, 0.3)' },
              { offset: 1, color: 'rgba(103, 194, 58, 0.1)' }
            ]
          }
        }
      },
      {
        name: '销售金额',
        type: 'bar',
        yAxisIndex: 1,
        data: salesAmounts,
        itemStyle: {
          color: '#e6a23c'
        }
      }
    ]
  }

  chartInstance.setOption(option)
}

// 加载数据总览
const loadDashboard = async (period = 'today') => {
  try {
    const response = await getDashboard(period)
    if (response.code === 200 && response.data) {
      stats.value = response.data
      selectedPeriod.value = response.data.period || period
      
      // 数据加载后更新图表
      await nextTick()
      if (stats.value.daily_sales && stats.value.daily_sales.length > 0) {
        initChart()
      }
    } else {
      ElMessage.error(response.message || '获取数据失败')
    }
  } catch (error) {
    console.error('获取数据总览失败:', error)
    ElMessage.error('获取数据总览失败，请稍后再试')
  }
}

// 时间范围改变
const handlePeriodChange = (period) => {
  loadDashboard(period)
}

onMounted(() => {
  loadDashboard('today')
})

// 组件卸载时销毁图表
onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  if (chartInstance) {
    chartInstance.dispose()
    chartInstance = null
  }
})
</script>

<style scoped>
.dashboard {
  padding: 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
  margin: 0;
}

.period-selector {
  display: flex;
  align-items: center;
}

.stat-card {
  margin-bottom: 20px;
  transition: all 0.3s;
}

.stat-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.stat-content {
  display: flex;
  align-items: center;
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 24px;
  margin-right: 15px;
  flex-shrink: 0;
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 28px;
  font-weight: bold;
  color: #303133;
  margin-bottom: 5px;
  line-height: 1.2;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  line-height: 1.2;
  display: flex;
  align-items: center;
  gap: 8px;
}

.growth-rate {
  font-size: 12px;
  font-weight: 500;
}

.growth-rate.positive {
  color: #67c23a;
}

.growth-rate.negative {
  color: #f56c6c;
}

.card-header {
  font-size: 16px;
  font-weight: 500;
}

.top-products {
  min-height: 200px;
}

.product-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.product-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background-color: #f5f7fa;
  border-radius: 8px;
  transition: all 0.3s;
}

.product-item:hover {
  background-color: #ecf5ff;
  transform: translateX(4px);
}

.product-rank {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 14px;
  flex-shrink: 0;
}

.product-info {
  flex: 1;
}

.product-name {
  font-size: 15px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 6px;
}

.product-stats {
  display: flex;
  gap: 16px;
  font-size: 13px;
  color: #909399;
}

.recent-orders {
  min-height: 200px;
}

.daily-sales {
  min-height: 200px;
}

.chart-container {
  width: 100%;
  height: 400px;
  margin-bottom: 20px;
}

.empty-data {
  text-align: center;
  padding: 40px;
  color: #909399;
  font-size: 14px;
}

.cost-price {
  color: #409eff;
  font-weight: 500;
}
</style>
