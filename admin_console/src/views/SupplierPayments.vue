<template>
  <div class="supplier-payments-container">
    <el-card>
      <div class="page-header">
        <el-button v-if="supplierId" @click="goBack" :icon="ArrowLeft">返回</el-button>
        <h2 class="page-title">供应商付款统计</h2>
      </div>

      <!-- 筛选条件（仅在详情页显示） -->
      <div class="filter-section" v-if="supplierId">
        <el-form :inline="true" :model="filterForm">
          <el-form-item label="时间范围">
            <el-select v-model="filterForm.timeRange" @change="handleTimeRangeChange" style="width: 150px">
              <el-option label="全部" value="" />
              <el-option label="今日" value="today" />
              <el-option label="最近7天" value="week" />
              <el-option label="最近30天" value="month" />
            </el-select>
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="filterForm.status" @change="loadData" style="width: 150px">
              <el-option label="全部" value="" />
              <el-option label="待付款" value="pending" />
              <el-option label="已付款" value="paid" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="loadData">查询</el-button>
            <el-button @click="resetFilter">重置</el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- 统计列表表格 -->
      <el-card v-if="!supplierId" class="stats-table-card">
        <el-table :data="statsList" stripe border v-loading="loading" empty-text="暂无数据">
          <el-table-column prop="supplier_id" label="供应商ID" width="100" align="center" />
          <el-table-column prop="supplier_name" label="供应商名称" min-width="150" />
          <el-table-column prop="total_amount" label="应付款总额" align="right" width="150">
            <template #default="scope">
              <span class="total-amount">¥{{ formatMoney(scope.row.total_amount) }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="order_count" label="订单数量" width="120" align="center" />
          <el-table-column prop="pending_amount" label="待付款" align="right" width="150">
            <template #default="scope">
              <span class="pending-amount">¥{{ formatMoney(scope.row.pending_amount) }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="paid_amount" label="已付款" align="right" width="150">
            <template #default="scope">
              <span class="paid-amount">¥{{ formatMoney(scope.row.paid_amount) }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="payment_status" label="付款状态" width="120" align="center">
            <template #default="scope">
              <el-tag :type="getPaymentStatusType(scope.row.payment_status)">
                {{ formatPaymentStatus(scope.row.payment_status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120" align="center" fixed="right">
            <template #default="scope">
              <el-button type="primary" size="small" @click="viewDetail(scope.row.supplier_id)">
                查看详情
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

      <!-- 详细清单 -->
      <el-card v-if="supplierId && detailData" class="detail-card">
        <template #header>
          <div class="detail-header">
            <h3>{{ detailData.supplier_name }} - 付款清单</h3>
            <div class="header-actions">
              <div class="header-stats">
                <span>总金额：<strong class="total">¥{{ formatMoney(detailData.total_amount) }}</strong></span>
                <span>订单数：<strong>{{ detailData.order_count }}</strong></span>
              </div>
              <el-button type="primary" @click="showPaymentDialog">
                <el-icon><Plus /></el-icon>
                标记付款
              </el-button>
            </div>
          </div>
        </template>

        <el-table :data="detailData.orders" stripe @selection-change="handleSelectionChange">
          <el-table-column type="selection" width="55" />
          <el-table-column type="expand">
            <template #default="scope">
              <el-table :data="scope.row.items" size="small" border>
                <el-table-column type="selection" width="55" />
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
                <el-table-column prop="is_paid" label="付款状态" width="100" align="center">
                  <template #default="scope">
                    <el-tag :type="scope.row.is_paid ? 'success' : 'danger'" size="small">
                      {{ scope.row.is_paid ? '已付款' : '待付款' }}
                    </el-tag>
                  </template>
                </el-table-column>
              </el-table>
            </template>
          </el-table-column>
          <el-table-column prop="order_id" label="订单ID" width="120" align="center" />
          <el-table-column prop="address_name" label="客户地址名称" min-width="150" />
          <el-table-column prop="order_date" label="订单日期" width="180" />
          <el-table-column prop="pickup_date" label="取货日期" width="180" />
          <el-table-column prop="status" label="订单状态" width="100">
            <template #default="scope">
              <el-tag :type="getStatusType(scope.row.status)">
                {{ formatStatus(scope.row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="payment_status" label="付款状态" width="120" align="center">
            <template #default="scope">
              <el-tag :type="getPaymentStatusType(scope.row.payment_status)">
                {{ formatPaymentStatus(scope.row.payment_status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="total_cost" label="金额" align="right" width="150">
            <template #default="scope">
              <strong class="total">¥{{ formatMoney(scope.row.total_cost) }}</strong>
            </template>
          </el-table-column>
        </el-table>

        <!-- 选中清单小计 -->
        <div class="selected-summary" v-if="selectedOrders.length > 0">
          <div class="summary-content">
            <span class="summary-label">已选中 <strong>{{ selectedOrders.length }}</strong> 个订单</span>
            <span class="summary-amount">
              小计：<strong class="total">¥{{ formatMoney(selectedTotalAmount) }}</strong>
            </span>
          </div>
        </div>

        <!-- 分页 -->
        <div class="pagination-container">
          <el-pagination
            v-model:current-page="detailPagination.pageNum"
            v-model:page-size="detailPagination.pageSize"
            :page-sizes="[10, 20, 50, 100]"
            :total="detailPagination.total"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="handleDetailSizeChange"
            @current-change="handleDetailPageChange"
          />
        </div>
      </el-card>

      <!-- 付款标记对话框 -->
      <el-dialog v-model="paymentDialogVisible" title="标记付款" width="800px">
        <el-form :model="paymentForm" :rules="paymentRules" ref="paymentFormRef" label-width="120px">
          <el-form-item label="供应商" prop="supplier_id">
            <el-input :value="detailData?.supplier_name" disabled />
          </el-form-item>
          <el-form-item label="付款日期" prop="payment_date">
            <el-date-picker
              v-model="paymentForm.payment_date"
              type="date"
              placeholder="选择付款日期"
              format="YYYY-MM-DD"
              value-format="YYYY-MM-DD"
              style="width: 100%"
            />
          </el-form-item>
          <el-form-item label="付款方式" prop="payment_method">
            <el-select v-model="paymentForm.payment_method" placeholder="选择付款方式" style="width: 100%">
              <el-option label="银行转账" value="bank_transfer" />
              <el-option label="现金" value="cash" />
              <el-option label="支付宝" value="alipay" />
              <el-option label="微信" value="wechat" />
            </el-select>
          </el-form-item>
          <el-form-item label="付款账户">
            <el-input v-model="paymentForm.payment_account" placeholder="请输入付款账户" />
          </el-form-item>
          <el-form-item label="付款凭证">
            <el-input v-model="paymentForm.payment_receipt" placeholder="请输入付款凭证URL" />
          </el-form-item>
          <el-form-item label="备注">
            <el-input v-model="paymentForm.remark" type="textarea" :rows="3" placeholder="请输入备注" />
          </el-form-item>
          <el-form-item label="选中商品">
            <el-table :data="selectedItems" border size="small" max-height="300">
              <el-table-column prop="product_name" label="商品名称" />
              <el-table-column prop="spec_name" label="规格" />
              <el-table-column prop="quantity" label="数量" align="right" />
              <el-table-column prop="cost_price" label="成本价" align="right">
                <template #default="scope">¥{{ formatMoney(scope.row.cost_price) }}</template>
              </el-table-column>
              <el-table-column prop="subtotal" label="小计" align="right">
                <template #default="scope">
                  <strong>¥{{ formatMoney(scope.row.subtotal) }}</strong>
                </template>
              </el-table-column>
            </el-table>
            <div class="total-amount">
              <strong>合计金额：¥{{ formatMoney(calculatedAmount) }}</strong>
            </div>
          </el-form-item>
          <el-form-item label="付款金额" prop="payment_amount">
            <el-input-number
              v-model="paymentForm.payment_amount"
              :precision="2"
              :min="0"
              style="width: 100%"
            />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="paymentDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="submitPayment" :loading="submitting">确认付款</el-button>
        </template>
      </el-dialog>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Plus } from '@element-plus/icons-vue'
import { getSupplierPaymentsStats, getSupplierPaymentDetail, createSupplierPayment } from '../api/suppliers'

const route = useRoute()
const router = useRouter()

// 供应商ID（用于显示详细清单）
// 从菜单进入时 supplierId 为 null，显示所有供应商的统计列表
// 点击"查看详情"时设置 supplierId，显示该供应商的详细清单
const supplierId = ref(null)
const statsList = ref([])
const detailData = ref(null)
const selectedOrders = ref([])
const paymentDialogVisible = ref(false)
const submitting = ref(false)
const paymentFormRef = ref(null)
const loading = ref(false)

// 分页
// 统计列表分页
const pagination = reactive({
  pageNum: 1,
  pageSize: 20,
  total: 0
})

// 详情列表分页
const detailPagination = reactive({
  pageNum: 1,
  pageSize: 20,
  total: 0
})

const filterForm = reactive({
  timeRange: '', // 统计列表默认显示全部时间
  status: ''
})

const paymentForm = reactive({
  supplier_id: null,
  payment_date: new Date().toISOString().split('T')[0],
  payment_amount: 0,
  payment_method: '',
  payment_account: '',
  payment_receipt: '',
  remark: ''
})

const paymentRules = {
  payment_date: [{ required: true, message: '请选择付款日期', trigger: 'change' }],
  payment_method: [{ required: true, message: '请选择付款方式', trigger: 'change' }],
  payment_amount: [{ required: true, message: '请输入付款金额', trigger: 'blur' }]
}

// 选中的商品列表
const selectedItems = computed(() => {
  const items = []
  selectedOrders.value.forEach(order => {
    if (order.items && order.items.length > 0) {
      order.items.forEach(item => {
        items.push({
          order_id: order.order_id,
          order_item_id: item.order_item_id || item.product_id, // 如果没有order_item_id，使用product_id作为临时标识
          product_id: item.product_id,
          product_name: item.product_name,
          spec_name: item.spec_name,
          quantity: item.quantity,
          cost_price: item.cost_price,
          subtotal: item.subtotal
        })
      })
    }
  })
  return items
})

// 计算总金额
const calculatedAmount = computed(() => {
  return selectedItems.value.reduce((sum, item) => sum + (item.subtotal || 0), 0)
})

// 选中订单的小计金额
const selectedTotalAmount = computed(() => {
  return selectedOrders.value.reduce((sum, order) => sum + (Number(order.total_cost) || 0), 0)
})

// 格式化金额
const formatMoney = (amount) => {
  if (!amount) return '0.00'
  return parseFloat(amount).toFixed(2)
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
    'completed': '已收款',
    'cancelled': '已取消',
    'shipped': '已送达'
  }
  return statusMap[status] || status || '-'
}

// 获取状态标签类型
const getStatusType = (status) => {
  const typeMap = {
    'pending': 'danger',
    'pending_delivery': 'danger',
    'pending_pickup': 'warning',
    'delivering': 'primary',
    'delivered': 'warning',
    'shipped': 'warning',
    'paid': 'success',
    'completed': 'success',
    'cancelled': 'info'
  }
  return typeMap[status] || 'info'
}

// 格式化付款状态
const formatPaymentStatus = (status) => {
  const statusMap = {
    'pending': '待付款',
    'partial': '部分已付款',
    'paid': '已付款'
  }
  return statusMap[status] || status || '-'
}

// 获取付款状态标签类型
const getPaymentStatusType = (status) => {
  const typeMap = {
    'pending': 'danger',
    'partial': 'warning',
    'paid': 'success'
  }
  return typeMap[status] || 'info'
}

// 返回
const goBack = () => {
  if (supplierId.value) {
    // 如果是在详情页，返回列表页
    supplierId.value = null
    loadData()
  } else {
    // 如果是在列表页，返回供应商管理页面
    router.push('/suppliers')
  }
}

// 加载数据
const loadData = async () => {
  loading.value = true
  try {
    if (supplierId.value) {
      // 加载详细清单
      const params = {
        page: detailPagination.pageNum,
        page_size: detailPagination.pageSize
      }
      if (filterForm.status) {
        params.status = filterForm.status
      }
      if (filterForm.timeRange) {
        params.time_range = filterForm.timeRange
      }
      const response = await getSupplierPaymentDetail(supplierId.value, params)
      if (response.code === 200) {
        detailData.value = response.data
        // 更新分页信息
        if (response.data.total !== undefined) {
          detailPagination.total = response.data.total || 0
        } else {
          detailPagination.total = response.data.orders?.length || 0
        }
        // 为每个订单添加order_item_id（如果缺少）
        if (detailData.value.orders) {
          detailData.value.orders.forEach(order => {
            if (order.items) {
              order.items.forEach((item, index) => {
                if (!item.order_item_id) {
                  // 如果没有order_item_id，使用临时标识（实际应该从后端返回）
                  item.order_item_id = `${order.order_id}_${item.product_id}_${index}`
                }
              })
            }
          })
        }
      } else {
        ElMessage.error(response.message || '加载失败')
      }
    } else {
      // 加载统计列表
      const response = await getSupplierPaymentsStats({
        time_range: filterForm.timeRange,
        status: filterForm.status,
        page: pagination.pageNum,
        page_size: pagination.pageSize
      })
      if (response.code === 200) {
        if (response.data && response.data.list) {
          statsList.value = response.data.list || []
          pagination.total = response.data.total || 0
        } else {
          // 兼容旧格式
          statsList.value = response.data || []
          pagination.total = statsList.value.length
        }
      } else {
        ElMessage.error(response.message || '加载失败')
      }
    }
  } catch (error) {
    console.error('加载数据失败:', error)
    ElMessage.error('加载数据失败')
  } finally {
    loading.value = false
  }
}

// 统计列表分页变化
const handlePageChange = (page) => {
  pagination.pageNum = page
  loadData()
}

const handleSizeChange = (size) => {
  pagination.pageSize = size
  pagination.pageNum = 1
  loadData()
}

// 详情列表分页变化
const handleDetailPageChange = (page) => {
  detailPagination.pageNum = page
  loadData()
}

const handleDetailSizeChange = (size) => {
  detailPagination.pageSize = size
  detailPagination.pageNum = 1
  loadData()
}

// 查看详情
const viewDetail = (id) => {
  supplierId.value = id
  detailPagination.pageNum = 1 // 重置到第一页
  loadData()
}

// 时间范围改变
const handleTimeRangeChange = () => {
  pagination.pageNum = 1
  loadData()
}

// 重置筛选
const resetFilter = () => {
  filterForm.timeRange = '' // 重置为全部时间
  filterForm.status = ''
  pagination.pageNum = 1
  loadData()
}

// 选择变化
const handleSelectionChange = (selection) => {
  selectedOrders.value = selection
}

// 显示付款对话框
const showPaymentDialog = () => {
  if (selectedOrders.value.length === 0) {
    ElMessage.warning('请先选择要付款的订单')
    return
  }
  
  // 收集所有选中的商品
  const allItems = []
  selectedOrders.value.forEach(order => {
    if (order.items && order.items.length > 0) {
      order.items.forEach(item => {
        allItems.push({
          order_id: order.order_id,
          order_item_id: item.order_item_id || `${order.order_id}_${item.product_id}_${item.spec_name}`, // 临时标识
          product_id: item.product_id,
          product_name: item.product_name,
          spec_name: item.spec_name,
          quantity: item.quantity,
          cost_price: item.cost_price,
          subtotal: item.subtotal
        })
      })
    }
  })
  
  if (allItems.length === 0) {
    ElMessage.warning('选中的订单中没有商品')
    return
  }
  
  paymentForm.supplier_id = supplierId.value
  paymentForm.payment_amount = calculatedAmount.value
  paymentDialogVisible.value = true
}

// 提交付款
const submitPayment = async () => {
  if (!paymentFormRef.value) return
  
  try {
    await paymentFormRef.value.validate()
    
    if (selectedItems.value.length === 0) {
      ElMessage.warning('请选择要付款的商品')
      return
    }
    
    // 需要从订单中获取真实的order_item_id
    // 这里需要重新查询订单详情以获取order_item_id
    const orderItems = []
    for (const order of selectedOrders.value) {
      const orderDetail = detailData.value.orders.find(o => o.order_id === order.order_id)
      if (orderDetail && orderDetail.items) {
        orderDetail.items.forEach(item => {
          orderItems.push({
            order_id: order.order_id,
            order_item_id: item.order_item_id || item.product_id, // 需要确保有order_item_id
            product_id: item.product_id,
            product_name: item.product_name,
            spec_name: item.spec_name || '',
            quantity: item.quantity,
            cost_price: item.cost_price,
            subtotal: item.subtotal
          })
        })
      }
    }
    
    if (orderItems.length === 0) {
      ElMessage.warning('无法获取商品信息，请刷新页面后重试')
      return
    }
    
    submitting.value = true
    
    const response = await createSupplierPayment({
      supplier_id: paymentForm.supplier_id,
      payment_date: paymentForm.payment_date,
      payment_amount: paymentForm.payment_amount,
      payment_method: paymentForm.payment_method || null,
      payment_account: paymentForm.payment_account || null,
      payment_receipt: paymentForm.payment_receipt || null,
      remark: paymentForm.remark || null,
      order_items: orderItems
    })
    
    if (response.code === 200) {
      ElMessage.success('付款记录创建成功')
      paymentDialogVisible.value = false
      // 重置表单
      paymentFormRef.value?.resetFields()
      selectedOrders.value = []
      // 重新加载数据
      await loadData()
    } else {
      // 业务逻辑错误，显示后端返回的错误信息
      ElMessage.error(response.message || '创建付款记录失败')
    }
  } catch (error) {
    console.error('提交付款失败:', error)
    // 提取后端返回的错误信息，优先显示后端返回的具体错误
    let errorMessage = '提交付款失败'
    if (error?.response?.data?.message) {
      // HTTP 错误响应中的错误信息
      errorMessage = error.response.data.message
    } else if (error?.response?.data?.error) {
      errorMessage = error.response.data.error
    } else if (error?.response?.data) {
      // 如果 data 是字符串，直接使用
      errorMessage = typeof error.response.data === 'string' ? error.response.data : (error.response.data.message || '提交付款失败')
    } else if (error?.message) {
      errorMessage = error.message
    } else if (typeof error === 'string') {
      errorMessage = error
    }
    ElMessage.error(errorMessage)
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  console.log('页面加载，supplierId:', supplierId.value, '路由路径:', route.path, '路由参数:', route.params)
  loadData()
})
</script>

<style scoped>
.supplier-payments-container {
  padding: 20px;
}

.page-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 20px;
}

.page-title {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
}

.filter-section {
  margin-bottom: 20px;
  padding: 16px;
  background: #f5f7fa;
  border-radius: 8px;
}

.stats-cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 20px;
  margin-bottom: 20px;
}

.stat-card {
  transition: all 0.3s;
}

.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
}

.stat-content {
  padding: 8px;
}

.stat-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #e4e7ed;
}

.stat-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.stat-info {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.stat-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.stat-item .label {
  color: #606266;
  font-size: 14px;
}

.stat-item .value {
  font-size: 16px;
  font-weight: 600;
}

.stat-item .value.total {
  color: #409eff;
  font-size: 20px;
}

.stat-item .value.pending {
  color: #e6a23c;
}

.stat-item .value.paid {
  color: #67c23a;
}

.detail-card {
  margin-top: 20px;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  gap: 20px;
}

.detail-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #303133;
  flex-shrink: 0;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 24px;
  flex-shrink: 0;
}

.header-stats {
  display: flex;
  align-items: center;
  gap: 24px;
  font-size: 14px;
  color: #606266;
}

.header-stats span {
  display: flex;
  align-items: center;
  white-space: nowrap;
}

.header-stats strong {
  font-size: 16px;
  font-weight: 600;
  margin-left: 4px;
}

.header-stats .total {
  color: #409eff;
  font-size: 20px;
  font-weight: 600;
}

.total-amount {
  margin-top: 12px;
  text-align: right;
  font-size: 16px;
  color: #409eff;
}

.stats-table-card {
  margin-bottom: 20px;
}

.stats-table-card .pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

/* 选中清单小计 */
.selected-summary {
  margin-top: 16px;
  padding: 16px 20px;
  background: linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 100%);
  border: 1px solid #bae6fd;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.1);
}

.summary-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 14px;
  color: #606266;
}

.summary-label {
  display: flex;
  align-items: center;
  gap: 4px;
}

.summary-label strong {
  color: #409eff;
  font-weight: 600;
}

.summary-amount {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  color: #303133;
}

.summary-amount .total {
  color: #409eff;
  font-size: 20px;
  font-weight: 600;
}

/* 待付款和已付款颜色 */
.pending-amount {
  color: #e6a23c;
  font-weight: 600;
}

.paid-amount {
  color: #67c23a;
  font-weight: 600;
}
</style>

