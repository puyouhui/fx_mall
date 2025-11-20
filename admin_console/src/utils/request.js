import axios from 'axios'

// 创建axios实例
const request = axios.create({
  baseURL: 'http://localhost:8082/api/mini', // 后端API基础URL
  timeout: 10000, // 请求超时时间
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
request.interceptors.request.use(
  config => {
    // 在发送请求之前做些什么
    const token = localStorage.getItem('token')
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`
    }
    return config
  },
  error => {
    // 处理请求错误
    console.error('请求错误:', error)
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  response => {
    // 对响应数据做点什么
    return response.data
  },
  error => {
    // 处理响应错误
    console.error('响应错误:', error)
    
    if (error.response) {
      switch (error.response.status) {
        case 401:
          // 未授权，跳转到登录页
          localStorage.removeItem('token')
          window.location.href = '/login'
          break
        case 403:
          // alert('您没有权限执行此操作')、
          console.log('您没有权限执行此操作');
          
          break
        case 500:
          // alert('服务器错误，请稍后再试')
          console.log('服务器错误，请稍后再试');
          break
        default:
          // alert(`请求失败: ${error.response.data.message || '未知错误'}`)
          console.log(`请求失败: ${error.response.data.message || '未知错误'}`);
          break
      }
    }
    
    return Promise.reject(error)
  }
)

export default request