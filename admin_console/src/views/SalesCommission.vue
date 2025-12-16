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
          <el-table
            v-loading="recordsLoading"
            :data="commissionRecords"
            border
            stripe
            class="records-table"
            empty-text="暂无数据"
            style="width: 100%"
          >
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
import { ElMessage } from 'element-plus'
import {
  getSalesCommissionStats,
  getSalesCommissions,
  getSalesCommissionConfig,
  updateSalesCommissionConfig,
  getSalesEmployees
} from '../api/salesCommission'

export default {
  name: 'SalesCommission',
  setup() {
    const loading = ref(false)
    const recordsLoading = ref(false)
    const configSaving = ref(false)
    const statsList = ref([])
    const commissionRecords = ref([])
    const salesEmployees = ref([])
    const selectedEmployeeCode = ref('')
    const selectedMonth = ref('')
    const activeTab = ref('records')
    const configDialogVisible = ref(false)
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
        const res = await getSalesCommissions(
          selectedEmployeeCode.value,
          selectedMonth.value || null,
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

    onMounted(() => {
      initMonth()
      loadSalesEmployees()
      loadStats()
    })

    return {
      loading,
      recordsLoading,
      configSaving,
      statsList,
      commissionRecords,
      salesEmployees,
      selectedEmployeeCode,
      selectedMonth,
      activeTab,
      configDialogVisible,
      configFormRef,
      recordsPagination,
      configForm,
      configRules,
      handleSearch,
      handleViewDetails,
      handleConfig,
      handleConfirmConfig,
      handleRecordsPageChange,
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

