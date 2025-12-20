import request from '../utils/request'

// 获取供应商商品列表（分页）
export function getProducts(params) {
  return request({
    url: '/products',
    method: 'get',
    params
  })
}

// 获取商品详情（供应商专用API）
export function getProductDetail(id) {
  return request({
    url: `/products/${id}`,
    method: 'get'
  })
}

