import request from '../utils/request'

/**
 * 获取所有价格反馈列表
 */
export const getPriceFeedbacks = (params) => {
  return request({
    url: '/admin/price-feedback',
    method: 'get',
    params
  })
}

/**
 * 更新价格反馈状态
 */
export const updatePriceFeedbackStatus = (id, data) => {
  return request({
    url: `/admin/price-feedback/${id}/status`,
    method: 'put',
    data
  })
}

