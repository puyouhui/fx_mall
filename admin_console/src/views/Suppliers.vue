<template>
  <div class="suppliers-container">
    <el-card>
      <h2 class="page-title">供应商管理</h2>

      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="primary" @click="handleAddSupplier">
            <el-icon>
              <Plus />
            </el-icon>
            新增供应商
          </el-button>
        </div>
      </div>

      <!-- 供应商列表 -->
      <el-card class="suppliers-card">
        <el-table :data="suppliersData" stripe>
          <el-table-column prop="id" label="ID" align="center" width="80" />
          <el-table-column prop="name" label="供应商名称" align="center" />
          <el-table-column prop="contact" label="联系人" align="center" />
          <el-table-column prop="phone" label="联系电话" align="center" />
          <el-table-column prop="email" label="邮箱" align="center" />
          <el-table-column prop="address" label="地址" align="center" show-overflow-tooltip />
          <el-table-column prop="username" label="登录账号" align="center" />
          <el-table-column prop="status" label="状态" align="center" width="100">
            <template #default="scope">
              <el-tag :type="scope.row.status === 1 ? 'success' : 'danger'">
                {{ scope.row.status === 1 ? '启用' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" align="center" width="180">
            <template #default="scope">
              {{ formatDate(scope.row.created_at) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" align="center" fixed="right" width="180">
            <template #default="scope">
              <el-button type="primary" size="small" @click="handleEditSupplier(scope.row)">
                编辑
              </el-button>
              <el-button 
                type="danger" 
                size="small" 
                :disabled="scope.row.username === 'self_operated'"
                @click="handleDeleteSupplier(scope.row.id)">
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </el-card>

    <!-- 新增/编辑供应商弹窗 -->
    <el-dialog v-model="dialogVisible" :title="dialogType === 'add' ? '新增供应商' : '编辑供应商'" width="600px">
      <el-form ref="supplierFormRef" :model="supplierForm" :rules="supplierRules" label-width="100px">
        <el-form-item label="供应商名称" prop="name">
          <el-input v-model="supplierForm.name" placeholder="请输入供应商名称" />
        </el-form-item>
        <el-form-item label="联系人" prop="contact">
          <el-input v-model="supplierForm.contact" placeholder="请输入联系人" />
        </el-form-item>
        <el-form-item label="联系电话" prop="phone">
          <el-input v-model="supplierForm.phone" placeholder="请输入联系电话" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="supplierForm.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="地址" prop="address">
          <el-input v-model="supplierForm.address" type="textarea" :rows="2" placeholder="请输入地址" />
        </el-form-item>
        <el-form-item label="登录账号" prop="username">
          <el-input v-model="supplierForm.username" placeholder="请输入登录账号" />
        </el-form-item>
        <el-form-item v-if="dialogType === 'add'" label="登录密码" prop="password">
          <el-input v-model="supplierForm.password" type="password" show-password placeholder="请输入登录密码（至少6位）" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-switch v-model="supplierForm.status" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { getAllSuppliers, createSupplier, updateSupplier, deleteSupplier } from '../api/suppliers'
import { formatDate } from '../utils/time-format'

const suppliersData = ref([])

// 弹窗相关
const dialogVisible = ref(false)
const dialogType = ref('add')
const supplierFormRef = ref(null)
const supplierForm = reactive({
  id: '',
  name: '',
  contact: '',
  phone: '',
  email: '',
  address: '',
  username: '',
  password: '',
  status: true
})

// 表单验证规则
const supplierRules = {
  name: [
    { required: true, message: '请输入供应商名称', trigger: 'blur' }
  ],
  username: [
    { required: true, message: '请输入登录账号', trigger: 'blur' },
    { min: 3, max: 20, message: '账号长度在 3 到 20 个字符', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入登录密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度在 6 到 20 个字符', trigger: 'blur' }
  ]
}

// 初始化数据
const initData = async () => {
  try {
    const response = await getAllSuppliers()
    if (response.code === 200 && response.data) {
      suppliersData.value = response.data
    }
  } catch (error) {
    console.error('加载数据失败:', error)
    ElMessage.error('加载数据失败')
  }
}

// 打开新增供应商弹窗
const handleAddSupplier = () => {
  dialogType.value = 'add'
  if (supplierFormRef.value) {
    supplierFormRef.value.resetFields()
  }
  Object.assign(supplierForm, {
    id: '',
    name: '',
    contact: '',
    phone: '',
    email: '',
    address: '',
    username: '',
    password: '',
    status: true
  })
  dialogVisible.value = true
}

// 打开编辑供应商弹窗
const handleEditSupplier = (row) => {
  dialogType.value = 'edit'
  Object.assign(supplierForm, {
    id: row.id,
    name: row.name,
    contact: row.contact || '',
    phone: row.phone || '',
    email: row.email || '',
    address: row.address || '',
    username: row.username,
    password: '', // 编辑时不显示密码
    status: row.status === 1 // 转换为布尔值供el-switch使用
  })
  dialogVisible.value = true
}

// 删除供应商
const handleDeleteSupplier = async (id) => {
  try {
    // 检查是否是自营供应商
    const supplier = suppliersData.value.find(s => s.id === id)
    if (supplier && supplier.username === 'self_operated') {
      ElMessage.warning('不能删除系统默认的"自营"供应商')
      return
    }

    await ElMessageBox.confirm('确定要删除这个供应商吗？', '确认删除', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    const response = await deleteSupplier(id)
    if (response.code === 200) {
      await initData()
      ElMessage.success('删除成功')
    } else {
      ElMessage.error(response.message || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

// 提交表单
const handleSubmit = async () => {
  try {
    await supplierFormRef.value.validate()

    const formData = {
      name: supplierForm.name,
      contact: supplierForm.contact,
      phone: supplierForm.phone,
      email: supplierForm.email,
      address: supplierForm.address,
      username: supplierForm.username,
      status: supplierForm.status ? 1 : 0
    }

    if (dialogType.value === 'add') {
      formData.password = supplierForm.password
      const response = await createSupplier(formData)
      if (response.code === 200) {
        ElMessage.success('创建成功')
        dialogVisible.value = false
        await initData()
      } else {
        ElMessage.error(response.message || '创建失败')
      }
    } else {
      const response = await updateSupplier(supplierForm.id, formData)
      if (response.code === 200) {
        ElMessage.success('更新成功')
        dialogVisible.value = false
        await initData()
      } else {
        ElMessage.error(response.message || '更新失败')
      }
    }
  } catch (error) {
    if (error !== false) {
      console.error('提交失败:', error)
    }
  }
}

onMounted(() => {
  initData()
})
</script>

<style scoped>
.suppliers-container {
  padding: 20px;
}

.page-title {
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 20px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  margin-bottom: 20px;
}

.suppliers-card {
  margin-top: 20px;
}
</style>

