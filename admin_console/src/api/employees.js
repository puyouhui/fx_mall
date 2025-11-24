import request from '../utils/request'

export function getEmployees(params) {
  return request({
    url: '/admin/employees',
    method: 'get',
    params
  })
}

export function getEmployeeDetail(id) {
  return request({
    url: `/admin/employees/${id}`,
    method: 'get'
  })
}

export function createEmployee(data) {
  return request({
    url: '/admin/employees',
    method: 'post',
    data
  })
}

export function updateEmployee(id, data) {
  return request({
    url: `/admin/employees/${id}`,
    method: 'put',
    data
  })
}

export function deleteEmployee(id) {
  return request({
    url: `/admin/employees/${id}`,
    method: 'delete'
  })
}

