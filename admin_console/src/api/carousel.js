import request from '../utils/request'

// 获取轮播图列表
export function getCarouselList() {
  return request({
    url: '/admin/carousels',
    method: 'get'
  })
}

// 获取轮播图详情
export function getCarouselDetail(id) {
  return request({
    url: `/admin/carousels/${id}`,
    method: 'get'
  })
}

// 创建轮播图
export function createCarousel(data) {
  return request({
    url: '/admin/carousels',
    method: 'post',
    data
  })
}

// 更新轮播图
export function updateCarousel(id, data) {
  return request({
    url: `/admin/carousels/${id}`,
    method: 'put',
    data
  })
}

// 删除轮播图
export function deleteCarousel(id) {
  return request({
    url: `/admin/carousels/${id}`,
    method: 'delete'
  })
}

// 上传轮播图图片
export function uploadCarouselImage(data) {
  return request({
    url: '/admin/carousels/upload',
    method: 'post',
    headers: {
      'Content-Type': 'multipart/form-data'
    },
    data
  })
}