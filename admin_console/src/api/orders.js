import request from '../utils/request'

// 获取订单列表
export const getOrders = async (params = {}) => {
  try {
    const res = await request.get('/admin/orders', { params })
    // request.get 已经通过拦截器返回了 response.data，所以 res 就是后端返回的完整响应
    return res
  } catch (error) {
    console.error('获取订单列表失败:', error)
    throw error
  }
}

// 获取订单详情
export const getOrderDetail = async (id) => {
  try {
    const res = await request.get(`/admin/orders/${id}`)
    // request.get 已经通过拦截器返回了 response.data，所以 res 就是后端返回的完整响应
    return res
  } catch (error) {
    console.error('获取订单详情失败:', error)
    throw error
  }
}

// 更新订单状态
export const updateOrderStatus = async (id, status) => {
  try {
    const res = await request.put(`/admin/orders/${id}/status`, { status })
    return res
  } catch (error) {
    console.error('更新订单状态失败:', error)
    throw error
  }
}

// 管理员手动退款（用于支付回调未同步等异常场景）
export const manualRefundOrder = async (id) => {
  try {
    const res = await request.post(`/admin/orders/${id}/manual-refund`)
    return res
  } catch (error) {
    console.error('手动退款失败:', error)
    throw error
  }
}

// 售后退款（指定金额、原因、是否取消订单）
export const refundOrderWithDetails = async (id, data) => {
  try {
    const res = await request.post(`/admin/orders/${id}/refund-with-details`, data)
    return res
  } catch (error) {
    console.error('售后退款失败:', error)
    throw error
  }
}

// 强制重新计算订单利润
export const recalculateOrderProfit = async (id) => {
  try {
    const res = await request.post(`/admin/orders/${id}/recalculate-profit`)
    return res
  } catch (error) {
    console.error('重新计算订单利润失败:', error)
    throw error
  }
}

