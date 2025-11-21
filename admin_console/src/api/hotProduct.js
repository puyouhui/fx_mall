import request from '../utils/request'

// 获取所有热销产品（管理后台）
export function getAllHotProducts() {
  return request({
    url: '/admin/hot-products',
    method: 'get'
  })
}

// 创建热销产品关联
export function createHotProduct(data) {
  return request({
    url: '/admin/hot-products',
    method: 'post',
    data
  })
}

// 更新热销产品关联
export function updateHotProduct(id, data) {
  return request({
    url: `/admin/hot-products/${id}`,
    method: 'put',
    data
  })
}

// 删除热销产品关联
export function deleteHotProduct(id) {
  return request({
    url: `/admin/hot-products/${id}`,
    method: 'delete'
  })
}

// 批量更新热销产品排序
export function updateHotProductSort(items) {
  return request({
    url: '/admin/hot-products/sort',
    method: 'put',
    data: { items }
  })
}

