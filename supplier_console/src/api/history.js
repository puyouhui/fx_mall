import request from '../utils/request'

// 获取历史记录（按天）
export function getHistoryByDate(params) {
  return request({
    url: '/history',
    method: 'get',
    params
  })
}

// 获取某天的货物详情
export function getHistoryDetail(date) {
  return request({
    url: `/history/${date}`,
    method: 'get'
  })
}

