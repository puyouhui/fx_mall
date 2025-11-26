import request from '../utils/request'

// 获取优惠券列表
export const getCoupons = async () => {
  try {
    const { data } = await request.get('/admin/coupons')
    return data
  } catch (error) {
    console.error('获取优惠券列表失败:', error)
    throw error
  }
}

// 获取优惠券详情
export const getCouponById = async (id) => {
  try {
    const { data } = await request.get(`/admin/coupons/${id}`)
    return data
  } catch (error) {
    console.error('获取优惠券详情失败:', error)
    throw error
  }
}

// 创建优惠券
export const createCoupon = async (couponData) => {
  try {
    const { data } = await request.post('/admin/coupons', couponData)
    return data
  } catch (error) {
    console.error('创建优惠券失败:', error)
    throw error
  }
}

// 更新优惠券
export const updateCoupon = async (id, couponData) => {
  try {
    const { data } = await request.put(`/admin/coupons/${id}`, couponData)
    return data
  } catch (error) {
    console.error('更新优惠券失败:', error)
    throw error
  }
}

// 删除优惠券
export const deleteCoupon = async (id) => {
  try {
    const { data } = await request.delete(`/admin/coupons/${id}`)
    return data
  } catch (error) {
    console.error('删除优惠券失败:', error)
    throw error
  }
}

// 发放优惠券给用户
export const issueCouponToUser = async (issueData) => {
  try {
    const { data } = await request.post('/admin/coupons/issue', issueData)
    return data
  } catch (error) {
    console.error('发放优惠券失败:', error)
    throw error
  }
}

