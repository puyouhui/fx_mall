import request from '../utils/request'

// 获取商品列表
export const getProducts = async (params = {}) => {
  try {
    const { data } = await request.get('/admin/products', { params })
    return data
  } catch (error) {
    console.error('获取商品列表失败:', error)
    // 返回模拟数据
    return {
      code: 200,
      data: {
        list: [
          {
            id: 1,
            name: '智能手机',
            description: '高性能智能手机，配备高清摄像头',
            original_price: 3299.99,
            price: 2999.99,
            category_id: 1,
            is_special: true,
            images: ['/static/test/product1-1.jpg'],
            specs: [{ name: '品牌', value: '知名品牌' }],
            status: 1,
            created_at: new Date(),
            updated_at: new Date()
          },
          {
            id: 2,
            name: '笔记本电脑',
            description: '高性能笔记本电脑',
            original_price: 6299.99,
            price: 5999.99,
            category_id: 1,
            is_special: true,
            images: ['/static/test/product2-1.jpg'],
            specs: [{ name: '型号', value: 'Pro 14' }],
            status: 1,
            created_at: new Date(),
            updated_at: new Date()
          },
          {
            id: 3,
            name: '智能手表',
            description: '多功能智能手表',
            original_price: 999.99,
            price: 899.99,
            category_id: 1,
            is_special: true,
            images: ['/static/test/product3-1.jpg'],
            specs: [{ name: '颜色', value: '黑色' }],
            status: 1,
            created_at: new Date(),
            updated_at: new Date()
          },
          {
            id: 4,
            name: '时尚台灯',
            description: '护眼台灯',
            original_price: 149.99,
            price: 129.99,
            category_id: 2,
            is_special: true,
            images: ['/static/test/product4-1.jpg'],
            specs: [{ name: '材质', value: '金属' }],
            status: 1,
            created_at: new Date(),
            updated_at: new Date()
          },
          {
            id: 5,
            name: '休闲T恤',
            description: '纯棉透气T恤',
            original_price: 109.99,
            price: 89.99,
            category_id: 3,
            is_special: true,
            images: ['/static/test/product5-1.jpg'],
            specs: [{ name: '材质', value: '纯棉' }],
            status: 1,
            created_at: new Date(),
            updated_at: new Date()
          }
        ],
        total: 24,
        pageNum: params.pageNum || 1,
        pageSize: params.pageSize || 10
      },
      message: 'success'
    }
  }
}

// 创建商品
export const createProduct = async (productData) => {
  try {
    const { data } = await request.post('/admin/products', productData)
    return data
  } catch (error) {
    console.error('创建商品失败:', error)
    throw error
  }
}

// 更新商品
export const updateProduct = async (id, productData) => {
  try {
    const { data } = await request.put(`/admin/products/${id}`, productData)
    return data
  } catch (error) {
    console.error('更新商品失败:', error)
    throw error
  }
}

// 删除商品
export const deleteProduct = async (id) => {
  try {
    const { data } = await request.delete(`/admin/products/${id}`)
    return data
  } catch (error) {
    console.error('删除商品失败:', error)
    throw error
  }
}