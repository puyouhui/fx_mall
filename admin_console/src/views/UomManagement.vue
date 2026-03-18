<template>
  <div class="uom-container">
    <el-card>
      <h2 class="page-title">计量单位管理</h2>
      <p class="page-desc">管理单位类别及其下的换算单位。选择「大于基准单位」或「小于基准单位」并输入换算数量，系统自动计算比例。</p>

      <div class="toolbar">
        <el-button type="primary" @click="handleAddCategory">
          <el-icon><Plus /></el-icon>
          新增单位类别
        </el-button>
      </div>

      <el-table :data="categories" row-key="id" border>
        <!-- <el-table-column prop="id" label="ID" width="80" /> -->
        <el-table-column prop="name" label="类别名称" min-width="150" />
        <el-table-column prop="base_unit" label="基准单位" min-width="120">
          <template #default="{ row }">
            {{ row.base_unit ? row.base_unit.name : '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="units" label="单位标签" min-width="220">
          <template #default="{ row }">
            <template v-if="(row.units || []).length">
              <el-tag
                v-for="unit in row.units"
                :key="unit.id"
                size="small"
                :type="unit.is_base === 1 ? 'success' : 'info'"
                style="margin-right: 4px; margin-bottom: 4px;"
              >
                {{ unit.name }}<span v-if="unit.is_base === 1">（基准）</span>
              </el-tag>
            </template>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" min-width="160">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" min-width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEditCategory(row)">编辑</el-button>
            <el-button type="danger" link size="small" :disabled="isDefaultCategory(row)" @click="handleDeleteCategory(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 单位类别弹窗（新增/编辑，内含单位管理） -->
    <el-dialog
      v-model="categoryDialogVisible"
      :title="categoryDialogType === 'add' ? '新增单位类别' : '编辑单位类别'"
      width="560px"
      :close-on-click-modal="false"
    >
      <el-form ref="categoryFormRef" :model="categoryForm" :rules="categoryRules" label-width="100px">
        <el-form-item label="类别名称" prop="name">
          <el-input v-model="categoryForm.name" placeholder="如：10瓶/件" />
        </el-form-item>
      </el-form>

      <div class="units-section">
        <div class="units-header">
          <span>单位列表</span>
          <el-button type="primary" size="small" @click="handleAddUnitInDialog">
            <el-icon><Plus /></el-icon>
            添加单位
          </el-button>
        </div>
        <el-table :data="dialogUnits" size="small" border max-height="240">
          <el-table-column prop="name" label="单位名称" width="120" />
          <el-table-column prop="ratio_desc" label="换算关系" min-width="140">
            <template #default="scope">
              <span v-if="scope.row.is_base === 1">
                <el-tag type="success" size="small">基准单位</el-tag>
              </span>
              <span v-else>{{ getRatioDesc(scope.row) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120">
            <template #default="scope">
              <el-button type="primary" link size="small" @click="handleEditUnitInDialog(scope.row)">编辑</el-button>
              <el-button type="danger" link size="small" :disabled="scope.row.is_base === 1" @click="handleDeleteUnitInDialog(scope.row)">
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <template #footer>
        <el-button @click="categoryDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitCategory">确定</el-button>
      </template>
    </el-dialog>

    <!-- 单位表单弹窗（在类别弹框内使用） -->
    <el-dialog
      v-model="unitDialogVisible"
      :title="unitDialogType === 'add' ? '添加单位' : '编辑单位'"
      width="440px"
      append-to-body
    >
      <el-form ref="unitFormRef" :model="unitForm" :rules="unitRules" label-width="120px">
        <el-form-item label="单位名称" prop="name">
          <el-input v-model="unitForm.name" placeholder="如：瓶、件" />
        </el-form-item>
        <el-form-item label="单位类型" prop="size_type">
          <el-radio-group v-model="unitForm.size_type">
            <el-radio label="base">基准单位</el-radio>
            <el-radio label="bigger">大于基准单位</el-radio>
            <el-radio label="smaller">小于基准单位</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-alert v-if="(unitForm.size_type === 'bigger' || unitForm.size_type === 'smaller') && !baseUnitName" type="warning" :closable="false" show-icon style="margin-bottom:12px">
          请先添加基准单位
        </el-alert>
        <el-form-item v-if="unitForm.size_type === 'bigger' && baseUnitName" label="换算数量" prop="convert_num">
          <div class="convert-input">
            <span>1 {{ unitForm.name || '该单位' }} = </span>
            <el-input-number v-model="unitForm.convert_num" :min="1" :max="999999" :precision="0" />
            <span> {{ baseUnitName }}（基准单位）</span>
          </div>
          <div class="form-tip">即 1 个该单位 = 多少基准单位，如 1件=10瓶 则输入 10</div>
        </el-form-item>
        <el-form-item v-if="unitForm.size_type === 'smaller' && baseUnitName" label="换算数量" prop="convert_num">
          <div class="convert-input">
            <el-input-number v-model="unitForm.convert_num" :min="1" :max="999999" :precision="0" />
            <span> {{ unitForm.name || '该单位' }} = 1 {{ baseUnitName }}（基准单位）</span>
          </div>
          <div class="form-tip">即 多少该单位 = 1 个基准单位，如 2半瓶=1瓶 则输入 2</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="unitDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitUnitInDialog">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import {
  getUomCategories,
  createUomCategory,
  updateUomCategory,
  deleteUomCategory,
  createUomUnit,
  updateUomUnit,
  deleteUomUnit
} from '../api/uom'
import { formatDate } from '../utils/time-format'

const categories = ref([])

const categoryDialogVisible = ref(false)
const categoryDialogType = ref('add')
const categoryFormRef = ref(null)
const categoryForm = reactive({
  id: null,
  name: ''
})
const categoryRules = {
  name: [
    { required: true, message: '请输入类别名称', trigger: 'blur' },
    { max: 64, message: '不超过64字符', trigger: 'blur' }
  ]
}

// 弹框内的单位列表（新增时为本地列表，编辑时为接口数据副本）
const dialogUnits = ref([])

const unitDialogVisible = ref(false)
const unitDialogType = ref('add')
const unitFormRef = ref(null)
const unitEditIndex = ref(-1) // 编辑时的索引，-1 表示新增
const unitForm = reactive({
  id: null,
  category_id: null,
  name: '',
  ratio: 1,
  is_base: 0,
  size_type: 'base',
  convert_num: 1
})
const baseUnitName = computed(() => {
  const base = dialogUnits.value.find((u) => u.is_base === 1)
  return base ? base.name : ''
})
const unitRules = {
  name: [
    { required: true, message: '请输入单位名称', trigger: 'blur' },
    { max: 32, message: '不超过32字符', trigger: 'blur' }
  ]
}

const getRatioDesc = (row) => {
  const base = dialogUnits.value.find((u) => u.is_base === 1)
  const baseName = base ? base.name : '基准'
  if (row.ratio >= 1) {
    return `1${row.name}=${row.ratio}${baseName}`
  }
  const n = 1 / row.ratio
  return `${n}${row.name}=1${baseName}`
}

const isDefaultCategory = (row) => {
  return row.name === '件' && (row.units || []).length === 1 && (row.units || [])[0]?.name === '件'
}

const loadData = async () => {
  try {
    const res = await getUomCategories()
    if (res.code === 200 && res.data) {
      categories.value = res.data
    }
  } catch (e) {
    console.error(e)
    ElMessage.error('加载数据失败')
  }
}

const handleAddCategory = () => {
  categoryDialogType.value = 'add'
  Object.assign(categoryForm, { id: null, name: '' })
  dialogUnits.value = []
  categoryDialogVisible.value = true
}

const handleEditCategory = (row) => {
  categoryDialogType.value = 'edit'
  Object.assign(categoryForm, { id: row.id, name: row.name })
  dialogUnits.value = (row.units || []).map((u) => ({ ...u }))
  categoryDialogVisible.value = true
}

const handleSubmitCategory = async () => {
  try {
    await categoryFormRef.value?.validate()
    if (categoryDialogType.value === 'add') {
      const res = await createUomCategory({ name: categoryForm.name })
      if (res.code !== 200) {
        ElMessage.error(res.message || '创建失败')
        return
      }
      const newCategoryId = res.data?.id
      if (!newCategoryId) {
        ElMessage.error('创建类别成功但未返回ID')
        categoryDialogVisible.value = false
        await loadData()
        return
      }
      for (const u of dialogUnits.value) {
        await createUomUnit({
          category_id: newCategoryId,
          name: u.name,
          ratio: u.is_base === 1 ? 1 : u.ratio,
          is_base: u.is_base
        })
      }
      ElMessage.success('创建成功')
      categoryDialogVisible.value = false
      await loadData()
    } else {
      const res = await updateUomCategory(categoryForm.id, { name: categoryForm.name })
      if (res.code === 200) {
        ElMessage.success('更新成功')
        categoryDialogVisible.value = false
        await loadData()
      } else {
        ElMessage.error(res.message || '更新失败')
      }
    }
  } catch (e) {
    if (e !== false) console.error(e)
  }
}

const handleDeleteCategory = async (row) => {
  if (isDefaultCategory(row)) {
    ElMessage.warning('默认「件」单位类别用于兼容老数据，不可删除')
    return
  }
  try {
    await ElMessageBox.confirm('删除单位类别会同时删除其下所有单位，确定删除？', '确认删除', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const res = await deleteUomCategory(row.id)
    if (res.code === 200) {
      ElMessage.success('删除成功')
      await loadData()
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

const computeRatioFromForm = () => {
  if (unitForm.size_type === 'base') return 1
  if (unitForm.size_type === 'bigger') return unitForm.convert_num >= 1 ? unitForm.convert_num : 1
  if (unitForm.size_type === 'smaller') return unitForm.convert_num >= 1 ? 1 / unitForm.convert_num : 1
  return 1
}

const initUnitFormFromUnit = (unit) => {
  const isBase = unit.is_base === 1
  let sizeType = 'base'
  let convertNum = 1
  if (!isBase) {
    if (unit.ratio >= 1) {
      sizeType = 'bigger'
      convertNum = Math.round(unit.ratio)
    } else {
      sizeType = 'smaller'
      convertNum = Math.round(1 / unit.ratio)
    }
  }
  return { size_type: sizeType, convert_num: convertNum }
}

const handleAddUnitInDialog = () => {
  unitEditIndex.value = -1
  unitDialogType.value = 'add'
  Object.assign(unitForm, {
    id: null,
    category_id: categoryForm.id,
    name: '',
    ratio: 1,
    is_base: 0,
    size_type: 'base',
    convert_num: 1
  })
  unitDialogVisible.value = true
}

const handleEditUnitInDialog = (unit) => {
  const idx = dialogUnits.value.indexOf(unit)
  unitEditIndex.value = idx
  unitDialogType.value = 'edit'
  const { size_type, convert_num } = initUnitFormFromUnit(unit)
  Object.assign(unitForm, {
    id: unit.id || null,
    category_id: categoryForm.id,
    name: unit.name,
    ratio: unit.ratio,
    is_base: unit.is_base,
    size_type,
    convert_num
  })
  unitDialogVisible.value = true
}

const handleSubmitUnitInDialog = async () => {
  try {
    if (unitForm.size_type !== 'base') {
      if (!baseUnitName.value) {
        ElMessage.warning('请先添加基准单位')
        return
      }
      const n = unitForm.convert_num
      if (!n || n < 1) {
        ElMessage.warning('换算数量必须大于等于 1')
        return
      }
    }
    await unitFormRef.value?.validate()
    const isBase = unitForm.size_type === 'base'
    const ratio = isBase ? 1 : computeRatioFromForm()
    if (isBase) {
      dialogUnits.value.forEach((u) => { u.is_base = 0 })
    }
    const item = {
      name: unitForm.name,
      ratio,
      is_base: isBase ? 1 : 0
    }
    if (unitDialogType.value === 'add') {
      if (categoryDialogType.value === 'add') {
        dialogUnits.value.push({ ...item })
      } else {
        const res = await createUomUnit({
          category_id: categoryForm.id,
          ...item
        })
        if (res.code === 200 && res.data) {
          dialogUnits.value.push({ ...res.data })
        } else {
          ElMessage.error(res.message || '添加失败')
          return
        }
      }
      ElMessage.success('添加成功')
    } else {
      if (categoryDialogType.value === 'add') {
        const u = dialogUnits.value[unitEditIndex.value]
        Object.assign(u, { ...item })
      } else {
        const res = await updateUomUnit(unitForm.id, { ...item })
        if (res.code !== 200) {
          ElMessage.error(res.message || '更新失败')
          return
        }
        const u = dialogUnits.value[unitEditIndex.value]
        Object.assign(u, { ...item })
      }
      ElMessage.success('更新成功')
    }
    unitDialogVisible.value = false
  } catch (e) {
    if (e !== false) console.error(e)
  }
}

const handleDeleteUnitInDialog = async (unit) => {
  if (unit.is_base === 1) {
    ElMessage.warning('不能删除基准单位')
    return
  }
  try {
    await ElMessageBox.confirm(`确定删除单位「${unit.name}」？`, '确认删除', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    if (categoryDialogType.value === 'add') {
      dialogUnits.value = dialogUnits.value.filter((u) => u !== unit)
      ElMessage.success('已移除')
    } else {
      const res = await deleteUomUnit(unit.id)
      if (res.code === 200) {
        dialogUnits.value = dialogUnits.value.filter((u) => u.id !== unit.id)
        ElMessage.success('删除成功')
      } else {
        ElMessage.error(res.message || '删除失败')
      }
    }
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.uom-container { padding: 0 0 20px 0; }
.page-title { font-size: 24px; margin-bottom: 8px; color: #333; }
.page-desc { color: #666; font-size: 13px; margin-bottom: 20px; }
.toolbar { margin-bottom: 16px; }
.units-section { margin-top: 16px; padding-top: 16px; border-top: 1px solid #ebeef5; }
.units-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; font-weight: 500; }
.convert-input { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.form-tip { font-size: 12px; color: #999; margin-top: 6px; }
</style>
