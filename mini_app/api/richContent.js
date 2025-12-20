import { get } from './request'

/**
 * 获取已发布的富文本内容列表
 * @param {Object} params - 查询参数
 * @param {number} params.page - 页码
 * @param {number} params.page_size - 每页数量
 * @param {string} params.content_type - 内容类型
 */
export function getRichContentList(params) {
  return get('/rich-contents', params)
}

/**
 * 获取已发布的富文本内容详情
 * @param {number} id - 富文本内容ID
 */
export function getRichContentDetail(id) {
  return get(`/rich-contents/${id}`)
}

