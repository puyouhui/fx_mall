import request from '../utils/request'

// 获取所有供应商
export function getAllSuppliers() {
  return request({
    url: '/admin/suppliers',
    method: 'get'
  })
}

// 获取供应商详情
export function getSupplierById(id) {
  return request({
    url: `/admin/suppliers/${id}`,
    method: 'get'
  })
}

// 创建供应商
export function createSupplier(data) {
  return request({
    url: '/admin/suppliers',
    method: 'post',
    data
  })
}

// 更新供应商
export function updateSupplier(id, data) {
  return request({
    url: `/admin/suppliers/${id}`,
    method: 'put',
    data
  })
}

// 删除供应商
export function deleteSupplier(id) {
  return request({
    url: `/admin/suppliers/${id}`,
    method: 'delete'
  })
}

