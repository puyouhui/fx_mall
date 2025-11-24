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

