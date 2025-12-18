import request from '../utils/request'

// 获取所有热门搜索关键词
export function getAllHotSearchKeywords() {
  return request({
    url: '/admin/hot-search-keywords',
    method: 'get'
  })
}

// 创建热门搜索关键词
export function createHotSearchKeyword(data) {
  return request({
    url: '/admin/hot-search-keywords',
    method: 'post',
    data
  })
}

// 更新热门搜索关键词
export function updateHotSearchKeyword(id, data) {
  return request({
    url: `/admin/hot-search-keywords/${id}`,
    method: 'put',
    data
  })
}

// 删除热门搜索关键词
export function deleteHotSearchKeyword(id) {
  return request({
    url: `/admin/hot-search-keywords/${id}`,
    method: 'delete'
  })
}


