import request from '../utils/request'

// 获取所有系统设置
export function getSystemSettings() {
  return request({
    url: '/admin/settings',
    method: 'get'
  })
}

// 更新系统设置
export function updateSystemSettings(settings) {
  return request({
    url: '/admin/settings',
    method: 'put',
    data: { settings }
  })
}

// 获取地图设置
export function getMapSettings() {
  return request({
    url: '/admin/settings/map',
    method: 'get'
  })
}

// 更新地图设置
export function updateMapSettings(data) {
  return request({
    url: '/admin/settings/map',
    method: 'put',
    data
  })
}

// 获取WebSocket配置
export function getWebSocketConfig() {
  return request({
    url: '/admin/settings/websocket',
    method: 'get'
  })
}

