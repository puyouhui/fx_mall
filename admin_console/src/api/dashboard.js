import request from '../utils/request'

// 获取仪表盘统计数据
export const getDashboardStats = (params = {}) => {
  return request.get('/admin/dashboard/stats', { params })
}