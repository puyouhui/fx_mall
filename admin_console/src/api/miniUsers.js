import request from '../utils/request'

export function getMiniUsers(params) {
  return request({
    url: '/admin/mini-app/users',
    method: 'get',
    params
  })
}

