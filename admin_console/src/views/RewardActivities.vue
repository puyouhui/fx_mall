<template>
  <div class="reward-activities-container">
    <el-card>
      <h2 class="page-title">奖励活动管理</h2>

      <!-- Tab切换：活动列表 / 奖励记录 -->
      <el-tabs v-model="activeTab" class="reward-tabs">
        <!-- 活动列表 -->
        <el-tab-pane label="活动列表" name="activities">
          <div class="toolbar">
            <el-select
              v-model="filterActivityType"
              placeholder="活动类型"
              clearable
              style="width: 150px; margin-right: 10px;"
              @change="loadActivities"
            >
              <el-option label="拉新活动" value="referral" />
              <el-option label="新客奖励" value="new_customer" />
            </el-select>
            <el-button type="primary" @click="handleAddActivity">
              <el-icon><Plus /></el-icon>
              添加活动
            </el-button>
          </div>

          <el-card class="activities-card">
            <el-table :data="activities" stripe v-loading="activitiesLoading" style="width: 100%">
              <el-table-column prop="id" label="ID" align="center" width="80" />
              <el-table-column prop="activity_name" label="活动名称" min-width="150" />
              <el-table-column label="活动类型" align="center" width="120">
                <template #default="{ row }">
                  <el-tag :type="row.activity_type === 'referral' ? 'success' : 'warning'">
                    {{ row.activity_type === 'referral' ? '拉新活动' : '新客奖励' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="奖励类型" align="center" width="100">
                <template #default="{ row }">
                  <el-tag :type="getRewardTypeTag(row.reward_type)">
                    {{ getRewardTypeText(row.reward_type) }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="奖励值" align="center" width="180">
                <template #default="{ row }">
                  <span v-if="row.reward_type === 'points'">{{ row.reward_value }} 积分</span>
                  <span v-else-if="row.reward_type === 'coupon'">
                    优惠券ID:
                    <span v-if="Array.isArray(row.coupon_ids) && row.coupon_ids.length">
                      {{ row.coupon_ids.join(',') }}
                    </span>
                    <span v-else-if="row.coupon_id">
                      {{ row.coupon_id }}
                    </span>
                    <span v-else>--</span>
                  </span>
                  <span v-else>¥{{ (row.reward_value || 0).toFixed(2) }}</span>
                </template>
              </el-table-column>
              <el-table-column label="是否启用" align="center" width="100">
                <template #default="{ row }">
                  <el-switch
                    v-model="row.is_enabled"
                    @change="handleToggleEnable(row)"
                    :loading="row._toggleLoading"
                  />
                </template>
              </el-table-column>
              <el-table-column label="创建时间" min-width="180">
                <template #default="{ row }">
                  <span>{{ formatDate(row.created_at) }}</span>
                </template>
              </el-table-column>
              <el-table-column label="操作" align="center" width="180" fixed="right">
                <template #default="{ row }">
                  <el-button type="primary" link size="small" @click="handleEditActivity(row)">
                    编辑
                  </el-button>
                  <el-button type="danger" link size="small" @click="handleDeleteActivity(row)">
                    删除
                  </el-button>
                </template>
              </el-table-column>
            </el-table>

            <div style="margin-top: 12px; text-align: right;" v-if="activitiesTotal > 0">
              <el-pagination
                v-model:current-page="activitiesPageNum"
                v-model:page-size="activitiesPageSize"
                :total="activitiesTotal"
                :page-sizes="[10, 20, 50, 100]"
                layout="total, sizes, prev, pager, next, jumper"
                @size-change="loadActivities"
                @current-change="loadActivities"
              />
            </div>
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

    <!-- 活动表单弹框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="activityFormRef"
        :model="activityForm"
        :rules="activityRules"
        label-width="120px"
      >
        <el-form-item label="活动名称" prop="activity_name">
          <el-input
            v-model="activityForm.activity_name"
            placeholder="请输入活动名称"
            maxlength="100"
            show-word-limit
          />
        </el-form-item>

        <el-form-item label="活动类型" prop="activity_type">
          <el-radio-group v-model="activityForm.activity_type">
            <el-radio label="referral">拉新活动（奖励老客户）</el-radio>
            <el-radio label="new_customer">新客奖励（奖励新客户）</el-radio>
          </el-radio-group>
          <div style="color: #909399; font-size: 12px; margin-top: 4px;">
            <div>拉新活动：老用户推荐新用户首次下单完成付款时，奖励给老用户</div>
            <div>新客奖励：新用户首次下单完成付款时，奖励给新用户</div>
          </div>
        </el-form-item>

        <el-form-item label="是否启用" prop="is_enabled">
          <el-switch
            v-model="activityForm.is_enabled"
            active-text="启用"
            inactive-text="禁用"
          />
        </el-form-item>

        <el-form-item label="奖励类型" prop="reward_type">
          <el-radio-group v-model="activityForm.reward_type" @change="handleRewardTypeChange">
            <el-radio label="points">积分</el-radio>
            <el-radio label="coupon">优惠券</el-radio>
            <el-radio label="amount">金额</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item
          v-if="activityForm.reward_type === 'points'"
          label="奖励积分"
          prop="reward_value"
        >
          <el-input-number
            v-model="activityForm.reward_value"
            :min="0"
            :precision="0"
            placeholder="请输入奖励积分数量"
            style="width: 200px;"
          />
        </el-form-item>

        <el-form-item
          v-if="activityForm.reward_type === 'coupon'"
          label="优惠券"
          prop="coupon_ids"
        >
          <el-select
            v-model="activityForm.coupon_ids"
            placeholder="请选择优惠券（可多选）"
            filterable
            multiple
            collapse-tags
            collapse-tags-tooltip
            style="width: 300px;"
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
        </el-form-item>

        <el-form-item
          v-if="activityForm.reward_type === 'amount'"
          label="奖励金额"
          prop="reward_value"
        >
          <el-input-number
            v-model="activityForm.reward_value"
            :min="0"
            :precision="2"
            placeholder="请输入奖励金额"
            style="width: 200px;"
          />
          <span style="margin-left: 8px;">元</span>
        </el-form-item>

        <el-form-item label="活动说明" prop="description">
          <el-input
            v-model="activityForm.description"
            type="textarea"
            :rows="3"
            placeholder="请输入活动说明"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveActivity" :loading="saving">
          保存
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { getRewardActivities, createRewardActivity, updateRewardActivity, deleteRewardActivity } from '../api/rewardActivities'
import { getReferralRewards } from '../api/referralReward'
import { getCoupons as getAllCoupons } from '../api/coupons'

const activeTab = ref('activities')
const dialogVisible = ref(false)
const dialogTitle = computed(() => editingActivity.value ? '编辑活动' : '添加活动')
const activityFormRef = ref(null)
const saving = ref(false)
const activitiesLoading = ref(false)
const rewardsLoading = ref(false)
const availableCoupons = ref([])
const editingActivity = ref(null)

// 活动列表相关
const activities = ref([])
const activitiesTotal = ref(0)
const activitiesPageNum = ref(1)
const activitiesPageSize = ref(10)
const filterActivityType = ref('')

// 奖励记录相关
const rewards = ref([])
const rewardsTotal = ref(0)
const rewardsPageNum = ref(1)
const rewardsPageSize = ref(10)
const searchKeyword = ref('')
const filterStatus = ref('')

// 活动表单
const activityForm = reactive({
  activity_name: '',
  activity_type: 'referral',
  is_enabled: true,  // 默认启用
  reward_type: 'points',
  reward_value: 0,
  coupon_ids: [],
  description: ''
})

// 表单验证规则
const activityRules = {
  activity_name: [
    { required: true, message: '请输入活动名称', trigger: 'blur' }
  ],
  activity_type: [
    { required: true, message: '请选择活动类型', trigger: 'change' }
  ],
  reward_type: [
    { required: true, message: '请选择奖励类型', trigger: 'change' }
  ],
  reward_value: [
    { required: true, message: '请输入奖励值', trigger: 'blur' },
    { type: 'number', min: 0, message: '奖励值必须大于等于0', trigger: 'blur' }
  ],
  coupon_ids: [
    { required: true, message: '请至少选择一张优惠券', trigger: 'change' }
  ]
}

// 加载活动列表
const loadActivities = async () => {
  try {
    activitiesLoading.value = true
    const params = {
      page_num: activitiesPageNum.value,
      page_size: activitiesPageSize.value
    }

    if (filterActivityType.value) {
      params.activity_type = filterActivityType.value
    }

    const res = await getRewardActivities(params)
    if (res.code === 200 && res.data) {
      activities.value = (res.data.list || []).map(item => ({
        ...item,
        _toggleLoading: false
      }))
      activitiesTotal.value = res.data.total || 0
    }
  } catch (error) {
    console.error('加载活动列表失败:', error)
    ElMessage.error('加载活动列表失败')
  } finally {
    activitiesLoading.value = false
  }
}

// 加载可用优惠券列表
const loadAvailableCoupons = async () => {
  try {
    const res = await getAllCoupons()
    // getAllCoupons 在 api/coupons.js 中直接返回的是 data（数组），不是 { code, data }
    const coupons = Array.isArray(res) ? res : (Array.isArray(res?.data) ? res.data : [])
    availableCoupons.value = coupons.filter(c => c.status === 1)
  } catch (error) {
    console.error('加载优惠券列表失败:', error)
    ElMessage.error('加载优惠券列表失败')
  }
}

// 奖励类型改变
const handleRewardTypeChange = (value) => {
  if (value !== 'coupon') {
    activityForm.coupon_ids = []
  }
  if (value === 'coupon' && availableCoupons.value.length === 0) {
    loadAvailableCoupons()
  }
}

// 添加活动
const handleAddActivity = () => {
  editingActivity.value = null
  resetActivityForm()
  dialogVisible.value = true
  loadAvailableCoupons()
}

// 编辑活动
const handleEditActivity = (activity) => {
  editingActivity.value = activity
  Object.assign(activityForm, {
    activity_name: activity.activity_name,
    activity_type: activity.activity_type,
    is_enabled: activity.is_enabled,
    reward_type: activity.reward_type,
    reward_value: activity.reward_value,
    coupon_ids: Array.isArray(activity.coupon_ids)
      ? [...activity.coupon_ids]
      : (activity.coupon_id ? [activity.coupon_id] : []),
    description: activity.description || ''
  })
  dialogVisible.value = true
  if (activity.reward_type === 'coupon') {
    loadAvailableCoupons()
  }
}

// 重置活动表单
const resetActivityForm = () => {
  Object.assign(activityForm, {
    activity_name: '',
    activity_type: 'referral',
    is_enabled: true,  // 默认启用
    reward_type: 'points',
    reward_value: 0,
    coupon_ids: [],
    description: ''
  })
  if (activityFormRef.value) {
    activityFormRef.value.clearValidate()
  }
}

// 保存活动
const handleSaveActivity = async () => {
  if (!activityFormRef.value) return

  try {
    await activityFormRef.value.validate()
    
    // 如果奖励类型是coupon，验证coupon_ids
    if (activityForm.reward_type === 'coupon' && (!Array.isArray(activityForm.coupon_ids) || activityForm.coupon_ids.length === 0)) {
      ElMessage.error('请至少选择一张优惠券')
      return
    }

    saving.value = true

    const formData = {
      activity_name: activityForm.activity_name,
      activity_type: activityForm.activity_type,
      is_enabled: activityForm.is_enabled,
      reward_type: activityForm.reward_type,
      reward_value: activityForm.reward_value,
      coupon_ids: activityForm.reward_type === 'coupon' ? activityForm.coupon_ids : [],
      description: activityForm.description
    }

    let res
    if (editingActivity.value) {
      // 更新
      res = await updateRewardActivity(editingActivity.value.id, formData)
    } else {
      // 创建
      res = await createRewardActivity(formData)
    }

    console.log('保存活动响应:', res)
    console.log('响应 code:', res?.code)
    console.log('响应 message:', res?.message)

    // 检查响应数据
    if (res && (res.code === 200 || res.code === '200')) {
      ElMessage.success(editingActivity.value ? '更新成功' : '创建成功')
      dialogVisible.value = false
      loadActivities()
    } else {
      // 如果 code 不是 200，显示错误消息
      const errorMsg = res?.message || '操作失败'
      console.error('操作失败，响应:', res)
      ElMessage.error(errorMsg)
    }
  } catch (error) {
    console.error('保存活动异常:', error)
    // 处理错误响应
    let errorMsg = '保存活动失败'
    if (error.response && error.response.data) {
      const errorData = error.response.data
      errorMsg = errorData.message || (typeof errorData === 'string' ? errorData : '操作失败')
    } else if (error.message) {
      errorMsg = error.message
    }
    ElMessage.error(errorMsg)
  } finally {
    saving.value = false
  }
}

// 删除活动
const handleDeleteActivity = async (activity) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除活动"${activity.activity_name}"吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const res = await deleteRewardActivity(activity.id)
    if (res.code === 200) {
      ElMessage.success('删除成功')
      loadActivities()
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除活动失败:', error)
      ElMessage.error('删除活动失败')
    }
  }
}

// 切换启用状态
const handleToggleEnable = async (activity) => {
  try {
    activity._toggleLoading = true
    const formData = {
      activity_name: activity.activity_name,
      activity_type: activity.activity_type,
      is_enabled: activity.is_enabled,
      reward_type: activity.reward_type,
      reward_value: activity.reward_value,
      coupon_ids: Array.isArray(activity.coupon_ids)
        ? activity.coupon_ids
        : (activity.coupon_id ? [activity.coupon_id] : []),
      description: activity.description || ''
    }

    const res = await updateRewardActivity(activity.id, formData)
    if (res.code === 200) {
      ElMessage.success(activity.is_enabled ? '已启用' : '已禁用')
    } else {
      // 回滚状态
      activity.is_enabled = !activity.is_enabled
      ElMessage.error(res.message || '操作失败')
    }
  } catch (error) {
    // 回滚状态
    activity.is_enabled = !activity.is_enabled
    console.error('切换启用状态失败:', error)
    ElMessage.error('操作失败')
  } finally {
    activity._toggleLoading = false
  }
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
  loadActivities()
  loadRewards()
})
</script>

<style scoped>
.reward-activities-container {
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

.toolbar {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
}

.activities-card,
.rewards-card {
  margin-top: 20px;
}
</style>
