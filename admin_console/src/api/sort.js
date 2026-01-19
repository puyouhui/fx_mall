// sort.js - 排序管理API

import request from '../utils/request'

/**
 * 批量更新分类排序
 * @param {Array} items - 排序项数组，每个元素包含 {id, sort}
 * @returns Promise
 */
export function batchUpdateCategorySort(items) {
  return request({
    url: '/admin/categories/sort',
    method: 'put',
    data: { items }
  })
}

/**
 * 批量更新商品排序
 * @param {Array} items - 排序项数组，每个元素包含 {id, sort}
 * @returns Promise
 */
export function batchUpdateProductSort(items) {
  return request({
    url: '/admin/products/sort',
    method: 'put',
    data: { items }
  })
}

/**
 * 获取所有精选商品（用于排序管理）
 * @returns Promise
 */
export function getAllSpecialProducts() {
  return request({
    url: '/admin/special-products',
    method: 'get'
  })
}

/**
 * 批量更新精选商品排序
 * @param {Array} items - 排序项数组，每个元素包含 {id, special_sort}
 * @returns Promise
 */
export function batchUpdateSpecialProductSort(items) {
  return request({
    url: '/admin/special-products/sort',
    method: 'put',
    data: { items }
  })
}
