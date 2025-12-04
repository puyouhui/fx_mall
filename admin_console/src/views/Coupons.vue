<template>
  <div class="coupons-container">
    <el-card>
      <h2 class="page-title">优惠券管理</h2>

      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-right">
          <el-button type="primary" @click="handleAddCoupon">
            <el-icon>
              <Plus />
            </el-icon>
            新增优惠券
          </el-button>
        </div>
      </div>

      <!-- 列表 Tab：优惠券列表 / 发放记录 / 使用记录 -->
      <el-tabs v-model="activeTab" class="coupons-tabs">
        <el-tab-pane label="优惠券列表" name="coupons">
          <el-card class="coupons-card">
            <el-table :data="coupons" stripe v-loading="loading">
              <el-table-column prop="id" label="ID" align="center" width="80" />
              <el-table-column prop="name" label="优惠券名称" min-width="150" />
              <el-table-column label="类型" align="center" width="120">
                <template #default="{ row }">
                  <el-tag :type="row.type === 'delivery_fee' ? 'success' : 'warning'">
                    {{ row.type === 'delivery_fee' ? '配送费券' : '金额券' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="优惠值" align="center" width="120">
                <template #default="{ row }">
                  <span v-if="row.type === 'delivery_fee'">免配送费</span>
                  <span v-else>¥{{ (row.discount_value || 0).toFixed(2) }}</span>
                </template>
              </el-table-column>
              <el-table-column label="使用条件" align="center" min-width="150">
                <template #default="{ row }">
                  <div v-if="(row.min_amount || 0) > 0">满¥{{ (row.min_amount || 0).toFixed(2) }}可用</div>
                  <div v-else>无门槛</div>
                  <div v-if="row.category_ids && row.category_ids.length > 0" style="color: #909399; font-size: 12px;">
                    指定分类
                  </div>
                </template>
              </el-table-column>
              <el-table-column label="已发放/已使用" align="center" width="140">
                <template #default="{ row }">
                  <div>{{ row.issued_count || 0 }}/{{ row.used_count || 0 }}</div>
                  <div v-if="row.total_count > 0" style="color: #909399; font-size: 12px;">
                    总量: {{ row.total_count }}
                  </div>
                </template>
              </el-table-column>

              <el-table-column label="数量限制" align="center" width="140">
                <template #default="{ row }">
                  <div v-if="row.total_count > 0">限制数量: {{ row.total_count }}</div>
                  <div v-else>不限制</div>
                  <div v-if="row.total_count > 0" style="color: #909399; font-size: 12px;">
                    总量: {{ row.total_count }}
                  </div>
                </template>
              </el-table-column>
              <el-table-column label="有效期" align="center" min-width="200">
                <template #default="{ row }">
                  <div>{{ row.valid_from ? formatDate(row.valid_from) : '-' }}</div>
                  <div style="color: #909399; font-size: 12px;">至 {{ row.valid_to ? formatDate(row.valid_to) : '-' }}</div>
                </template>
              </el-table-column>
              <el-table-column label="状态" align="center" width="100">
                <template #default="{ row }">
                  <el-tag :type="row.status === 1 ? 'success' : 'info'">
                    {{ row.status === 1 ? '启用' : '禁用' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="操作" fixed="right" align="center" width="250">
                <template #default="{ row }">
                  <el-button type="success" size="small" @click="handleIssueCoupon(row)">
                    发放
                  </el-button>
                  <el-button type="primary" size="small" @click="handleEditCoupon(row)">
                    编辑
                  </el-button>
                  <el-button type="danger" size="small" @click="handleDeleteCoupon(row.id)">
                    删除
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-tab-pane>

        <el-tab-pane label="发放记录" name="issues">
          <div class="toolbar" style="margin-bottom: 12px;">
            <el-input
              v-model="issueKeyword"
              placeholder="按优惠券名称 / 发放人 / 原因搜索"
              clearable
              style="width: 260px; margin-right: 10px;"
              @keyup.enter="loadCouponIssues"
            />
            <el-button type="primary" @click="loadCouponIssues">搜索</el-button>
          </div>
          <el-card class="coupons-card">
            <el-table :data="couponIssues" stripe v-loading="issuesLoading">
              <el-table-column prop="id" label="ID" align="center" width="80" />
              <el-table-column prop="coupon_name" label="优惠券名称" min-width="160" />
              <el-table-column label="用户" min-width="200">
                <template #default="{ row }">
                  <div>
                    <span>{{ row.user_name || '-' }}</span>
                    <span v-if="row.user_code" style="margin-left: 8px; color: #909399; font-size: 12px;">
                      (编号 {{ row.user_code }})
                    </span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column prop="quantity" label="数量" align="center" width="80" />
              <el-table-column prop="reason" label="发放原因" min-width="160" />
              <el-table-column label="发放人" min-width="160">
                <template #default="{ row }">
                  <span>{{ row.operator_name || '-' }}</span>
                  <el-tag
                    v-if="row.operator_type"
                    size="small"
                    style="margin-left: 6px;"
                  >
                    {{ row.operator_type === 'employee' ? '员工' : '管理员' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="到期时间" min-width="180">
                <template #default="{ row }">
                  <span v-if="row.expires_at">{{ formatDate(row.expires_at) }}</span>
                  <span v-else>--</span>
                </template>
              </el-table-column>
              <el-table-column label="发放时间" min-width="180">
                <template #default="{ row }">
                  <span>{{ formatDate(row.created_at) }}</span>
                </template>
              </el-table-column>
            </el-table>
            <div style="margin-top: 12px; text-align: right;" v-if="issuesTotal > 0">
              <el-pagination
                v-model:current-page="issuesPageNum"
                v-model:page-size="issuesPageSize"
                :total="issuesTotal"
                layout="prev, pager, next, jumper"
                @current-change="loadCouponIssues"
                @size-change="loadCouponIssues"
              />
            </div>
          </el-card>
        </el-tab-pane>

        <el-tab-pane label="使用记录" name="usages">
          <el-card class="coupons-card">
            <div class="search-bar" style="margin-bottom: 16px;">
              <el-input
                v-model="usageKeyword"
                placeholder="搜索用户姓名、手机号或订单号"
                style="width: 300px; margin-right: 12px;"
                clearable
                @clear="loadCouponUsages"
                @keyup.enter="loadCouponUsages"
              >
                <template #prefix>
                  <el-icon><Search /></el-icon>
                </template>
              </el-input>
              <el-button type="primary" @click="loadCouponUsages">搜索</el-button>
            </div>
            <el-table :data="couponUsages" stripe v-loading="usagesLoading">
              <el-table-column prop="id" label="ID" align="center" width="80" />
              <el-table-column label="用户信息" min-width="150">
                <template #default="{ row }">
                  <div>{{ row.user_name || '未知' }}</div>
                  <div style="color: #909399; font-size: 12px;">{{ row.user_phone || '-' }}</div>
                </template>
              </el-table-column>
              <el-table-column label="优惠券信息" min-width="200">
                <template #default="{ row }">
                  <div>{{ row.coupon_name || '-' }}</div>
                  <div style="color: #909399; font-size: 12px;">
                    <el-tag :type="row.coupon_type === 'delivery_fee' ? 'success' : 'warning'" size="small">
                      {{ row.coupon_type === 'delivery_fee' ? '配送费券' : '金额券' }}
                    </el-tag>
                    <span v-if="row.coupon_type === 'amount'" style="margin-left: 8px;">
                      ¥{{ (row.discount_value || 0).toFixed(2) }}
                    </span>
                  </div>
                </template>
              </el-table-column>
              <el-table-column label="订单信息" min-width="150">
                <template #default="{ row }">
                  <div v-if="row.order_number">
                    <el-link type="primary" :underline="false">{{ row.order_number }}</el-link>
                  </div>
                  <div v-else style="color: #909399;">-</div>
                </template>
              </el-table-column>
              <el-table-column label="使用时间" min-width="180">
                <template #default="{ row }">
                  <span>{{ formatDate(row.used_at) }}</span>
                </template>
              </el-table-column>
              <el-table-column label="发放时间" min-width="180">
                <template #default="{ row }">
                  <span>{{ formatDate(row.created_at) }}</span>
                </template>
              </el-table-column>
            </el-table>
            <div style="margin-top: 12px; text-align: right;" v-if="usagesTotal > 0">
              <el-pagination
                v-model:current-page="usagesPageNum"
                v-model:page-size="usagesPageSize"
                :total="usagesTotal"
                layout="prev, pager, next, jumper"
                @current-change="loadCouponUsages"
                @size-change="loadCouponUsages"
              />
            </div>
          </el-card>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 新增/编辑优惠券弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogType === 'add' ? '新增优惠券' : '编辑优惠券'"
      width="700px"
      destroy-on-close
    >
      <el-form
        ref="couponFormRef"
        :model="couponForm"
        :rules="couponRules"
        label-width="120px"
      >
        <el-form-item label="优惠券名称" prop="name">
          <el-input v-model="couponForm.name" placeholder="请输入优惠券名称" />
        </el-form-item>

        <el-form-item label="优惠券类型" prop="type">
          <el-radio-group v-model="couponForm.type" @change="handleTypeChange">
            <el-radio-button label="delivery_fee">配送费券</el-radio-button>
            <el-radio-button label="amount">金额券</el-radio-button>
          </el-radio-group>
        </el-form-item>

        <el-form-item
          v-if="couponForm.type === 'amount'"
          label="优惠金额"
          prop="discount_value"
        >
          <el-input-number
            v-model="couponForm.discount_value"
            :min="0.01"
            :precision="2"
            :step="1"
            controls-position="right"
            style="width: 100%"
          />
          <span class="input-addon">元</span>
        </el-form-item>

        <el-form-item
          v-else
          label="优惠说明"
        >
          <el-text type="info">配送费券将全免配送费</el-text>
        </el-form-item>

        <el-form-item label="使用门槛" prop="min_amount">
          <el-radio-group v-model="useThreshold" @change="handleThresholdChange">
            <el-radio-button :label="true">满额可用</el-radio-button>
            <el-radio-button :label="false">无门槛</el-radio-button>
          </el-radio-group>
          <el-input-number
            v-if="useThreshold"
            v-model="couponForm.min_amount"
            :min="0.01"
            :precision="2"
            :step="10"
            controls-position="right"
            style="width: 200px; margin-left: 10px;"
            placeholder="最低使用金额"
          />
          <span v-if="useThreshold" class="input-addon">元</span>
        </el-form-item>

        <el-form-item label="适用分类">
          <el-checkbox v-model="useCategoryLimit">指定分类</el-checkbox>
          <el-cascader
            v-if="useCategoryLimit"
            v-model="couponForm.category_ids"
            :options="treeCategories"
            :props="{ checkStrictly: true, label: 'name', value: 'id', children: 'children', multiple: true }"
            placeholder="选择适用分类（可多选）"
            clearable
            style="width: 100%; margin-top: 10px;"
            collapse-tags
            collapse-tags-tooltip
          />
          <el-text v-else type="info" style="margin-left: 10px;">全品类可用</el-text>
        </el-form-item>

        <el-form-item label="发放数量" prop="total_count">
          <el-radio-group v-model="useCountLimit" @change="handleCountLimitChange">
            <el-radio-button :label="true">限制数量</el-radio-button>
            <el-radio-button :label="false">不限制</el-radio-button>
          </el-radio-group>
          <el-input-number
            v-if="useCountLimit"
            v-model="couponForm.total_count"
            :min="1"
            :step="10"
            controls-position="right"
            style="width: 200px; margin-left: 10px;"
            placeholder="发放总数"
          />
        </el-form-item>

        <el-form-item label="有效期" prop="valid_from">
          <el-date-picker
            v-model="dateRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DD HH:mm:ss"
            style="width: 100%"
            @change="handleDateRangeChange"
          />
        </el-form-item>

        <el-form-item label="状态" prop="status">
          <el-switch
            v-model="couponForm.status"
            :active-value="1"
            :inactive-value="0"
          />
        </el-form-item>

        <el-form-item label="说明">
          <el-input
            v-model="couponForm.description"
            type="textarea"
            :rows="3"
            placeholder="可选，优惠券使用说明"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 发放优惠券弹窗 -->
    <el-dialog
      v-model="issueDialogVisible"
      title="发放优惠券"
      width="600px"
      destroy-on-close
    >
      <el-form label-width="100px">
        <el-form-item label="优惠券信息">
          <div v-if="currentCoupon">
            <div><strong>名称：</strong>{{ currentCoupon.name }}</div>
            <div style="margin-top: 8px;">
              <strong>类型：</strong>
              <el-tag :type="currentCoupon.type === 'delivery_fee' ? 'success' : 'warning'" style="margin-left: 8px;">
                {{ currentCoupon.type === 'delivery_fee' ? '配送费券' : '金额券' }}
              </el-tag>
            </div>
            <div style="margin-top: 8px;" v-if="currentCoupon.type === 'amount'">
              <strong>优惠金额：</strong>¥{{ (currentCoupon.discount_value || 0).toFixed(2) }}
            </div>
          </div>
        </el-form-item>

        <el-form-item label="选择用户" required>
          <el-select
            v-model="selectedUserId"
            filterable
            remote
            reserve-keyword
            placeholder="请输入用户ID、姓名、手机号、地址搜索"
            :remote-method="searchUsers"
            :loading="userSearchLoading"
            style="width: 100%"
            clearable
          >
            <el-option
              v-for="user in userOptions"
              :key="user.id"
              :label="getUserLabel(user)"
              :value="user.id"
            >
              <div>
                <div style="font-weight: 500; margin-bottom: 4px;">
                  {{ getUserLabel(user) }}
                </div>
                <div style="color: #909399; font-size: 12px; line-height: 1.5;">
                  <div v-if="user.default_address">
                    <span>地址：</span>{{ user.default_address.address || '-' }}
                  </div>
                  <div v-if="user.phone">
                    <span>电话：</span>{{ user.phone }}
                  </div>
                </div>
              </div>
            </el-option>
          </el-select>
        </el-form-item>

        <el-form-item label="发放数量" required>
          <el-input-number
            v-model="issueQuantity"
            :min="1"
            :max="100"
            :step="1"
            controls-position="right"
            style="width: 100%"
          />
          <div style="margin-top: 8px; color: #909399; font-size: 12px;">默认1张，最多可发放100张</div>
        </el-form-item>

        <el-form-item label="发放原因" required>
          <el-radio-group v-model="issueReason">
            <el-radio-button label="潜在客户" />
            <el-radio-button label="优质客户" />
            <el-radio-button label="老客户关怀" />
            <el-radio-button label="活动赠送" />
            <el-radio-button label="售后补偿" />
          </el-radio-group>
        </el-form-item>

        <el-form-item label="有效期设置">
          <el-radio-group v-model="issueExpireType" @change="handleIssueExpireTypeChange">
            <el-radio label="none">不限制</el-radio>
            <el-radio label="days">N天后过期</el-radio>
            <el-radio label="date">指定日期</el-radio>
          </el-radio-group>
          <div v-if="issueExpireType === 'days'" style="margin-top: 10px;">
            <el-input-number
              v-model="issueExpiresIn"
              :min="1"
              :max="365"
              :step="1"
              controls-position="right"
              placeholder="请输入天数"
              style="width: 100%"
            />
            <div style="margin-top: 8px; color: #909399; font-size: 12px;">从发放时开始计算，N天后过期</div>
          </div>
          <div v-if="issueExpireType === 'date'" style="margin-top: 10px;">
            <el-date-picker
              v-model="issueExpiresAt"
              type="datetime"
              placeholder="选择过期日期"
              format="YYYY-MM-DD HH:mm:ss"
              value-format="YYYY-MM-DD HH:mm:ss"
              style="width: 100%"
            />
          </div>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="issueDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="issuing" @click="handleIssueSubmit">确定发放</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search } from '@element-plus/icons-vue'
import { getCoupons, createCoupon, updateCoupon, deleteCoupon, issueCouponToUser, getCouponIssues, getCouponUsages } from '../api/coupons'
import { getCategoryList } from '../api/category'
import { getMiniUsers } from '../api/miniUsers'
import { formatDate } from '../utils/time-format'

const loading = ref(false)
const saving = ref(false)
const coupons = ref([])
const dialogVisible = ref(false)
const dialogType = ref('add')
const couponFormRef = ref(null)
const treeCategories = ref([])
const useThreshold = ref(false)
const useCategoryLimit = ref(false)
const useCountLimit = ref(false)
const dateRange = ref([])

// 发放优惠券相关
const issueDialogVisible = ref(false)
const currentCoupon = ref(null)
const selectedUserId = ref(null)
const issueQuantity = ref(1)
const issueExpireType = ref('none') // none, days, date
const issueExpiresIn = ref(30) // 天数
const issueExpiresAt = ref(null) // 指定日期
const userOptions = ref([])
const userSearchLoading = ref(false)
const issuing = ref(false)
const issueReason = ref('潜在客户')

// 发放记录相关
const activeTab = ref('coupons')
const couponIssues = ref([])
const issuesLoading = ref(false)
const issuesPageNum = ref(1)
const issuesPageSize = ref(20)
const issuesTotal = ref(0)
const issueKeyword = ref('')

// 使用记录相关
const couponUsages = ref([])
const usagesLoading = ref(false)
const usagesPageNum = ref(1)
const usagesPageSize = ref(20)
const usagesTotal = ref(0)
const usageKeyword = ref('')

const couponForm = reactive({
  id: null,
  name: '',
  type: 'delivery_fee',
  discount_value: 0,
  min_amount: 0,
  category_ids: [],
  total_count: 0,
  status: 1,
  valid_from: '',
  valid_to: '',
  description: ''
})

const couponRules = {
  name: [
    { required: true, message: '请输入优惠券名称', trigger: 'blur' }
  ],
  type: [
    { required: true, message: '请选择优惠券类型', trigger: 'change' }
  ],
  discount_value: [
    { required: true, message: '请输入优惠金额', trigger: 'blur' },
    { type: 'number', min: 0.01, message: '优惠金额必须大于0', trigger: 'blur' }
  ],
  valid_from: [
    { required: true, message: '请选择有效期', trigger: 'change' }
  ]
}

// 加载分类数据
const loadCategories = async () => {
  try {
    const response = await getCategoryList()
    if (response.code === 200 && response.data) {
      const flattenCategories = (categories) => {
        let result = []
        categories.forEach(category => {
          result.push({
            id: category.id,
            name: category.name,
            parent_id: category.parent_id || 0,
            children: category.children ? flattenCategories(category.children) : []
          })
          if (category.children && category.children.length > 0) {
            result = result.concat(flattenCategories(category.children))
          }
        })
        return result
      }
      treeCategories.value = flattenCategories(response.data)
    }
  } catch (error) {
    console.error('加载分类失败:', error)
  }
}

// 加载优惠券列表
const loadCoupons = async () => {
  loading.value = true
  try {
    const response = await getCoupons()
    console.log('获取优惠券响应:', response)
    
    // 检查响应格式
    if (!response) {
      ElMessage.error('获取优惠券列表失败：响应为空')
      coupons.value = []
      return
    }
    
    // 如果响应直接是数组（某些情况下可能直接返回数组）
    if (Array.isArray(response)) {
      coupons.value = response.map(coupon => ({
        ...coupon,
        type: coupon.type || '',
        discount_value: coupon.discount_value || 0,
        min_amount: coupon.min_amount || 0,
        category_ids: coupon.category_ids || [],
        total_count: coupon.total_count || 0,
        used_count: coupon.used_count || 0,
        status: coupon.status !== undefined ? coupon.status : 1,
        // 保持时间字符串格式，避免时区转换问题
        valid_from: coupon.valid_from || null,
        valid_to: coupon.valid_to || null
      }))
      return
    }
    
    // 标准响应格式：{ code, data, message }
    if (response.code === 200) {
      // 如果 data 为 null 或空，设置为空数组
      if (response.data && Array.isArray(response.data)) {
        coupons.value = response.data.map(coupon => ({
          ...coupon,
          type: coupon.type || '',
          discount_value: coupon.discount_value || 0,
          min_amount: coupon.min_amount || 0,
          category_ids: coupon.category_ids || [],
          total_count: coupon.total_count || 0,
          used_count: coupon.used_count || 0,
          status: coupon.status !== undefined ? coupon.status : 1,
          // 保持时间字符串格式，避免时区转换问题
          valid_from: coupon.valid_from || null,
          valid_to: coupon.valid_to || null
        }))
      } else {
        coupons.value = []
      }
    } else {
      console.error('获取优惠券列表失败，响应码:', response.code, '消息:', response.message)
      ElMessage.error(response.message || '未能获取到优惠券列表')
      coupons.value = []
    }
  } catch (error) {
    console.error('加载优惠券列表失败:', error)
    ElMessage.error('加载优惠券列表失败: ' + (error.response?.data?.message || error.message || '未知错误'))
    coupons.value = []
  } finally {
    loading.value = false
  }
}

// 处理类型变化
const handleTypeChange = () => {
  if (couponForm.type === 'delivery_fee') {
    couponForm.discount_value = 0
  }
}

// 处理门槛变化
const handleThresholdChange = (value) => {
  if (!value) {
    couponForm.min_amount = 0
  }
}

// 处理数量限制变化
const handleCountLimitChange = (value) => {
  if (!value) {
    couponForm.total_count = 0
  }
}

// 处理日期范围变化
const handleDateRangeChange = (dates) => {
  if (dates && dates.length === 2) {
    couponForm.valid_from = dates[0]
    couponForm.valid_to = dates[1]
  } else {
    couponForm.valid_from = ''
    couponForm.valid_to = ''
  }
}

// 打开新增弹窗
const handleAddCoupon = () => {
  dialogType.value = 'add'
  resetForm()
  dialogVisible.value = true
}

// 打开编辑弹窗
const handleEditCoupon = (coupon) => {
  dialogType.value = 'edit'
  resetForm()
  
  couponForm.id = coupon.id
  couponForm.name = coupon.name
  couponForm.type = coupon.type
  couponForm.discount_value = coupon.discount_value || 0
  couponForm.min_amount = coupon.min_amount || 0
  couponForm.category_ids = coupon.category_ids || []
  couponForm.total_count = coupon.total_count || 0
  couponForm.status = coupon.status
  couponForm.description = coupon.description || ''
  
  // 设置日期范围
  if (coupon.valid_from && coupon.valid_to) {
    // 将时间字符串转换为日期选择器需要的格式
    // 如果是 ISO 格式（如 2025-11-26T01:00:00Z），需要提取本地时间部分
    const formatTimeString = (timeStr) => {
      if (!timeStr) return ''
      // 如果是 ISO 格式，提取日期和时间部分
      if (timeStr.includes('T')) {
        // 处理 ISO 格式：2025-11-26T01:00:00Z 或 2025-11-26T01:00:00+08:00
        const datePart = timeStr.split('T')[0]
        const timePart = timeStr.split('T')[1].split(/[Z+-]/)[0] // 提取时间部分，去掉时区信息
        return `${datePart} ${timePart}`
      }
      // 如果已经是 YYYY-MM-DD HH:mm:ss 格式，直接返回
      return timeStr
    }
    
    dateRange.value = [
      formatTimeString(coupon.valid_from),
      formatTimeString(coupon.valid_to)
    ]
    couponForm.valid_from = dateRange.value[0]
    couponForm.valid_to = dateRange.value[1]
  }
  
  // 设置开关状态
  useThreshold.value = coupon.min_amount > 0
  useCategoryLimit.value = coupon.category_ids && coupon.category_ids.length > 0
  useCountLimit.value = coupon.total_count > 0
  
  dialogVisible.value = true
}

// 重置表单
const resetForm = () => {
  couponForm.id = null
  couponForm.name = ''
  couponForm.type = 'delivery_fee'
  couponForm.discount_value = 0
  couponForm.min_amount = 0
  couponForm.category_ids = []
  couponForm.total_count = 0
  couponForm.status = 1
  couponForm.valid_from = ''
  couponForm.valid_to = ''
  couponForm.description = ''
  useThreshold.value = false
  useCategoryLimit.value = false
  useCountLimit.value = false
  dateRange.value = []
  couponFormRef.value?.resetFields()
}

// 提交表单
const handleSubmit = async () => {
  try {
    await couponFormRef.value.validate()
    
    // 验证日期范围
    if (!couponForm.valid_from || !couponForm.valid_to) {
      ElMessage.error('请选择有效期')
      return
    }
    
    // 验证金额券的优惠值
    if (couponForm.type === 'amount' && couponForm.discount_value <= 0) {
      ElMessage.error('金额券的优惠金额必须大于0')
      return
    }
    
    // 如果没有选择分类限制，清空分类ID
    if (!useCategoryLimit.value) {
      couponForm.category_ids = []
    }
    
    saving.value = true
    
    const submitData = {
      name: couponForm.name,
      type: couponForm.type,
      discount_value: couponForm.type === 'delivery_fee' ? 0 : couponForm.discount_value,
      min_amount: useThreshold.value ? couponForm.min_amount : 0,
      category_ids: useCategoryLimit.value ? couponForm.category_ids : [],
      total_count: useCountLimit.value ? couponForm.total_count : 0,
      status: couponForm.status,
      valid_from: couponForm.valid_from,
      valid_to: couponForm.valid_to,
      description: couponForm.description
    }
    
    if (dialogType.value === 'add') {
      await createCoupon(submitData)
      ElMessage.success('创建成功')
    } else {
      await updateCoupon(couponForm.id, submitData)
      ElMessage.success('更新成功')
    }
    
    dialogVisible.value = false
    loadCoupons()
  } catch (error) {
    if (error.response && error.response.data) {
      ElMessage.error(error.response.data.message || '操作失败')
    } else {
      ElMessage.error('操作失败，请稍后再试')
    }
  } finally {
    saving.value = false
  }
}

// 删除优惠券
const handleDeleteCoupon = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除这个优惠券吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await deleteCoupon(id)
    ElMessage.success('删除成功')
    loadCoupons()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败，请稍后再试')
    }
  }
}

// 打开发放优惠券弹窗
const handleIssueCoupon = (coupon) => {
  currentCoupon.value = coupon
  selectedUserId.value = null
  issueQuantity.value = 1
  issueExpireType.value = 'none'
  issueExpiresIn.value = 30
  issueExpiresAt.value = null
  userOptions.value = []
  issueReason.value = '潜在客户'
  issueDialogVisible.value = true
}

// 有效期类型改变
const handleIssueExpireTypeChange = () => {
  if (issueExpireType.value === 'none') {
    issueExpiresIn.value = 30
    issueExpiresAt.value = null
  }
}

// 搜索用户
const searchUsers = async (query) => {
  if (!query || query.trim() === '') {
    userOptions.value = []
    return
  }
  
  userSearchLoading.value = true
  try {
    const response = await getMiniUsers({
      pageNum: 1,
      pageSize: 20,
      keyword: query.trim()
    })
    if (response.code === 200 && Array.isArray(response.data)) {
      userOptions.value = response.data
    } else {
      userOptions.value = []
    }
  } catch (error) {
    console.error('搜索用户失败:', error)
    userOptions.value = []
  } finally {
    userSearchLoading.value = false
  }
}

// 获取用户显示标签
const getUserLabel = (user) => {
  const parts = []
  if (user.name) parts.push(user.name)
  if (user.phone) parts.push(user.phone)
  if (user.user_code) parts.push(`用户${user.user_code}`)
  return parts.length > 0 ? parts.join(' - ') : `用户ID: ${user.id}`
}

// 提交发放
const handleIssueSubmit = async () => {
  if (!selectedUserId.value) {
    ElMessage.warning('请选择要发放的用户')
    return
  }
  
  if (!currentCoupon.value) {
    ElMessage.error('优惠券信息错误')
    return
  }
  
  if (issueQuantity.value < 1) {
    ElMessage.warning('发放数量必须大于0')
    return
  }

  if (!issueReason.value) {
    ElMessage.warning('请选择发放原因')
    return
  }
  
  issuing.value = true
  try {
    const issueData = {
      coupon_id: currentCoupon.value.id,
      user_id: selectedUserId.value,
      quantity: issueQuantity.value,
      reason: issueReason.value
    }
    
    // 添加有效期参数
    if (issueExpireType.value === 'days') {
      issueData.expires_in = issueExpiresIn.value
    } else if (issueExpireType.value === 'date' && issueExpiresAt.value) {
      issueData.expires_at = issueExpiresAt.value
    }
    
    await issueCouponToUser(issueData)
    ElMessage.success(`成功发放 ${issueQuantity.value} 张优惠券`)
    issueDialogVisible.value = false
    // 刷新优惠券列表，更新已发放数量
    loadCoupons()
  } catch (error) {
    const errorMsg = error.response?.data?.message || error.message || '发放失败'
    ElMessage.error(errorMsg)
  } finally {
    issuing.value = false
  }
}

// 加载优惠券发放记录
const loadCouponIssues = async () => {
  issuesLoading.value = true
  try {
    const params = {
      pageNum: issuesPageNum.value,
      pageSize: issuesPageSize.value
    }
    if (issueKeyword.value && issueKeyword.value.trim() !== '') {
      params.keyword = issueKeyword.value.trim()
    }

    const response = await getCouponIssues(params)
    console.log(response)
    if (response && response.code === 200 && response.data) {
      const { list, total, pageNum, pageSize } = response.data
      couponIssues.value = Array.isArray(list) ? list : []
      issuesTotal.value = total || 0
      issuesPageNum.value = pageNum || 1
      issuesPageSize.value = pageSize || 20
    } else if (Array.isArray(response)) {
      // 兼容直接返回数组的情况
      couponIssues.value = response
      issuesTotal.value = response.length
    } else {
      couponIssues.value = []
      issuesTotal.value = 0
    }
  } catch (error) {
    console.error('获取优惠券发放记录失败:', error)
    couponIssues.value = []
    issuesTotal.value = 0
  } finally {
    issuesLoading.value = false
  }
}

// 加载优惠券使用记录
const loadCouponUsages = async () => {
  usagesLoading.value = true
  try {
    const params = {
      pageNum: usagesPageNum.value,
      pageSize: usagesPageSize.value
    }
    if (usageKeyword.value && usageKeyword.value.trim() !== '') {
      params.keyword = usageKeyword.value.trim()
    }

    const response = await getCouponUsages(params)
    if (response && response.code === 200 && response.data) {
      const { list, total, pageNum, pageSize } = response.data
      couponUsages.value = Array.isArray(list) ? list : []
      usagesTotal.value = total || 0
      usagesPageNum.value = pageNum || 1
      usagesPageSize.value = pageSize || 20
    } else {
      couponUsages.value = []
      usagesTotal.value = 0
    }
  } catch (error) {
    console.error('获取优惠券使用记录失败:', error)
    ElMessage.error('获取优惠券使用记录失败')
    couponUsages.value = []
    usagesTotal.value = 0
  } finally {
    usagesLoading.value = false
  }
}

// 监听tab切换
watch(activeTab, (newTab) => {
  if (newTab === 'usages') {
    loadCouponUsages()
  }
})

onMounted(() => {
  loadCategories()
  loadCoupons()
  loadCouponIssues()
})
</script>

<style scoped>
.coupons-container {
  padding: 20px;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 20px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  margin-bottom: 20px;
}

.toolbar-right {
  display: flex;
  gap: 10px;
}

.coupons-card {
  margin-top: 20px;
}

.input-addon {
  margin-left: 8px;
  color: #909399;
  font-size: 14px;
}
</style>

