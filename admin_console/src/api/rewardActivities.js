import request from '../utils/request'

// 获取奖励活动列表
export const getRewardActivities = async (params) => {
  try {
    const res = await request.get('/admin/reward-activities', { params })
    return res
  } catch (error) {
    console.error('获取奖励活动列表失败:', error)
    throw error
  }
}

// 获取单个奖励活动
export const getRewardActivity = async (id) => {
  try {
    const res = await request.get(`/admin/reward-activities/${id}`)
    return res
  } catch (error) {
    console.error('获取奖励活动失败:', error)
    throw error
  }
}

// 创建奖励活动
export const createRewardActivity = async (activityData) => {
  try {
    const res = await request.post('/admin/reward-activities', activityData)
    return res
  } catch (error) {
    console.error('创建奖励活动失败:', error)
    throw error
  }
}

// 更新奖励活动
export const updateRewardActivity = async (id, activityData) => {
  try {
    const res = await request.put(`/admin/reward-activities/${id}`, activityData)
    return res
  } catch (error) {
    console.error('更新奖励活动失败:', error)
    throw error
  }
}

// 删除奖励活动
export const deleteRewardActivity = async (id) => {
  try {
    const res = await request.delete(`/admin/reward-activities/${id}`)
    return res
  } catch (error) {
    console.error('删除奖励活动失败:', error)
    throw error
  }
}
