import request from '../utils/request'

// 获取供应商订单列表（分页）
export function getOrders(params) {
  return request({
    url: '/orders',
    method: 'get',
    params
  })
}

// 获取订单详情
export function getOrderDetail(id) {
  return request({
    url: `/orders/${id}`,
    method: 'get'
  })
}

