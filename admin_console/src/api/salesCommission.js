import request from '../utils/request'

// 获取所有销售员的分成统计（管理员）
export function getSalesCommissionStats(employeeCode, month) {
  const params = {}
  if (employeeCode) {
    params.employee_code = employeeCode
  }
  if (month) {
    params.month = month
  }
  return request({
    url: '/admin/sales-commission/stats',
    method: 'get',
    params
  })
}

// 获取销售员的分成记录列表（管理员）
export function getSalesCommissions(employeeCode, month, pageNum, pageSize) {
  const params = {
    employee_code: employeeCode,
    pageNum: pageNum || 1,
    pageSize: pageSize || 10
  }
  if (month) {
    params.month = month
  }
  return request({
    url: '/admin/sales-commission/list',
    method: 'get',
    params
  })
}

// 获取销售员的分成配置（管理员）
export function getSalesCommissionConfig(employeeCode) {
  return request({
    url: '/admin/sales-commission/config',
    method: 'get',
    params: {
      employee_code: employeeCode
    }
  })
}

// 更新销售员的分成配置（管理员）
export function updateSalesCommissionConfig(data) {
  return request({
    url: '/admin/sales-commission/config',
    method: 'put',
    data
  })
}

// 获取所有销售员列表（用于下拉选择）
export function getSalesEmployees() {
  return request({
    url: '/admin/employees/sales',
    method: 'get'
  })
}

