import request from '../utils/request'

// 获取所有供应商合作申请列表
export function getSupplierApplications(params) {
  return request({
    url: '/admin/supplier-applications',
    method: 'get',
    params
  })
}

// 更新供应商合作申请状态
export function updateSupplierApplicationStatus(id, data) {
  return request({
    url: `/admin/supplier-applications/${id}/status`,
    method: 'put',
    data
  })
}

