import request from '../utils/request'

// 获取配送员收入统计（管理员）
export function getDeliveryIncomeStats(employeeCode) {
  const params = {}
  if (employeeCode) {
    params.employee_code = employeeCode
  }
  return request({
    url: '/admin/delivery-income/stats',
    method: 'get',
    params
  })
}

// 批量结算配送费
export function batchSettleDeliveryFees(data) {
  return request({
    url: '/admin/delivery-income/settle',
    method: 'post',
    data
  })
}

