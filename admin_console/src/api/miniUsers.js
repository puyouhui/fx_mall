import request from '../utils/request'

export function getMiniUsers(params) {
  return request({
    url: '/admin/mini-app/users',
    method: 'get',
    params
  })
}

export function getMiniUserDetail(id) {
  return request({
    url: `/admin/mini-app/users/${id}`,
    method: 'get'
  })
}

export function updateMiniUser(id, data) {
  return request({
    url: `/admin/mini-app/users/${id}`,
    method: 'put',
    data
  })
}

export function getAdminAddressDetail(id) {
  return request({
    url: `/admin/mini-app/addresses/${id}`,
    method: 'get'
  })
}

export function updateAdminAddress(id, data) {
  return request({
    url: `/admin/mini-app/addresses/${id}`,
    method: 'put',
    data
  })
}

export function getSalesEmployees() {
  return request({
    url: '/admin/employees/sales',
    method: 'get'
  })
}

// 获取用户优惠券列表
export function getUserCoupons(userId) {
  return request({
    url: `/admin/mini-app/users/${userId}/coupons`,
    method: 'get'
  })
}

// 上传用户头像
export function uploadUserAvatar(userId, file) {
  const formData = new FormData()
  formData.append('file', file)
  return request({
    url: `/admin/mini-app/users/${userId}/avatar`,
    method: 'post',
    headers: {
      'Content-Type': 'multipart/form-data'
    },
    data: formData
  })
}

