import request from '../utils/request'

// 获取已付款清单
export function getPaidItems(params = {}) {
  return request({
    url: '/payments/paid',
    method: 'get',
    params
  })
}

// 获取待付款清单
export function getPendingItems(params = {}) {
  return request({
    url: '/payments/pending',
    method: 'get',
    params
  })
}

// 获取对账统计
export function getPaymentStats(params = {}) {
  return request({
    url: '/payments/stats',
    method: 'get',
    params
  })
}

