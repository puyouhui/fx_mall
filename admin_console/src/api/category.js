import request from '../utils/request'

// 获取分类列表
export function getCategoryList() {
  return request({
    url: '/admin/categories',
    method: 'get'
  })
}

// 创建分类
export function createCategory(data) {
  console.log('原始数据:', data);
  // 移除id和sort字段并将status布尔值转换为整型
  const { id, sort, ...categoryData } = data;
  // 如果status是布尔值，转换为1或0
  if (typeof categoryData.status === 'boolean') {
    console.log('转换前status类型:', typeof categoryData.status, '值:', categoryData.status);
    categoryData.status = categoryData.status ? 1 : 0;
    console.log('转换后status类型:', typeof categoryData.status, '值:', categoryData.status);
  }
  console.log('发送到后端的数据:', categoryData);
  return request({
    url: '/admin/categories',
    method: 'post',
    data: categoryData
  })
}

// 更新分类
export function updateCategory(id, data) {
  console.log('原始数据:', data);
  // 移除排序字段并将status布尔值转换为整型
  const { sort, ...categoryData } = data;
  // 如果status是布尔值，转换为1或0
  if (typeof categoryData.status === 'boolean') {
    console.log('转换前status类型:', typeof categoryData.status, '值:', categoryData.status);
    categoryData.status = categoryData.status ? 1 : 0;
    console.log('转换后status类型:', typeof categoryData.status, '值:', categoryData.status);
  }
  console.log('发送到后端的数据:', categoryData);
  return request({
    url: `/admin/categories/${id}`,
    method: 'put',
    data: categoryData
  })
}

// 删除分类
export function deleteCategory(id) {
  return request({
    url: `/admin/categories/${id}`,
    method: 'delete'
  })
}

// 上传分类图标
export function uploadCategoryImage(file) {
  const formData = new FormData()
  formData.append('file', file)
  
  return request({
    url: '/admin/categories/upload',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}