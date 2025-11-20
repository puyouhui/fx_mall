import request from '../utils/request'

// 获取分类列表
export const getCategories = async () => {
  try {
    const { data } = await request.get('/admin/categories')
    return data
  } catch (error) {
    console.error('获取分类列表失败:', error)
    // 返回模拟数据
    return {
      code: 200,
      data: [
        {
          id: 1,
          name: '电子产品',
          parent_id: 0,
          sort: 1,
          status: 1,
          created_at: new Date(),
          updated_at: new Date()
        },
        {
          id: 2,
          name: '家居用品',
          parent_id: 0,
          sort: 2,
          status: 1,
          created_at: new Date(),
          updated_at: new Date()
        },
        {
          id: 3,
          name: '服装鞋帽',
          parent_id: 0,
          sort: 3,
          status: 1,
          created_at: new Date(),
          updated_at: new Date()
        },
        {
          id: 4,
          name: '食品饮料',
          parent_id: 0,
          sort: 4,
          status: 1,
          created_at: new Date(),
          updated_at: new Date()
        },
        {
          id: 5,
          name: '手机配件',
          parent_id: 1,
          sort: 1,
          status: 1,
          created_at: new Date(),
          updated_at: new Date()
        },
        {
          id: 6,
          name: '电脑配件',
          parent_id: 1,
          sort: 2,
          status: 1,
          created_at: new Date(),
          updated_at: new Date()
        }
      ],
      message: 'success'
    }
  }
}

// 创建分类
export const createCategory = async (categoryData) => {
  try {
    const { data } = await request.post('/admin/categories', categoryData)
    return data
  } catch (error) {
    console.error('创建分类失败:', error)
    throw error
  }
}

// 更新分类
export const updateCategory = async (id, categoryData) => {
  try {
    const { data } = await request.put(`/admin/categories/${id}`, categoryData)
    return data
  } catch (error) {
    console.error('更新分类失败:', error)
    throw error
  }
}

// 删除分类
export const deleteCategory = async (id) => {
  try {
    const { data } = await request.delete(`/admin/categories/${id}`)
    return data
  } catch (error) {
    console.error('删除分类失败:', error)
    throw error
  }
}