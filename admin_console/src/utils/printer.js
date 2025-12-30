// 打印机配置管理工具

const PRINTER_ADDRESS_KEY = 'printer_address'
const DEFAULT_PRINTER_ADDRESS = 'http://198.18.0.1:17521'

/**
 * 判断是否为本地地址
 * @param {string} address 地址
 * @returns {boolean} 是否为本地地址
 */
function isLocalAddress(address) {
  try {
    const url = new URL(address)
    const hostname = url.hostname.toLowerCase()
    
    // 检查是否为本地地址
    return (
      hostname === 'localhost' ||
      hostname === '127.0.0.1' ||
      hostname.startsWith('192.168.') ||
      hostname.startsWith('10.') ||
      hostname.startsWith('172.16.') ||
      hostname.startsWith('172.17.') ||
      hostname.startsWith('172.18.') ||
      hostname.startsWith('172.19.') ||
      hostname.startsWith('172.20.') ||
      hostname.startsWith('172.21.') ||
      hostname.startsWith('172.22.') ||
      hostname.startsWith('172.23.') ||
      hostname.startsWith('172.24.') ||
      hostname.startsWith('172.25.') ||
      hostname.startsWith('172.26.') ||
      hostname.startsWith('172.27.') ||
      hostname.startsWith('172.28.') ||
      hostname.startsWith('172.29.') ||
      hostname.startsWith('172.30.') ||
      hostname.startsWith('172.31.') ||
      hostname.startsWith('198.18.') ||
      hostname.startsWith('198.19.')
    )
  } catch (e) {
    return false
  }
}

/**
 * 获取打印机地址
 * 支持两种方式：
 * 1. 直接连接本地打印机客户端（例如：http://198.18.0.1:17521）
 * 2. 通过中转服务连接（推荐，例如：https://mall.sscchh.com:17521）
 * 
 * 如果使用 HTTPS 页面，建议使用中转服务（node-hiprint-transit）来解决混合内容问题
 * @returns {string} 打印机地址
 */
export function getPrinterAddress() {
  let address = localStorage.getItem(PRINTER_ADDRESS_KEY) || DEFAULT_PRINTER_ADDRESS
  
  // 如果地址已经是 https://，直接返回（这是中转服务地址）
  if (address.startsWith('https://')) {
    return address
  }
  
  // 如果当前页面使用 HTTPS，且打印机地址使用 http://
  if (window.location.protocol === 'https:' && address.startsWith('http://')) {
    // 如果是本地地址，HTTPS 页面无法直接连接
    // 建议使用中转服务，但如果用户配置的是中转服务地址（https://），这里不会执行
    // 如果用户配置的是本地地址，这里会尝试转换为 https://，但通常本地客户端不支持
    // 更好的方案是提示用户使用中转服务
    const isLocal = isLocalAddress(address)
    if (isLocal) {
      // 本地地址在 HTTPS 页面下无法连接，保持原地址（连接会失败，但可以提示用户使用中转服务）
      console.warn('⚠️ HTTPS 页面无法直接连接本地打印机，建议使用中转服务（node-hiprint-transit）')
      return address
    } else {
      // 非本地地址，尝试转换为 https://
      address = address.replace('http://', 'https://')
    }
  }
  
  return address
}

/**
 * 检查是否可以在当前环境下连接打印机
 * @returns {object} { canConnect: boolean, reason?: string, suggestion?: string }
 */
export function canConnectToPrinter() {
  const address = localStorage.getItem(PRINTER_ADDRESS_KEY) || DEFAULT_PRINTER_ADDRESS
  const isHttps = window.location.protocol === 'https:'
  const isLocal = isLocalAddress(address)
  
  if (isHttps && isLocal && address.startsWith('http://')) {
    return {
      canConnect: false,
      reason: 'HTTPS 页面无法连接到本地 HTTP 打印机。',
      suggestion: '建议使用中转服务（node-hiprint-transit）：1) 在服务器上部署中转服务并配置 HTTPS；2) 在系统设置中配置中转服务地址（例如：https://mall.sscchh.com:17521）；3) 本地打印机客户端连接到中转服务。'
    }
  }
  
  return { canConnect: true }
}

/**
 * 保存打印机地址
 * @param {string} address 打印机地址
 */
export function setPrinterAddress(address) {
  if (address && address.trim()) {
    localStorage.setItem(PRINTER_ADDRESS_KEY, address.trim())
  } else {
    localStorage.removeItem(PRINTER_ADDRESS_KEY)
  }
}

/**
 * 获取默认打印机地址
 * @returns {string} 默认打印机地址
 */
export function getDefaultPrinterAddress() {
  return DEFAULT_PRINTER_ADDRESS
}

/**
 * 判断当前是否为线上环境（使用中转服务）
 * @returns {boolean} 是否为线上环境
 */
export function isOnlineEnvironment() {
  const address = localStorage.getItem(PRINTER_ADDRESS_KEY) || DEFAULT_PRINTER_ADDRESS
  // 如果地址是 https://，说明使用的是中转服务（线上环境）
  return address.startsWith('https://')
}

/**
 * 判断当前是否为本地环境（直接连接）
 * @returns {boolean} 是否为本地环境
 */
export function isLocalEnvironment() {
  return !isOnlineEnvironment()
}

// 存储客户端列表（用于线上环境）
let cachedClients = null
let clientListPromise = null

/**
 * 获取客户端列表（仅线上环境需要）
 * @param {object} hiprintInstance hiprint 实例（从调用处传入）
 * @returns {Promise<string|null>} 返回第一个客户端的 clientId，如果没有则返回 null
 */
export async function getFirstClientId(hiprintInstance) {
  const isOnline = isOnlineEnvironment()
  if (!isOnline) {
    // 本地环境不需要 client
    return null
  }

  // 如果已经有缓存的客户端列表，直接返回第一个
  if (cachedClients && Object.keys(cachedClients).length > 0) {
    const firstClientId = Object.keys(cachedClients)[0]
    console.log('使用缓存的客户端:', firstClientId)
    return firstClientId
  }

  // 如果正在获取客户端列表，等待完成
  if (clientListPromise) {
    await clientListPromise
    if (cachedClients && Object.keys(cachedClients).length > 0) {
      return Object.keys(cachedClients)[0]
    }
    return null
  }

  // 开始获取客户端列表
  clientListPromise = new Promise((resolve) => {
    try {
      // 检查 hiprint 是否已初始化
      if (!hiprintInstance || !hiprintInstance.hiwebSocket || !hiprintInstance.hiwebSocket.socket) {
        console.warn('hiprint 未初始化，无法获取客户端列表')
        resolve(null)
        return
      }

      const socket = hiprintInstance.hiwebSocket.socket
      
      // 监听客户端列表返回
      const clientsHandler = (clients) => {
        console.log('📋 获取到客户端列表:', clients)
        cachedClients = clients
        
        // 获取第一个客户端 ID
        const firstClientId = Object.keys(clients)[0]
        if (firstClientId) {
          console.log('✅ 选择第一个客户端:', firstClientId)
          socket.off('clients', clientsHandler)
          resolve(firstClientId)
        } else {
          console.warn('⚠️ 没有可用的客户端')
          socket.off('clients', clientsHandler)
          resolve(null)
        }
      }

      // 注册监听器
      socket.on('clients', clientsHandler)

      // 请求客户端列表
      console.log('📡 请求客户端列表...')
      socket.emit('getClients')

      // 设置超时（5秒）
      setTimeout(() => {
        socket.off('clients', clientsHandler)
        if (!cachedClients) {
          console.warn('⚠️ 获取客户端列表超时')
          resolve(null)
        }
      }, 5000)
    } catch (error) {
      console.error('获取客户端列表失败:', error)
      resolve(null)
    }
  })

  return await clientListPromise
}

/**
 * 清除客户端列表缓存（当连接断开或重新连接时调用）
 */
export function clearClientCache() {
  cachedClients = null
  clientListPromise = null
  console.log('🗑️ 已清除客户端列表缓存')
}

/**
 * 获取打印选项（根据环境自动调整）
 * @param {object} options 打印选项（如 printer, client 等）
 * @param {object} hiprintInstance hiprint 实例（可选，如果未提供则尝试从全局获取）
 * @returns {Promise<object>} 调整后的打印选项
 */
export async function getPrintOptions(options = {}, hiprintInstance = null) {
  const isOnline = isOnlineEnvironment()
  const address = localStorage.getItem(PRINTER_ADDRESS_KEY) || DEFAULT_PRINTER_ADDRESS
  
  if (isOnline) {
    // 线上环境（使用中转服务）
    console.log('📡 使用线上中转服务打印，地址:', address)
    
    // 根据 README.md，使用中转服务时，打印需要指定 client 参数
    // socket.to(options.client).emit("news", { ...options, replyId: socket.id })
    
    // 如果没有指定 client，尝试获取第一个客户端
    if (!options.client) {
      // 尝试从全局获取 hiprint（如果未传入）
      let hiprint = hiprintInstance
      if (!hiprint && typeof window !== 'undefined') {
        // 尝试从全局导入
        try {
          const { hiprint: hiprintGlobal } = await import('vue-plugin-hiprint')
          hiprint = hiprintGlobal
        } catch (e) {
          console.warn('无法获取 hiprint 实例')
        }
      }
      
      if (hiprint) {
        const firstClientId = await getFirstClientId(hiprint)
        if (firstClientId) {
          options.client = firstClientId
          console.log('✅ 已自动选择客户端:', firstClientId)
        } else {
          console.warn('⚠️ 无法获取客户端，打印可能会失败')
        }
      } else {
        console.warn('⚠️ 无法获取 hiprint 实例，无法自动选择客户端')
      }
    }
    
    return {
      ...options,
    }
  } else {
    // 本地环境（直接连接）
    console.log('🖨️ 使用本地打印机打印，地址:', address)
    return options
  }
}

