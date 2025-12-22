// 打印机配置管理工具

const PRINTER_ADDRESS_KEY = 'printer_address'
const DEFAULT_PRINTER_ADDRESS = 'http://198.18.0.1:17521'

/**
 * 获取打印机地址
 * @returns {string} 打印机地址
 */
export function getPrinterAddress() {
  const address = localStorage.getItem(PRINTER_ADDRESS_KEY)
  return address || DEFAULT_PRINTER_ADDRESS
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

