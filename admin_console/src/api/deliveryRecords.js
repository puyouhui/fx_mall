import request from '../utils/request'

// 获取配送记录列表
export const getDeliveryRecords = async (params = {}) => {
  try {
    const res = await request.get('/admin/delivery-records', { params })
    return res
  } catch (error) {
    console.error('获取配送记录列表失败:', error)
    throw error
  }
}

// 获取配送记录详情
export const getDeliveryRecordDetail = async (id) => {
  try {
    const res = await request.get(`/admin/delivery-records/${id}`)
    return res
  } catch (error) {
    console.error('获取配送记录详情失败:', error)
    throw error
  }
}

// 根据订单ID获取配送记录
export const getDeliveryRecordByOrderId = async (orderId) => {
  try {
    const res = await request.get(`/admin/delivery-records/order/${orderId}`)
    return res
  } catch (error) {
    console.error('获取订单配送记录失败:', error)
    throw error
  }
}

