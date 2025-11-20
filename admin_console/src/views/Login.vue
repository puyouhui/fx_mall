<template>
  <div class="login-container">
    <div class="login-box">
      <div class="login-header">
        <h2>云鹿进货管理后台</h2>
        <p>请输入您的账号和密码登录</p>
      </div>
      <el-form ref="loginFormRef" :model="loginForm" :rules="loginRules" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="loginForm.username" placeholder="请输入用户名" :prefix-icon="User" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="请输入密码"
            :prefix-icon="Lock"
            show-password
          />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            :loading="loading"
            style="width: 100%;"
            @click="handleLogin"
          >
            登录
          </el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import { login } from '../api/auth'

const router = useRouter()
const loginFormRef = ref(null)
const loading = ref(false)
const loginForm = reactive({
  username: 'admin',
  password: 'admin123'
})

const loginRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 2, max: 20, message: '用户名长度在 2 到 20 个字符', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度在 6 到 20 个字符', trigger: 'blur' }
  ]
}

// 处理登录
const handleLogin = async () => {
  try {
    // 验证表单
    await loginFormRef.value.validate()
    loading.value = true
    
    // 调用真实的登录API
    const response = await login(loginForm)
    

    console.log(response);
    console.log(response.data,response);
    
    // 保存token和用户名到本地存储
    if (response.data && response.data.token) {
      localStorage.setItem('token', response.data.token)
      localStorage.setItem('username', loginForm.username)
      
      ElMessage.success('登录成功')
      router.push('/dashboard')
    } else {
      ElMessage.error('登录失败，返回数据格式不正确')
    }
  } catch (error) {
    loading.value = false
    if (error.response && error.response.data && error.response.data.message) {
      ElMessage.error(error.response.data.message || '登录失败，请稍后再试')
    } else {
      ElMessage.error('登录失败，请稍后再试')
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  height: 100vh;
  min-width: 100%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  justify-content: center;
  align-items: center;
}

.login-box {
  width: 400px;
  background: #fff;
  border-radius: 8px;
  padding: 40px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
}

.login-header {
  text-align: center;
  margin-bottom: 30px;
}

.login-header h2 {
  color: #333;
  margin-bottom: 10px;
}

.login-header p {
  color: #999;
  font-size: 14px;
}
</style>