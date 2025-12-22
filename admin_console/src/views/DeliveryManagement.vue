<template>
  <div class="delivery-management-page">
    <el-card class="delivery-card" shadow="never">
      <div class="card-header">
        <div class="title">
          <span class="main">配送管理</span>
          <span class="sub">查看配送员的订单信息</span>
        </div>
        <div class="date-filter">
          <div class="quick-date-buttons">
            <el-button size="small" :type="activeQuickDate === 'today' ? 'primary' : 'default'"
              @click="handleQuickDateSelect('today')">
              今日
            </el-button>
            <el-button size="small" :type="activeQuickDate === '3days' ? 'primary' : 'default'"
              @click="handleQuickDateSelect('3days')">
              近3日
            </el-button>
            <el-button size="small" :type="activeQuickDate === '7days' ? 'primary' : 'default'"
              @click="handleQuickDateSelect('7days')">
              近7日
            </el-button>
          </div>
          <el-date-picker v-model="dateRange" type="daterange" range-separator="至" start-placeholder="开始日期"
            end-placeholder="结束日期" format="YYYY-MM-DD" value-format="YYYY-MM-DD" @change="handleDateChange"
            style="width: 300px; margin-left: 10px;" />
          <el-button type="primary" @click="handleSearch" style="margin-left: 10px;">查询</el-button>
        </div>
      </div>

      <div class="content-wrapper">
        <!-- 左侧配送员列表 -->
        <div class="delivery-list-panel">
          <div class="panel-header">
            <span class="panel-title">配送员列表</span>
          </div>
          <div class="panel-content">
            <el-radio-group v-model="selectedDeliveryId" @change="handleDeliverySelectionChange">
              <div v-for="employee in deliveryEmployees" :key="employee.id" class="delivery-item"
                :class="{ 'is-selected': selectedDeliveryId === employee.id }">
                <el-radio :label="employee.id" class="delivery-radio">
                  <div class="delivery-info">
                    <div class="delivery-name">{{ employee.name || employee.employee_code }}</div>
                    <div class="delivery-code">员工码: {{ employee.employee_code }}</div>
                    <!-- <div class="delivery-phone">手机: {{ employee.phone }}</div> -->
                  </div>
                </el-radio>
              </div>
            </el-radio-group>
            <el-empty v-if="deliveryEmployees.length === 0" description="暂无配送员" :image-size="100" />
          </div>
        </div>

        <!-- 右侧订单列表 -->
        <div class="orders-panel">
          <div class="panel-header">
            <span class="panel-title">
              订单列表
              <span v-if="selectedDeliveryId" class="selected-count">
                (已选择 1 位配送员)
              </span>
            </span>
            <div class="panel-actions">
              <el-button type="primary" :disabled="!selectedDeliveryId" @click="handlePrintPickupList"
                :loading="printLoading">
                打印取货单
              </el-button>
              <el-button type="success" :disabled="!selectedDeliveryId" @click="handleOpenPrintOrderDialog"
                :loading="printOrderLoading" style="margin-left: 10px;">
                打印订单小票
              </el-button>
            </div>
          </div>
          <div class="panel-content">
            <el-table v-loading="loading" :data="orders" border stripe class="orders-table" empty-text="请选择配送员或暂无订单数据"
              row-key="id">
              <el-table-column prop="order_number" label="订单编号" width="180" />
              <el-table-column label="用户信息" min-width="180">
                <template #default="scope">
                  <div v-if="scope.row.user">
                    <div>{{ scope.row.user.name || '未命名' }}</div>
                    <div style="color: #909399; font-size: 12px;">用户{{ scope.row.user.user_code || scope.row.user_id }}
                    </div>
                  </div>
                  <span v-else>用户ID: {{ scope.row.user_id }}</span>
                </template>
              </el-table-column>
              <el-table-column label="配送员" width="150">
                <template #default="scope">
                  <div v-if="scope.row.delivery_employee">
                    <el-tag size="small" type="success">
                      {{ scope.row.delivery_employee.name || scope.row.delivery_employee.employee_code }}
                    </el-tag>
                    <div v-if="scope.row.delivery_employee.employee_code"
                      style="color: #909399; font-size: 11px; margin-top: 2px;">
                      {{ scope.row.delivery_employee.employee_code }}
                    </div>
                  </div>
                  <span v-else style="color: #c0c4cc;">未分配</span>
                </template>
              </el-table-column>
              <el-table-column label="收货地址" min-width="200">
                <template #default="scope">
                  <div v-if="scope.row.address">
                    <div>{{ scope.row.address.name || '-' }}</div>
                    <div style="color: #909399; font-size: 12px;">{{ scope.row.address.address || '-' }}</div>
                  </div>
                  <span v-else>-</span>
                </template>
              </el-table-column>
              <el-table-column prop="status" label="订单状态" width="120">
                <template #default="scope">
                  <el-tag :type="getStatusType(scope.row.status)">
                    {{ formatStatus(scope.row.status) }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="商品件数" width="120" align="center">
                <template #default="scope">
                  <el-button type="primary" link @click="handleViewOrderItems(scope.row.id)"
                    :disabled="!scope.row.item_count || scope.row.item_count === 0">
                    {{ scope.row.item_count || 0 }} 件
                  </el-button>
                </template>
              </el-table-column>
              <el-table-column label="金额信息" min-width="150">
                <template #default="scope">
                  <div style="color: #ff4d4f; font-weight: 600; font-size: 14px;">
                    实付: ¥{{ formatMoney(scope.row.total_amount) }}
                  </div>
                  <div style="color: #909399; font-size: 12px; margin-top: 4px;">
                    <span v-if="scope.row.delivery_fee > 0">
                      配送费: ¥{{ formatMoney(scope.row.delivery_fee) }}
                    </span>
                    <span v-else style="color: #67c23a;">
                      配送费: 免费配送
                    </span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column prop="created_at" label="下单时间" width="160">
                <template #default="scope">
                  {{ formatDate(scope.row.created_at) }}
                </template>
              </el-table-column>
              <el-table-column label="操作" width="120" fixed="right">
                <template #default="scope">
                  <el-button type="primary" link @click="handleViewDetail(scope.row.id)">
                    查看详情
                  </el-button>
                </template>
              </el-table-column>
            </el-table>

            <div class="pagination" v-if="pagination.total > 0">
              <el-pagination background layout="total, prev, pager, next, jumper" :page-size="pagination.pageSize"
                :current-page="pagination.pageNum" :total="pagination.total" @current-change="handlePageChange" />
            </div>
          </div>
        </div>
      </div>
    </el-card>

    <!-- 订单详情对话框 -->
    <el-dialog v-model="detailDialogVisible" title="订单详情" width="80%" :close-on-click-modal="false">
      <div v-loading="detailLoading" v-if="orderDetail">
        <el-tabs v-model="activeTab">
          <el-tab-pane label="基本信息" name="basic">
            <!-- 订单基本信息 -->
            <el-descriptions :column="2" border style="margin-bottom: 20px;">
              <el-descriptions-item label="订单ID">{{ orderDetail.order?.id }}</el-descriptions-item>
              <el-descriptions-item label="订单编号">{{ orderDetail.order?.order_number || '-' }}</el-descriptions-item>
              <el-descriptions-item label="订单状态">
                <el-tag :type="getStatusType(orderDetail.order?.status)">
                  {{ formatStatus(orderDetail.order?.status) }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="下单时间">{{ formatDate(orderDetail.order?.created_at) }}</el-descriptions-item>
              <el-descriptions-item label="更新时间">{{ formatDate(orderDetail.order?.updated_at) }}</el-descriptions-item>
            </el-descriptions>

            <!-- 用户信息 -->
            <el-divider content-position="left">用户信息</el-divider>
            <el-descriptions :column="2" border style="margin-bottom: 20px;" v-if="orderDetail.user">
              <el-descriptions-item label="用户ID">{{ orderDetail.user.id }}</el-descriptions-item>
              <el-descriptions-item label="用户编号">用户{{ orderDetail.user.user_code || '-' }}</el-descriptions-item>
              <el-descriptions-item label="姓名">{{ orderDetail.user.name || '-' }}</el-descriptions-item>
              <el-descriptions-item label="手机号">{{ orderDetail.user.phone || '-' }}</el-descriptions-item>
            </el-descriptions>

            <!-- 收货地址 -->
            <el-divider content-position="left">收货地址</el-divider>
            <el-descriptions :column="2" border style="margin-bottom: 20px;" v-if="orderDetail.address">
              <el-descriptions-item label="地址名称">{{ orderDetail.address.name || '-' }}</el-descriptions-item>
              <el-descriptions-item label="联系人">{{ orderDetail.address.contact || '-' }}</el-descriptions-item>
              <el-descriptions-item label="手机号">{{ orderDetail.address.phone || '-' }}</el-descriptions-item>
              <el-descriptions-item label="详细地址" :span="2">{{ orderDetail.address.address || '-'
              }}</el-descriptions-item>
            </el-descriptions>

            <!-- 订单明细 -->
            <el-divider content-position="left">订单明细</el-divider>
            <el-table :data="orderDetail.order_items" border stripe style="margin-bottom: 20px;">
              <el-table-column prop="product_name" label="商品名称" min-width="150" />
              <el-table-column prop="spec_name" label="规格" width="120" />
              <el-table-column prop="quantity" label="数量" width="80" align="center" />
              <el-table-column prop="unit_price" label="单价" width="100" align="right">
                <template #default="scope">
                  ¥{{ formatMoney(scope.row.unit_price) }}
                </template>
              </el-table-column>
              <el-table-column prop="subtotal" label="小计" width="100" align="right">
                <template #default="scope">
                  ¥{{ formatMoney(scope.row.subtotal) }}
                </template>
              </el-table-column>
            </el-table>
          </el-tab-pane>

          <!-- 金额信息标签页 -->
          <el-tab-pane label="金额信息" name="amount">
            <el-descriptions :column="1" border>
              <el-descriptions-item label="商品金额">
                ¥{{ formatMoney(orderDetail.order?.goods_amount) }}
              </el-descriptions-item>
              <el-descriptions-item label="配送费">
                ¥{{ formatMoney(orderDetail.order?.delivery_fee) }}
              </el-descriptions-item>
              <el-descriptions-item label="实付金额">
                <span style="color: #ff4d4f; font-weight: 600; font-size: 16px;">
                  ¥{{ formatMoney(orderDetail.order?.total_amount) }}
                </span>
              </el-descriptions-item>
            </el-descriptions>
          </el-tab-pane>
        </el-tabs>
      </div>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 打印订单小票对话框 -->
    <el-dialog v-model="printOrderDialogVisible" title="打印订单小票" width="600px" :close-on-click-modal="false">
      <div class="print-order-dialog-content">
        <div class="dialog-title">请选择要打印的订单状态：</div>
        <el-radio-group v-model="selectedPrintOrderStatus" class="print-status-radio-group">
          <div class="status-radio-wrapper" :class="{ 'is-checked': selectedPrintOrderStatus === 'pending_pickup' }"
            @click="selectedPrintOrderStatus = 'pending_pickup'">
            <el-radio label="pending_pickup" class="status-radio-item">
              <div class="status-radio-content">
                <div class="status-label">待取货订单</div>
                <div class="status-desc">仅打印待取货状态的订单</div>
              </div>
            </el-radio>
          </div>
          <div class="status-radio-wrapper" :class="{ 'is-checked': selectedPrintOrderStatus === 'delivering' }"
            @click="selectedPrintOrderStatus = 'delivering'">
            <el-radio label="delivering" class="status-radio-item">
              <div class="status-radio-content">
                <div class="status-label">配送中订单</div>
                <div class="status-desc">仅打印配送中状态的订单</div>
              </div>
            </el-radio>
          </div>
          <div class="status-radio-wrapper"
            :class="{ 'is-checked': selectedPrintOrderStatus === 'pending_pickup_delivering' }"
            @click="selectedPrintOrderStatus = 'pending_pickup_delivering'">
            <el-radio label="pending_pickup_delivering" class="status-radio-item">
              <div class="status-radio-content">
                <div class="status-label">待取货+配送中</div>
                <div class="status-desc">打印待取货和配送中两种状态的订单</div>
              </div>
            </el-radio>
          </div>
          <div class="status-radio-wrapper status-radio-warning"
            :class="{ 'is-checked': selectedPrintOrderStatus === 'all' }" @click="selectedPrintOrderStatus = 'all'">
            <el-radio label="all" class="status-radio-item">
              <div class="status-radio-content">
                <div class="status-label">
                  全部订单
                  <el-tag type="warning" size="small" style="margin-left: 8px;">打印量大</el-tag>
                </div>
                <div class="status-desc">打印选中日期范围内的所有订单</div>
              </div>
            </el-radio>
          </div>
        </el-radio-group>
      </div>
      <template #footer>
        <el-button @click="printOrderDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleConfirmPrintOrders" :loading="printOrderLoading">
          确认打印
        </el-button>
      </template>
    </el-dialog>

    <!-- 商品列表对话框 -->
    <el-dialog v-model="itemsDialogVisible" title="订单商品列表" width="70%" :close-on-click-modal="false">
      <div v-loading="itemsLoading">
        <el-table :data="orderItems" border stripe v-if="orderItems.length > 0">
          <el-table-column type="index" label="序号" width="60" align="center" />
          <el-table-column label="商品图片" width="100" align="center">
            <template #default="scope">
              <el-image v-if="scope.row.image" :src="scope.row.image"
                style="width: 60px; height: 60px; border-radius: 4px;" fit="cover"
                :preview-src-list="[scope.row.image]" />
              <span v-else style="color: #909399;">无图片</span>
            </template>
          </el-table-column>
          <el-table-column prop="product_name" label="商品名称" min-width="150" />
          <el-table-column prop="spec_name" label="规格" width="120" />
          <el-table-column prop="quantity" label="数量" width="80" align="center" />
          <el-table-column prop="unit_price" label="单价" width="100" align="right">
            <template #default="scope">
              ¥{{ formatMoney(scope.row.unit_price) }}
            </template>
          </el-table-column>
          <el-table-column prop="subtotal" label="小计" width="100" align="right">
            <template #default="scope">
              ¥{{ formatMoney(scope.row.subtotal) }}
            </template>
          </el-table-column>
        </el-table>
        <el-empty v-else description="暂无商品数据" />
      </div>
      <template #footer>
        <el-button @click="itemsDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getEmployees } from '../api/employees'
import { getOrders, getOrderDetail } from '../api/orders'
import { getProductDetail } from '../api/product'
import { getAllSuppliers } from '../api/suppliers'
import { hiprint } from 'vue-plugin-hiprint'

const loading = ref(false)
const deliveryEmployees = ref([])
const selectedDeliveryId = ref(null) // 单选配送员ID
const orders = ref([])
const dateRange = ref([])
const activeQuickDate = ref('3days') // 当前激活的快速日期选择：today, 3days, 7days

const pagination = reactive({
  pageNum: 1,
  pageSize: 20,
  total: 0
})

// 订单详情相关
const detailDialogVisible = ref(false)
const detailLoading = ref(false)
const orderDetail = ref(null)
const activeTab = ref('basic')

// 商品列表相关
const itemsDialogVisible = ref(false)
const itemsLoading = ref(false)
const orderItems = ref([])

// 打印相关
const printLoading = ref(false)
const printOrderLoading = ref(false)
const printOrderDialogVisible = ref(false)
const selectedPrintOrderStatus = ref('pending_pickup')

// 初始化日期为近3日
const initDateRange = () => {
  setDateRangeForDays(3)
  activeQuickDate.value = '3days'
}

// 设置日期范围为指定天数（从今天往前推）
const setDateRangeForDays = (days) => {
  const today = new Date()
  const endDate = new Date(today)
  endDate.setHours(23, 59, 59, 999)

  const startDate = new Date(today)
  startDate.setDate(startDate.getDate() - (days - 1))
  startDate.setHours(0, 0, 0, 0)

  const startStr = startDate.toISOString().split('T')[0]
  const endStr = endDate.toISOString().split('T')[0]
  dateRange.value = [startStr, endStr]
}

// 处理快速日期选择
const handleQuickDateSelect = (type) => {
  activeQuickDate.value = type

  if (type === 'today') {
    const today = new Date()
    const todayStr = today.toISOString().split('T')[0]
    dateRange.value = [todayStr, todayStr]
  } else if (type === '3days') {
    setDateRangeForDays(3)
  } else if (type === '7days') {
    setDateRangeForDays(7)
  }

  // 直接查询
  pagination.pageNum = 1
  loadOrders()
}

// 加载配送员列表
const loadDeliveryEmployees = async () => {
  try {
    const res = await getEmployees({ is_delivery: true })
    let employeeList = []

    if (res && res.code === 200 && res.data) {
      if (res.data.list && Array.isArray(res.data.list)) {
        employeeList = res.data.list
      } else if (Array.isArray(res.data)) {
        employeeList = res.data
      }
    } else if (Array.isArray(res)) {
      employeeList = res
    } else if (res && res.list && Array.isArray(res.list)) {
      employeeList = res.list
    }

    // 只显示启用的配送员
    deliveryEmployees.value = employeeList.filter(emp => emp.is_delivery && emp.status)
  } catch (error) {
    console.error('获取配送员列表失败:', error)
    ElMessage.error('获取配送员列表失败，请稍后再试')
  }
}

// 加载订单列表
const loadOrders = async () => {
  if (!selectedDeliveryId.value) {
    orders.value = []
    pagination.total = 0
    return
  }

  loading.value = true
  try {
    const params = {
      pageNum: pagination.pageNum,
      pageSize: pagination.pageSize,
      // 配送员ID（单选）
      delivery_employee_ids: selectedDeliveryId.value.toString(),
      // 包含所有配送相关状态（逗号分隔）
      status: 'pending_delivery,pending_pickup,delivering,delivered,paid'
    }

    // 添加日期筛选
    if (dateRange.value && dateRange.value.length === 2) {
      params.start_date = dateRange.value[0]
      params.end_date = dateRange.value[1]
    }

    const res = await getOrders(params)

    let orderList = []
    let total = 0

    if (res) {
      if (res.code === 200 && res.data) {
        orderList = res.data.list || []
        total = res.data.total || 0
      } else if (res.list && Array.isArray(res.list)) {
        orderList = res.list
        total = res.total || 0
      } else if (Array.isArray(res)) {
        orderList = res
        total = res.length
      } else if (res.data && Array.isArray(res.data)) {
        orderList = res.data
        total = res.total || res.data.length
      }
    }

    orders.value = Array.isArray(orderList) ? [...orderList] : []
    pagination.total = Number(total) || 0
  } catch (error) {
    console.error('获取订单失败:', error)
    orders.value = []
    pagination.total = 0
    ElMessage.error('获取订单列表失败，请稍后再试')
  } finally {
    loading.value = false
  }
}

// 处理配送员选择变化
const handleDeliverySelectionChange = () => {
  pagination.pageNum = 1
  loadOrders()
}

// 处理日期变化
const handleDateChange = () => {
  // 如果手动选择日期，清除快速选择状态
  activeQuickDate.value = ''
  pagination.pageNum = 1
  loadOrders()
}

// 处理搜索
const handleSearch = () => {
  pagination.pageNum = 1
  loadOrders()
}

// 处理分页变化
const handlePageChange = (page) => {
  pagination.pageNum = page
  loadOrders()
}

// 查看订单详情
const handleViewDetail = async (id) => {
  detailDialogVisible.value = true
  detailLoading.value = true
  orderDetail.value = null
  activeTab.value = 'basic'

  try {
    const res = await getOrderDetail(id)
    if (res && res.code === 200) {
      orderDetail.value = res.data
    } else {
      ElMessage.error(res?.message || '获取订单详情失败')
      detailDialogVisible.value = false
    }
  } catch (error) {
    console.error('获取订单详情失败:', error)
    ElMessage.error('获取订单详情失败，请稍后再试')
    detailDialogVisible.value = false
  } finally {
    detailLoading.value = false
  }
}

// 打开打印订单小票对话框
const handleOpenPrintOrderDialog = () => {
  if (!selectedDeliveryId.value) {
    ElMessage.warning('请先选择配送员')
    return
  }
  selectedPrintOrderStatus.value = 'pending_pickup'
  printOrderDialogVisible.value = true
}

// 确认打印订单小票
const handleConfirmPrintOrders = async () => {
  // 如果选择全部订单，显示警告
  if (selectedPrintOrderStatus.value === 'all') {
    try {
      await ElMessageBox.confirm(
        '您选择了打印全部订单，可能打印量比较大，是否继续？',
        '警告',
        {
          confirmButtonText: '确认继续',
          cancelButtonText: '取消',
          type: 'warning'
        }
      )
    } catch {
      // 用户取消
      return
    }
  }

  printOrderDialogVisible.value = false
  await printOrdersByStatus()
}

// 根据选择的状态打印订单小票
const printOrdersByStatus = async () => {
  if (!selectedDeliveryId.value) {
    ElMessage.warning('请先选择配送员')
    return
  }

  // 必须选择日期范围
  if (!dateRange.value || dateRange.value.length !== 2) {
    ElMessage.warning('请先选择日期范围')
    return
  }

  printOrderLoading.value = true
  try {
    // 根据选择的状态确定要查询的状态列表
    let statusList = []
    if (selectedPrintOrderStatus.value === 'pending_pickup') {
      statusList = ['pending_pickup']
    } else if (selectedPrintOrderStatus.value === 'delivering') {
      statusList = ['delivering']
    } else if (selectedPrintOrderStatus.value === 'pending_pickup_delivering') {
      statusList = ['pending_pickup', 'delivering']
    } else if (selectedPrintOrderStatus.value === 'all') {
      // 全部订单，不传status参数
      statusList = []
    }

    // 获取订单列表（必须使用日期范围）
    const params = {
      pageNum: 1,
      pageSize: 1000, // 获取足够多的数据
      delivery_employee_ids: selectedDeliveryId.value.toString(),
      start_date: dateRange.value[0],
      end_date: dateRange.value[1]
    }

    if (statusList.length > 0) {
      params.status = statusList.join(',')
    }

    const res = await getOrders(params)
    let orderList = []

    if (res) {
      if (res.code === 200 && res.data) {
        orderList = res.data.list || []
      } else if (res.list && Array.isArray(res.list)) {
        orderList = res.list
      } else if (Array.isArray(res)) {
        orderList = res
      } else if (res.data && Array.isArray(res.data)) {
        orderList = res.data
      }
    }

    if (orderList.length === 0) {
      ElMessage.warning('没有找到符合条件的订单')
      return
    }

    // 获取每个订单的详情并打印
    let printCount = 0
    for (const order of orderList) {
      try {
        const detailRes = await getOrderDetail(order.id)
        if (detailRes && detailRes.code === 200 && detailRes.data) {
          await printOrderTicket(detailRes.data)
          printCount++
          // 每个订单之间稍作延迟，避免打印过快
          await new Promise(resolve => setTimeout(resolve, 300))
        }
      } catch (error) {
        console.error(`打印订单 ${order.id} 失败:`, error)
      }
    }

    ElMessage.success(`已打印 ${printCount} 张订单小票`)
  } catch (error) {
    console.error('打印订单小票失败:', error)
    ElMessage.error('打印订单小票失败，请稍后再试')
  } finally {
    printOrderLoading.value = false
  }
}

// 打印订单小票（格式参照Orders.vue中的executePrint）
const printOrderTicket = async (orderData) => {
  // 检查 hiprint 是否初始化
  if (!hiprint) {
    ElMessage.error('打印功能未初始化，请刷新页面重试')
    return
  }

  try {
    // 创建打印模板
    const hiprintTemplate = new hiprint.PrintTemplate()

    // 添加打印面板（80mm宽度）
    const panel = hiprintTemplate.addPrintPanel({
      width: 80,
      height: 350,
      paperFooter: 0,
      paperHeader: 0,
      paperNumberLeft: 0,
      paperNumberRight: 0,
      paperNumberFormat: ' ',
    })

    let currentTop = 5

    const order = orderData.order || orderData
    const user = orderData.user || orderData
    const address = orderData.address || orderData
    const orderItems = orderData.order_items || []
    const orderNumber = order.order_number || orderData.order_number || '-'
    const orderTime = order.created_at || orderData.created_at
    const timeStr = orderTime ? formatDate(orderTime) : '-'
    const goodsAmount = order.goods_amount || 0
    const deliveryFee = order.delivery_fee || 0
    const urgentFee = order.urgent_fee || 0
    const couponDiscount = order.coupon_discount || 0
    const totalAmount = order.total_amount || 0
    const hidePrice = order.hide_price || orderData.hide_price || false
    const remark = order.remark || orderData.remark || '' // 订单备注
    const trustReceipt = order.trust_receipt || orderData.trust_receipt || false // 信任签收
    const requirePhoneContact = order.require_phone_contact || orderData.require_phone_contact || false // 要求电话联系

    // 格式化价格
    const formatPrice = (amount) => {
      return hidePrice ? '**' : formatMoney(amount)
    }

    // 格式化电话号码：只显示后四位，其他用*代替
    const formatPhone = (phone) => {
      if (!phone) return '-'
      const phoneStr = String(phone)
      if (phoneStr.length <= 4) return phoneStr
      const lastFour = phoneStr.slice(-4)
      const stars = '*'.repeat(phoneStr.length - 4)
      return stars + lastFour
    }

    // 订单标题
    const title = hidePrice ? "云鹿进货（环保票）" : "云鹿进货"
    panel.addPrintText({
      options: {
        width: 220,
        height: 20,
        top: currentTop,
        left: 0,
        title: title,
        textAlign: "center",
        fontSize: 14,
        fontWeight: "bold"
      },
    })
    currentTop += 30

    // 订单编号
    panel.addPrintText({
      options: {
        width: 300,
        top: currentTop,
        left: 0,
        title: `订单号：${orderNumber}`,
        textAlign: "left",
        fontSize: 10
      },
    })
    currentTop += 15

    // 下单时间
    panel.addPrintText({
      options: {
        width: 300,
        top: currentTop,
        left: 0,
        title: `下单时间：${timeStr}`,
        textAlign: "left",
        fontSize: 9
      },
    })
    currentTop += 15

    // 分隔线
    panel.addPrintText({
      options: {
        width: 230,
        top: currentTop,
        left: 0,
        title: "-------------------------------------------",
        textAlign: "center",
        fontSize: 9
      },
    })
    currentTop += 15

    // 用户信息
    if (user) {
      // 判断用户名，如果没有用户名使用用户编号
      const customerName = user.name 
        ? user.name 
        : (user.user_code ? `用户${user.user_code}` : (user.user_id ? `用户${user.user_id}` : '-'))
      
      panel.addPrintText({
        options: {
          width: 300,
          top: currentTop,
          left: 0,
          title: `客户：${customerName}`,
          textAlign: "left",
          fontSize: 11
        },
      })
      currentTop += 15

      if (user.phone) {
        panel.addPrintText({
          options: {
            width: 300,
            top: currentTop,
            left: 0,
            title: `电话：${formatPhone(user.phone)}`,
            textAlign: "left",
            fontSize: 11
          },
        })
        currentTop += 15
      }
    }

    // 收货地址
    if (address && address.address) {
      const addressText = `地址：${address.address}`
      // 估算文本行数：每行约20个字符（根据宽度230和字体大小9估算）
      const estimatedLines = Math.ceil(addressText.length / 20)
      const textHeight = Math.max(15, estimatedLines * 15) // 最小15px，每行15px
      
      panel.addPrintText({
        options: {
          width: 230,
          height: textHeight, // 设置高度以容纳多行文本
          top: currentTop,
          left: 0,
          title: addressText,
          textAlign: "left",
          fontSize: 9,
          lineHeight: 15 // 设置行高，确保换行时有足够间距
        },
      })
      currentTop += textHeight + 5 // 根据实际文本高度调整间距
    }

    // 信任签收和电话联系提示
    // 只有当有信任签收或备注时，才显示分割线（如果只有电话联系，不显示分割线）
    if (trustReceipt || (remark && remark.trim())) {
      // 添加分割线，和客户信息分割开
      currentTop += 5
      panel.addPrintText({
        options: {
          width: 230,
          top: currentTop,
          left: 0,
          title: "-------------------------------------------",
          textAlign: "center",
          fontSize: 9
        },
      })
      currentTop += 15
    }

    if (trustReceipt || requirePhoneContact) {
      if (trustReceipt) {
        panel.addPrintText({
          options: {
            width: 230,
            top: currentTop,
            left: 0,
            title: '注意：客户已开启信任签收',
            textAlign: "left",
            fontSize: 10,
            fontWeight: "bold"
          },
        })
        currentTop += 18
      }
      if (requirePhoneContact) {
        panel.addPrintText({
          options: {
            width: 230,
            top: currentTop,
            left: 0,
            title: '注意：配送前需电话联系',
            textAlign: "left",
            fontSize: 10,
            fontWeight: "bold"
          },
        })
        currentTop += 18
      }
    }

    // 订单备注（如果有备注，字体要大一点，因为备注很重要）
    if (remark && remark.trim()) {
      currentTop += 5
      const remarkText = `订单备注：${remark.trim()}`
      // 估算文本行数：每行约18个字符（根据宽度230和字体大小12估算）
      const estimatedLines = Math.ceil(remarkText.length / 18)
      const textHeight = Math.max(20, estimatedLines * 20) // 最小20px，每行20px
      
      panel.addPrintText({
        options: {
          width: 230,
          height: textHeight,
          top: currentTop,
          left: 0,
          title: remarkText,
          textAlign: "left",
          fontSize: 12, // 字体大一点，因为备注很重要
          fontWeight: "bold",
          lineHeight: 20 // 设置行高
        },
      })
      currentTop += textHeight + 5
    }

    // 分隔线
    currentTop += 3
    panel.addPrintText({
      options: {
        width: 230,
        top: currentTop,
        left: 0,
        title: "-------------------------------------------",
        textAlign: "center",
        fontSize: 9
      },
    })
    currentTop += 15

    // 商品列表
    if (orderItems.length > 0) {
      orderItems.forEach((item) => {
        const quantity = item.quantity || 0
        const unitPrice = item.unit_price || 0
        const subtotal = item.subtotal || (quantity * unitPrice)

        const productName = `${item.product_name || ''} ${item.spec_name || ''}`.trim()
        const productNameText = productName + ' ' + ' X ' + quantity
        // 估算文本行数：每行约18个字符（根据宽度230和字体大小11估算）
        const estimatedLines = Math.ceil(productNameText.length / 18)
        const textHeight = Math.max(18, estimatedLines * 18) // 最小18px，每行18px
        
        panel.addPrintText({
          options: {
            width: 230,
            height: textHeight, // 设置高度以容纳多行文本
            top: currentTop,
            left: 0,
            title: productNameText,
            textAlign: "left",
            fontSize: 11,
            fontWeight: "bold",
            lineHeight: 18 // 设置行高，确保换行时有足够间距
          },
        })
        currentTop += textHeight + 3 // 根据实际文本高度调整间距

        panel.addPrintText({
          options: {
            width: 230,
            top: currentTop,
            left: 0,
            title: `  ${quantity} × ¥${formatPrice(unitPrice)} = ¥${formatPrice(subtotal)}`,
            textAlign: "left",
            fontSize: 10
          },
        })
        currentTop += 20
      })
    }

    // 分隔线
    currentTop += 5
    panel.addPrintText({
      options: {
        width: 230,
        top: currentTop,
        left: 0,
        title: "-------------------------------------------",
        textAlign: "center",
        fontSize: 9
      },
    })
    currentTop += 22

    // 金额汇总
    panel.addPrintText({
      options: {
        width: 220,
        top: currentTop,
        left: 0,
        title: `商品金额：¥${formatPrice(goodsAmount)}`,
        textAlign: "right",
        fontSize: 10
      },
    })
    currentTop += 20

    // 配送费
    panel.addPrintText({
      options: {
        width: 220,
        top: currentTop,
        left: 0,
        title: deliveryFee > 0 ? `配送费：¥${formatPrice(deliveryFee)}` : '免配送费',
        textAlign: "right",
        fontSize: 10
      },
    })
    currentTop += 20

    if (urgentFee > 0) {
      panel.addPrintText({
        options: {
          width: 220,
          top: currentTop,
          left: 0,
          title: `加急费：¥${formatPrice(urgentFee)}`,
          textAlign: "right",
          fontSize: 10
        },
      })
      currentTop += 20
    }

    // 优惠券
    if (couponDiscount > 0) {
      panel.addPrintText({
        options: {
          width: 220,
          top: currentTop,
          left: 0,
          title: `共计优惠：-¥${formatPrice(couponDiscount)}`,
          textAlign: "right",
          fontSize: 10
        },
      })
      currentTop += 20
    }

    // 实付金额
    panel.addPrintText({
      options: {
        width: 220,
        top: currentTop,
        left: 0,
        title: `实付金额：¥${formatPrice(totalAmount)}`,
        textAlign: "right",
        fontSize: 12,
        fontWeight: "bold"
      },
    })
    currentTop += 30

    // 订单编号条形码
    if (orderNumber && orderNumber !== '-') {
      panel.addPrintText({
        options: {
          width: 200,
          height: 45,
          top: currentTop,
          left: 15,
          title: orderNumber,
          textType: "barcode",
        },
      })
      currentTop += 60
    }

    // 底部感谢文字
    currentTop += 10
    panel.addPrintText({
      options: {
        width: 220,
        top: currentTop,
        left: 0,
        title: "微信搜索\"云鹿进货\"小程序，",
        textAlign: "center",
        fontSize: 11
      },
    })
    currentTop += 20

    panel.addPrintText({
      options: {
        width: 220,
        top: currentTop,
        left: 0,
        title: "了解更多优惠产品！",
        textAlign: "center",
        fontSize: 11
      },
    })

    // 执行打印
    hiprintTemplate.print2(panel)
  } catch (error) {
    console.error('打印订单小票失败:', error)
    throw error
  }
}

// 查看订单商品列表
const handleViewOrderItems = async (id) => {
  itemsDialogVisible.value = true
  itemsLoading.value = true
  orderItems.value = []

  try {
    const res = await getOrderDetail(id)
    if (res && res.code === 200 && res.data && res.data.order_items) {
      orderItems.value = res.data.order_items
    } else {
      ElMessage.error('获取商品列表失败')
      itemsDialogVisible.value = false
    }
  } catch (error) {
    console.error('获取商品列表失败:', error)
    ElMessage.error('获取商品列表失败，请稍后再试')
    itemsDialogVisible.value = false
  } finally {
    itemsLoading.value = false
  }
}

// 打印取货单
const handlePrintPickupList = async () => {
  if (!selectedDeliveryId.value) {
    ElMessage.warning('请先选择配送员')
    return
  }

  // 必须选择日期范围
  if (!dateRange.value || dateRange.value.length !== 2) {
    ElMessage.warning('请先选择日期范围')
    return
  }

  printLoading.value = true
  try {
    // 获取选中配送员信息
    const selectedEmployee = deliveryEmployees.value.find(emp => emp.id === selectedDeliveryId.value)
    if (!selectedEmployee) {
      ElMessage.error('未找到配送员信息')
      return
    }

    // 获取待取货订单（必须使用日期范围）
    const params = {
      pageNum: 1,
      pageSize: 1000, // 获取足够多的数据
      delivery_employee_ids: selectedDeliveryId.value.toString(),
      status: 'pending_pickup',
      start_date: dateRange.value[0],
      end_date: dateRange.value[1]
    }

    const res = await getOrders(params)
    let orderList = []

    if (res) {
      if (res.code === 200 && res.data) {
        orderList = res.data.list || []
      } else if (res.list && Array.isArray(res.list)) {
        orderList = res.list
      } else if (Array.isArray(res)) {
        orderList = res
      } else if (res.data && Array.isArray(res.data)) {
        orderList = res.data
      }
    }

    if (orderList.length === 0) {
      ElMessage.warning('该配送员暂无待取货订单')
      return
    }

    // 获取所有订单的详情和商品信息
    const allItems = []
    for (const order of orderList) {
      try {
        const detailRes = await getOrderDetail(order.id)
        if (detailRes && detailRes.code === 200 && detailRes.data && detailRes.data.order_items) {
          const items = detailRes.data.order_items.map(item => ({
            ...item,
            order_id: order.id,
            order_number: order.order_number || detailRes.data.order?.order_number
          }))
          allItems.push(...items)
        }
      } catch (error) {
        console.error(`获取订单 ${order.id} 详情失败:`, error)
      }
    }

    if (allItems.length === 0) {
      ElMessage.warning('未找到待取货商品')
      return
    }

    // 获取商品的供应商信息并分组
    // 先收集所有唯一的商品ID
    const productIds = [...new Set(allItems.map(item => item.product_id))]

    // 批量查询商品信息（获取供应商ID）
    const productSupplierMap = new Map()
    for (const productId of productIds) {
      try {
        const productRes = await getProductDetail(productId)
        if (productRes && productRes.code === 200 && productRes.data) {
          // 尝试从不同可能的字段获取供应商ID
          const supplierId = productRes.data.supplier_id || productRes.data.supplier?.id || 0
          productSupplierMap.set(productId, supplierId)
        }
      } catch (error) {
        console.error(`获取商品 ${productId} 详情失败:`, error)
        productSupplierMap.set(productId, 0) // 默认供应商ID为0
      }
    }

    // 获取供应商名称
    const supplierIds = [...new Set(Array.from(productSupplierMap.values()))].filter(id => id > 0)
    const supplierNameMap = new Map()
    supplierNameMap.set(0, '未知供应商') // 默认供应商

    // 获取供应商列表以获取名称
    try {
      const suppliersRes = await getAllSuppliers()
      if (suppliersRes && suppliersRes.code === 200 && suppliersRes.data) {
        const suppliers = Array.isArray(suppliersRes.data) ? suppliersRes.data : (suppliersRes.data.list || [])
        suppliers.forEach(supplier => {
          if (supplier.id) {
            supplierNameMap.set(supplier.id, supplier.name || `供应商${supplier.id}`)
          }
        })
      }
    } catch (error) {
      console.error('获取供应商列表失败:', error)
    }

    // 为没有名称的供应商设置默认名称
    supplierIds.forEach(id => {
      if (!supplierNameMap.has(id)) {
        supplierNameMap.set(id, `供应商${id}`)
      }
    })

    // 按供应商分组商品
    const supplierItemsMap = new Map()
    for (const item of allItems) {
      const supplierId = productSupplierMap.get(item.product_id) || 0
      const supplierName = supplierNameMap.get(supplierId) || '未知供应商'

      if (!supplierItemsMap.has(supplierId)) {
        supplierItemsMap.set(supplierId, {
          supplierId,
          supplierName,
          items: []
        })
      }

      supplierItemsMap.get(supplierId).items.push({
        product_name: item.product_name,
        spec_name: item.spec_name,
        quantity: item.quantity,
        order_number: item.order_number
      })
    }

    // 为每个供应商打印取货单
    for (const [supplierId, supplierData] of supplierItemsMap) {
      // 合并相同商品和规格的商品
      const mergedItems = mergeItemsByProductAndSpec(supplierData.items)
      const mergedSupplierData = {
        ...supplierData,
        items: mergedItems
      }
      await printPickupListForSupplier(mergedSupplierData, selectedEmployee)
      // 每个供应商之间稍作延迟，避免打印过快
      await new Promise(resolve => setTimeout(resolve, 500))
    }

    ElMessage.success(`已打印 ${supplierItemsMap.size} 张取货单`)
  } catch (error) {
    console.error('打印取货单失败:', error)
    ElMessage.error('打印取货单失败，请稍后再试')
  } finally {
    printLoading.value = false
  }
}

// 合并相同商品和规格的商品（不同规格分开）
const mergeItemsByProductAndSpec = (items) => {
  const mergedMap = new Map()

  items.forEach(item => {
    // 使用商品名称和规格作为合并的key
    const key = `${item.product_name}|||${item.spec_name || ''}`

    if (mergedMap.has(key)) {
      // 如果已存在，累加数量
      mergedMap.get(key).quantity += item.quantity
    } else {
      // 如果不存在，创建新项
      mergedMap.set(key, {
        product_name: item.product_name,
        spec_name: item.spec_name,
        quantity: item.quantity
      })
    }
  })

  // 转换为数组并返回
  return Array.from(mergedMap.values())
}

// 为单个供应商打印取货单
const printPickupListForSupplier = async (supplierData, employee) => {
  // 检查 hiprint 是否初始化
  if (!hiprint) {
    ElMessage.error('打印功能未初始化，请刷新页面重试')
    return
  }

  try {
    // 创建打印模板
    const hiprintTemplate = new hiprint.PrintTemplate()

    // 添加打印面板（80mm宽度）
    const panel = hiprintTemplate.addPrintPanel({
      width: 80,
      height: 400,
      paperFooter: 0,
      paperHeader: 0,
      paperNumberLeft: 0,
      paperNumberRight: 0,
      paperNumberFormat: ' ',
    })

    let currentTop = 5

    // 标题
    panel.addPrintText({
      options: {
        width: 230,
        height: 20,
        top: currentTop,
        left: 0,
        title: '配送员提货单',
        textAlign: 'center',
        fontSize: 16,
        fontWeight: 'bold'
      }
    })
    currentTop += 25


    // 供应商信息
    panel.addPrintText({
      options: {
        width: 230,
        top: currentTop,
        left: 0,
        title: `供应商：${supplierData.supplierName}`,
        textAlign: 'left',
        fontSize: 11,
        fontWeight: 'bold'
      }
    })
    currentTop += 18

    // 配送员信息
    panel.addPrintText({
      options: {
        width: 230,
        top: currentTop,
        left: 0,
        title: `配送员：${employee.name || employee.employee_code}`,
        textAlign: 'left',
        fontSize: 11
      }
    })
    currentTop += 18

    panel.addPrintText({
      options: {
        width: 230,
        top: currentTop,
        left: 0,
        title: `工号：${employee.employee_code}`,
        textAlign: 'left',
        fontSize: 11
      }
    })
    currentTop += 15



    // 分隔线
    panel.addPrintText({
      options: {
        width: 230,
        top: currentTop,
        left: 0,
        title: '-------------------------------------------',
        textAlign: 'center',
        fontSize: 9
      }
    })
    currentTop += 15

    // 商品列表
    let totalQuantity = 0
    supplierData.items.forEach((item, index) => {
      // 商品名称和规格
      const nameText = `${item.product_name}${item.spec_name ? ` (${item.spec_name})` : ''}`
      // 估算文本可能需要的高度（考虑换行情况）
      // 假设每行大约12px高度，加上行间距，给足够的空间
      const estimatedLines = Math.ceil((nameText.length + 5) / 20) // 大约每行20个字符
      const textHeight = Math.max(18, estimatedLines * 18) // 每行至少18px高度

      panel.addPrintText({
        options: {
          width: 220,
          height: textHeight, // 设置高度以支持换行
          top: currentTop,
          left: 0,
          title: `${nameText} × ${item.quantity}`,
          textAlign: 'left',
          fontSize: 12,
          fontWeight: 'bold',
          lineHeight: 18 // 设置行高，让换行后的文字有足够间距
        }
      })
      currentTop += textHeight + 5  // 根据实际高度调整间距
      totalQuantity += item.quantity
    })

    // 分隔线
    currentTop += 5
    panel.addPrintText({
      options: {
        width: 230,
        top: currentTop,
        left: 0,
        title: '-------------------------------------------',
        textAlign: 'center',
        fontSize: 9
      }
    })
    currentTop += 18

    // 合计
    panel.addPrintText({
      options: {
        width: 230,
        top: currentTop,
        left: 0,
        title: `合计：${totalQuantity}件`,
        textAlign: 'right',
        fontSize: 12,
        fontWeight: 'bold'
      }
    })
    currentTop += 20

    // 打印时间
    const printTime = new Date().toLocaleString('zh-CN')
    panel.addPrintText({
      options: {
        width: 230,
        top: currentTop,
        left: 0,
        title: `打印时间：${printTime}`,
        textAlign: 'left',
        fontSize: 9
      }
    })

    // 执行打印
    hiprintTemplate.print2(panel)
  } catch (error) {
    console.error('打印失败:', error)
    throw error
  }
}

// 格式化日期
const formatDate = (value) => {
  if (!value) return '-'
  return new Date(value).toLocaleString('zh-CN')
}

// 格式化金额
const formatMoney = (value) => {
  if (value === null || value === undefined) return '0.00'
  const num = Number(value)
  if (isNaN(num)) return '0.00'
  return num.toFixed(2)
}

// 格式化状态
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
  return statusMap[status] || status
}

// 获取状态类型
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

// 初始化
onMounted(() => {
  initDateRange()
  loadDeliveryEmployees()
})
</script>

<style scoped>
.delivery-management-page {
  padding: 0;
}

.delivery-card {
  border-radius: 8px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 20px;
  border-bottom: 1px solid #e6e6e6;
}

.title {
  display: flex;
  flex-direction: column;
}

.title .main {
  font-size: 20px;
  font-weight: 600;
  color: #333;
  margin-bottom: 4px;
}

.title .sub {
  font-size: 14px;
  color: #909399;
}

.date-filter {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
}

.quick-date-buttons {
  display: flex;
  gap: 8px;
}

.content-wrapper {
  display: flex;
  gap: 20px;
  min-height: 600px;
}

.delivery-list-panel {
  width: 300px;
  border: 1px solid #e6e6e6;
  border-radius: 8px;
  background: #fafafa;
  display: flex;
  flex-direction: column;
}

.orders-panel {
  flex: 1;
  border: 1px solid #e6e6e6;
  border-radius: 8px;
  background: #fff;
  display: flex;
  flex-direction: column;
}

.panel-header {
  padding: 16px 20px;
  border-bottom: 1px solid #e6e6e6;
  background: #fff;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-radius: 8px 8px 0 0;
}

.panel-title {
  font-size: 16px;
  font-weight: 600;
  color: #333;
}

.panel-actions {
  display: flex;
  align-items: center;
}

.selected-count {
  font-size: 12px;
  color: #909399;
  font-weight: normal;
  margin-left: 8px;
}

.panel-content {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
}

.delivery-item {
  padding: 12px 16px;
  margin-bottom: 8px;
  background: #fff;
  border: 1px solid #e6e6e6;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.3s;
}

.delivery-item:hover {
  border-color: #409eff;
  background: #f0f9ff;
}

.delivery-item.is-selected {
  border-color: #409eff;
  background: #e6f4ff;
}

.delivery-radio {
  width: 100%;
}

.delivery-info {
  margin-left: 8px;
}

.delivery-name {
  font-size: 15px;
  font-weight: 600;
  color: #333;
  margin-bottom: 4px;
}

.delivery-code,
.delivery-phone {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
}

/* 打印订单小票对话框样式 */
.print-order-dialog-content {
  padding: 10px 0;
}

.dialog-title {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid #ebeef5;
}

.print-status-radio-group {
  display: flex;
  flex-direction: column;
  gap: 12px;
  width: 100%;
}

.status-radio-wrapper {
  width: 100%;
  height: 60px;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  transition: all 0.3s;
  background: #fff;
  cursor: pointer;
  padding: 0;
  margin: 0;
}

.status-radio-wrapper:hover {
  border-color: #409eff;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.1);
}

.status-radio-wrapper.is-checked {
  border-color: #409eff;
  background: #f0f9ff;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.15);
}

.status-radio-item {
  width: 100%;
  height: 100%;
  margin: 0;
  padding: 0;
}

.status-radio-item :deep(.el-radio) {
  width: 100%;
  margin: 0;
  height: auto;
}

.status-radio-item :deep(.el-radio__input) {
  position: absolute;
  left: 16px;
  top: 50%;
  transform: translateY(-50%);
}

.status-radio-item :deep(.el-radio__label) {
  width: 100%;
  /* padding: 16px 20px; */
  padding-left: 40px;
  margin-left: 0;
  cursor: pointer;
}

.status-radio-item :deep(.el-radio__input.is-checked) .el-radio__inner {
  border-color: #409eff;
  background: #409eff;
}

.status-radio-content {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.status-label {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
  display: flex;
  align-items: center;
}

.status-desc {
  font-size: 13px;
  color: #909399;
  line-height: 1.5;
}

.status-radio-warning {
  border-color: #e6a23c;
}

.status-radio-warning:hover {
  border-color: #e6a23c;
  box-shadow: 0 2px 8px rgba(230, 162, 60, 0.15);
}

.status-radio-warning.is-checked {
  border-color: #e6a23c;
  background: #fef0e6;
  box-shadow: 0 2px 8px rgba(230, 162, 60, 0.2);
}

.status-radio-warning :deep(.el-radio__input.is-checked) .el-radio__inner {
  border-color: #e6a23c;
  background: #e6a23c;
}

.orders-table {
  width: 100%;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>
