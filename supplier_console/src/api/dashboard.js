import request from '../utils/request'

// 获取供应商数据总览
// period: today, 7days, month, year
export function getDashboard(period = 'today') {
  return request({
    url: '/dashboard',
    method: 'get',
    params: {
      period
    }
  })
}

