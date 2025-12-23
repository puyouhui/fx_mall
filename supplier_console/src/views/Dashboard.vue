<template>
  <div class="dashboard">
    <div class="page-header">
      <h2 class="page-title">数据总览</h2>
    </div>
    
    <!-- 今日已取货统计 -->
    <el-row :gutter="20" style="margin-bottom: 20px;">
      <el-col :span="8">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background-color: #67c23a;">
              <el-icon><Box /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ pickedStats.item_count || 0 }}</div>
              <div class="stat-label">今日已取货件数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background-color: #67c23a;">
              <el-icon><ShoppingBag /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ pickedStats.goods_count || 0 }}</div>
              <div class="stat-label">今日已取货种类数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background-color: #67c23a;">
              <el-icon><Money /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">¥{{ formatAmount(pickedStats.total_amount) }}</div>
              <div class="stat-label">今日已取货金额</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 第四行：今日待备货列表和今日已取货列表 -->
    <el-row :gutter="20">
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>待备货货物列表</span>
              <el-tag type="danger" size="small">
                共 {{ pendingGoods.length }} 种
              </el-tag>
            </div>
          </template>
          <div class="goods-list">
            <div v-if="pendingGoods.length === 0" class="empty-data">暂无待备货货物</div>
            <el-table v-else :data="pendingGoods" style="width: 100%" stripe max-height="400">
              <el-table-column prop="product_name" label="商品名称" min-width="150" align="center" show-overflow-tooltip />
              <el-table-column prop="spec_name" label="规格" width="120" align="center" />
              <el-table-column label="数量" width="100" align="center">
                <template #default="scope">
                  {{ scope.row.quantity }} 件
                </template>
              </el-table-column>
              <el-table-column label="成本价" width="120" align="center">
                <template #default="scope">
                  <span class="cost-price">¥{{ formatAmount(scope.row.cost_price) }}</span>
                </template>
              </el-table-column>
              <el-table-column label="小计" width="120" align="center">
                <template #default="scope">
                  <span class="cost-price">¥{{ formatAmount(scope.row.total_cost) }}</span>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>今日已取货货物列表</span>
              <el-tag type="success" size="small">
                共 {{ pickedGoods.length }} 种
              </el-tag>
            </div>
          </template>
          <div class="goods-list">
            <div v-if="pickedGoods.length === 0" class="empty-data">暂无已取货货物</div>
            <el-table v-else :data="pickedGoods" style="width: 100%" stripe max-height="400">
              <el-table-column prop="product_name" label="商品名称" min-width="150" align="center" show-overflow-tooltip />
              <el-table-column prop="spec_name" label="规格" width="120" align="center" />
              <el-table-column label="数量" width="100" align="center">
                <template #default="scope">
                  {{ scope.row.quantity }} 件
                </template>
              </el-table-column>
              <el-table-column label="成本价" width="120" align="center">
                <template #default="scope">
                  <span class="cost-price">¥{{ formatAmount(scope.row.cost_price) }}</span>
                </template>
              </el-table-column>
              <el-table-column label="小计" width="120" align="center">
                <template #default="scope">
                  <span class="cost-price">¥{{ formatAmount(scope.row.total_cost) }}</span>
                </template>
              </el-table-column>
            </el-table>
            <div v-if="pickedGoods.length > 0" class="total-amount">
              <span class="total-label">总金额：</span>
              <span class="total-value">¥{{ formatAmount(pickedStats.total_amount) }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { ShoppingBag, Box, Money, Clock, TrendCharts } from '@element-plus/icons-vue'
import { getTodayGoodsStats, getTodayPendingGoods, getTodayPickedGoods } from '../api/goods'

// 总统计
const totalStats = ref({
  total_products: 0,
  total_item_count: 0,
  total_amount: 0,
  total_goods_count: 0
})

// 待备货统计
const pendingStats = ref({
  item_count: 0,
  goods_count: 0,
  total_amount: 0
})

// 已取货统计
const pickedStats = ref({
  item_count: 0,
  goods_count: 0,
  total_amount: 0
})

// 待备货货物列表
const pendingGoods = ref([])

// 已取货货物列表
const pickedGoods = ref([])

// 格式化金额
const formatAmount = (amount) => {
  if (!amount) return '0.00'
  return Number(amount).toFixed(2)
}

// 加载今日货物统计
const loadTodayStats = async () => {
  try {
    const response = await getTodayGoodsStats()
    if (response.code === 200 && response.data) {
      totalStats.value = response.data.total || {}
      pendingStats.value = response.data.pending || {}
      pickedStats.value = response.data.picked || {}
    } else {
      ElMessage.error(response.message || '获取统计数据失败')
    }
  } catch (error) {
    console.error('获取统计数据失败:', error)
    ElMessage.error('获取统计数据失败，请稍后再试')
  }
}

// 加载今日待备货货物列表
const loadPendingGoods = async () => {
  try {
    const response = await getTodayPendingGoods()
    if (response.code === 200 && response.data) {
      pendingGoods.value = response.data.list || []
    } else {
      ElMessage.error(response.message || '获取待备货货物列表失败')
    }
  } catch (error) {
    console.error('获取待备货货物列表失败:', error)
    ElMessage.error('获取待备货货物列表失败，请稍后再试')
  }
}

// 加载今日已取货货物列表
const loadPickedGoods = async () => {
  try {
    const response = await getTodayPickedGoods()
    if (response.code === 200 && response.data) {
      pickedGoods.value = response.data.list || []
    } else {
      ElMessage.error(response.message || '获取已取货货物列表失败')
    }
  } catch (error) {
    console.error('获取已取货货物列表失败:', error)
    ElMessage.error('获取已取货货物列表失败，请稍后再试')
  }
}

// 加载所有数据
const loadAllData = async () => {
  await Promise.all([
    loadTodayStats(),
    loadPendingGoods(),
    loadPickedGoods()
  ])
}

onMounted(() => {
  loadAllData()
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
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 16px;
  font-weight: 500;
}

.goods-list {
  min-height: 200px;
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

.total-amount {
  margin-top: 15px;
  padding: 15px;
  text-align: right;
  border-top: 1px solid #e4e7ed;
  background-color: #f5f7fa;
}

.total-label {
  font-size: 16px;
  color: #606266;
  margin-right: 10px;
}

.total-value {
  font-size: 20px;
  font-weight: bold;
  color: #409eff;
}
</style>
