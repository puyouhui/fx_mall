<template>
  <div class="payments-container">
    <div class="page-header">
      <h2 class="page-title">付款对账</h2>
    </div>

    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="8">
        <el-card class="stat-card total-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon">
              <el-icon><Money /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-label">总应付款</div>
              <div class="stat-value">¥{{ formatMoney(stats.total_amount) }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card class="stat-card paid-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon">
              <el-icon><CircleCheck /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-label">已付款</div>
              <div class="stat-value">¥{{ formatMoney(stats.paid_amount) }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card class="stat-card pending-card" shadow="hover">
          <div class="stat-content">
            <div class="stat-icon">
              <el-icon><Clock /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-label">待付款</div>
              <div class="stat-value">¥{{ formatMoney(stats.pending_amount) }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 筛选条件 -->
    <el-card class="filter-card">
      <el-form :inline="true" :model="filterForm">
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
            @change="handleDateRangeChange"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadData">查询</el-button>
          <el-button @click="resetFilter">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 标签页 -->
    <el-tabs v-model="activeTab" @tab-change="handleTabChange">
      <!-- 已付款清单 -->
      <el-tab-pane label="已付款清单" name="paid">
        <el-card>
          <el-table :data="paidList" stripe border>
            <el-table-column type="expand">
              <template #default="scope">
                <el-table :data="scope.row.items" size="small" border>
                  <el-table-column prop="product_name" label="商品名称" />
                  <el-table-column prop="spec_name" label="规格" />
                  <el-table-column prop="quantity" label="数量" align="right" />
                  <el-table-column prop="cost_price" label="成本价" align="right">
                    <template #default="scope">
                      ¥{{ formatMoney(scope.row.cost_price) }}
                    </template>
                  </el-table-column>
                  <el-table-column prop="subtotal" label="小计" align="right">
                    <template #default="scope">
                      <strong>¥{{ formatMoney(scope.row.subtotal) }}</strong>
                    </template>
                  </el-table-column>
                </el-table>
              </template>
            </el-table-column>
            <el-table-column prop="payment_date" label="付款日期" width="120" />
            <el-table-column prop="payment_amount" label="付款金额" align="right" width="150">
              <template #default="scope">
                <strong class="paid-amount">¥{{ formatMoney(scope.row.payment_amount) }}</strong>
              </template>
            </el-table-column>
            <el-table-column prop="payment_method" label="付款方式" width="120">
              <template #default="scope">
                {{ formatPaymentMethod(scope.row.payment_method) }}
              </template>
            </el-table-column>
            <el-table-column prop="payment_account" label="付款账户" width="150" />
            <el-table-column prop="created_at" label="创建时间" width="180" />
            <el-table-column prop="remark" label="备注" show-overflow-tooltip />
          </el-table>
        </el-card>
      </el-tab-pane>

      <!-- 待付款清单 -->
      <el-tab-pane label="待付款清单" name="pending">
        <el-card>
          <el-table :data="pendingList" stripe border>
            <el-table-column type="expand">
              <template #default="scope">
                <el-table :data="scope.row.items" size="small" border>
                  <el-table-column prop="product_name" label="商品名称" />
                  <el-table-column prop="spec_name" label="规格" />
                  <el-table-column prop="quantity" label="数量" align="right" />
                  <el-table-column prop="cost_price" label="成本价" align="right">
                    <template #default="scope">
                      ¥{{ formatMoney(scope.row.cost_price) }}
                    </template>
                  </el-table-column>
                  <el-table-column prop="subtotal" label="小计" align="right">
                    <template #default="scope">
                      <strong>¥{{ formatMoney(scope.row.subtotal) }}</strong>
                    </template>
                  </el-table-column>
                </el-table>
              </template>
            </el-table-column>
            <el-table-column prop="pickup_date" label="取货日期" width="180" />
            <el-table-column prop="total_cost" label="应付款金额" align="right" width="150">
              <template #default="scope">
                <strong class="pending-amount">¥{{ formatMoney(scope.row.total_cost) }}</strong>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <!-- 产品统计 -->
      <el-tab-pane label="产品统计" name="products">
        <el-card>
          <el-table :data="productStats" stripe border>
            <el-table-column prop="product_name" label="商品名称" />
            <el-table-column prop="spec_name" label="规格" />
            <el-table-column prop="total_quantity" label="总数量" align="right" />
            <el-table-column prop="paid_quantity" label="已付款数量" align="right" />
            <el-table-column prop="pending_quantity" label="待付款数量" align="right" />
            <el-table-column prop="total_amount" label="总金额" align="right" width="150">
              <template #default="scope">
                <strong>¥{{ formatMoney(scope.row.total_amount) }}</strong>
              </template>
            </el-table-column>
            <el-table-column prop="paid_amount" label="已付款金额" align="right" width="150">
              <template #default="scope">
                <strong class="paid-amount">¥{{ formatMoney(scope.row.paid_amount) }}</strong>
              </template>
            </el-table-column>
            <el-table-column prop="pending_amount" label="待付款金额" align="right" width="150">
              <template #default="scope">
                <strong class="pending-amount">¥{{ formatMoney(scope.row.pending_amount) }}</strong>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Money, CircleCheck, Clock } from '@element-plus/icons-vue'
import { getPaidItems, getPendingItems, getPaymentStats } from '../api/payments'

const activeTab = ref('paid')
const dateRange = ref([])
const stats = ref({
  total_amount: 0,
  paid_amount: 0,
  pending_amount: 0
})
const paidList = ref([])
const pendingList = ref([])
const productStats = ref([])

const filterForm = reactive({
  start_date: '',
  end_date: ''
})

// 格式化金额
const formatMoney = (amount) => {
  if (!amount) return '0.00'
  return parseFloat(amount).toFixed(2)
}

// 格式化付款方式
const formatPaymentMethod = (method) => {
  const methodMap = {
    'bank_transfer': '银行转账',
    'cash': '现金',
    'alipay': '支付宝',
    'wechat': '微信'
  }
  return methodMap[method] || method || '-'
}

// 日期范围改变
const handleDateRangeChange = (dates) => {
  if (dates && dates.length === 2) {
    filterForm.start_date = dates[0]
    filterForm.end_date = dates[1]
  } else {
    filterForm.start_date = ''
    filterForm.end_date = ''
  }
}

// 标签页切换
const handleTabChange = (tabName) => {
  loadData()
}

// 加载数据
const loadData = async () => {
  try {
    const params = {}
    if (filterForm.start_date) {
      params.start_date = filterForm.start_date
    }
    if (filterForm.end_date) {
      params.end_date = filterForm.end_date
    }

    // 加载统计
    const statsRes = await getPaymentStats(params)
    if (statsRes.code === 200) {
      stats.value = statsRes.data || stats.value
    }

    // 根据当前标签页加载对应数据
    if (activeTab.value === 'paid') {
      const paidRes = await getPaidItems(params)
      if (paidRes.code === 200) {
        paidList.value = paidRes.data || []
      }
    } else if (activeTab.value === 'pending') {
      const pendingRes = await getPendingItems(params)
      if (pendingRes.code === 200) {
        pendingList.value = pendingRes.data?.orders || []
      }
    } else if (activeTab.value === 'products') {
      // 从已付款和待付款数据中统计产品
      await loadProductStats()
    }
  } catch (error) {
    console.error('加载数据失败:', error)
    ElMessage.error('加载数据失败')
  }
}

// 加载产品统计
const loadProductStats = async () => {
  try {
    const params = {}
    if (filterForm.start_date) {
      params.start_date = filterForm.start_date
    }
    if (filterForm.end_date) {
      params.end_date = filterForm.end_date
    }

    // 获取已付款和待付款数据
    const [paidRes, pendingRes] = await Promise.all([
      getPaidItems(params),
      getPendingItems(params)
    ])

    // 统计产品数据
    const productMap = new Map()

    // 处理已付款数据
    if (paidRes.code === 200 && paidRes.data) {
      paidRes.data.forEach(payment => {
        if (payment.items) {
          payment.items.forEach(item => {
            const key = `${item.product_id}_${item.spec_name || ''}`
            if (!productMap.has(key)) {
              productMap.set(key, {
                product_id: item.product_id,
                product_name: item.product_name,
                spec_name: item.spec_name || '',
                total_quantity: 0,
                paid_quantity: 0,
                pending_quantity: 0,
                total_amount: 0,
                paid_amount: 0,
                pending_amount: 0
              })
            }
            const stat = productMap.get(key)
            stat.paid_quantity += item.quantity
            stat.paid_amount += item.subtotal
            stat.total_quantity += item.quantity
            stat.total_amount += item.subtotal
          })
        }
      })
    }

    // 处理待付款数据
    if (pendingRes.code === 200 && pendingRes.data?.orders) {
      pendingRes.data.orders.forEach(order => {
        if (order.items) {
          order.items.forEach(item => {
            const key = `${item.product_id}_${item.spec_name || ''}`
            if (!productMap.has(key)) {
              productMap.set(key, {
                product_id: item.product_id,
                product_name: item.product_name,
                spec_name: item.spec_name || '',
                total_quantity: 0,
                paid_quantity: 0,
                pending_quantity: 0,
                total_amount: 0,
                paid_amount: 0,
                pending_amount: 0
              })
            }
            const stat = productMap.get(key)
            stat.pending_quantity += item.quantity
            stat.pending_amount += item.subtotal
            stat.total_quantity += item.quantity
            stat.total_amount += item.subtotal
          })
        }
      })
    }

    productStats.value = Array.from(productMap.values())
  } catch (error) {
    console.error('加载产品统计失败:', error)
    ElMessage.error('加载产品统计失败')
  }
}

// 重置筛选
const resetFilter = () => {
  dateRange.value = []
  filterForm.start_date = ''
  filterForm.end_date = ''
  loadData()
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.payments-container {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
  margin: 0;
}

.stats-row {
  margin-bottom: 20px;
}

.stat-card {
  transition: all 0.3s;
}

.stat-card:hover {
  transform: translateY(-4px);
}

.stat-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  color: #fff;
}

.total-card .stat-icon {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.paid-card .stat-icon {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.pending-card .stat-icon {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
}

.stat-info {
  flex: 1;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-bottom: 8px;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #303133;
}

.filter-card {
  margin-bottom: 20px;
}

.paid-amount {
  color: #67c23a;
}

.pending-amount {
  color: #e6a23c;
}
</style>

