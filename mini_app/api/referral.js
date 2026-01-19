// referral.js - 拉新用户API

import { get, post } from './request';

/**
 * 获取我拉取的用户列表
 * @param {string} token - 用户token
 * @param {Object} params - 查询参数
 * @param {number} params.page_num - 页码
 * @param {number} params.page_size - 每页数量
 * @returns Promise
 */
export const getReferralUsers = (token, params = {}) => {
  return get('/mini-app/users/referrals', params, {
    header: {
      'Authorization': `Bearer ${token}`
    }
  });
};

/**
 * 获取拉新活动说明
 * @param {string} token - 用户token
 * @returns Promise
 */
export const getReferralActivityInfo = (token) => {
  return get('/mini-app/users/referrals/activity-info', {}, {
    header: {
      'Authorization': `Bearer ${token}`
    }
  });
};

/**
 * 获取拉新统计数据
 * @param {string} token - 用户token
 * @returns Promise
 */
export const getReferralStats = (token) => {
  return get('/mini-app/users/referrals/stats', {}, {
    header: {
      'Authorization': `Bearer ${token}`
    }
  });
};
