import axios from 'axios'

// 根据环境自动选择 API 地址
// 开发环境使用 localhost，生产环境使用相对路径（通过 Nginx 代理）
const getBaseURL = () => {
  // 如果是开发环境（localhost 或 127.0.0.1）
  if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
    return 'http://localhost:8082/api/mini' // 开发环境
  }
  // 生产环境使用相对路径，通过 Nginx 代理到后端
  // 注意：后端 Nginx 配置为 /api_mall/，所以这里使用 /api_mall/mini
  return '/api_mall/mini'
}

// 创建独立的axios实例，不添加token
const mobileRequest = axios.create({
  baseURL: getBaseURL(),
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 响应拦截器
mobileRequest.interceptors.response.use(
  response => {
    const data = response.data
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
    console.error('响应错误:', error)
    if (error.response) {
      const errorData = error.response.data || {}
      const errorMessage = errorData.message || '请求失败'
      const customError = new Error(errorMessage)
      customError.response = error.response
      return Promise.reject(customError)
    }
    return Promise.reject(error)
  }
)

// 获取移动端待备货货物列表（不需要token）
export function getMobilePendingGoods(supplierName, supplierId) {
  return mobileRequest({
    url: '/mobile/pending-goods',
    method: 'get',
    params: {
      name: supplierName,
      ID: supplierId
    }
  })
}

