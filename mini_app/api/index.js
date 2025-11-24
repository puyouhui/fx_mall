// index.js - 小程序首页API

import { get, post, put, del as deleteRequest, BASE_URL } from './request';

/**
 * 获取轮播图数据
 * @returns Promise 轮播图数据
 */
export const getCarousels = () => {
  return get('/carousels', {
    type: 'mini' // 指定小程序类型的轮播图
  }).catch(error => {
    console.error('获取轮播图失败:', error);
    // 返回模拟数据作为后备
    return {
      code: 200,
      data: [
        { id: 1, image: '/static/test/carousel1.png', link: 'product/1', title: '限时特惠' },
        { id: 2, image: '/static/test/carousel1.png', link: 'category/1', title: '新品上市' },
        { id: 3, image: '/static/test/carousel1.png', link: 'product/3', title: '热销推荐' }
      ]
    };
  });
};

/**
 * 获取分类数据
 * @returns Promise 分类数据
 */
export const getCategories = () => {
  return get('/categories').catch(error => {
    console.error('获取分类失败:', error);
    // 返回模拟数据作为后备
    return {
      code: 200,
      data: [
        { id: 1, name: '电子产品', icon: '/static/test/category1.png' },
        { id: 2, name: '家居用品', icon: '/static/test/category1.png' },
        { id: 3, name: '服装鞋帽', icon: '/static/test/category1.png' },
        { id: 4, name: '食品饮料', icon: '/static/test/category1.png' }
      ]
    };
  });
};

/**
 * 获取特价商品数据
 * @param {Object} params - 查询参数
 * @param {number} params.pageNum - 页码
 * @param {number} params.pageSize - 每页数量
 * @returns Promise 特价商品数据
 */
export const getSpecialProducts = (params = { pageNum: 1, pageSize: 10 }) => {
  return get('/products/special', params).catch(error => {
    console.error('获取特价商品失败:', error);
    // 返回模拟数据作为后备
    return {
      code: 200,
      data: [
        {
          id: 1,
          name: '智能手机',
          description: '高性能智能手机，配备高清摄像头',
          categoryId: 1,
          price: 2999.99,
          isSpecial: true,
          images: ['/static/test/product1-1.jpg', '/static/test/product1-1.jpg'],
          specs: [
            { id: 1, name: '8GB+128GB', price: 2999.99 },
            { id: 2, name: '8GB+256GB', price: 3299.99 }
          ]
        },
        {
          id: 3,
          name: '智能手表',
          description: '多功能智能手表，支持健康监测',
          categoryId: 1,
          price: 899.99,
          isSpecial: true,
          images: ['/static/test/product1-1.jpg']
        },
        {
          id: 4,
          name: '时尚台灯',
          description: '护眼台灯，调节亮度',
          categoryId: 2,
          price: 129.99,
          isSpecial: true,
          images: ['/static/test/product1-1.jpg'],
          specs: [
            { id: 1, name: '白色', price: 129.99 },
            { id: 2, name: '黑色', price: 139.99 }
          ]
        },
        {
          id: 6,
          name: '休闲T恤',
          description: '纯棉透气T恤，舒适百搭',
          categoryId: 3,
          price: 89.99,
          isSpecial: true,
          images: ['/static/test/product1-1.jpg'],
          specs: [
            { id: 1, name: 'M码', price: 89.99 },
            { id: 2, name: 'L码', price: 99.99 },
            { id: 3, name: 'XL码', price: 109.99 }
          ]
        }
      ]
    };
  });
};

/**
 * 获取热销商品数据
 * @returns Promise 热销商品数据
 */
export const getHotProducts = () => {
  return get('/products/hot').catch(error => {
    console.error('获取热销商品失败:', error);
    // 返回空数据作为后备，避免页面报错
    return {
      code: 200,
      data: []
    };
  });
};

/**
 * 小程序登录，获取并保存唯一ID
 * @param {string} code - wx.login 返回的code
 */
export const miniLogin = (code) => {
  return post('/auth/login', { code });
};

/**
 * 更新小程序用户身份类型
 * @param {'retail'|'wholesale'} userType
 * @param {string} token
 */
export const updateMiniUserType = (userType, token) => {
  return put('/mini-app/users/type', { user_type: userType }, {
    header: {
      'content-type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {})
    }
  });
};

/**
 * 获取当前登录用户信息
 * @param {string} token - 用户token
 */
export const getMiniUserInfo = (token) => {
  return get('/mini-app/users/info', {}, {
    header: {
      'Authorization': `Bearer ${token}`
    }
  });
};

/**
 * 更新用户姓名
 * @param {string} name - 用户姓名
 * @param {string} token - 用户token
 * @returns Promise
 */
export const updateMiniUserName = (name, token) => {
  return put('/mini-app/users/name', { name }, {
    header: {
      'Authorization': `Bearer ${token}`
    }
  });
};

/**
 * 更新用户电话
 * @param {string} phone - 用户电话
 * @param {string} token - 用户token
 * @returns Promise
 */
export const updateMiniUserPhone = (phone, token) => {
  return put('/mini-app/users/phone', { phone }, {
    header: {
      'Authorization': `Bearer ${token}`
    }
  });
};

/**
 * 更新小程序用户资料（创建或更新地址）
 * @param {Object} profileData - 资料数据
 * @param {number} profileData.address_id - 地址ID，为空表示新增，不为空表示编辑
 * @param {string} profileData.name - 店铺名称
 * @param {string} profileData.contact - 联系人
 * @param {string} profileData.phone - 手机号码
 * @param {string} profileData.address - 地址
 * @param {string} profileData.storeType - 店铺类型
 * @param {string} profileData.salesCode - 业务员代码
 * @param {number} profileData.latitude - 纬度
 * @param {number} profileData.longitude - 经度
 * @param {boolean} profileData.is_default - 是否设置为默认地址
 * @param {string} token - 用户token
 */
export const updateMiniUserProfile = (profileData, token) => {
  return put('/mini-app/users/profile', profileData, {
    header: {
      ...(token ? { Authorization: `Bearer ${token}` } : {})
    }
  });
};

/**
 * 获取用户的所有地址
 * @param {string} token - 用户token
 */
export const getMiniUserAddresses = (token) => {
  return get('/mini-app/users/addresses', {}, {
    header: {
      ...(token ? { Authorization: `Bearer ${token}` } : {})
    }
  });
};

/**
 * 获取用户的默认地址
 * @param {string} token - 用户token
 */
export const getMiniUserDefaultAddress = (token) => {
  return get('/mini-app/users/addresses/default', {}, {
    header: {
      ...(token ? { Authorization: `Bearer ${token}` } : {})
    }
  });
};

/**
 * 获取采购单（购物车）列表
 */
export const getPurchaseList = (token) => {
  return get('/mini-app/users/purchase-list', {}, {
    header: {
      ...(token ? { Authorization: `Bearer ${token}` } : {})
    }
  });
};

/**
 * 添加商品到采购单
 */
export const addPurchaseListItem = (data, token) => {
  return post('/mini-app/users/purchase-list', data, {
    header: {
      ...(token ? { Authorization: `Bearer ${token}` } : {})
    }
  });
};

/**
 * 更新采购单项数量
 */
export const updatePurchaseListItem = (id, data, token) => {
  return put(`/mini-app/users/purchase-list/${id}`, data, {
    header: {
      ...(token ? { Authorization: `Bearer ${token}` } : {})
    }
  });
};

/**
 * 删除采购单项
 */
export const deletePurchaseListItem = (id, token) => {
  return deleteRequest(`/mini-app/users/purchase-list/${id}`, {}, {
    header: {
      ...(token ? { Authorization: `Bearer ${token}` } : {})
    }
  });
};

/**
 * 清空采购单
 */
export const clearPurchaseList = (token) => {
  return deleteRequest('/mini-app/users/purchase-list', {}, {
    header: {
      ...(token ? { Authorization: `Bearer ${token}` } : {})
    }
  });
};

/**
 * 删除地址
 * @param {number} addressId - 地址ID
 * @param {string} token - 用户token
 */
export const deleteMiniUserAddress = (addressId, token) => {
  return deleteRequest(`/mini-app/users/addresses/${addressId}`, {}, {
    header: {
      ...(token ? { Authorization: `Bearer ${token}` } : {})
    }
  });
};

/**
 * 设置默认地址
 * @param {number} addressId - 地址ID
 * @param {string} token - 用户token
 */
export const setDefaultMiniUserAddress = (addressId, token) => {
  return put(`/mini-app/users/addresses/${addressId}/default`, {}, {
    header: {
      ...(token ? { Authorization: `Bearer ${token}` } : {})
    }
  });
};

/**
 * 上传小程序用户头像
 * @param {string} filePath - 图片文件路径
 * @param {string} token - 用户token
 */
export const uploadMiniUserAvatar = (filePath, token) => {
  return new Promise((resolve, reject) => {
    uni.uploadFile({
      url: BASE_URL + '/mini-app/users/avatar',
      filePath: filePath,
      name: 'file',
      header: {
        ...(token ? { Authorization: `Bearer ${token}` } : {})
      },
      success: (res) => {
        try {
          const data = JSON.parse(res.data);
          if (data.code === 200) {
            resolve(data);
          } else {
            reject(new Error(data.message || '上传失败'));
          }
        } catch (error) {
          reject(new Error('解析响应失败'));
        }
      },
      fail: (err) => {
        reject(err);
      }
    });
  });
};

/**
 * 上传地址头像（门头照片）
 * @param {string} filePath - 图片文件路径
 * @param {string} token - 用户token
 * @returns Promise 上传结果
 */
export const uploadAddressAvatar = (filePath, token) => {
  return new Promise((resolve, reject) => {
    uni.uploadFile({
      url: BASE_URL + '/mini-app/users/addresses/avatar',
      filePath: filePath,
      name: 'file',
      header: {
        ...(token ? { Authorization: `Bearer ${token}` } : {})
      },
      success: (res) => {
        try {
          const data = JSON.parse(res.data);
          if (data.code === 200) {
            resolve(data);
          } else {
            reject(new Error(data.message || '上传失败'));
          }
        } catch (error) {
          reject(new Error('解析响应失败'));
        }
      },
      fail: (err) => {
        reject(err);
      }
    });
  });
};