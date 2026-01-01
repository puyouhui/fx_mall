import request from '../utils/request'

// 获取图片列表（支持分页和目录过滤）
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

// 上传图片（支持目录分类）
export function uploadImage(formData, category = 'others') {
  // 将category添加到FormData
  formData.append('category', category)
  return request({
    url: '/admin/images/upload',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

