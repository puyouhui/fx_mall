import request from '../utils/request'

/**
 * 获取富文本内容列表
 * @param {Object} params - 查询参数
 * @param {number} params.page - 页码
 * @param {number} params.page_size - 每页数量
 * @param {string} params.content_type - 内容类型
 * @param {string} params.status - 状态
 */
export function getRichContentList(params) {
  return request({
    url: '/admin/rich-contents',
    method: 'get',
    params
  })
}

/**
 * 获取富文本内容详情
 * @param {number} id - 富文本内容ID
 */
export function getRichContent(id) {
  return request({
    url: `/admin/rich-contents/${id}`,
    method: 'get'
  })
}

/**
 * 创建富文本内容
 * @param {Object} data - 富文本内容数据
 * @param {string} data.title - 标题
 * @param {string} data.content - HTML内容
 * @param {string} data.content_type - 内容类型
 */
export function createRichContent(data) {
  return request({
    url: '/admin/rich-contents',
    method: 'post',
    data
  })
}

/**
 * 更新富文本内容
 * @param {number} id - 富文本内容ID
 * @param {Object} data - 更新数据
 */
export function updateRichContent(id, data) {
  return request({
    url: `/admin/rich-contents/${id}`,
    method: 'put',
    data
  })
}

/**
 * 发布富文本内容
 * @param {number} id - 富文本内容ID
 */
export function publishRichContent(id) {
  return request({
    url: `/admin/rich-contents/${id}/publish`,
    method: 'put'
  })
}

/**
 * 归档富文本内容
 * @param {number} id - 富文本内容ID
 */
export function archiveRichContent(id) {
  return request({
    url: `/admin/rich-contents/${id}/archive`,
    method: 'put'
  })
}

/**
 * 删除富文本内容
 * @param {number} id - 富文本内容ID
 */
export function deleteRichContent(id) {
  return request({
    url: `/admin/rich-contents/${id}`,
    method: 'delete'
  })
}

/**
 * 上传图片（富文本编辑器使用）
 * @param {FormData} formData - 包含图片文件的FormData
 */
export function uploadImage(formData) {
  // 富文本编辑器的图片存到rich-content目录
  formData.append('category', 'rich-content')
  return request({
    url: '/admin/images/upload', // 使用支持目录分类的上传接口
    method: 'post',
    headers: {
      'Content-Type': 'multipart/form-data'
    },
    data: formData
  })
}

