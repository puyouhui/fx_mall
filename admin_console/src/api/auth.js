import request from '../utils/request'

/**
 * 管理员登录接口
 * @param {Object} params - 登录参数
 * @returns {Promise}
 */
export function login(params) {
  return request({
    url: '/admin/login',
    method: 'post',
    data: params
  })
}

/**
 * 管理员退出登录接口
 * @returns {Promise}
 */
export function logout() {
  return request({
    url: '/admin/logout',
    method: 'post'
  })
}

/**
 * 获取管理员信息接口
 * @returns {Promise}
 */
export function getAdminInfo() {
  return request({
    url: '/admin/info',
    method: 'get'
  })
}