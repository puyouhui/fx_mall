import axios from 'axios'

// 根据环境自动选择 API 地址
// 开发环境使用 localhost，生产环境使用相对路径（通过 Nginx 代理）
const getBaseURL = () => {
  // 如果是开发环境（localhost 或 127.0.0.1）
  if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
    // return 'http://localhost:8082/api/mini'
    return 'https://mall.sscchh.com/api_mall/mini' // 生产环境
  }
  // 生产环境使用相对路径，通过 Nginx 代理到后端
  // 注意：后端 Nginx 配置为 /api_mall/，所以这里使用 /api_mall/mini
  return '/api_mall/mini'
}

// 创建axios实例
const request = axios.create({
  baseURL: getBaseURL(),
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