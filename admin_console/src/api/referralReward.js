import request from '../utils/request'

// 获取推荐奖励活动配置
export const getReferralRewardConfig = async () => {
  try {
    const { data } = await request.get('/admin/referral-reward/config')
    return data
  } catch (error) {
    console.error('获取推荐奖励活动配置失败:', error)
    throw error
  }
}

// 更新推荐奖励活动配置
export const updateReferralRewardConfig = async (configData) => {
  try {
    const { data } = await request.put('/admin/referral-reward/config', configData)
    return data
  } catch (error) {
    console.error('更新推荐奖励活动配置失败:', error)
    throw error
  }
}

// 获取推荐奖励记录列表
export const getReferralRewards = async (params) => {
  try {
    const res = await request.get('/admin/referral-reward/rewards', { params })
    return res
  } catch (error) {
    console.error('获取推荐奖励记录失败:', error)
    throw error
  }
}

