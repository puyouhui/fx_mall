// products.js - 商品相关API

import { get } from './request';

/**
 * 获取商品详情
 * @param {number} productId - 商品ID
 * @returns Promise 商品详情数据
 */
export const getProductDetail = (productId) => {
  return get(`/products/${productId}`).catch(error => {
    console.error('获取商品详情失败:', error);
    // 返回模拟数据作为后备
    let mockData;
    
    // 对于ID为2的商品，提供更完整的规格信息
    if (parseInt(productId) === 2) {
      mockData = {
        code: 200,
        data: {
          id: 2,
          name: '抽纸-竹江BAMBOO精装抽纸件/100包',
          description: '描述1111',
          price: 199,
          originalPrice: 299,
          categoryId: 14,
          categoryName: '商品分类',
          isSpecial: true,
          images: [
            'http://113.44.164.151:9000/selected/product_1758766782.jpg'
          ],
          specifications: [
            { name: '包装', value: '提' },
            { name: '单位', value: '件' },
            { name: '材质', value: '竹浆' },
            { name: '规格', value: '3层*100抽*30包' }
          ],
          specs: [
            { id: 5, name: '3提装', description: '家庭装', price: 199 },
            { id: 6, name: '5提装', description: '量贩装', price: 299 }
          ],
          stock: 100,
          sales: 50,
          details: '描述1111',
          createdAt: '2025-09-25T10:25:01+08:00',
          updatedAt: '2025-09-25T11:09:59+08:00'
        }
      };
    } else {
      // 其他ID的商品数据
      mockData = {
        code: 200,
        data: {
          id: productId,
          name: '智能手机',
          description: '高性能智能手机，配备高清摄像头和大容量电池',
          categoryId: 1,
          categoryName: '电子产品',
          price: 2999.99,
          originalPrice: 3999.99,
          isSpecial: true,
          stock: 100,
          sales: 500,
          images: [
            'https://mall.sscchh.com/minio/fengxing/products/product_1769156291.jpg',
            'https://mall.sscchh.com/minio/fengxing/products/product_1769156291.jpg',
            'https://mall.sscchh.com/minio/fengxing/products/product_1769156291.jpg'
          ],
          specifications: [
            { name: '颜色', value: '黑色' },
            { name: '内存', value: '128GB' },
            { name: '屏幕尺寸', value: '6.7英寸' },
            { name: '处理器', value: '最新一代芯片' }
          ],
          specs: [
            { id: 1, name: '8GB+128GB', description: '标准版', price: 2999.99 },
            { id: 2, name: '8GB+256GB', description: '高配版', price: 3299.99 }
          ],
          details: '这是一款高性能智能手机，具有出色的拍照和游戏性能。'
        }
      };
    }
    
    return mockData;
  });
};

/**
 * 获取分类商品
 * @param {Object} params - 查询参数
 * @param {number} params.categoryId - 分类ID
 * @param {number} params.pageNum - 页码
 * @param {number} params.pageSize - 每页数量
 * @returns Promise 商品列表数据
 */
export const getProductsByCategory = (params = { categoryId: 1, pageNum: 1, pageSize: 10 }) => {
  console.log('调用getProductsByCategory，参数:', params);
  return get('/products/category', params).catch(error => {
    console.error('获取分类商品失败:', error);
    // 返回模拟数据作为后备
    return {
      code: 200,
      data: {
        list: [
          {
            id: 1,
            name: '智能手机',
            price: 2999.99,
            isSpecial: true,
            images: ['https://mall.sscchh.com/minio/fengxing/products/product_1769156291.jpg'],
            specs: [
              { id: 1, name: '8GB+128GB', price: 2999.99 },
              { id: 2, name: '8GB+256GB', price: 3299.99 }
            ]
          },
          {
            id: 3,
            name: '智能手表',
            price: 899.99,
            isSpecial: false,
            images: ['https://mall.sscchh.com/minio/fengxing/products/product_1769156291.jpg']
          }
        ],
        total: 2,
        pageNum: 1,
        pageSize: 10
      }
    };
  });
};

/**
 * 搜索商品建议
 * @param {string} keyword - 搜索关键词
 * @param {number} limit - 返回数量限制（可选，默认10）
 * @returns Promise 商品名称列表
 */
export const searchProductSuggestions = (keyword, limit = 10) => {
  const params = {
    keyword: keyword,
    limit: limit
  };
  return get('/products/search/suggestions', params).catch(error => {
    console.error('获取搜索建议失败:', error);
    // 返回空数组
    return {
      code: 200,
      data: []
    };
  });
};

/**
 * 搜索商品
 * @param {string} keyword - 搜索关键词
 * @param {number} pageNum - 页码（可选）
 * @param {number} pageSize - 每页数量（可选）
 * @returns Promise 搜索结果
 */
export const searchProducts = (keyword, pageNum = 1, pageSize = 10) => {
  const params = {
    keyword: keyword,
    pageNum: pageNum,
    pageSize: pageSize
  };
  return get('/products/search', params).catch(error => {
    console.error('搜索商品失败:', error);
    // 返回空结果
    return {
      code: 200,
      data: {
        list: [],
        total: 0,
        pageNum: 1,
        pageSize: 10
      }
    };
  });
};

/**
 * 获取热门搜索关键词
 * @returns Promise 热门搜索关键词列表
 */
export const getHotSearchKeywords = () => {
  return get('/hot-search-keywords').catch(error => {
    console.error('获取热门搜索关键词失败:', error);
    return {
      code: 200,
      data: []
    };
  });
};