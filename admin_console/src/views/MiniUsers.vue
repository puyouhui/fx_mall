<template>
  <div class="mini-users-page">
    <el-card class="mini-users-card" shadow="never">
      <div class="card-header">
        <div class="title">
          <span class="main">小程序用户</span>
          <span class="sub">查看登录用户唯一ID与基础信息</span>
        </div>
        <div class="actions">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索唯一ID / 姓名 / 电话"
            clearable
            @keyup.enter="handleSearch"
          />
          <el-button type="primary" @click="handleSearch">搜索</el-button>
        </div>
      </div>

      <el-table
        v-loading="loading"
        :data="users"
        border
        stripe
        class="mini-users-table"
      >
        <!-- <el-table-column prop="unique_id" label="唯一ID" min-width="220" /> -->
        <el-table-column prop="id" label="ID" min-width="60" />
        <el-table-column prop="name" label="名称" min-width="120">
          <template #default="scope">
            {{ scope.row.name || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="contact" label="联系人" min-width="120">
          <template #default="scope">
            {{ scope.row.contact || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="phone" label="手机号" min-width="130">
          <template #default="scope">
            {{ scope.row.phone || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="address" label="地址" min-width="180" show-overflow-tooltip>
          <template #default="scope">
            {{ scope.row.address || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="经纬度" min-width="140">
          <template #default="scope">
            <span v-if="scope.row.latitude && scope.row.longitude">
              {{ scope.row.latitude }}, {{ scope.row.longitude }}
            </span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="user_type" label="用户类型" min-width="120">
          <template #default="scope">
            <el-tag :type="scope.row.user_type === 'wholesale' ? 'warning' : 'success'">
              {{ formatUserType(scope.row.user_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="profile_completed" label="资料完善" min-width="110">
          <template #default="scope">
            <el-tag :type="scope.row.profile_completed ? 'success' : 'info'">
              {{ scope.row.profile_completed ? '已完善' : '未完善' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="sales_code" label="销售员代码" min-width="120">
          <template #default="scope">
            {{ scope.row.sales_code || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="store_type" label="店铺类型" min-width="120">
          <template #default="scope">
            {{ scope.row.store_type || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="首次登录时间" min-width="160">
          <template #default="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <!-- <el-table-column prop="updated_at" label="最近更新时间" min-width="160">
          <template #default="scope">
            {{ formatDate(scope.row.updated_at) }}
          </template>
        </el-table-column> -->
      </el-table>

      <div class="pagination">
        <el-pagination
          background
          layout="total, prev, pager, next, jumper"
          :page-size="pagination.pageSize"
          :current-page="pagination.pageNum"
          :total="pagination.total"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getMiniUsers } from '../api/miniUsers'

const loading = ref(false)
const users = ref([])
const searchKeyword = ref('')
const pagination = reactive({
  pageNum: 1,
  pageSize: 10,
  total: 0
})

const loadUsers = async () => {
  loading.value = true
  try {
    const res = await getMiniUsers({
      pageNum: pagination.pageNum,
      pageSize: pagination.pageSize,
      keyword: searchKeyword.value
    })
    if (res.code === 200) {
      users.value = Array.isArray(res.data) ? res.data : []
      pagination.total = res.total || users.value.length
    }
  } catch (error) {
    console.error('获取用户失败:', error)
    ElMessage.error('获取用户列表失败，请稍后再试')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.pageNum = 1
  loadUsers()
}

const handlePageChange = (page) => {
  pagination.pageNum = page
  loadUsers()
}

const formatDate = (value) => {
  if (!value) return '-'
  return new Date(value).toLocaleString()
}

const formatUserType = (type) => {
  if (type === 'wholesale') return '批发用户'
  if (type === 'retail') return '零售用户'
  return '未选择'
}

onMounted(() => {
  loadUsers()
})
</script>

<style scoped>
.mini-users-page {
  padding: 20px 0;
}

.mini-users-card {
  border: none;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.04);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 12px;
}

.title .main {
  font-size: 20px;
  font-weight: 600;
  margin-right: 12px;
}

.title .sub {
  color: #909399;
  font-size: 14px;
}

.actions {
  display: flex;
  gap: 12px;
  min-width: 320px;
}

.mini-users-table {
  margin-top: 10px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>

