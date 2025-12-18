<template>
  <div class="hot-search-keywords-container">
    <el-card>
      <h2 class="page-title">热门搜索关键词管理</h2>

      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-right">
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            添加关键词
          </el-button>
        </div>
      </div>

      <!-- 关键词列表 -->
      <el-table :data="keywords" stripe v-loading="loading">
        <el-table-column type="index" label="序号" width="60" align="center" />
        <el-table-column prop="keyword" label="关键词" min-width="150" />
        <el-table-column prop="sort" label="排序" width="100" align="center" />
        <el-table-column label="状态" width="100" align="center">
          <template #default="scope">
            <el-switch
              v-model="scope.row.status"
              :active-value="1"
              :inactive-value="0"
              @change="handleStatusChange(scope.row)"
            />
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="180" align="center">
          <template #default="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" fixed="right" width="150" align="center">
          <template #default="scope">
            <el-button type="primary" size="small" @click="handleEdit(scope.row)">
              编辑
            </el-button>
            <el-button type="danger" size="small" @click="handleDelete(scope.row.id)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 空状态 -->
      <div v-if="!loading && keywords.length === 0" class="empty-state">
        <el-empty description="暂无热门搜索关键词" />
      </div>
    </el-card>

    <!-- 添加/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑关键词' : '添加关键词'" width="400px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="关键词" required>
          <el-input v-model="form.keyword" placeholder="请输入关键词" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" :max="999" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitLoading">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getAllHotSearchKeywords,
  createHotSearchKeyword,
  updateHotSearchKeyword,
  deleteHotSearchKeyword
} from '../api/hotSearchKeywords'

const keywords = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitLoading = ref(false)
const editId = ref(null)

const form = ref({
  keyword: '',
  sort: 0,
  status: 1
})

// 格式化日期
const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN')
}

// 加载关键词列表
const loadKeywords = async () => {
  loading.value = true
  try {
    const res = await getAllHotSearchKeywords()
    if (res.code === 200) {
      keywords.value = res.data || []
    }
  } catch (error) {
    console.error('加载关键词失败:', error)
    ElMessage.error('加载关键词失败')
  } finally {
    loading.value = false
  }
}

// 添加关键词
const handleAdd = () => {
  isEdit.value = false
  editId.value = null
  form.value = { keyword: '', sort: 0, status: 1 }
  dialogVisible.value = true
}

// 编辑关键词
const handleEdit = (row) => {
  isEdit.value = true
  editId.value = row.id
  form.value = {
    keyword: row.keyword,
    sort: row.sort,
    status: row.status
  }
  dialogVisible.value = true
}

// 提交表单
const handleSubmit = async () => {
  if (!form.value.keyword.trim()) {
    ElMessage.warning('请输入关键词')
    return
  }

  submitLoading.value = true
  try {
    if (isEdit.value) {
      await updateHotSearchKeyword(editId.value, form.value)
      ElMessage.success('更新成功')
    } else {
      await createHotSearchKeyword(form.value)
      ElMessage.success('添加成功')
    }
    dialogVisible.value = false
    loadKeywords()
  } catch (error) {
    console.error('操作失败:', error)
    ElMessage.error('操作失败')
  } finally {
    submitLoading.value = false
  }
}

// 状态变更
const handleStatusChange = async (row) => {
  try {
    await updateHotSearchKeyword(row.id, {
      keyword: row.keyword,
      sort: row.sort,
      status: row.status
    })
    ElMessage.success('状态更新成功')
  } catch (error) {
    console.error('状态更新失败:', error)
    ElMessage.error('状态更新失败')
    loadKeywords()
  }
}

// 删除关键词
const handleDelete = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除该关键词吗？', '提示', {
      type: 'warning'
    })
    await deleteHotSearchKeyword(id)
    ElMessage.success('删除成功')
    loadKeywords()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

onMounted(() => {
  loadKeywords()
})
</script>

<style scoped>
.hot-search-keywords-container {
  padding: 20px;
}

.page-title {
  margin: 0 0 20px 0;
  font-size: 18px;
  font-weight: 600;
}

.toolbar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 20px;
}

.empty-state {
  padding: 40px 0;
}
</style>


