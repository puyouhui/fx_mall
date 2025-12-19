<template>
  <div class="sales-commission-page">
    <el-card class="commission-card" shadow="never">
      <div class="card-header">
        <div class="title">
          <span class="main">销售分成管理</span>
          <span class="sub">查看和管理销售员的利润分成</span>
        </div>
        <div class="actions">
          <el-select
            v-model="selectedEmployeeCode"
            placeholder="选择销售员"
            clearable
            filterable
            style="width: 200px; margin-right: 10px;"
            @change="handleSearch"
          >
            <el-option
              v-for="emp in salesEmployees"
              :key="emp.employee_code"
              :label="`${emp.employee_code} - ${emp.name || '未命名'}`"
              :value="emp.employee_code"
            />
          </el-select>
          <el-date-picker
            v-model="selectedMonth"
            type="month"
            placeholder="选择月份"
            format="YYYY-MM"
            value-format="YYYY-MM"
            style="width: 150px; margin-right: 10px;"
            @change="handleSearch"
          />
          <el-select
            v-model="selectedStatus"
            placeholder="状态筛选"
            clearable
            style="width: 150px; margin-right: 10px;"
            @change="handleSearch"
          >
            <el-option label="全部" value="all" />
            <el-option label="已计入" value="accounted" />
            <el-option label="已结算" value="settled" />
            <el-option label="未计入" value="unaccounted" />
            <el-option label="未结算" value="unsettled" />
          </el-select>
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
        </div>
      </div>

      <!-- 统计卡片 -->
      <div class="stats-cards" v-if="statsList.length > 0">
        <el-row :gutter="20">
          <el-col :span="6" v-for="stat in statsList" :key="stat.employee_code">
            <el-card shadow="hover" class="stat-card">
              <div class="stat-header">
                <span class="employee-name">{{ getEmployeeName(stat.employee_code) }}</span>
                <el-tag :type="getTierTagType(stat.tier_level)" size="small">
                  {{ getTierText(stat.tier_level) }}
                </el-tag>
              </div>
              <div class="stat-content">
                <div class="stat-item">
                  <span class="label">总销售额：</span>
                  <span class="value">¥{{ formatMoney(stat.total_sales_amount) }}</span>
                </div>
                <div class="stat-item">
                  <span class="label">有效订单：</span>
                  <span class="value">{{ stat.total_valid_orders }} 单</span>
                </div>
                <div class="stat-item">
                  <span class="label">新客户数：</span>
                  <span class="value">{{ stat.total_new_customers }} 人</span>
                </div>
                <div class="stat-item">
                  <span class="label">总利润：</span>
                  <span class="value profit">¥{{ formatMoney(stat.total_profit) }}</span>
                </div>
                <div class="stat-item total">
                  <span class="label">总分成：</span>
                  <span class="value total-commission">¥{{ formatMoney(stat.total_commission) }}</span>
                </div>
              </div>
              <div class="stat-actions">
                <el-button type="primary" size="small" @click="handleViewDetails(stat)">
                  查看详情
                </el-button>
                <el-button type="info" size="small" @click="handleConfig(stat)">
                  配置
                </el-button>
              </div>
            </el-card>
          </el-col>
        </el-row>
      </div>

      <!-- 详情表格 -->
      <el-tabs v-model="activeTab" class="detail-tabs" v-if="selectedEmployeeCode">
        <el-tab-pane label="分成记录" name="records">
          <div class="table-toolbar">
            <div class="toolbar-left">
              <el-button
                type="primary"
                :disabled="selectedRecords.length === 0"
                @click="handleBatchAccount"
              >
                批量计入 ({{ selectedRecords.length }})
              </el-button>
              <el-button
                type="success"
                :disabled="selectedRecords.length === 0"
                @click="handleBatchSettle"
              >
                批量结算 ({{ selectedRecords.length }})
              </el-button>
              <el-button
                type="warning"
                :disabled="!selectedEmployeeCode"
                @click="handleBatchAccountByDate"
              >
                按日期批量计入
              </el-button>
              <el-button
                type="info"
                :disabled="!selectedEmployeeCode"
                @click="handleBatchSettleByDate"
              >
                按日期批量结算
              </el-button>
              <el-button
                type="danger"
                :disabled="selectedRecords.length === 0"
                @click="handleBatchCancelAccount"
              >
                取消计入 ({{ selectedRecords.length }})
              </el-button>
              <el-button
                type="warning"
                :disabled="selectedRecords.length === 0"
                @click="handleBatchResetAccount"
              >
                重新计入 ({{ selectedRecords.length }})
              </el-button>
            </div>
          </div>
          <el-table
            v-loading="recordsLoading"
            :data="commissionRecords"
            border
            stripe
            class="records-table"
            empty-text="暂无数据"
            style="width: 100%"
            @selection-change="handleSelectionChange"
          >
            <el-table-column type="selection" width="55" align="center" />
            <el-table-column prop="order_number" label="订单编号" align="center" min-width="180" />
            <el-table-column prop="order_date" label="订单日期" align="center" min-width="120">
              <template #default="scope">
                {{ formatDate(scope.row.order_date) }}
              </template>
            </el-table-column>
            <el-table-column prop="settlement_date" label="结算日期" align="center" min-width="120">
              <template #default="scope">
                {{ scope.row.settlement_date ? formatDate(scope.row.settlement_date) : '-' }}
              </template>
            </el-table-column>
            <el-table-column label="订单金额" align="center" min-width="120">
              <template #default="scope">
                ¥{{ formatMoney(scope.row.order_amount) }}
              </template>
            </el-table-column>
            <el-table-column label="订单利润" align="center" min-width="120">
              <template #default="scope">
                <span :class="scope.row.order_profit > 0 ? 'profit' : ''">
                  ¥{{ formatMoney(scope.row.order_profit) }}
                </span>
              </template>
            </el-table-column>
            <el-table-column label="基础提成" align="center" min-width="120">
              <template #default="scope">
                ¥{{ formatMoney(scope.row.base_commission) }}
              </template>
            </el-table-column>
            <el-table-column label="新客激励" align="center" min-width="120">
              <template #default="scope">
                <span v-if="scope.row.is_new_customer_order" class="new-customer">
                  ¥{{ formatMoney(scope.row.new_customer_bonus) }}
                </span>
                <span v-else>-</span>
              </template>
            </el-table-column>
            <el-table-column label="阶梯提成" align="center" min-width="120">
              <template #default="scope">
                <span v-if="scope.row.tier_commission > 0">
                  ¥{{ formatMoney(scope.row.tier_commission) }}
                  <el-tag :type="getTierTagType(scope.row.tier_level)" size="small" style="margin-left: 5px;">
                    {{ getTierText(scope.row.tier_level) }}
                  </el-tag>
                </span>
                <span v-else>-</span>
              </template>
            </el-table-column>
            <el-table-column label="总分成" align="center" min-width="120">
              <template #default="scope">
                <span class="total-commission">¥{{ formatMoney(scope.row.total_commission) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="是否有效" align="center" min-width="100">
              <template #default="scope">
                <el-tag :type="scope.row.is_valid_order ? 'success' : 'info'">
                  {{ scope.row.is_valid_order ? '是' : '否' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="计入状态" align="center" min-width="120">
              <template #default="scope">
                <el-tag v-if="scope.row.is_accounted_cancelled" type="danger">
                  计入已取消
                </el-tag>
                <el-tag v-else-if="scope.row.is_accounted" type="success">
                  已计入
                </el-tag>
                <el-tag v-else type="info">
                  未计入
                </el-tag>
                <div v-if="scope.row.accounted_at && !scope.row.is_accounted_cancelled" style="margin-top: 4px; font-size: 12px; color: #909399;">
                  {{ formatDate(scope.row.accounted_at) }}
                </div>
              </template>
            </el-table-column>
            <el-table-column label="结算状态" align="center" min-width="120">
              <template #default="scope">
                <el-tag :type="scope.row.is_settled ? 'success' : 'warning'">
                  {{ scope.row.is_settled ? '已结算' : '未结算' }}
                </el-tag>
                <div v-if="scope.row.settled_at" style="margin-top: 4px; font-size: 12px; color: #909399;">
                  {{ formatDate(scope.row.settled_at) }}
                </div>
              </template>
            </el-table-column>
            <el-table-column label="操作" align="center" min-width="220" fixed="right">
              <template #default="scope">
                <el-button
                  v-if="!scope.row.is_accounted && !scope.row.is_accounted_cancelled"
                  type="primary"
                  size="small"
                  @click="handleAccount(scope.row)"
                >
                  计入
                </el-button>
                <el-button
                  v-if="scope.row.is_accounted && !scope.row.is_settled && !scope.row.is_accounted_cancelled"
                  type="success"
                  size="small"
                  @click="handleSettle(scope.row)"
                >
                  结算
                </el-button>
                <el-button
                  v-if="scope.row.is_accounted && !scope.row.is_settled && !scope.row.is_accounted_cancelled"
                  type="danger"
                  size="small"
                  @click="handleCancelAccount(scope.row)"
                >
                  取消计入
                </el-button>
                <el-button
                  v-if="scope.row.is_accounted_cancelled"
                  type="warning"
                  size="small"
                  @click="handleResetAccount(scope.row)"
                >
                  重新计入
                </el-button>
                <span v-if="scope.row.is_settled" style="color: #909399;">已结算</span>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination">
            <el-pagination
              background
              layout="total, prev, pager, next, jumper"
              :page-size="recordsPagination.pageSize"
              :current-page="recordsPagination.pageNum"
              :total="recordsPagination.total"
              @current-change="handleRecordsPageChange"
            />
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 按日期批量计入对话框 -->
    <el-dialog
      v-model="accountByDateDialogVisible"
      title="按日期批量计入"
      width="500px"
    >
      <el-form :model="accountByDateForm" label-width="120px">
        <el-form-item label="销售员">
          <el-input :value="getEmployeeName(selectedEmployeeCode)" disabled />
        </el-form-item>
        <el-form-item label="开始日期" required>
          <el-date-picker
            v-model="accountByDateForm.startDate"
            type="date"
            placeholder="选择开始日期"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="结束日期" required>
          <el-date-picker
            v-model="accountByDateForm.endDate"
            type="date"
            placeholder="选择结束日期"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="accountByDateDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleConfirmAccountByDate" :loading="accounting">
          确认计入
        </el-button>
      </template>
    </el-dialog>

    <!-- 按日期批量结算对话框 -->
    <el-dialog
      v-model="settleByDateDialogVisible"
      title="按日期批量结算"
      width="500px"
    >
      <el-form :model="settleByDateForm" label-width="120px">
        <el-form-item label="销售员">
          <el-input :value="getEmployeeName(selectedEmployeeCode)" disabled />
        </el-form-item>
        <el-form-item label="开始日期" required>
          <el-date-picker
            v-model="settleByDateForm.startDate"
            type="date"
            placeholder="选择开始日期"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="结束日期" required>
          <el-date-picker
            v-model="settleByDateForm.endDate"
            type="date"
            placeholder="选择结束日期"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="settleByDateDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleConfirmSettleByDate" :loading="settling">
          确认结算
        </el-button>
      </template>
    </el-dialog>

    <!-- 配置对话框 -->
    <el-dialog
      v-model="configDialogVisible"
      title="销售分成配置"
      width="600px"
    >
      <el-form :model="configForm" label-width="180px" :rules="configRules" ref="configFormRef">
        <el-form-item label="销售员工号">
          <el-input v-model="configForm.employee_code" disabled />
        </el-form-item>
        <el-form-item label="基础提成比例" prop="base_commission_rate">
          <el-input-number
            v-model="configForm.base_commission_rate"
            :min="0"
            :max="1"
            :step="0.01"
            :precision="4"
            style="width: 100%"
          />
          <span style="margin-left: 10px; color: #909399;">
            (当前: {{ (configForm.base_commission_rate * 100).toFixed(2) }}%)
          </span>
        </el-form-item>
        <el-form-item label="新客开发激励比例" prop="new_customer_bonus_rate">
          <el-input-number
            v-model="configForm.new_customer_bonus_rate"
            :min="0"
            :max="1"
            :step="0.01"
            :precision="4"
            style="width: 100%"
          />
          <span style="margin-left: 10px; color: #909399;">
            (当前: {{ (configForm.new_customer_bonus_rate * 100).toFixed(2) }}%)
          </span>
        </el-form-item>
        <el-divider />
        <el-form-item label="阶梯1阈值" prop="tier1_threshold">
          <el-input-number
            v-model="configForm.tier1_threshold"
            :min="0"
            :step="1000"
            :precision="2"
            style="width: 100%"
          />
          <span style="margin-left: 10px; color: #909399;">元</span>
        </el-form-item>
        <el-form-item label="阶梯1提成比例" prop="tier1_rate">
          <el-input-number
            v-model="configForm.tier1_rate"
            :min="0"
            :max="1"
            :step="0.01"
            :precision="4"
            style="width: 100%"
          />
          <span style="margin-left: 10px; color: #909399;">
            (当前: {{ (configForm.tier1_rate * 100).toFixed(2) }}%)
          </span>
        </el-form-item>
        <el-form-item label="阶梯2阈值" prop="tier2_threshold">
          <el-input-number
            v-model="configForm.tier2_threshold"
            :min="0"
            :step="1000"
            :precision="2"
            style="width: 100%"
          />
          <span style="margin-left: 10px; color: #909399;">元</span>
        </el-form-item>
        <el-form-item label="阶梯2提成比例" prop="tier2_rate">
          <el-input-number
            v-model="configForm.tier2_rate"
            :min="0"
            :max="1"
            :step="0.01"
            :precision="4"
            style="width: 100%"
          />
          <span style="margin-left: 10px; color: #909399;">
            (当前: {{ (configForm.tier2_rate * 100).toFixed(2) }}%)
          </span>
        </el-form-item>
        <el-form-item label="阶梯3阈值" prop="tier3_threshold">
          <el-input-number
            v-model="configForm.tier3_threshold"
            :min="0"
            :step="1000"
            :precision="2"
            style="width: 100%"
          />
          <span style="margin-left: 10px; color: #909399;">元</span>
        </el-form-item>
        <el-form-item label="阶梯3提成比例" prop="tier3_rate">
          <el-input-number
            v-model="configForm.tier3_rate"
            :min="0"
            :max="1"
            :step="0.01"
            :precision="4"
            style="width: 100%"
          />
          <span style="margin-left: 10px; color: #909399;">
            (当前: {{ (configForm.tier3_rate * 100).toFixed(2) }}%)
          </span>
        </el-form-item>
        <el-divider />
        <el-form-item label="最小利润阈值" prop="min_profit_threshold">
          <el-input-number
            v-model="configForm.min_profit_threshold"
            :min="0"
            :step="1"
            :precision="2"
            style="width: 100%"
          />
          <span style="margin-left: 10px; color: #909399;">元（订单利润需大于此值才计入有效订单）</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="configDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleConfirmConfig" :loading="configSaving">
          保存
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { reactive, ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getSalesCommissionStats,
  getSalesCommissions,
  getSalesCommissionConfig,
  updateSalesCommissionConfig,
  getSalesEmployees,
  accountSalesCommissions,
  settleSalesCommissions,
  cancelAccountSalesCommissions,
  resetAccountSalesCommissions
} from '../api/salesCommission'

export default {
  name: 'SalesCommission',
  setup() {
    const loading = ref(false)
    const recordsLoading = ref(false)
    const configSaving = ref(false)
    const accounting = ref(false)
    const settling = ref(false)
    const cancelingAccount = ref(false)
    const resettingAccount = ref(false)
    const statsList = ref([])
    const commissionRecords = ref([])
    const salesEmployees = ref([])
    const selectedEmployeeCode = ref('')
    const selectedMonth = ref('')
    const selectedStatus = ref('')
    const dateRange = ref(null)
    const selectedRecords = ref([])
    const activeTab = ref('records')
    const configDialogVisible = ref(false)
    const accountByDateDialogVisible = ref(false)
    const settleByDateDialogVisible = ref(false)
    const configFormRef = ref(null)

    const recordsPagination = reactive({
      pageNum: 1,
      pageSize: 10,
      total: 0
    })

    const configForm = reactive({
      employee_code: '',
      base_commission_rate: 0.45,
      new_customer_bonus_rate: 0.20,
      tier1_threshold: 50000,
      tier1_rate: 0.05,
      tier2_threshold: 100000,
      tier2_rate: 0.10,
      tier3_threshold: 200000,
      tier3_rate: 0.20,
      min_profit_threshold: 5.00
    })

    const accountByDateForm = reactive({
      startDate: '',
      endDate: ''
    })

    const settleByDateForm = reactive({
      startDate: '',
      endDate: ''
    })

    const configRules = {
      base_commission_rate: [
        { required: true, message: '请输入基础提成比例', trigger: 'blur' }
      ],
      new_customer_bonus_rate: [
        { required: true, message: '请输入新客开发激励比例', trigger: 'blur' }
      ],
      tier1_threshold: [
        { required: true, message: '请输入阶梯1阈值', trigger: 'blur' }
      ],
      tier1_rate: [
        { required: true, message: '请输入阶梯1提成比例', trigger: 'blur' }
      ],
      tier2_threshold: [
        { required: true, message: '请输入阶梯2阈值', trigger: 'blur' }
      ],
      tier2_rate: [
        { required: true, message: '请输入阶梯2提成比例', trigger: 'blur' }
      ],
      tier3_threshold: [
        { required: true, message: '请输入阶梯3阈值', trigger: 'blur' }
      ],
      tier3_rate: [
        { required: true, message: '请输入阶梯3提成比例', trigger: 'blur' }
      ],
      min_profit_threshold: [
        { required: true, message: '请输入最小利润阈值', trigger: 'blur' }
      ]
    }

    // 初始化月份为当前月份
    const initMonth = () => {
      const now = new Date()
      selectedMonth.value = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
    }

    // 加载销售员列表
    const loadSalesEmployees = async () => {
      try {
        const res = await getSalesEmployees()
        if (res.code === 200) {
          salesEmployees.value = res.data || []
        }
      } catch (error) {
        console.error('获取销售员列表失败:', error)
      }
    }

    // 加载统计数据
    const loadStats = async () => {
      loading.value = true
      try {
        const res = await getSalesCommissionStats(
          selectedEmployeeCode.value || null,
          selectedMonth.value || null
        )
        if (res.code === 200) {
          if (Array.isArray(res.data)) {
            statsList.value = res.data
          } else if (res.data) {
            statsList.value = [res.data]
          } else {
            statsList.value = []
          }
        } else {
          ElMessage.error(res.message || '获取统计数据失败')
          statsList.value = []
        }
      } catch (error) {
        console.error('获取统计数据失败:', error)
        ElMessage.error('获取统计数据失败: ' + (error.message || '未知错误'))
        statsList.value = []
      } finally {
        loading.value = false
      }
    }

    // 加载分成记录
    const loadRecords = async () => {
      if (!selectedEmployeeCode.value) {
        commissionRecords.value = []
        return
      }

      recordsLoading.value = true
      try {
        const startDate = dateRange.value && dateRange.value[0] ? dateRange.value[0] : null
        const endDate = dateRange.value && dateRange.value[1] ? dateRange.value[1] : null
        const res = await getSalesCommissions(
          selectedEmployeeCode.value,
          selectedMonth.value || null,
          selectedStatus.value || null,
          startDate,
          endDate,
          recordsPagination.pageNum,
          recordsPagination.pageSize
        )
        if (res.code === 200) {
          commissionRecords.value = res.data?.list || []
          recordsPagination.total = res.data?.total || 0
        } else {
          ElMessage.error(res.message || '获取分成记录失败')
          commissionRecords.value = []
          recordsPagination.total = 0
        }
      } catch (error) {
        console.error('获取分成记录失败:', error)
        ElMessage.error('获取分成记录失败: ' + (error.message || '未知错误'))
        commissionRecords.value = []
        recordsPagination.total = 0
      } finally {
        recordsLoading.value = false
      }
    }

    // 搜索
    const handleSearch = () => {
      loadStats()
      if (selectedEmployeeCode.value) {
        recordsPagination.pageNum = 1
        loadRecords()
      }
    }

    // 查看详情
    const handleViewDetails = (stat) => {
      selectedEmployeeCode.value = stat.employee_code
      activeTab.value = 'records'
      recordsPagination.pageNum = 1
      loadRecords()
    }

    // 配置
    const handleConfig = async (stat) => {
      configForm.employee_code = stat.employee_code
      // 获取当前配置
      try {
        const res = await getSalesCommissionConfig(stat.employee_code)
        if (res.code === 200 && res.data) {
          configForm.base_commission_rate = res.data.base_commission_rate || 0.45
          configForm.new_customer_bonus_rate = res.data.new_customer_bonus_rate || 0.20
          configForm.tier1_threshold = res.data.tier1_threshold || 50000
          configForm.tier1_rate = res.data.tier1_rate || 0.05
          configForm.tier2_threshold = res.data.tier2_threshold || 100000
          configForm.tier2_rate = res.data.tier2_rate || 0.10
          configForm.tier3_threshold = res.data.tier3_threshold || 200000
          configForm.tier3_rate = res.data.tier3_rate || 0.20
          configForm.min_profit_threshold = res.data.min_profit_threshold || 5.00
        }
        configDialogVisible.value = true
      } catch (error) {
        console.error('获取配置失败:', error)
        ElMessage.warning('获取配置失败，使用默认值')
        configDialogVisible.value = true
      }
    }

    // 确认配置
    const handleConfirmConfig = async () => {
      if (!configFormRef.value) return

      try {
        await configFormRef.value.validate()
        configSaving.value = true

        const res = await updateSalesCommissionConfig(configForm)
        if (res.code === 200) {
          ElMessage.success('配置保存成功')
          configDialogVisible.value = false
          // 重新加载统计数据
          loadStats()
        } else {
          ElMessage.error(res.message || '保存配置失败')
        }
      } catch (error) {
        if (error !== false) {
          console.error('保存配置失败:', error)
          ElMessage.error('保存配置失败: ' + (error.message || '未知错误'))
        }
      } finally {
        configSaving.value = false
      }
    }

    // 记录分页变化
    const handleRecordsPageChange = (page) => {
      recordsPagination.pageNum = page
      loadRecords()
    }

    // 格式化日期
    const formatDate = (value) => {
      if (!value) return '-'
      return new Date(value).toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
      })
    }

    // 格式化金额
    const formatMoney = (value) => {
      if (value === null || value === undefined) return '0.00'
      const num = Number(value)
      if (isNaN(num)) return '0.00'
      return num.toFixed(2)
    }

    // 获取销售员姓名
    const getEmployeeName = (employeeCode) => {
      const emp = salesEmployees.value.find(e => e.employee_code === employeeCode)
      return emp ? (emp.name || employeeCode) : employeeCode
    }

    // 获取阶梯标签类型
    const getTierTagType = (tierLevel) => {
      if (tierLevel >= 3) return 'danger'
      if (tierLevel >= 2) return 'warning'
      if (tierLevel >= 1) return 'success'
      return 'info'
    }

    // 获取阶梯文本
    const getTierText = (tierLevel) => {
      if (tierLevel >= 3) return '阶梯3'
      if (tierLevel >= 2) return '阶梯2'
      if (tierLevel >= 1) return '阶梯1'
      return '未达标'
    }

    // 表格选择变化
    const handleSelectionChange = (selection) => {
      selectedRecords.value = selection
    }

    // 单笔计入
    const handleAccount = async (row) => {
      try {
        await ElMessageBox.confirm(
          `确认计入订单 ${row.order_number} 的分成？`,
          '确认计入',
          {
            confirmButtonText: '确认',
            cancelButtonText: '取消',
            type: 'warning'
          }
        )

        accounting.value = true
        const res = await accountSalesCommissions({
          commission_ids: [row.id]
        })
        if (res.code === 200) {
          ElMessage.success(`计入成功，共计入 ${res.affected || 1} 条记录`)
          loadRecords()
        } else {
          ElMessage.error(res.message || '计入失败')
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error('计入失败:', error)
          ElMessage.error('计入失败: ' + (error.message || '未知错误'))
        }
      } finally {
        accounting.value = false
      }
    }

    // 单笔结算
    const handleSettle = async (row) => {
      try {
        await ElMessageBox.confirm(
          `确认结算订单 ${row.order_number} 的分成？`,
          '确认结算',
          {
            confirmButtonText: '确认',
            cancelButtonText: '取消',
            type: 'warning'
          }
        )

        settling.value = true
        const res = await settleSalesCommissions({
          commission_ids: [row.id]
        })
        if (res.code === 200) {
          ElMessage.success(`结算成功，共结算 ${res.affected || 1} 条记录`)
          loadRecords()
        } else {
          ElMessage.error(res.message || '结算失败')
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error('结算失败:', error)
          ElMessage.error('结算失败: ' + (error.message || '未知错误'))
        }
      } finally {
        settling.value = false
      }
    }

    // 批量计入
    const handleBatchAccount = async () => {
      if (selectedRecords.value.length === 0) {
        ElMessage.warning('请先选择要计入的记录')
        return
      }

      try {
        await ElMessageBox.confirm(
          `确认计入选中的 ${selectedRecords.value.length} 条记录？`,
          '批量计入',
          {
            confirmButtonText: '确认',
            cancelButtonText: '取消',
            type: 'warning'
          }
        )

        accounting.value = true
        const commissionIds = selectedRecords.value.map(r => r.id)
        const res = await accountSalesCommissions({
          commission_ids: commissionIds
        })
        if (res.code === 200) {
          ElMessage.success(`计入成功，共计入 ${res.affected || 0} 条记录`)
          selectedRecords.value = []
          loadRecords()
        } else {
          ElMessage.error(res.message || '计入失败')
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error('批量计入失败:', error)
          ElMessage.error('批量计入失败: ' + (error.message || '未知错误'))
        }
      } finally {
        accounting.value = false
      }
    }

    // 批量结算
    const handleBatchSettle = async () => {
      if (selectedRecords.value.length === 0) {
        ElMessage.warning('请先选择要结算的记录')
        return
      }

      // 检查是否都已计入
      const notAccounted = selectedRecords.value.filter(r => !r.is_accounted)
      if (notAccounted.length > 0) {
        ElMessage.warning(`有 ${notAccounted.length} 条记录未计入，无法结算`)
        return
      }

      try {
        await ElMessageBox.confirm(
          `确认结算选中的 ${selectedRecords.value.length} 条记录？`,
          '批量结算',
          {
            confirmButtonText: '确认',
            cancelButtonText: '取消',
            type: 'warning'
          }
        )

        settling.value = true
        const commissionIds = selectedRecords.value.map(r => r.id)
        const res = await settleSalesCommissions({
          commission_ids: commissionIds
        })
        if (res.code === 200) {
          ElMessage.success(`结算成功，共结算 ${res.affected || 0} 条记录`)
          selectedRecords.value = []
          loadRecords()
        } else {
          ElMessage.error(res.message || '结算失败')
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error('批量结算失败:', error)
          ElMessage.error('批量结算失败: ' + (error.message || '未知错误'))
        }
      } finally {
        settling.value = false
      }
    }

    // 按日期批量计入
    const handleBatchAccountByDate = () => {
      if (!selectedEmployeeCode.value) {
        ElMessage.warning('请先选择销售员')
        return
      }
      accountByDateForm.startDate = ''
      accountByDateForm.endDate = ''
      accountByDateDialogVisible.value = true
    }

    // 确认按日期批量计入
    const handleConfirmAccountByDate = async () => {
      if (!accountByDateForm.startDate || !accountByDateForm.endDate) {
        ElMessage.warning('请选择开始日期和结束日期')
        return
      }

      try {
        await ElMessageBox.confirm(
          `确认计入 ${accountByDateForm.startDate} 至 ${accountByDateForm.endDate} 期间的所有未计入记录？`,
          '按日期批量计入',
          {
            confirmButtonText: '确认',
            cancelButtonText: '取消',
            type: 'warning'
          }
        )

        accounting.value = true
        const res = await accountSalesCommissions({
          employee_code: selectedEmployeeCode.value,
          start_date: accountByDateForm.startDate,
          end_date: accountByDateForm.endDate
        })
        if (res.code === 200) {
          ElMessage.success(`计入成功，共计入 ${res.affected || 0} 条记录`)
          accountByDateDialogVisible.value = false
          loadRecords()
        } else {
          ElMessage.error(res.message || '计入失败')
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error('按日期批量计入失败:', error)
          ElMessage.error('按日期批量计入失败: ' + (error.message || '未知错误'))
        }
      } finally {
        accounting.value = false
      }
    }

    // 按日期批量结算
    const handleBatchSettleByDate = () => {
      if (!selectedEmployeeCode.value) {
        ElMessage.warning('请先选择销售员')
        return
      }
      settleByDateForm.startDate = ''
      settleByDateForm.endDate = ''
      settleByDateDialogVisible.value = true
    }

    // 确认按日期批量结算
    const handleConfirmSettleByDate = async () => {
      if (!settleByDateForm.startDate || !settleByDateForm.endDate) {
        ElMessage.warning('请选择开始日期和结束日期')
        return
      }

      try {
        await ElMessageBox.confirm(
          `确认结算 ${settleByDateForm.startDate} 至 ${settleByDateForm.endDate} 期间的所有已计入但未结算记录？`,
          '按日期批量结算',
          {
            confirmButtonText: '确认',
            cancelButtonText: '取消',
            type: 'warning'
          }
        )

        settling.value = true
        const res = await settleSalesCommissions({
          employee_code: selectedEmployeeCode.value,
          start_date: settleByDateForm.startDate,
          end_date: settleByDateForm.endDate
        })
        if (res.code === 200) {
          ElMessage.success(`结算成功，共结算 ${res.affected || 0} 条记录`)
          settleByDateDialogVisible.value = false
          loadRecords()
        } else {
          ElMessage.error(res.message || '结算失败')
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error('按日期批量结算失败:', error)
          ElMessage.error('按日期批量结算失败: ' + (error.message || '未知错误'))
        }
      } finally {
        settling.value = false
      }
    }

    // 单笔取消计入
    const handleCancelAccount = async (row) => {
      try {
        await ElMessageBox.confirm(
          `确认取消计入订单 ${row.order_number} 的分成？取消后该订单将标记为"计入已取消"。`,
          '确认取消计入',
          {
            confirmButtonText: '确认',
            cancelButtonText: '取消',
            type: 'warning'
          }
        )

        cancelingAccount.value = true
        const res = await cancelAccountSalesCommissions({
          commission_ids: [row.id]
        })
        if (res.code === 200) {
          ElMessage.success(`取消计入成功，共取消 ${res.affected || 1} 条记录`)
          loadRecords()
        } else {
          ElMessage.error(res.message || '取消计入失败')
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error('取消计入失败:', error)
          ElMessage.error('取消计入失败: ' + (error.message || '未知错误'))
        }
      } finally {
        cancelingAccount.value = false
      }
    }

    // 批量取消计入
    const handleBatchCancelAccount = async () => {
      if (selectedRecords.value.length === 0) {
        ElMessage.warning('请先选择要取消计入的记录')
        return
      }

      // 检查是否都已计入且未结算
      const invalidRecords = selectedRecords.value.filter(r => !r.is_accounted || r.is_settled || r.is_accounted_cancelled)
      if (invalidRecords.length > 0) {
        ElMessage.warning(`有 ${invalidRecords.length} 条记录不符合条件（只能取消已计入未结算的记录）`)
        return
      }

      try {
        await ElMessageBox.confirm(
          `确认取消计入选中的 ${selectedRecords.value.length} 条记录？取消后这些订单将标记为"计入已取消"。`,
          '批量取消计入',
          {
            confirmButtonText: '确认',
            cancelButtonText: '取消',
            type: 'warning'
          }
        )

        cancelingAccount.value = true
        const commissionIds = selectedRecords.value.map(r => r.id)
        const res = await cancelAccountSalesCommissions({
          commission_ids: commissionIds
        })
        if (res.code === 200) {
          ElMessage.success(`取消计入成功，共取消 ${res.affected || 0} 条记录`)
          selectedRecords.value = []
          loadRecords()
        } else {
          ElMessage.error(res.message || '取消计入失败')
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error('批量取消计入失败:', error)
          ElMessage.error('批量取消计入失败: ' + (error.message || '未知错误'))
        }
      } finally {
        cancelingAccount.value = false
      }
    }

    // 单笔重新计入
    const handleResetAccount = async (row) => {
      try {
        await ElMessageBox.confirm(
          `确认重新计入订单 ${row.order_number} 的分成？这将把该订单从"计入已取消"状态恢复为"已计入"状态。`,
          '确认重新计入',
          {
            confirmButtonText: '确认',
            cancelButtonText: '取消',
            type: 'warning'
          }
        )

        resettingAccount.value = true
        const res = await resetAccountSalesCommissions({
          commission_ids: [row.id]
        })
        if (res.code === 200) {
          ElMessage.success(`重新计入成功，共重新计入 ${res.affected || 1} 条记录`)
          loadRecords()
        } else {
          ElMessage.error(res.message || '重新计入失败')
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error('重新计入失败:', error)
          ElMessage.error('重新计入失败: ' + (error.message || '未知错误'))
        }
      } finally {
        resettingAccount.value = false
      }
    }

    // 批量重新计入
    const handleBatchResetAccount = async () => {
      if (selectedRecords.value.length === 0) {
        ElMessage.warning('请先选择要重新计入的记录')
        return
      }

      // 检查是否都是已取消的记录
      const invalidRecords = selectedRecords.value.filter(r => !r.is_accounted_cancelled)
      if (invalidRecords.length > 0) {
        ElMessage.warning(`有 ${invalidRecords.length} 条记录不符合条件（只能重新计入已取消的记录）`)
        return
      }

      try {
        await ElMessageBox.confirm(
          `确认重新计入选中的 ${selectedRecords.value.length} 条记录？这将把这些订单从"计入已取消"状态恢复为"已计入"状态。`,
          '批量重新计入',
          {
            confirmButtonText: '确认',
            cancelButtonText: '取消',
            type: 'warning'
          }
        )

        resettingAccount.value = true
        const commissionIds = selectedRecords.value.map(r => r.id)
        const res = await resetAccountSalesCommissions({
          commission_ids: commissionIds
        })
        if (res.code === 200) {
          ElMessage.success(`重新计入成功，共重新计入 ${res.affected || 0} 条记录`)
          selectedRecords.value = []
          loadRecords()
        } else {
          ElMessage.error(res.message || '重新计入失败')
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error('批量重新计入失败:', error)
          ElMessage.error('批量重新计入失败: ' + (error.message || '未知错误'))
        }
      } finally {
        resettingAccount.value = false
      }
    }

    onMounted(() => {
      initMonth()
      loadSalesEmployees()
      loadStats()
    })

    return {
      loading,
      recordsLoading,
      configSaving,
      accounting,
      settling,
      cancelingAccount,
      resettingAccount,
      statsList,
      commissionRecords,
      salesEmployees,
      selectedEmployeeCode,
      selectedMonth,
      selectedStatus,
      dateRange,
      selectedRecords,
      activeTab,
      configDialogVisible,
      accountByDateDialogVisible,
      settleByDateDialogVisible,
      configFormRef,
      recordsPagination,
      configForm,
      accountByDateForm,
      settleByDateForm,
      configRules,
      handleSearch,
      handleViewDetails,
      handleConfig,
      handleConfirmConfig,
      handleRecordsPageChange,
      handleSelectionChange,
      handleAccount,
      handleSettle,
      handleBatchAccount,
      handleBatchSettle,
      handleBatchAccountByDate,
      handleConfirmAccountByDate,
      handleBatchSettleByDate,
      handleConfirmSettleByDate,
      handleCancelAccount,
      handleBatchCancelAccount,
      handleResetAccount,
      handleBatchResetAccount,
      formatDate,
      formatMoney,
      getEmployeeName,
      getTierTagType,
      getTierText
    }
  }
}
</script>

<style scoped>
.sales-commission-page {
  padding: 20px;
}

.commission-card {
  min-height: calc(100vh - 100px);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.title {
  display: flex;
  flex-direction: column;
}

.main {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
}

.sub {
  font-size: 14px;
  color: #909399;
  margin-top: 4px;
}

.actions {
  display: flex;
  align-items: center;
}

.stats-cards {
  margin-bottom: 30px;
}

.stat-card {
  margin-bottom: 20px;
}

.stat-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  padding-bottom: 10px;
  border-bottom: 1px solid #ebeef5;
}

.employee-name {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.stat-content {
  margin-bottom: 15px;
}

.stat-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  font-size: 14px;
}

.stat-item.total {
  margin-top: 10px;
  padding-top: 10px;
  border-top: 2px solid #ebeef5;
  font-size: 16px;
  font-weight: 600;
}

.stat-item .label {
  color: #606266;
}

.stat-item .value {
  color: #303133;
  font-weight: 500;
}

.stat-item .value.profit {
  color: #67c23a;
}

.stat-item .value.total-commission {
  color: #409eff;
  font-size: 18px;
}

.stat-actions {
  display: flex;
  gap: 10px;
  margin-top: 15px;
}

.detail-tabs {
  margin-top: 30px;
}

.table-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  padding: 15px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.toolbar-left {
  display: flex;
  gap: 10px;
}

.records-table {
  margin-top: 20px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.profit {
  color: #67c23a;
  font-weight: 600;
}

.new-customer {
  color: #e6a23c;
  font-weight: 600;
}

.total-commission {
  color: #409eff;
  font-weight: 600;
  font-size: 16px;
}
</style>

