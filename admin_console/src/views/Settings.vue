<template>
  <div class="settings-container">
    <h1 class="page-title">系统设置</h1>
    
    <el-card class="settings-card">
      <template #header>
        <div class="card-header">
          <el-icon><Lock /></el-icon>
          <span>修改密码</span>
        </div>
      </template>
      
      <el-form
        ref="passwordFormRef"
        :model="passwordForm"
        :rules="passwordRules"
        label-width="120px"
        class="password-form"
      >
        <el-form-item label="原密码" prop="old_password">
          <el-input
            v-model="passwordForm.old_password"
            type="password"
            placeholder="请输入原密码"
            show-password
            :prefix-icon="Lock"
            style="width: 400px"
          />
        </el-form-item>
        
        <el-form-item label="新密码" prop="new_password">
          <el-input
            v-model="passwordForm.new_password"
            type="password"
            placeholder="请输入新密码（至少6位）"
            show-password
            :prefix-icon="Lock"
            style="width: 400px"
          />
        </el-form-item>
        
        <el-form-item label="确认新密码" prop="confirm_password">
          <el-input
            v-model="passwordForm.confirm_password"
            type="password"
            placeholder="请再次输入新密码"
            show-password
            :prefix-icon="Lock"
            style="width: 400px"
          />
        </el-form-item>
        
        <el-form-item>
          <el-button
            type="primary"
            :loading="loading"
            @click="handleChangePassword"
          >
            确认修改
          </el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Lock } from '@element-plus/icons-vue'
import { changePassword } from '../api/auth'

const passwordFormRef = ref(null)
const loading = ref(false)

const passwordForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

// 自定义验证规则：确认密码
const validateConfirmPassword = (rule, value, callback) => {
  if (value === '') {
    callback(new Error('请再次输入新密码'))
  } else if (value !== passwordForm.new_password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const passwordRules = {
  old_password: [
    { required: true, message: '请输入原密码', trigger: 'blur' }
  ],
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少为6位', trigger: 'blur' }
  ],
  confirm_password: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' }
  ]
}

// 处理修改密码
const handleChangePassword = async () => {
  try {
    // 验证表单
    await passwordFormRef.value.validate()
    
    loading.value = true
    
    // 调用API修改密码
    const response = await changePassword({
      old_password: passwordForm.old_password,
      new_password: passwordForm.new_password
    })
    
    if (response.code === 200) {
      ElMessage.success('密码修改成功，请重新登录')
      
      // 询问是否立即退出登录
      ElMessageBox.confirm(
        '密码已修改成功，为了安全起见，建议您重新登录。是否立即退出登录？',
        '提示',
        {
          confirmButtonText: '立即退出',
          cancelButtonText: '稍后退出',
          type: 'success'
        }
      ).then(() => {
        // 退出登录
        localStorage.removeItem('token')
        localStorage.removeItem('username')
        window.location.href = '/login'
      }).catch(() => {
        // 用户选择稍后退出，重置表单
        handleReset()
      })
    } else {
      ElMessage.error(response.message || '密码修改失败')
    }
  } catch (error) {
    console.error('修改密码失败:', error)
    if (error.response && error.response.data) {
      ElMessage.error(error.response.data.message || '密码修改失败')
    } else if (error.message) {
      ElMessage.error(error.message)
    } else {
      ElMessage.error('密码修改失败，请稍后再试')
    }
  } finally {
    loading.value = false
  }
}

// 重置表单
const handleReset = () => {
  passwordFormRef.value?.resetFields()
  passwordForm.old_password = ''
  passwordForm.new_password = ''
  passwordForm.confirm_password = ''
}
</script>

<style scoped>
.settings-container {
  padding: 20px;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 20px;
}

.settings-card {
  max-width: 800px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.password-form {
  padding: 20px 0;
}

:deep(.el-form-item__label) {
  font-weight: 500;
}

:deep(.el-input__wrapper) {
  border-radius: 4px;
}
</style>

