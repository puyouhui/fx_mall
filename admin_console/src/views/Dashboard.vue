<template>
  <div class="dashboard-container">
    <h1 class="page-title">商品选购管理系统</h1>
    
    <!-- 顶部快捷数据卡片 -->
    <div class="quick-stats">
      <div class="quick-stat-item">
        <div class="quick-stat-label">今日销量</div>
        <div class="quick-stat-value">¥{{ dailySales }}</div>
        <div class="quick-stat-change">
          <el-icon><ArrowUp /></el-icon>
          <span class="text-success">7.8%</span>
        </div>
      </div>
      <div class="quick-stat-item">
        <div class="quick-stat-label">待发货</div>
        <div class="quick-stat-value">{{ pendingOrders }}</div>
        <div class="quick-stat-more">
          <el-button type="text" size="small">查看详情</el-button>
        </div>
      </div>
      <div class="quick-stat-item">
        <div class="quick-stat-label">特价商品</div>
        <div class="quick-stat-value">{{ specialProductsCount }}</div>
        <div class="quick-stat-more">
          <el-button type="text" size="small">查看详情</el-button>
        </div>
      </div>
      <div class="quick-stat-item">
        <div class="quick-stat-label">待处理退款</div>
        <div class="quick-stat-value">{{ refundRequests }}</div>
        <div class="quick-stat-change">
          <el-icon><ArrowDown /></el-icon>
          <span class="text-danger">4.2%</span>
        </div>
      </div>
      <div class="quick-stat-item">
        <div class="quick-stat-label">访客数</div>
        <div class="quick-stat-value">{{ visitorsCount }}</div>
        <div class="quick-stat-change">
          <el-icon><ArrowUp /></el-icon>
          <span class="text-success">12.5%</span>
        </div>
      </div>
    </div>
    
    <!-- 主要内容区域 -->
    <div class="main-content">
      <!-- 左侧区域：销售统计和趋势 -->
      <div class="left-section">
        <!-- 总销售额卡片 -->
        <el-card class="sales-overview">
          <div class="sales-header">
            <h3>总销售额 (今日)</h3>
            <div class="time-range">
              <el-radio-group v-model="dateRange">
                <el-radio-button label="今日">今日</el-radio-button>
                <el-radio-button label="本周">本周</el-radio-button>
                <el-radio-button label="本月">本月</el-radio-button>
                <el-radio-button label="全年">全年</el-radio-button>
              </el-radio-group>
            </div>
          </div>
          <div class="sales-amount">
            <span class="amount-value">{{ totalSales }}</span>
            <span class="amount-change">
              <el-icon><ArrowUp /></el-icon>
              <span class="text-success">+16.8%</span>
              <span class="amount-compare">较昨日</span>
            </span>
          </div>
          
          <!-- 销售趋势图表 -->
          <div class="sales-chart">
            <canvas ref="salesChart"></canvas>
          </div>
        </el-card>
      </div>
      
      <!-- 右侧区域：销售排行和商家信息 -->
      <div class="right-section">
        <!-- 销售排行榜 -->
        <el-card class="sales-ranking">
          <h3 class="section-title">销售排行榜</h3>
          <div class="ranking-tabs">
            <el-tabs v-model="rankingType" type="border-card">
              <el-tab-pane label="今日" name="today">
                <el-table :data="salesRankingToday" size="small" show-header="false">
                  <el-table-column type="index" width="40" />
                  <el-table-column prop="name" label="商品名称" width="180">
                    <template #default="scope">
                      <div class="ranking-product">
                        <img v-if="scope.row.image" :src="scope.row.image" alt="商品图片" class="ranking-product-image">
                        <span class="ranking-product-name">{{ scope.row.name }}</span>
                      </div>
                    </template>
                  </el-table-column>
                  <el-table-column prop="sales" label="销售额" align="right">
                    <template #default="scope">
                      <span class="ranking-sales">{{ scope.row.sales }}</span>
                    </template>
                  </el-table-column>
                </el-table>
              </el-tab-pane>
              <el-tab-pane label="本周" name="week">
                <el-table :data="salesRankingWeek" size="small" show-header="false">
                  <el-table-column type="index" width="40" />
                  <el-table-column prop="name" label="商品名称" width="180">
                    <template #default="scope">
                      <div class="ranking-product">
                        <img v-if="scope.row.image" :src="scope.row.image" alt="商品图片" class="ranking-product-image">
                        <span class="ranking-product-name">{{ scope.row.name }}</span>
                      </div>
                    </template>
                  </el-table-column>
                  <el-table-column prop="sales" label="销售额" align="right">
                    <template #default="scope">
                      <span class="ranking-sales">{{ scope.row.sales }}</span>
                    </template>
                  </el-table-column>
                </el-table>
              </el-tab-pane>
              <el-tab-pane label="本月" name="month">
                <el-table :data="salesRankingMonth" size="small" show-header="false">
                  <el-table-column type="index" width="40" />
                  <el-table-column prop="name" label="商品名称" width="180">
                    <template #default="scope">
                      <div class="ranking-product">
                        <img v-if="scope.row.image" :src="scope.row.image" alt="商品图片" class="ranking-product-image">
                        <span class="ranking-product-name">{{ scope.row.name }}</span>
                      </div>
                    </template>
                  </el-table-column>
                  <el-table-column prop="sales" label="销售额" align="right">
                    <template #default="scope">
                      <span class="ranking-sales">{{ scope.row.sales }}</span>
                    </template>
                  </el-table-column>
                </el-table>
              </el-tab-pane>
            </el-tabs>
          </div>
        </el-card>
        
        <!-- 商家信息卡片 -->
        <el-card class="merchant-info">
          <div class="merchant-header">
            <!-- <div class="merchant-logo">
              <img src="/static/logo.png" alt="商家Logo" class="logo-img">
            </div> -->
            <div class="merchant-details">
              <div class="merchant-name">进货管理后台</div>
              <div class="merchant-stats">
                <span class="stat-item">商品评分: <span class="score">4.8</span></span>
                <span class="stat-item">服务态度: <span class="score">4.9</span></span>
                <span class="stat-item">物流速度: <span class="score">4.7</span></span>
              </div>
              <div class="merchant-balance">
                保证金: <span class="balance-amount">¥100,000.00</span>
              </div>
            </div>
          </div>
          <div class="merchant-actions">
            <el-button type="primary" size="small">查看详情</el-button>
          </div>
        </el-card>
        
        <!-- 常用功能快捷入口 -->
        <el-card class="quick-functions">
          <h3 class="section-title">常用功能</h3>
          <div class="functions-grid">
            <div class="function-item">
              <div class="function-icon">
                <el-icon><Plus /></el-icon>
              </div>
              <div class="function-label">发布商品</div>
            </div>
            <div class="function-item">
              <div class="function-icon">
                <el-icon><Box /></el-icon>
              </div>
              <div class="function-label">商品管理</div>
            </div>
            <div class="function-item">
              <div class="function-icon">
                <el-icon><VideoPlay /></el-icon>
              </div>
              <div class="function-label">活动管理</div>
            </div>
            <div class="function-item">
              <div class="function-icon">
                <el-icon><User /></el-icon>
              </div>
              <div class="function-label">客户管理</div>
            </div>
          </div>
        </el-card>
      </div>
    </div>
    
    <!-- 推广活动卡片 -->
    <!-- <el-card class="promotion-card">
      <div class="promotion-content">
        <div class="promotion-text">
          <h3>优质商家扶持计划</h3>
          <p>专享流量倾斜，提升店铺曝光率</p>
        </div>
        <el-button type="primary">立即申请</el-button>
      </div>
    </el-card> -->
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import { 
  ShoppingBag, Grid, ShoppingCart, Money, ArrowUp, ArrowDown, 
  Plus, Box, VideoPlay, User
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { getDashboardData } from '../api/dashboard'
import { getProducts } from '../api/products'
import { getCategories } from '../api/categories'

// 导入Chart.js用于销售趋势图表
import Chart from 'chart.js/auto'

// 数据初始化
const productCount = ref(0)
const categoryCount = ref(0)
const orderCount = ref(0)
const totalSales = ref('¥0.00')
const dailySales = ref('6,987.35')
const pendingOrders = ref(804)
const specialProductsCount = ref(19)
const refundRequests = ref(27)
const visitorsCount = ref(8)

// 图表和时间范围相关
const dateRange = ref('今日')
const salesChart = ref(null)
const chartInstance = ref(null)

// 销售排行榜相关
const rankingType = ref('today')
const salesRankingToday = ref([])
const salesRankingWeek = ref([])
const salesRankingMonth = ref([])

// 初始化页面数据
onMounted(() => {
  fetchDashboardData()
})

// 获取仪表盘数据
const fetchDashboardData = async () => {
  try {
    // 获取商品数据
    const productsRes = await getProducts({ pageNum: 1, pageSize: 1000 })
    if (productsRes.code === 200) {
      productCount.value = productsRes.data.total
      
      // 计算特价商品数量
      const specialCount = productsRes.data.list.filter(p => p.is_special).length
      specialProductsCount.value = specialCount
      
      // 获取分类数据
      const categoriesRes = await getCategories()
      if (categoriesRes.code === 200) {
        categoryCount.value = categoriesRes.data.length
        
        // 生成销售排行榜数据
        generateSalesRankingData(productsRes.data.list)
      }
    }
    
    // 使用模拟的订单和销售额数据
    generateOrderAndSalesData()
    
    // 等待DOM更新后初始化图表
    await nextTick()
    initSalesChart()
    
  } catch (error) {
    console.error('获取数据失败:', error)
    ElMessage.error('获取数据失败，请稍后再试')
    
    // 失败时使用备用数据
    useFallbackData()
  }
}

// 生成销售排行榜数据
const generateSalesRankingData = (products) => {
  // 生成今日销售排行
  const todayRanking = products
    .map(product => ({
      id: product.id,
      name: product.name,
      image: product.images && product.images.length > 0 ? product.images[0] : '',
      sales: `¥${Math.floor(Math.random() * 10000 + 1000).toLocaleString('zh-CN')}.00`
    }))
    .sort((a, b) => {
      const salesA = parseInt(a.sales.replace(/[^\d]/g, ''))
      const salesB = parseInt(b.sales.replace(/[^\d]/g, ''))
      return salesB - salesA
    })
    .slice(0, 5)
  
  salesRankingToday.value = todayRanking
  
  // 生成本周销售排行（在今日基础上增加一些数值）
  salesRankingWeek.value = todayRanking.map(item => ({
    ...item,
    sales: `¥${Math.floor(parseInt(item.sales.replace(/[^\d]/g, '')) * (1 + Math.random() * 2)).toLocaleString('zh-CN')}.00`
  }))
  
  // 生成本月销售排行（在本周基础上增加更多数值）
  salesRankingMonth.value = salesRankingWeek.value.map(item => ({
    ...item,
    sales: `¥${Math.floor(parseInt(item.sales.replace(/[^\d]/g, '')) * (2 + Math.random() * 3)).toLocaleString('zh-CN')}.00`
  }))
}

// 生成订单和销售额数据
const generateOrderAndSalesData = () => {
  // 基于商品数量计算合理的订单数量
  const baseOrderCount = Math.floor(productCount.value * 3.5) // 假设每个商品平均有3.5个订单
  orderCount.value = baseOrderCount + Math.floor(Math.random() * 100) // 增加一些随机性
  
  // 计算销售额（基于商品数量和平均客单价）
  const avgOrderValue = 350 // 假设平均订单金额为350元
  const salesAmount = orderCount.value * avgOrderValue + Math.floor(Math.random() * 5000)
  totalSales.value = `¥${salesAmount.toLocaleString('zh-CN')}.00`
}

// 初始化销售趋势图表
const initSalesChart = () => {
  if (!salesChart.value) return
  
  // 销毁已存在的图表实例
  if (chartInstance.value) {
    chartInstance.value.destroy()
  }
  
  // 模拟销售趋势数据
  const labels = ['12-01', '12-02', '12-03', '12-04', '12-05', '12-06', '12-07']
  const salesData = [6500, 7800, 8900, 7200, 8100, 9500, 8800]
  
  // 创建图表
  chartInstance.value = new Chart(salesChart.value, {
    type: 'line',
    data: {
      labels: labels,
      datasets: [{
        label: '销售额',
        data: salesData,
        borderColor: '#2E74FF',
        backgroundColor: 'rgba(46, 116, 255, 0.1)',
        borderWidth: 2,
        fill: true,
        tension: 0.4,
        pointBackgroundColor: '#2E74FF',
        pointBorderColor: '#fff',
        pointBorderWidth: 2,
        pointRadius: 4,
        pointHoverRadius: 6
      }]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          display: false
        },
        tooltip: {
          mode: 'index',
          intersect: false,
          backgroundColor: '#fff',
          titleColor: '#333',
          bodyColor: '#666',
          borderColor: '#e0e0e0',
          borderWidth: 1,
          padding: 10,
          cornerRadius: 4
        }
      },
      scales: {
        x: {
          grid: {
            display: false
          },
          ticks: {
            color: '#666'
          }
        },
        y: {
          grid: {
            color: 'rgba(0, 0, 0, 0.05)'
          },
          ticks: {
            color: '#666',
            callback: function(value) {
              return '¥' + value.toLocaleString()
            }
          }
        }
      }
    }
  })
}

// 使用备用数据
const useFallbackData = () => {
  // 当API调用失败时使用的备用数据
  productCount.value = 24
  categoryCount.value = 6
  orderCount.value = 85
  totalSales.value = '¥12800.50'
  dailySales.value = '6,987.35'
  pendingOrders.value = 804
  specialProductsCount.value = 19
  refundRequests.value = 27
  visitorsCount.value = 8
  
  // 备用排行榜数据
  salesRankingToday.value = [
    { id: 1, name: '智能手机', image: '/static/test/product1-1.jpg', sales: '¥12,500.00' },
    { id: 3, name: '智能手表', image: '/static/test/product3-1.jpg', sales: '¥8,900.00' },
    { id: 2, name: '笔记本电脑', image: '/static/test/product2-1.jpg', sales: '¥6,500.00' },
    { id: 4, name: '时尚台灯', image: '/static/test/product4-1.jpg', sales: '¥5,800.00' },
    { id: 5, name: '休闲T恤', image: '/static/test/product5-1.jpg', sales: '¥4,200.00' }
  ]
  
  // 等待DOM更新后初始化图表
  nextTick(() => {
    initSalesChart()
  })
}
</script>

<style scoped>
.dashboard-container {
  padding: 20px;
  height: 100%;
  background-color: #f5f7fa;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 20px;
  color: #2E74FF;
}

/* 顶部快捷数据卡片 */
.quick-stats {
  display: flex;
  gap: 16px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.quick-stat-item {
  flex: 1;
  min-width: 180px;
  padding: 16px;
  background-color: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  display: flex;
  flex-direction: column;
}

.quick-stat-label {
  font-size: 14px;
  color: #657288;
  margin-bottom: 8px;
}

.quick-stat-value {
  font-size: 24px;
  font-weight: 600;
  color: #2E74FF;
  margin-bottom: 8px;
}

.quick-stat-change {
  display: flex;
  align-items: center;
  font-size: 12px;
}

.quick-stat-more {
  margin-top: auto;
}

.text-success {
  color: #67c23a;
}

.text-danger {
  color: #f56c6c;
}

/* 主要内容区域 */
.main-content {
  display: flex;
  gap: 20px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.left-section {
  flex: 2;
  min-width: 400px;
}

.right-section {
  flex: 1;
  min-width: 300px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* 销售概览卡片 */
.sales-overview {
  padding: 20px;
  height: fit-content;
}

.sales-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.sales-header h3 {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin: 0;
}

.time-range .el-radio-button__inner {
  border-color: #dcdfe6;
  color: #606266;
}

.time-range .el-radio-button__orig-radio:checked + .el-radio-button__inner {
  background-color: #2E74FF;
  border-color: #2E74FF;
  color: #fff;
}

.sales-amount {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: 20px;
}

.amount-value {
  font-size: 32px;
  font-weight: 600;
  color: #2E74FF;
}

.amount-change {
  display: flex;
  align-items: center;
  font-size: 14px;
}

.amount-compare {
  color: #606266;
  margin-left: 8px;
}

.sales-chart {
  height: 240px;
}

/* 销售排行榜 */
.sales-ranking {
  padding: 20px;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 16px;
}

.ranking-tabs .el-tabs__header {
  margin-bottom: 16px;
}

.ranking-tabs .el-tabs__nav-wrap::after {
  background-color: transparent;
}

.ranking-tabs .el-tabs__item {
  color: #606266;
}

.ranking-tabs .el-tabs__item.is-active {
  color: #2E74FF;
}

.ranking-tabs .el-tabs__active-bar {
  background-color: #2E74FF;
}

.ranking-product {
  display: flex;
  align-items: center;
  gap: 8px;
}

.ranking-product-image {
  width: 32px;
  height: 32px;
  border-radius: 4px;
  object-fit: cover;
}

.ranking-product-name {
  font-size: 14px;
  color: #303133;
}

.ranking-sales {
  font-size: 14px;
  font-weight: 500;
  color: #2E74FF;
}

/* 商家信息卡片 */
.merchant-info {
  padding: 20px;
}

.merchant-header {
  display: flex;
  gap: 16px;
  align-items: center;
  margin-bottom: 16px;
}

.merchant-logo {
  width: 60px;
  height: 60px;
  background-color: #f0f2f5;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.logo-img {
  width: 40px;
  height: 40px;
  object-fit: contain;
}

.merchant-details {
  flex: 1;
}

.merchant-name {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 8px;
}

.merchant-stats {
  display: flex;
  gap: 16px;
  margin-bottom: 8px;
}

.stat-item {
  font-size: 12px;
  color: #606266;
}

.score {
  color: #2E74FF;
  font-weight: 500;
}

.merchant-balance {
  font-size: 12px;
  color: #606266;
}

.balance-amount {
  color: #f56c6c;
  font-weight: 500;
}

.merchant-actions {
  display: flex;
  justify-content: flex-end;
}

/* 常用功能快捷入口 */
.quick-functions {
  padding: 20px;
}

.functions-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.function-item {
  text-align: center;
  padding: 16px;
  background-color: #f0f7ff;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.function-item:hover {
  background-color: #e0ebff;
  transform: translateY(-2px);
}

.function-icon {
  width: 40px;
  height: 40px;
  background-color: #2E74FF;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 8px;
  color: #fff;
  font-size: 20px;
}

.function-label {
  font-size: 14px;
  color: #303133;
}

/* 推广活动卡片 */
.promotion-card {
  background: linear-gradient(135deg, #2E74FF, #4a90e2);
  color: #fff;
  padding: 0;
  border-radius: 8px;
  overflow: hidden;
}

.promotion-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24px;
}

.promotion-text h3 {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 8px;
}

.promotion-text p {
  font-size: 14px;
  opacity: 0.9;
  margin: 0;
}

.promotion-card .el-button {
  background-color: #fff;
  color: #2E74FF;
  border: none;
}

.promotion-card .el-button:hover {
  background-color: #f0f7ff;
}

/* 响应式设计 */
@media (max-width: 1200px) {
  .main-content {
    flex-direction: column;
  }
  
  .left-section,
  .right-section {
    min-width: auto;
  }
}

@media (max-width: 768px) {
  .dashboard-container {
    padding: 16px;
  }
  
  .page-title {
    font-size: 20px;
  }
  
  .quick-stats {
    flex-direction: column;
  }
  
  .quick-stat-item {
    min-width: auto;
  }
  
  .functions-grid {
    grid-template-columns: 1fr;
  }
  
  .promotion-content {
    flex-direction: column;
    text-align: center;
    gap: 16px;
  }
}
</style>