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

