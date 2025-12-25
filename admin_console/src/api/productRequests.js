import request from '../utils/request'

// 获取所有新品需求列表
export function getProductRequests(params) {
  return request({
    url: '/admin/product-requests',
    method: 'get',
    params
  })
}

// 更新新品需求状态
export function updateProductRequestStatus(id, data) {
  return request({
    url: `/admin/product-requests/${id}/status`,
    method: 'put',
    data
  })
}

