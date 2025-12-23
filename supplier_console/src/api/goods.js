import request from '../utils/request'

// 获取今日待备货货物列表
export function getTodayPendingGoods(params) {
  return request({
    url: '/goods/today/pending',
    method: 'get',
    params
  })
}

// 获取今日已取货货物列表
export function getTodayPickedGoods(params) {
  return request({
    url: '/goods/today/picked',
    method: 'get',
    params
  })
}

// 获取今日货物统计
export function getTodayGoodsStats() {
  return request({
    url: '/goods/today/stats',
    method: 'get'
  })
}

