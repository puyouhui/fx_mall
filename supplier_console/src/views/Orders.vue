<template>
  <div class="orders-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <div class="header-title">
            <span>订单管理</span>
            <el-tag type="info" size="small" style="margin-left: 10px;">
              共 {{ pagination.total }} 单
            </el-tag>
          </div>
          <div class="header-actions">
            <el-radio-group v-model="statusFilter" @change="handleStatusChange" size="default">
              <el-radio-button label="">全部</el-radio-button>
              <el-radio-button label="pending_pickup">待取货</el-radio-button>
              <el-radio-button label="picked">已取货</el-radio-button>
            </el-radio-group>
          </div>
        </div>
      </template>

      <el-table :data="orders" v-loading="loading" stripe>
        <el-table-column prop="user_code" label="客户编号" align="center">
          <template #default="scope">
            <span>{{ scope.row.user_code || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="order_number" label="订单号" align="center" />
        <el-table-column label="状态" align="center">
          <template #default="scope">
            <el-tag :type="scope.row.status === '已取货' ? 'success' : 'warning'">
              {{ scope.row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="商品数量" align="center">
          <template #default="scope">
            {{ scope.row.item_count || 0 }} 件
          </template>
        </el-table-column>
        <el-table-column label="成本总额" align="center">
          <template #default="scope">
            <span class="cost-price">¥{{ formatPrice(scope.row.total_cost) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" align="center">
          <template #default="scope">
            {{ formatDateTime(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" align="center" fixed="right">
          <template #default="scope">
            <el-button type="primary" size="small" link @click="handleViewDetail(scope.row)">
              查看详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination v-model:current-page="pagination.page" v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]" :total="pagination.total" layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange" @current-change="handlePageChange" />
      </div>
    </el-card>

    <!-- 订单详情抽屉 -->
    <el-drawer v-model="detailDrawerVisible" title="订单详情" :size="700" direction="rtl">
      <div v-if="currentOrder" class="order-detail">
        <!-- 订单基本信息 -->
        <div class="detail-section">
          <h3>订单信息</h3>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="客户编号">{{ currentOrder.user_code || '-' }}</el-descriptions-item>
            <el-descriptions-item label="订单号">{{ currentOrder.order_number }}</el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="currentOrder.status === '已取货' ? 'success' : 'warning'">
                {{ currentOrder.status }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="商品数量">{{ currentOrder.item_count }} 件</el-descriptions-item>
            <el-descriptions-item label="成本总额">
              <span class="cost-price">¥{{ formatPrice(currentOrder.total_cost) }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="创建时间">
              {{ formatDateTime(currentOrder.created_at) }}
            </el-descriptions-item>
            <el-descriptions-item label="更新时间">
              {{ formatDateTime(currentOrder.updated_at) }}
            </el-descriptions-item>
          </el-descriptions>
        </div>

        <!-- 商品明细 -->
        <div v-if="currentOrder.items && currentOrder.items.length > 0" class="detail-section">
          <h3>商品明细</h3>
          <el-table :data="currentOrder.items" border style="width: 100%">
            <el-table-column label="商品图片" width="100" align="center">
              <template #default="scope">
                <el-image v-if="scope.row.image" :src="scope.row.image" fit="cover"
                  style="width: 60px; height: 60px; border-radius: 4px;" :preview-src-list="[scope.row.image]"
                  :preview-teleported="true" />
                <span v-else class="no-image">暂无图片</span>
              </template>
            </el-table-column>
            <el-table-column prop="product_name" label="商品名称" min-width="150" align="center" show-overflow-tooltip />
            <el-table-column prop="spec_name" label="规格" width="120" align="center" />
            <el-table-column label="数量" width="80" align="center">
              <template #default="scope">
                {{ scope.row.quantity }}
              </template>
            </el-table-column>
            <el-table-column label="价格" width="120" align="center">
              <template #default="scope">
                <span class="cost-price">¥{{ formatPrice(scope.row.cost_price) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="小计" width="120" align="center">
              <template #default="scope">
                <span class="cost-price">¥{{ formatPrice(scope.row.item_cost) }}</span>
              </template>
            </el-table-column>
          </el-table>
        </div>
        <div v-else class="empty-data">暂无商品明细</div>
      </div>
      <div v-else class="loading-container">
        <el-skeleton :rows="5" animated />
      </div>
    </el-drawer>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getOrders, getOrderDetail } from '../api/orders'

const loading = ref(false)
const orders = ref([])
const statusFilter = ref('')
const detailDrawerVisible = ref(false)
const currentOrder = ref(null)

const pagination = ref({
  page: 1,
  pageSize: 20,
  total: 0
})

// 加载订单列表
const loadOrders = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.value.page,
      page_size: pagination.value.pageSize
    }

    if (statusFilter.value) {
      params.status = statusFilter.value
    }

    const response = await getOrders(params)

    if (response.code === 200 && response.data) {
      orders.value = response.data.list || []
      pagination.value.total = response.data.total || 0
    } else {
      ElMessage.error(response.message || '获取订单列表失败')
    }
  } catch (error) {
    console.error('获取订单列表失败:', error)
    ElMessage.error(error.message || '获取订单列表失败，请稍后再试')
  } finally {
    loading.value = false
  }
}

// 状态筛选改变
const handleStatusChange = () => {
  pagination.value.page = 1
  loadOrders()
}

// 查看订单详情
const handleViewDetail = async (order) => {
  detailDrawerVisible.value = true
  currentOrder.value = null

  try {
    const response = await getOrderDetail(order.id)
    if (response.code === 200 && response.data) {
      currentOrder.value = response.data
    } else {
      ElMessage.error(response.message || '获取订单详情失败')
      // 如果获取详情失败，使用列表中的基本信息
      currentOrder.value = {
        ...order,
        items: []
      }
    }
  } catch (error) {
    console.error('获取订单详情失败:', error)
    ElMessage.error(error.message || '获取订单详情失败，请稍后再试')
    // 如果获取详情失败，使用列表中的基本信息
    currentOrder.value = {
      ...order,
      items: []
    }
  }
}

// 分页大小改变
const handleSizeChange = (size) => {
  pagination.value.pageSize = size
  pagination.value.page = 1
  loadOrders()
}

// 页码改变
const handlePageChange = (page) => {
  pagination.value.page = page
  loadOrders()
}

// 格式化价格
const formatPrice = (price) => {
  if (price === undefined || price === null) return '0.00'
  return Number(price).toFixed(2)
}

// 格式化日期时间
const formatDateTime = (dateTime) => {
  if (!dateTime) return '-'
  const date = new Date(dateTime)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}`
}

onMounted(() => {
  loadOrders()
})
</script>

<style scoped>
.orders-page {
  padding: 0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-title {
  display: flex;
  align-items: center;
}

.header-actions {
  display: flex;
  align-items: center;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.cost-price {
  color: #409eff;
  font-weight: 500;
}

.order-detail {
  padding: 0;
}

.detail-section {
  margin-bottom: 30px;
}

.detail-section h3 {
  margin-bottom: 15px;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  border-left: 4px solid #409eff;
  padding-left: 10px;
}

.no-image {
  color: #909399;
  font-size: 12px;
}

.empty-data {
  text-align: center;
  padding: 40px;
  color: #909399;
  font-size: 16px;
}

.loading-container {
  padding: 20px;
}
</style>
