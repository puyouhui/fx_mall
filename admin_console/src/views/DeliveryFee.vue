<template>
  <div class="delivery-fee-container">
    <h1 class="page-title">配送费设置</h1>

    <el-row :gutter="20">
      <el-col :xs="24" :sm="24" :md="10">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>基础规则</span>
              <el-button :icon="Refresh" link @click="loadSettings" :loading="settingsLoading">
                刷新
              </el-button>
            </div>
          </template>
          <el-form
            ref="settingsFormRef"
            :model="settingsForm"
            :rules="settingsRules"
            label-width="120px"
            class="settings-form"
          >
            <el-form-item label="基础配送费" prop="base_fee">
              <el-input-number
                v-model="settingsForm.base_fee"
                :min="0"
                :precision="2"
                :step="1"
                controls-position="right"
                style="width: 240px"
              />
              <span class="input-addon">元</span>
            </el-form-item>
            <el-form-item label="免配送费阈值" prop="free_shipping_threshold">
              <el-input-number
                v-model="settingsForm.free_shipping_threshold"
                :min="0"
                :precision="2"
                :step="10"
                controls-position="right"
                style="width: 240px"
              />
              <span class="input-addon">元</span>
            </el-form-item>
            <el-form-item label="备注说明">
              <el-input
                v-model="settingsForm.description"
                type="textarea"
                placeholder="可选，例：满139元免配送费，仅限不含特殊商品的订单"
                :rows="3"
                maxlength="200"
                show-word-limit
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="savingSettings" @click="handleSaveSettings">
                保存设置
              </el-button>
              <el-button :disabled="savingSettings" @click="resetSettings">
                重置
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="24" :md="14">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>不参与免配送费规则的对象</span>
              <div class="card-actions">
                <el-button type="primary" :icon="Plus" @click="openCreateDialog">
                  新增排除项
                </el-button>
                <el-button circle :icon="Refresh" @click="loadExclusions" :loading="exclusionLoading" />
              </div>
            </div>
          </template>

          <el-table :data="exclusions" v-loading="exclusionLoading" border>
            <el-table-column prop="item_type" label="类型" width="120" align="center">
              <template #default="{ row }">
                <el-tag :type="row.item_type === 'product' ? 'warning' : 'info'">
                  {{ row.item_type === 'product' ? '商品' : '分类' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="对象" min-width="200">
              <template #default="{ row }">
                <div class="target-name">
                  <span v-if="row.item_type === 'category' && row.parent_category_name">
                    {{ row.parent_category_name }} /
                  </span>
                  <span>{{ row.target_name || '-' }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="免运数量" width="160" align="center">
              <template #default="{ row }">
                {{ row.min_quantity_for_free ?? '—' }}
              </template>
            </el-table-column>
            <el-table-column prop="remark" label="说明" min-width="180">
              <template #default="{ row }">
                {{ row.remark || '—' }}
              </template>
            </el-table-column>
            <el-table-column label="更新时间" width="160">
              <template #default="{ row }">
                {{ formatDate(row.updated_at) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="180" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" size="small" @click="openEditDialog(row)">
                  编辑
                </el-button>
                <el-button link type="danger" size="small" @click="handleDelete(row)">
                  删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-empty
            v-if="!exclusionLoading && exclusions.length === 0"
            description="暂无排除项，可点击右上角按钮新增"
            :image-size="120"
          />
        </el-card>
      </el-col>
    </el-row>

    <el-dialog
      v-model="exclusionDialogVisible"
      :title="dialogMode === 'create' ? '新增排除项' : '编辑排除项'"
      width="520px"
      destroy-on-close
    >
      <el-form
        ref="exclusionFormRef"
        :model="exclusionForm"
        :rules="exclusionRules"
        label-width="120px"
      >
        <el-form-item label="排除类型" prop="item_type">
          <el-radio-group v-model="exclusionForm.item_type" :disabled="dialogMode === 'edit'">
            <el-radio-button label="category">分类</el-radio-button>
            <el-radio-button label="product">商品</el-radio-button>
          </el-radio-group>
        </el-form-item>

        <el-form-item
          v-if="exclusionForm.item_type === 'category'"
          label="选择分类"
          prop="category_id"
        >
          <el-cascader
            v-model="exclusionForm.category_id"
            :options="categoryOptions"
            :props="cascaderProps"
            placeholder="可选择任意一级或二级分类"
            clearable
            style="width: 100%;"
          />
        </el-form-item>

        <el-form-item
          v-else
          label="选择商品"
          prop="product_id"
        >
          <el-select
            v-model="exclusionForm.product_id"
            filterable
            remote
            reserve-keyword
            placeholder="搜索商品名称关键字"
            :remote-method="fetchProductOptions"
            :loading="productLoading"
            style="width: 100%;"
          >
            <el-option
              v-for="item in productOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>

        <el-form-item
          v-if="['product', 'category'].includes(exclusionForm.item_type)"
          label="免配送数量"
          prop="min_quantity_for_free"
        >
          <el-input-number
            v-model="exclusionForm.min_quantity_for_free"
            :min="1"
            :step="1"
            :controls="true"
            style="width: 200px;"
            placeholder="可选，不填则永不免运"
          />
          <span class="input-addon">件</span>
        </el-form-item>

        <el-form-item label="备注说明">
          <el-input
            v-model="exclusionForm.remark"
            type="textarea"
            :rows="3"
            maxlength="150"
            show-word-limit
            placeholder="例如：易碎品，运输成本高"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="exclusionDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="savingExclusion" @click="handleSubmitExclusion">
          确认
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh } from '@element-plus/icons-vue'
import {
  getDeliveryFeeSettings,
  updateDeliveryFeeSettings,
  getDeliveryFeeExclusions,
  createDeliveryFeeExclusion,
  updateDeliveryFeeExclusion,
  deleteDeliveryFeeExclusion
} from '../api/deliveryFee'
import { getCategoryList } from '../api/category'
import { getProductList } from '../api/product'
import { formatDate } from '../utils/time-format'

const settingsFormRef = ref(null)
const settingsForm = reactive({
  base_fee: 0,
  free_shipping_threshold: 0,
  description: ''
})
const originSettings = reactive({
  base_fee: 0,
  free_shipping_threshold: 0,
  description: ''
})
const settingsLoading = ref(false)
const savingSettings = ref(false)

const exclusions = ref([])
const exclusionLoading = ref(false)

const exclusionDialogVisible = ref(false)
const exclusionFormRef = ref(null)
const dialogMode = ref('create')
const savingExclusion = ref(false)

const exclusionForm = reactive({
  id: null,
  item_type: 'category',
  category_id: null,
  product_id: null,
  min_quantity_for_free: null,
  remark: ''
})

const categoryOptions = ref([])
const cascaderProps = {
  checkStrictly: true,
  value: 'id',
  label: 'name',
  children: 'children',
  emitPath: false
}

const productOptions = ref([])
const fallbackProductOption = ref(null)
const productLoading = ref(false)

const settingsRules = {
  base_fee: [
    { required: true, message: '请输入基础配送费', trigger: 'blur' }
  ],
  free_shipping_threshold: [
    { required: true, message: '请输入免配送费阈值', trigger: 'blur' }
  ]
}

const exclusionRules = {
  item_type: [{ required: true, message: '请选择类型', trigger: 'change' }],
  category_id: [
    {
      required: true,
      message: '请选择需要排除的分类',
      trigger: 'change',
      validator: (_, value, cb) => {
        if (exclusionForm.item_type !== 'category') {
          cb()
          return
        }
        if (!value) {
          cb(new Error('请选择分类'))
        } else {
          cb()
        }
      }
    }
  ],
  product_id: [
    {
      required: true,
      message: '请选择需要排除的商品',
      trigger: 'change',
      validator: (_, value, cb) => {
        if (exclusionForm.item_type !== 'product') {
          cb()
          return
        }
        if (!value) {
          cb(new Error('请选择商品'))
        } else {
          cb()
        }
      }
    }
  ],
  min_quantity_for_free: [
    {
      trigger: 'change',
      validator: (_, value, cb) => {
        if (value !== null && value !== undefined && value <= 0) {
          cb(new Error('数量必须大于0'))
        } else {
          cb()
        }
      }
    }
  ]
}

const resetSettings = () => {
  Object.assign(settingsForm, originSettings)
}

const loadSettings = async () => {
  settingsLoading.value = true
  try {
    const response = await getDeliveryFeeSettings()
    if (response && response.code === 200) {
      const data = response.data || response
      originSettings.base_fee = Number(data.base_fee || 0)
      originSettings.free_shipping_threshold = Number(data.free_shipping_threshold || 0)
      originSettings.description = data.description || ''
      resetSettings()
    }
  } catch (error) {
    console.error('加载配送费设置失败：', error)
    ElMessage.error('加载配送费设置失败')
  } finally {
    settingsLoading.value = false
  }
}

const handleSaveSettings = async () => {
  try {
    await settingsFormRef.value.validate()
    savingSettings.value = true
    const payload = {
      base_fee: settingsForm.base_fee,
      free_shipping_threshold: settingsForm.free_shipping_threshold,
      description: settingsForm.description
    }
    const response = await updateDeliveryFeeSettings(payload)
    if (response && response.code === 200) {
      ElMessage.success('保存成功')
      Object.assign(originSettings, payload)
    } else {
      ElMessage.error(response?.message || '保存失败')
    }
  } catch (error) {
    if (error?.response?.data?.message) {
      ElMessage.error(error.response.data.message)
    } else if (error?.message && error?.message !== 'cancel') {
      ElMessage.error(error.message)
    }
  } finally {
    savingSettings.value = false
  }
}

const loadExclusions = async () => {
  exclusionLoading.value = true
  try {
    const response = await getDeliveryFeeExclusions()
    if (response && response.code === 200) {
      exclusions.value = Array.isArray(response.data) ? response.data : []
    } else {
      exclusions.value = []
    }
  } catch (error) {
    console.error('加载排除项失败：', error)
    ElMessage.error('加载排除项失败')
  } finally {
    exclusionLoading.value = false
  }
}

const normalizeCategoryTree = (data) => {
  if (!Array.isArray(data)) return []
  const hasChildren = data.some(item => Array.isArray(item.children) && item.children.length > 0)
  if (hasChildren) {
    return data.map(item => ({
      ...item,
      children: normalizeCategoryTree(item.children || [])
    }))
  }

  const map = {}
  data.forEach(item => {
    map[item.id] = { ...item, children: [] }
  })
  const tree = []
  data.forEach(item => {
    if (item.parent_id && item.parent_id !== 0 && map[item.parent_id]) {
      map[item.parent_id].children.push(map[item.id])
    } else {
      tree.push(map[item.id])
    }
  })
  return tree
}

const loadCategories = async () => {
  try {
    const response = await getCategoryList()
    const raw = response?.data || response || []
    categoryOptions.value = normalizeCategoryTree(raw)
  } catch (error) {
    console.error('加载分类失败：', error)
  }
}

const normalizeProductList = (res) => {
  if (!res) return []
  if (Array.isArray(res)) return res
  if (Array.isArray(res.data)) return res.data
  if (Array.isArray(res.list)) return res.list
  if (res.data && Array.isArray(res.data.list)) return res.data.list
  return []
}

const fetchProductOptions = async (keyword = '') => {
  productLoading.value = true
  try {
    const response = await getProductList({
      keyword,
      pageNum: 1,
      pageSize: 20
    })
    const list = normalizeProductList(response).map(item => ({
      label: item.name,
      value: Number(item.id)
    }))
    const exists = fallbackProductOption.value
      ? list.some(item => item.value === fallbackProductOption.value.value)
      : true
    productOptions.value = exists || !fallbackProductOption.value
      ? list
      : [fallbackProductOption.value, ...list]
  } catch (error) {
    console.error('搜索商品失败：', error)
  } finally {
    productLoading.value = false
  }
}

const resetExclusionForm = () => {
  Object.assign(exclusionForm, {
    id: null,
    item_type: 'category',
    category_id: null,
    product_id: null,
    min_quantity_for_free: null,
    remark: ''
  })
  fallbackProductOption.value = null
  productOptions.value = []
}

const openCreateDialog = () => {
  dialogMode.value = 'create'
  resetExclusionForm()
  exclusionDialogVisible.value = true
  fetchProductOptions()
}

const openEditDialog = (row) => {
  dialogMode.value = 'edit'
  resetExclusionForm()
  exclusionForm.id = row.id
  exclusionForm.item_type = row.item_type
  exclusionForm.remark = row.remark || ''
  if (row.item_type === 'category') {
    exclusionForm.category_id = row.target_id
  } else {
    exclusionForm.product_id = row.target_id
    exclusionForm.min_quantity_for_free = row.min_quantity_for_free || null
    const label = row.target_name || `商品#${row.target_id}`
    fallbackProductOption.value = { label, value: row.target_id }
    productOptions.value = [{ ...fallbackProductOption.value }]
    fetchProductOptions(label)
  }
  exclusionDialogVisible.value = true
}

watch(
  () => exclusionForm.item_type,
  (type) => {
    if (type === 'category') {
      exclusionForm.product_id = null
      exclusionForm.min_quantity_for_free = null
      fallbackProductOption.value = null
    } else {
      exclusionForm.category_id = null
      fetchProductOptions()
    }
  }
)

const handleSubmitExclusion = async () => {
  try {
    await exclusionFormRef.value.validate()
    savingExclusion.value = true

    const payload = {
      item_type: exclusionForm.item_type,
      target_id: exclusionForm.item_type === 'category' ? exclusionForm.category_id : exclusionForm.product_id,
      min_quantity_for_free: exclusionForm.min_quantity_for_free,
      remark: exclusionForm.remark
    }

    let response
    if (dialogMode.value === 'create') {
      response = await createDeliveryFeeExclusion(payload)
    } else {
      response = await updateDeliveryFeeExclusion(exclusionForm.id, payload)
    }

    if (response && response.code === 200) {
      ElMessage.success('保存成功')
      exclusionDialogVisible.value = false
      loadExclusions()
    } else {
      ElMessage.error(response?.message || '保存失败')
    }
  } catch (error) {
    if (error?.response?.data?.message) {
      ElMessage.error(error.response.data.message)
    } else if (error?.message && error?.message !== 'cancel') {
      ElMessage.error(error.message)
    }
  } finally {
    savingExclusion.value = false
  }
}

const handleDelete = (row) => {
  ElMessageBox.confirm(
    `确定要移除【${row.target_name || (row.item_type === 'product' ? '商品' : '分类')}】的配送费排除配置吗？`,
    '提示',
    {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    }
  )
    .then(async () => {
      try {
        const response = await deleteDeliveryFeeExclusion(row.id)
        if (response && response.code === 200) {
          ElMessage.success('删除成功')
          loadExclusions()
        } else {
          ElMessage.error(response?.message || '删除失败')
        }
      } catch (error) {
        console.error('删除排除项失败：', error)
        ElMessage.error('删除失败，请稍后再试')
      }
    })
    .catch(() => {})
}

onMounted(async () => {
  await Promise.all([loadSettings(), loadExclusions(), loadCategories()])
  fetchProductOptions()
})
</script>

<style scoped>
.delivery-fee-container {
  padding: 20px;
}

.page-title {
  font-size: 22px;
  font-weight: 600;
  margin-bottom: 20px;
  color: #303133;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
  color: #303133;
}

.card-actions {
  display: flex;
  gap: 10px;
}

.settings-form {
  padding-top: 10px;
}

.input-addon {
  margin-left: 8px;
  color: #909399;
  font-size: 13px;
}

.target-name {
  display: flex;
  gap: 6px;
  color: #303133;
}

.el-table {
  margin-bottom: 10px;
}
</style>

