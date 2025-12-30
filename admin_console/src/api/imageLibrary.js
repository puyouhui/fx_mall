import request from '../utils/request'

// 获取图片列表（支持分页）
export function getImageList(params = {}) {
  return request({
    url: '/admin/images',
    method: 'get',
    params
  })
}

// 批量删除图片
export function batchDeleteImages(imageUrls) {
  return request({
    url: '/admin/images/batch',
    method: 'delete',
    data: {
      imageUrls
    }
  })
}

// 上传图片（复用商品图片上传接口）
export function uploadImage(formData) {
  return request({
    url: '/admin/products/upload',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

