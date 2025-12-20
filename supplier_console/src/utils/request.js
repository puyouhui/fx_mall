import axios from 'axios'

// 创建axios实例
const request = axios.create({
  baseURL: 'http://localhost:8082/api/mini/supplier', // 供应商后台API基础URL
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
    const data = response.data
    // 如果后端返回的格式是 {code, data, message}，直接返回
    // 如果 code 不是 200，抛出错误
    if (data.code && data.code !== 200) {
      const error = new Error(data.message || '请求失败')
      error.response = {
        status: data.code,
        data: data
      }
      return Promise.reject(error)
    }
    return data
  },
  error => {
    // 处理响应错误
    console.error('响应错误:', error)
    
    if (error.response) {
      const errorData = error.response.data || {}
      const errorMessage = errorData.message || '请求失败'
      
      switch (error.response.status) {
        case 401:
          // 未授权，清除token并跳转到登录页
          localStorage.removeItem('token')
          localStorage.removeItem('supplierId')
          localStorage.removeItem('supplierName')
          // 只在非登录页面时跳转
          if (window.location.pathname !== '/login') {
            window.location.href = '/login'
          }
          break
        case 403:
          console.log('您没有权限执行此操作')
          break
        case 500:
          console.log('服务器错误，请稍后再试')
          break
        default:
          console.log(`请求失败: ${errorMessage}`)
          break
      }
      
      // 创建一个新的错误对象，包含后端返回的 message
      const customError = new Error(errorMessage)
      customError.response = error.response
      return Promise.reject(customError)
    }
    
    return Promise.reject(error)
  }
)

export default request

