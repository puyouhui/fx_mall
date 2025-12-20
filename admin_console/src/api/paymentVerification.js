import request from '../utils/request'

// 获取收款审核列表
export const getPaymentVerificationRequests = async (params = {}) => {
  try {
    const res = await request.get('/admin/payment-verification', { params })
    return res
  } catch (error) {
    console.error('获取收款审核列表失败:', error)
    throw error
  }
}

// 审核收款申请
export const reviewPaymentVerification = async (id, data) => {
  try {
    const res = await request.post('/admin/payment-verification/review', data)
    return res
  } catch (error) {
    console.error('审核收款申请失败:', error)
    throw error
  }
}

