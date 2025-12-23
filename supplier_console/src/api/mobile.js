import axios from 'axios'

// 创建独立的axios实例，不添加token
const mobileRequest = axios.create({
  baseURL: 'http://localhost:8082/api/mini',
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

