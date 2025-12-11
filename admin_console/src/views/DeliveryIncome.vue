<template>
  <div class="delivery-income-page">
    <el-card class="income-card" shadow="never">
      <div class="card-header">
        <div class="title">
          <span class="main">配送费结算管理</span>
          <span class="sub">统计和结算配送员的配送费</span>
        </div>
        <div class="actions">
          <el-input
            v-model="employeeCodeFilter"
            placeholder="配送员工号"
            clearable
            style="width: 200px; margin-right: 10px;"
            @keyup.enter="handleSearch"
          />
          <el-button type="primary" @click="handleSearch">搜索</el-button>
        </div>
      </div>

      <el-table
        v-loading="loading"
        :data="incomeStats"
        border
        stripe
        class="income-table"
        empty-text="暂无数据"
        style="width: 100%"
      >
        <el-table-column prop="employee_code" label="配送员工号" align="center" min-width="120" />
        <el-table-column prop="employee_name" label="配送员姓名" align="center" min-width="120">
          <template #default="scope">
            <span>{{ scope.row.employee_name || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="已结算配送费" align="center" min-width="140">
          <template #default="scope">
            <span style="color: #67c23a; font-weight: 600;">
              ¥{{ formatMoney(scope.row.settled_fee) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="未结算配送费" align="center" min-width="140">
          <template #default="scope">
            <span style="color: #e6a23c; font-weight: 600;">
              ¥{{ formatMoney(scope.row.unsettled_fee) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="总配送费" align="center" min-width="140">
          <template #default="scope">
            <span style="color: #409eff; font-weight: 600;">
              ¥{{ formatMoney(scope.row.total_fee) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="order_count" label="订单数量" align="center" min-width="120" />
        <el-table-column label="平均每单配送成本" align="center" min-width="160">
          <template #default="scope">
            <span style="color: #606266; font-weight: 600;">
              ¥{{ formatMoney(scope.row.avg_fee_per_order) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right" align="center">
          <template #default="scope">
            <el-button
              type="info"
              size="small"
              @click="handleViewRecords(scope.row)"
            >
              查看记录
            </el-button>
            <el-button
              type="primary"
              size="small"
              :disabled="scope.row.unsettled_fee <= 0"
              @click="handleSettle(scope.row)"
            >
              批量结算
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 批量结算对话框 -->
    <el-dialog
      v-model="settleDialogVisible"
      title="批量结算配送费"
      width="500px"
    >
      <el-form :model="settleForm" label-width="120px">
        <el-form-item label="配送员工号">
          <el-input v-model="settleForm.employee_code" disabled />
        </el-form-item>
        <el-form-item label="配送员姓名">
          <el-input v-model="settleForm.employee_name" disabled />
        </el-form-item>
        <el-form-item label="未结算金额">
          <span style="color: #e6a23c; font-weight: 600; font-size: 16px;">
            ¥{{ formatMoney(settleForm.unsettled_fee) }}
          </span>
        </el-form-item>
        <el-form-item label="结算日期" required>
          <el-date-picker
            v-model="settleForm.settlement_date"
            type="date"
            placeholder="选择结算日期"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="结算订单">
          <el-radio-group v-model="settleForm.settleAll">
            <el-radio :label="true">结算所有未结算订单</el-radio>
            <el-radio :label="false">指定订单ID（暂不支持）</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="settleDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleConfirmSettle" :loading="settling">
          确认结算
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { getDeliveryIncomeStats, batchSettleDeliveryFees } from '../api/deliveryIncome'

export default {
  name: 'DeliveryIncome',
  data() {
    return {
      loading: false,
      settling: false,
      incomeStats: [],
      employeeCodeFilter: '',
      settleDialogVisible: false,
      settleForm: {
        employee_code: '',
        employee_name: '',
        unsettled_fee: 0,
        settlement_date: '',
        settleAll: true,
        order_ids: []
      }
    }
  },
  mounted() {
    this.loadIncomeStats()
  },
  methods: {
    async loadIncomeStats() {
      this.loading = true
      try {
        const res = await getDeliveryIncomeStats(this.employeeCodeFilter || null)
        if (res.code === 200) {
          this.incomeStats = res.data || []
        } else {
          this.$message.error(res.message || '获取收入统计失败')
        }
      } catch (error) {
        this.$message.error('获取收入统计失败: ' + (error.message || '未知错误'))
      } finally {
        this.loading = false
      }
    },
    handleSearch() {
      this.loadIncomeStats()
    },
    handleViewRecords(row) {
      // 跳转到配送记录页面，并自动筛选该配送员
      this.$router.push({
        path: '/delivery-records',
        query: {
          keyword: row.employee_code
        }
      })
    },
    handleSettle(row) {
      if (row.unsettled_fee <= 0) {
        this.$message.warning('该配送员没有未结算的配送费')
        return
      }
      this.settleForm = {
        employee_code: row.employee_code,
        employee_name: row.employee_name || '-',
        unsettled_fee: row.unsettled_fee,
        settlement_date: new Date().toISOString().split('T')[0], // 默认今天
        settleAll: true,
        order_ids: []
      }
      this.settleDialogVisible = true
    },
    async handleConfirmSettle() {
      if (!this.settleForm.settlement_date) {
        this.$message.warning('请选择结算日期')
        return
      }

      this.settling = true
      try {
        const data = {
          employee_code: this.settleForm.employee_code,
          settlement_date: this.settleForm.settlement_date
        }
        if (!this.settleForm.settleAll && this.settleForm.order_ids.length > 0) {
          data.order_ids = this.settleForm.order_ids
        }

        const res = await batchSettleDeliveryFees(data)
        if (res.code === 200) {
          this.$message.success(`批量结算成功，共结算 ${res.data.settled_count} 个订单`)
          this.settleDialogVisible = false
          this.loadIncomeStats()
        } else {
          this.$message.error(res.message || '批量结算失败')
        }
      } catch (error) {
        this.$message.error('批量结算失败: ' + (error.message || '未知错误'))
      } finally {
        this.settling = false
      }
    },
    formatMoney(amount) {
      if (amount == null || amount === undefined) return '0.00'
      return Number(amount).toFixed(2)
    }
  }
}
</script>

<style scoped>
.delivery-income-page {
  padding: 20px;
  width: 100%;
  height: 100%;
}

.income-card {
  border-radius: 8px;
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-shrink: 0;
}

.title .main {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
  margin-right: 10px;
}

.title .sub {
  font-size: 14px;
  color: #909399;
}

.actions {
  display: flex;
  align-items: center;
}

.income-table {
  margin-top: 20px;
  width: 100%;
  flex: 1;
}

/* 确保表格内容居中 */
.income-table :deep(.el-table__cell) {
  text-align: center;
}

/* 表头也居中 */
.income-table :deep(.el-table__header-wrapper th) {
  text-align: center;
}
</style>

