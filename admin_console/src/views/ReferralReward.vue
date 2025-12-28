<template>
  <div class="referral-reward-container">
    <el-card>
      <h2 class="page-title">推荐奖励活动管理</h2>

      <!-- Tab切换：活动配置 / 奖励记录 -->
      <el-tabs v-model="activeTab" class="reward-tabs">
        <!-- 活动配置 -->
        <el-tab-pane label="活动配置" name="config">
          <el-card class="config-card">
            <el-form
              ref="configFormRef"
              :model="configForm"
              :rules="configRules"
              label-width="150px"
              style="max-width: 800px;"
            >
              <el-form-item label="是否启用" prop="is_enabled">
                <el-switch
                  v-model="configForm.is_enabled"
                  active-text="启用"
                  inactive-text="禁用"
                />
                <div style="color: #909399; font-size: 12px; margin-top: 4px;">
                  启用后，老用户推荐的新用户首次下单完成付款时，将自动发放奖励给老用户
                </div>
              </el-form-item>

              <el-form-item label="奖励类型" prop="reward_type">
                <el-radio-group v-model="configForm.reward_type" @change="handleRewardTypeChange">
                  <el-radio label="points">积分</el-radio>
                  <el-radio label="coupon">优惠券</el-radio>
                  <el-radio label="amount">金额</el-radio>
                </el-radio-group>
              </el-form-item>

              <el-form-item
                v-if="configForm.reward_type === 'points'"
                label="奖励积分"
                prop="reward_value"
              >
                <el-input-number
                  v-model="configForm.reward_value"
                  :min="0"
                  :precision="0"
                  placeholder="请输入奖励积分数量"
                  style="width: 200px;"
                />
                <div style="color: #909399; font-size: 12px; margin-top: 4px;">
                  新用户首次下单完成付款后，将给推荐人发放对应数量的积分
                </div>
              </el-form-item>

              <el-form-item
                v-if="configForm.reward_type === 'coupon'"
                label="优惠券"
                prop="coupon_id"
              >
                <el-select
                  v-model="configForm.coupon_id"
                  placeholder="请选择优惠券"
                  filterable
                  style="width: 300px;"
                  @change="handleCouponChange"
                >
                  <el-option
                    v-for="coupon in availableCoupons"
                    :key="coupon.id"
                    :label="`${coupon.name} (${coupon.type === 'delivery_fee' ? '配送费券' : '金额券'})`"
                    :value="coupon.id"
                  />
                </el-select>
                <el-button
                  type="primary"
                  link
                  style="margin-left: 10px;"
                  @click="loadAvailableCoupons"
                >
                  刷新优惠券列表
                </el-button>
                <div style="color: #909399; font-size: 12px; margin-top: 4px;">
                  新用户首次下单完成付款后，将给推荐人发放一张选中的优惠券
                </div>
              </el-form-item>

              <el-form-item
                v-if="configForm.reward_type === 'amount'"
                label="奖励金额"
                prop="reward_value"
              >
                <el-input-number
                  v-model="configForm.reward_value"
                  :min="0"
                  :precision="2"
                  placeholder="请输入奖励金额"
                  style="width: 200px;"
                />
                <span style="margin-left: 8px;">元</span>
                <div style="color: #909399; font-size: 12px; margin-top: 4px;">
                  新用户首次下单完成付款后，将给推荐人发放对应金额（需要配合其他系统实现实际发放）
                </div>
              </el-form-item>

              <el-form-item label="活动说明" prop="description">
                <el-input
                  v-model="configForm.description"
                  type="textarea"
                  :rows="3"
                  placeholder="请输入活动说明"
                  maxlength="500"
                  show-word-limit
                />
              </el-form-item>

              <el-form-item>
                <el-button type="primary" @click="handleSaveConfig" :loading="saving">
                  保存配置
                </el-button>
                <el-button @click="handleResetConfig">重置</el-button>
              </el-form-item>
            </el-form>
          </el-card>
        </el-tab-pane>

        <!-- 奖励记录 -->
        <el-tab-pane label="奖励记录" name="rewards">
          <div class="toolbar" style="margin-bottom: 12px;">
            <el-input
              v-model="searchKeyword"
              placeholder="按推荐人/新用户/订单号搜索"
              clearable
              style="width: 260px; margin-right: 10px;"
              @keyup.enter="loadRewards"
            />
            <el-select
              v-model="filterStatus"
              placeholder="奖励状态"
              clearable
              style="width: 150px; margin-right: 10px;"
              @change="loadRewards"
            >
              <el-option label="待发放" value="pending" />
              <el-option label="已发放" value="completed" />
              <el-option label="发放失败" value="failed" />
            </el-select>
            <el-button type="primary" @click="loadRewards">搜索</el-button>
          </div>

          <el-card class="rewards-card">
            <el-table :data="rewards" stripe v-loading="rewardsLoading">
              <el-table-column prop="id" label="ID" align="center" width="80" />
              <el-table-column label="推荐人" min-width="150">
                <template #default="{ row }">
                  <div>
                    <span>{{ row.referrer_name || '-' }}</span>
                    <span v-if="row.referrer_code" style="margin-left: 8px; color: #909399; font-size: 12px;">
                      (编号 {{ row.referrer_code }})
                    </span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column label="新用户" min-width="150">
                <template #default="{ row }">
                  <div>
                    <span>{{ row.new_user_name || '-' }}</span>
                    <span v-if="row.new_user_code" style="margin-left: 8px; color: #909399; font-size: 12px;">
                      (编号 {{ row.new_user_code }})
                    </span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column prop="order_number" label="订单号" min-width="150" />
              <el-table-column label="奖励类型" align="center" width="100">
                <template #default="{ row }">
                  <el-tag :type="getRewardTypeTag(row.reward_type)">
                    {{ getRewardTypeText(row.reward_type) }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="奖励值" align="center" width="120">
                <template #default="{ row }">
                  <span v-if="row.reward_type === 'points'">{{ row.reward_value }} 积分</span>
                  <span v-else-if="row.reward_type === 'coupon'">优惠券ID: {{ row.coupon_id }}</span>
                  <span v-else>¥{{ (row.reward_value || 0).toFixed(2) }}</span>
                </template>
              </el-table-column>
              <el-table-column label="状态" align="center" width="100">
                <template #default="{ row }">
                  <el-tag :type="getStatusTag(row.status)">
                    {{ getStatusText(row.status) }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="发放时间" min-width="180">
                <template #default="{ row }">
                  <span v-if="row.reward_at">{{ formatDate(row.reward_at) }}</span>
                  <span v-else>--</span>
                </template>
              </el-table-column>
              <el-table-column label="创建时间" min-width="180">
                <template #default="{ row }">
                  <span>{{ formatDate(row.created_at) }}</span>
                </template>
              </el-table-column>
              <el-table-column label="备注" min-width="150" show-overflow-tooltip>
                <template #default="{ row }">
                  <span>{{ row.remark || '-' }}</span>
                </template>
              </el-table-column>
            </el-table>

            <div style="margin-top: 12px; text-align: right;" v-if="rewardsTotal > 0">
              <el-pagination
                v-model:current-page="rewardsPageNum"
                v-model:page-size="rewardsPageSize"
                :total="rewardsTotal"
                :page-sizes="[10, 20, 50, 100]"
                layout="total, sizes, prev, pager, next, jumper"
                @size-change="loadRewards"
                @current-change="loadRewards"
              />
            </div>
          </el-card>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getReferralRewardConfig, updateReferralRewardConfig, getReferralRewards } from '../api/referralReward'
import { getCoupons as getAllCoupons } from '../api/coupons'

const activeTab = ref('config')
const configFormRef = ref(null)
const saving = ref(false)
const rewardsLoading = ref(false)
const availableCoupons = ref([])

const configForm = reactive({
  id: 0,
  is_enabled: false,
  reward_type: 'points',
  reward_value: 0,
  coupon_id: null,
  description: ''
})

const configRules = {
  reward_type: [
    { required: true, message: '请选择奖励类型', trigger: 'change' }
  ],
  reward_value: [
    { required: true, message: '请输入奖励值', trigger: 'blur' },
    { type: 'number', min: 0, message: '奖励值必须大于等于0', trigger: 'blur' }
  ],
  coupon_id: [
    { required: true, message: '请选择优惠券', trigger: 'change' }
  ]
}

const rewards = ref([])
const rewardsTotal = ref(0)
const rewardsPageNum = ref(1)
const rewardsPageSize = ref(10)
const searchKeyword = ref('')
const filterStatus = ref('')

// 加载活动配置
const loadConfig = async () => {
  try {
    const res = await getReferralRewardConfig()
    if (res.code === 200 && res.data) {
      Object.assign(configForm, {
        id: res.data.id || 0,
        is_enabled: res.data.is_enabled || false,
        reward_type: res.data.reward_type || 'points',
        reward_value: res.data.reward_value || 0,
        coupon_id: res.data.coupon_id || null,
        description: res.data.description || ''
      })
    }
  } catch (error) {
    console.error('加载配置失败:', error)
    ElMessage.error('加载配置失败')
  }
}

// 加载可用优惠券列表
const loadAvailableCoupons = async () => {
  try {
    const res = await getAllCoupons()
    if (res && res.data) {
      // 只显示启用状态的优惠券
      const coupons = Array.isArray(res.data) ? res.data : []
      availableCoupons.value = coupons.filter(c => c.status === 1)
    }
  } catch (error) {
    console.error('加载优惠券列表失败:', error)
    ElMessage.error('加载优惠券列表失败')
  }
}

// 奖励类型改变
const handleRewardTypeChange = (value) => {
  if (value !== 'coupon') {
    configForm.coupon_id = null
  }
  if (value === 'coupon' && availableCoupons.value.length === 0) {
    loadAvailableCoupons()
  }
}

// 优惠券改变
const handleCouponChange = (value) => {
  // 可以在这里添加其他逻辑
}

// 保存配置
const handleSaveConfig = async () => {
  if (!configFormRef.value) return

  try {
    await configFormRef.value.validate()
    saving.value = true

    // 如果奖励类型是coupon，验证coupon_id
    if (configForm.reward_type === 'coupon' && !configForm.coupon_id) {
      ElMessage.error('请选择优惠券')
      saving.value = false
      return
    }

    const res = await updateReferralRewardConfig({
      id: configForm.id,
      is_enabled: configForm.is_enabled,
      reward_type: configForm.reward_type,
      reward_value: configForm.reward_value,
      coupon_id: configForm.reward_type === 'coupon' ? configForm.coupon_id : null,
      description: configForm.description
    })

    if (res.code === 200) {
      ElMessage.success('保存成功')
      loadConfig()
    } else {
      ElMessage.error(res.message || '保存失败')
    }
  } catch (error) {
    if (error !== false) {
      console.error('保存配置失败:', error)
      ElMessage.error('保存配置失败: ' + (error.message || '未知错误'))
    }
  } finally {
    saving.value = false
  }
}

// 重置配置
const handleResetConfig = () => {
  loadConfig()
}

// 加载奖励记录
const loadRewards = async () => {
  try {
    rewardsLoading.value = true
    const params = {
      page_num: rewardsPageNum.value,
      page_size: rewardsPageSize.value
    }

    if (filterStatus.value) {
      params.status = filterStatus.value
    }

    // 如果有关键词，尝试解析为referrer_id或new_user_id
    if (searchKeyword.value) {
      // 这里可以根据需要实现更复杂的搜索逻辑
      // 暂时只支持状态筛选
    }

    const res = await getReferralRewards(params)
    if (res.code === 200 && res.data) {
      rewards.value = res.data.list || []
      rewardsTotal.value = res.data.total || 0
    }
  } catch (error) {
    console.error('加载奖励记录失败:', error)
    ElMessage.error('加载奖励记录失败')
  } finally {
    rewardsLoading.value = false
  }
}

// 格式化日期
const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

// 获取奖励类型文本
const getRewardTypeText = (type) => {
  const map = {
    points: '积分',
    coupon: '优惠券',
    amount: '金额'
  }
  return map[type] || type
}

// 获取奖励类型标签
const getRewardTypeTag = (type) => {
  const map = {
    points: 'success',
    coupon: 'warning',
    amount: 'danger'
  }
  return map[type] || ''
}

// 获取状态文本
const getStatusText = (status) => {
  const map = {
    pending: '待发放',
    completed: '已发放',
    failed: '发放失败'
  }
  return map[status] || status
}

// 获取状态标签
const getStatusTag = (status) => {
  const map = {
    pending: 'info',
    completed: 'success',
    failed: 'danger'
  }
  return map[status] || ''
}

onMounted(() => {
  loadConfig()
  loadRewards()
  loadAvailableCoupons()
})
</script>

<style scoped>
.referral-reward-container {
  padding: 20px;
}

.page-title {
  margin: 0 0 20px 0;
  font-size: 20px;
  font-weight: 600;
}

.reward-tabs {
  margin-top: 20px;
}

.config-card,
.rewards-card {
  margin-top: 20px;
}

.toolbar {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
}
</style>

