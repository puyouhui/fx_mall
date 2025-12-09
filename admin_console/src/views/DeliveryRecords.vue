<template>
  <div class="delivery-records-page">
    <el-card class="delivery-records-card" shadow="never">
      <div class="card-header">
        <div class="title">
          <span class="main">配送记录</span>
          <span class="sub">查看所有配送完成记录</span>
        </div>
        <div class="actions">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索订单ID / 配送员员工码"
            clearable
            @keyup.enter="handleSearch"
            style="width: 200px; margin-right: 10px;"
          />
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
            style="width: 240px; margin-right: 10px;"
            @change="handleSearch"
          />
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </div>
      </div>

      <el-table
        v-loading="loading"
        :data="records"
        border
        stripe
        class="delivery-records-table"
        empty-text="暂无配送记录"
        row-key="id"
      >
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="order_id" label="订单ID" width="100" />
        <el-table-column label="订单编号" min-width="180">
          <template #default="scope">
            <el-button
              type="primary"
              link
              @click="handleViewOrder(scope.row.order_id)"
            >
              {{ scope.row.order_number || `订单#${scope.row.order_id}` }}
            </el-button>
          </template>
        </el-table-column>
        <el-table-column prop="delivery_employee_code" label="配送员员工码" width="140" />
        <el-table-column label="配送员姓名" width="120">
          <template #default="scope">
            {{ scope.row.delivery_employee_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="订单状态" width="120">
          <template #default="scope">
            <el-tag :type="getStatusType(scope.row.status)">
              {{ formatStatus(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="配送费" width="100" align="right">
          <template #default="scope">
            <span style="color: #f56c6c; font-weight: 600;">
              ¥{{ formatMoney(scope.row.delivery_fee || 0) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="completed_at" label="完成时间" min-width="160">
          <template #default="scope">
            {{ formatDateTime(scope.row.completed_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" min-width="160">
          <template #default="scope">
            {{ formatDateTime(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right" align="center">
          <template #default="scope">
            <el-button type="primary" link @click="handleViewDetail(scope.row)">
              详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.pageNum"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>

    <!-- 详情对话框 -->
    <el-dialog
      v-model="detailDialogVisible"
      title="配送记录详情"
      width="900px"
      :close-on-click-modal="false"
    >
      <div v-if="currentRecord" class="detail-content">
        <!-- 配送流程时间线 -->
        <el-divider content-position="left">配送流程</el-divider>
        <div class="process-timeline">
          <el-timeline>
            <el-timeline-item
              v-if="currentRecord.process_timeline?.created_at"
              timestamp=""
              placement="top"
              type="primary"
            >
              <el-card>
                <h4>订单创建</h4>
                <p>{{ formatDateTime(currentRecord.process_timeline.created_at) }}</p>
              </el-card>
            </el-timeline-item>
            <el-timeline-item
              v-if="currentRecord.process_timeline?.accepted_at"
              timestamp=""
              placement="top"
              type="success"
            >
              <el-card>
                <h4>配送员接单</h4>
                <p>{{ formatDateTime(currentRecord.process_timeline.accepted_at) }}</p>
                <p v-if="currentRecord.delivery_employee_name" style="color: #409eff; margin-top: 4px;">
                  配送员：{{ currentRecord.delivery_employee_name }} ({{ currentRecord.delivery_employee_code }})
                </p>
              </el-card>
            </el-timeline-item>
            <el-timeline-item
              v-if="currentRecord.process_timeline?.pickup_completed_at"
              timestamp=""
              placement="top"
              type="warning"
            >
              <el-card>
                <h4>取货完成</h4>
                <p>{{ formatDateTime(currentRecord.process_timeline.pickup_completed_at) }}</p>
              </el-card>
            </el-timeline-item>
            <el-timeline-item
              v-if="currentRecord.process_timeline?.delivering_started_at"
              timestamp=""
              placement="top"
              type="info"
            >
              <el-card>
                <h4>开始配送</h4>
                <p>{{ formatDateTime(currentRecord.process_timeline.delivering_started_at) }}</p>
              </el-card>
            </el-timeline-item>
            <el-timeline-item
              v-if="currentRecord.process_timeline?.delivering_completed_at"
              timestamp=""
              placement="top"
              type="success"
            >
              <el-card>
                <h4>配送完成</h4>
                <p>{{ formatDateTime(currentRecord.process_timeline.delivering_completed_at) }}</p>
              </el-card>
            </el-timeline-item>
          </el-timeline>
        </div>

        <!-- 基本信息 -->
        <el-divider content-position="left">基本信息</el-divider>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="记录ID">
            {{ currentRecord.id }}
          </el-descriptions-item>
          <el-descriptions-item label="订单ID">
            <el-button
              type="primary"
              link
              @click="handleViewOrder(currentRecord.order_id)"
            >
              {{ currentRecord.order_id }}
            </el-button>
          </el-descriptions-item>
          <el-descriptions-item label="订单编号">
            {{ currentRecord.order_number || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="订单创建时间">
            {{ currentRecord.order_created_at ? formatDateTime(currentRecord.order_created_at) : '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="配送员员工码">
            {{ currentRecord.delivery_employee_code }}
          </el-descriptions-item>
          <el-descriptions-item label="配送员姓名">
            {{ currentRecord.delivery_employee_name || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="配送费">
            <span style="color: #f56c6c; font-weight: 600;">¥{{ formatMoney(currentRecord.delivery_fee || 0) }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="完成时间">
            {{ formatDateTime(currentRecord.completed_at) }}
          </el-descriptions-item>
        </el-descriptions>

        <!-- 配送照片 -->
        <el-divider content-position="left">配送照片</el-divider>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="货物照片">
            <el-image
              v-if="currentRecord.product_image_url"
              :src="currentRecord.product_image_url"
              :preview-src-list="[currentRecord.product_image_url]"
              fit="cover"
              style="width: 200px; height: 200px; border-radius: 4px; cursor: pointer;"
              :preview-teleported="true"
            />
            <span v-else style="color: #909399;">无</span>
          </el-descriptions-item>
          <el-descriptions-item label="门牌照片">
            <el-image
              v-if="currentRecord.doorplate_image_url"
              :src="currentRecord.doorplate_image_url"
              :preview-src-list="[currentRecord.doorplate_image_url]"
              fit="cover"
              style="width: 200px; height: 200px; border-radius: 4px; cursor: pointer;"
              :preview-teleported="true"
            />
            <span v-else style="color: #909399;">无</span>
          </el-descriptions-item>
        </el-descriptions>

        <!-- 其他信息 -->
        <el-divider content-position="left">其他信息</el-divider>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="创建时间">
            {{ formatDateTime(currentRecord.created_at) }}
          </el-descriptions-item>
          <el-descriptions-item label="更新时间">
            {{ formatDateTime(currentRecord.updated_at) }}
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getDeliveryRecords, getDeliveryRecordByOrderId } from '../api/deliveryRecords'

const router = useRouter()

// 数据
const loading = ref(false)
const records = ref([])
const searchKeyword = ref('')
const dateRange = ref(null)
const detailDialogVisible = ref(false)
const currentRecord = ref(null)

// 分页
const pagination = reactive({
  pageNum: 1,
  pageSize: 20,
  total: 0
})

// 加载数据
const loadData = async () => {
  loading.value = true
  try {
    const params = {
      pageNum: pagination.pageNum,
      pageSize: pagination.pageSize
    }

    // 添加搜索关键词
    if (searchKeyword.value) {
      params.keyword = searchKeyword.value
    }

    // 添加日期范围
    if (dateRange.value && dateRange.value.length === 2) {
      params.startDate = dateRange.value[0]
      params.endDate = dateRange.value[1]
    }

    const res = await getDeliveryRecords(params)
    if (res.code === 200) {
      records.value = res.data?.list || []
      pagination.total = res.data?.total || 0
    } else {
      ElMessage.error(res.message || '获取配送记录失败')
    }
  } catch (error) {
    console.error('加载配送记录失败:', error)
    ElMessage.error('加载配送记录失败')
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  pagination.pageNum = 1
  loadData()
}

// 重置
const handleReset = () => {
  searchKeyword.value = ''
  dateRange.value = null
  pagination.pageNum = 1
  loadData()
}

// 分页变化
const handlePageChange = (page) => {
  pagination.pageNum = page
  loadData()
}

const handleSizeChange = (size) => {
  pagination.pageSize = size
  pagination.pageNum = 1
  loadData()
}

// 查看详情
const handleViewDetail = async (record) => {
  try {
    // 从后端获取完整详情（包括流程日志）
    const res = await getDeliveryRecordByOrderId(record.order_id)
    if (res.code === 200 && res.data) {
      currentRecord.value = res.data
      detailDialogVisible.value = true
    } else {
      // 如果获取失败，使用列表中的数据
      currentRecord.value = record
      detailDialogVisible.value = true
    }
  } catch (error) {
    console.error('获取配送记录详情失败:', error)
    // 如果获取失败，使用列表中的数据
    currentRecord.value = record
    detailDialogVisible.value = true
  }
}

// 查看订单
const handleViewOrder = (orderId) => {
  router.push({
    name: 'Orders',
    query: { orderId }
  })
}

// 格式化日期时间
const formatDateTime = (dateTime) => {
  if (!dateTime) return '-'
  const date = new Date(dateTime)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

// 格式化金额
const formatMoney = (amount) => {
  return (amount || 0).toFixed(2)
}

// 格式化订单状态
const formatStatus = (status) => {
  const statusMap = {
    'pending': '待配送',
    'pending_delivery': '待配送',
    'pending_pickup': '待取货',
    'delivering': '配送中',
    'delivered': '已送达',
    'paid': '已收款',
    'cancelled': '已取消'
  }
  return statusMap[status] || status || '-'
}

// 获取状态标签类型
const getStatusType = (status) => {
  const typeMap = {
    'pending': 'info',
    'pending_delivery': 'info',
    'pending_pickup': 'warning',
    'delivering': 'primary',
    'delivered': 'success',
    'paid': 'success',
    'cancelled': 'danger'
  }
  return typeMap[status] || 'info'
}

// 初始化
onMounted(() => {
  loadData()
})
</script>

<style scoped>
.delivery-records-page {
  padding: 20px;
}

.delivery-records-card {
  border-radius: 8px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 20px;
  border-bottom: 1px solid #ebeef5;
}

.title {
  display: flex;
  flex-direction: column;
}

.title .main {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 4px;
}

.title .sub {
  font-size: 14px;
  color: #909399;
}

.actions {
  display: flex;
  align-items: center;
}

.delivery-records-table {
  margin-top: 20px;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.detail-content {
  padding: 10px 0;
}
</style>

