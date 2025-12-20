<template>
  <div class="login-container">
    <!-- 左侧登录表单区域 -->
    <div class="login-left">
      <div class="logo">LOGO</div>
      <div class="login-form-wrapper">
        <h1 class="login-title">登录您的账号</h1>
        <el-form ref="loginFormRef" :model="loginForm" :rules="loginRules" class="login-form">
          <el-form-item prop="username">
            <el-input 
              v-model="loginForm.username" 
              placeholder="输入您的账号"
              size="large"
              class="login-input"
            />
          </el-form-item>
          <el-form-item prop="password">
            <el-input
              v-model="loginForm.password"
              type="password"
              placeholder="输入您的密码"
              size="large"
              show-password
              class="login-input"
            />
          </el-form-item>
          <div class="login-options">
            <el-checkbox v-model="rememberMe">记住我</el-checkbox>
            <a href="#" class="forgot-password" @click.prevent="handleForgotPassword">忘记密码?</a>
          </div>
          <el-form-item>
            <el-button
              type="primary"
              :loading="loading"
              class="login-button"
              size="large"
              @click="handleLogin"
            >
              使用账号登录
            </el-button>
          </el-form-item>
        </el-form>
      </div>
    </div>

    <!-- 右侧欢迎区域 -->
    <div class="login-right">
      <div class="welcome-content">
        <img src="../assets/Illustration.png" alt="Illustration" class="illustration" />
        <h2 class="welcome-title">管理控制后台</h2>
        <h3 class="welcome-subtitle">欢迎使用</h3>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { login } from '../api/auth'

const router = useRouter()
const loginFormRef = ref(null)
const loading = ref(false)
const rememberMe = ref(false)
const loginForm = reactive({
  username: 'admin',
  password: 'admin123'
})

const loginRules = {
  username: [
    { required: true, message: '请输入账号', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ]
}

// 处理忘记密码
const handleForgotPassword = () => {
  ElMessageBox.alert('如需重置密码，请联系管理员处理。', '忘记密码', {
    confirmButtonText: '确定',
    type: 'info'
  })
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
    console.log(response.data, response);
    
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
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  display: flex;
  margin: 0;
  padding: 0;
  overflow: hidden;
}

/* 左侧登录区域 */
.login-left {
  flex: 1;
  background: #ffffff;
  display: flex;
  flex-direction: column;
  padding: 40px 60px;
  min-width: 500px;
}

.logo {
  font-size: 24px;
  font-weight: bold;
  color: #333;
  margin-bottom: 40px;
}

.login-form-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  max-width: 400px;
  margin: 0 auto;
  width: 100%;
}

.login-title {
  font-size: 32px;
  font-weight: 600;
  color: #333;
  margin-bottom: 40px;
  text-align: center;
}

.login-form {
  width: 100%;
}

.login-input {
  width: 100%;
  margin-bottom: 20px;
}

.login-input :deep(.el-input__wrapper) {
  border-radius: 8px;
  padding: 12px 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  background-color: #f5f5f5;
}

.login-input :deep(.el-input__wrapper:hover) {
  background-color: #f5f5f5;
}

.login-input :deep(.el-input__wrapper.is-focus) {
  background-color: #f5f5f5;
}

.login-options {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
}

.forgot-password {
  color: #409eff;
  text-decoration: none;
  font-size: 14px;
}

.forgot-password:hover {
  text-decoration: underline;
}

.login-button {
  width: 100%;
  height: 48px;
  font-size: 16px;
  font-weight: 500;
  background-color: #0CAF60;
  border-color: #0CAF60;
  border-radius: 8px;
}

.login-button:hover {
  background-color: #0c9251;
  border-color: #0c9251;
}

/* 右侧欢迎区域 */
.login-right {
  flex: 1;
  background: linear-gradient(135deg, #0CAF60 0%, #5daf34 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 60px;
  min-width: 600px;
}

.welcome-content {
  text-align: center;
  color: #ffffff;
  max-width: 500px;
}

.illustration {
  width: 100%;
  max-width: 500px;
  height: auto;
  margin-bottom: 40px;
}

.welcome-title {
  font-size: 48px;
  font-weight: 700;
  margin-bottom: 20px;
  color: #ffffff;
}

.welcome-subtitle {
  font-size: 32px;
  font-weight: 600;
  margin-bottom: 30px;
  color: #ffffff;
}

.welcome-description {
  font-size: 16px;
  line-height: 1.8;
  color: rgba(255, 255, 255, 0.9);
  text-align: left;
}

/* 响应式设计 */
@media (max-width: 1024px) {
  .login-container {
    flex-direction: column;
  }

  .login-left {
    min-width: 100%;
    padding: 30px 40px;
  }

  .login-right {
    display: none;
  }
}
</style>
