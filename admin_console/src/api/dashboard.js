import request from '../utils/request'

// 获取仪表盘数据
export const getDashboardData = async () => {
  try {
    // 实际项目中这里应该调用真实的API
    // const { data } = await request.get('/dashboard/data')
    
    // 由于是模拟环境，返回模拟数据
    return {
      code: 200,
      data: {
        salesAmount: 128500,
        ordersCount: 2560,
        usersCount: 18900,
        productsCount: 1240,
        salesTrend: [12000, 15000, 13000, 18000, 16000, 22000, 25000, 28000, 23000, 30000, 32000, 35000],
        topProducts: [
          {
            id: '1',
            name: '智能手机Pro Max',
            sales: 12500,
            orders: 189
          },
          {
            id: '2',
            name: '无线蓝牙耳机',
            sales: 8900,
            orders: 256
          },
          {
            id: '3',
            name: '智能手表',
            sales: 6500,
            orders: 142
          },
          {
            id: '4',
            name: '超薄笔记本电脑',
            sales: 5800,
            orders: 89
          },
          {
            id: '5',
            name: '智能家居套装',
            sales: 4200,
            orders: 105
          }
        ]
      },
      message: 'success'
    }
  } catch (error) {
    console.error('获取仪表盘数据失败:', error)
    // 返回默认的模拟数据，以防API调用失败
    return {
      code: 200,
      data: {
        salesAmount: 128500,
        ordersCount: 2560,
        usersCount: 18900,
        productsCount: 1240,
        salesTrend: [12000, 15000, 13000, 18000, 16000, 22000, 25000, 28000, 23000, 30000, 32000, 35000],
        topProducts: [
          {
            id: '1',
            name: '智能手机Pro Max',
            sales: 12500,
            orders: 189
          },
          {
            id: '2',
            name: '无线蓝牙耳机',
            sales: 8900,
            orders: 256
          },
          {
            id: '3',
            name: '智能手表',
            sales: 6500,
            orders: 142
          },
          {
            id: '4',
            name: '超薄笔记本电脑',
            sales: 5800,
            orders: 89
          },
          {
            id: '5',
            name: '智能家居套装',
            sales: 4200,
            orders: 105
          }
        ]
      },
      message: 'success'
    }
  }
}