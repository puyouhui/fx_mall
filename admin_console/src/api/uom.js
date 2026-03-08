import request from '../utils/request'

// 获取默认「件」单位类别ID
export function getUomDefaultCategory() {
  return request({
    url: '/admin/uom/default-category',
    method: 'get'
  })
}

// 获取单位类别列表（含单位）
export function getUomCategories() {
  return request({
    url: '/admin/uom/categories',
    method: 'get'
  })
}

// 创建单位类别
export function createUomCategory(data) {
  return request({
    url: '/admin/uom/categories',
    method: 'post',
    data
  })
}

// 更新单位类别
export function updateUomCategory(id, data) {
  return request({
    url: `/admin/uom/categories/${id}`,
    method: 'put',
    data
  })
}

// 删除单位类别
export function deleteUomCategory(id) {
  return request({
    url: `/admin/uom/categories/${id}`,
    method: 'delete'
  })
}

// 获取单位列表（按类别）
export function getUomUnits(categoryId) {
  return request({
    url: '/admin/uom/units',
    method: 'get',
    params: { category_id: categoryId }
  })
}

// 创建单位
export function createUomUnit(data) {
  return request({
    url: '/admin/uom/units',
    method: 'post',
    data
  })
}

// 更新单位
export function updateUomUnit(id, data) {
  return request({
    url: `/admin/uom/units/${id}`,
    method: 'put',
    data
  })
}

// 删除单位
export function deleteUomUnit(id) {
  return request({
    url: `/admin/uom/units/${id}`,
    method: 'delete'
  })
}
