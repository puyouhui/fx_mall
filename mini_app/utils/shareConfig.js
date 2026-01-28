/**
 * 分享配置工具
 * 统一管理小程序的分享文字和封面图片
 */

// 默认分享配置
const DEFAULT_SHARE_CONFIG = {
  // 首页分享配置
  index: {
    title: '发现一个进货小程序，新用户享专属优惠，快来使用吧~',
    imageUrl: 'https://mall.sscchh.com/minio/fengxing/others/image_1769584856.png' // 可以替换为您的分享封面图
  },
  
  // 推荐页面分享配置
  referral: {
    title: '邀请您使用小程序快捷订货，新用户登录还可获得现金奖励~',
    imageUrl: 'https://mall.sscchh.com/minio/fengxing/others/image_1769584856.png' // 可以替换为推荐活动封面图
  },
  
  // 商品详情页分享配置（动态，使用商品信息）
  product: {
    title: '给你推荐了{productName}， 快来小程序选购吧~',
    imageUrl: '' // 使用商品图片
  },
  
  // 订单详情页分享配置
  order: {
    title: '订单详情 - 订单号：{orderNumber}',
    imageUrl: 'https://mall.sscchh.com/minio/fengxing/others/image_1769584856.png'
  }
};

/**
 * 获取分享配置
 * @param {string} type - 分享类型：'index' | 'referral' | 'product' | 'order'
 * @param {Object} options - 额外参数
 * @param {string} options.productName - 商品名称（product类型需要）
 * @param {string} options.productImage - 商品图片（product类型需要）
 * @param {string} options.orderNumber - 订单号（order类型需要）
 * @returns {Object} 分享配置 { title, imageUrl }
 */
export function getShareConfig(type, options = {}) {
  const config = DEFAULT_SHARE_CONFIG[type] || DEFAULT_SHARE_CONFIG.index;
  
  let title = config.title;
  let imageUrl = config.imageUrl;
  
  // 根据类型处理动态内容
  switch (type) {
    case 'product':
      if (options.productName) {
        title = title.replace('{productName}', options.productName);
      }
      if (options.productImage) {
        imageUrl = options.productImage;
      }
      break;
      
    case 'order':
      if (options.orderNumber) {
        title = title.replace('{orderNumber}', options.orderNumber);
      }
      break;
      
    case 'index':
      // 首页直接使用配置，无需特殊处理
      break;
      
    case 'referral':
      // 推荐页面直接使用配置，无需特殊处理
      break;
  }
  
  const result = {
    title: title,
    imageUrl: imageUrl || ''
  };
  
  // 调试信息
  console.log(`[shareConfig] 获取分享配置 - 类型: ${type}`, result);
  
  return result;
}

/**
 * 构建分享路径（带推荐者ID）
 * @param {string} basePath - 基础路径
 * @param {Object} params - 路径参数
 * @returns {string} 完整的分享路径
 */
export function buildSharePath(basePath, params = {}) {
  let path = basePath;
  const queryParams = [];
  
  // 添加推荐者ID
  const userInfo = uni.getStorageSync('miniUserInfo');
  const userId = userInfo?.id || userInfo?.ID;
  
  // 调试信息
  console.log('[buildSharePath] 用户信息:', userInfo);
  console.log('[buildSharePath] 提取的用户ID:', userId);
  
  if (userId) {
    queryParams.push(`referrer_id=${userId}`);
    console.log('[buildSharePath] 已添加推荐者ID:', userId);
  } else {
    console.warn('[buildSharePath] 未找到用户ID，无法添加推荐者ID');
  }
  
  // 添加其他参数
  Object.keys(params).forEach(key => {
    if (params[key] !== undefined && params[key] !== null) {
      queryParams.push(`${key}=${params[key]}`);
    }
  });
  
  if (queryParams.length > 0) {
    path += (path.includes('?') ? '&' : '?') + queryParams.join('&');
  }
  
  console.log('[buildSharePath] 最终分享路径:', path);
  
  return path;
}

/**
 * 获取默认分享图片
 * 如果配置的图片加载失败，可以使用这个作为后备
 */
export function getDefaultShareImage() {
  return 'https://mall.sscchh.com/minio/fengxing/products/product_1769156291.jpg';
}
