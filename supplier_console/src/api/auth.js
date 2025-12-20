import request from '../utils/request'

// 供应商登录
export function login(data) {
  return request({
    url: '/login',
    method: 'post',
    data
  })
}

// 获取供应商信息
export function getSupplierInfo() {
  return request({
    url: '/auth/info',
    method: 'get'
  })
}

// 退出登录
export function logout() {
  return request({
    url: '/auth/logout',
    method: 'post'
  })
}

