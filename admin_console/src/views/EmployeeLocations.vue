<template>
  <div class="employee-locations-page">
    <el-card class="locations-card" shadow="never">
      <template #header>
        <div class="card-header">
          <div class="title">
            <span class="main">员工位置</span>
            <span class="sub">实时查看配送员位置信息</span>
          </div>
          <div class="actions">
            <el-tag :type="wsConnected ? 'success' : 'danger'">
              {{ wsConnected ? '已连接' : '未连接' }}
            </el-tag>
            <el-button type="primary" @click="refreshLocations" :loading="loading">
              刷新位置
            </el-button>
          </div>
        </div>
      </template>

      <div class="locations-container">
        <!-- 左侧员工列表 -->
        <div class="locations-list">
          <el-table
            v-loading="loading"
            :data="locations"
            border
            stripe
            height="100%"
            @row-click="handleRowClick"
            highlight-current-row
          >
            <el-table-column prop="employee_code" label="员工码" min-width="100" />
            <el-table-column prop="name" label="姓名" min-width="100">
              <template #default="scope">
                {{ scope.row.name || '-' }}
              </template>
            </el-table-column>
            <el-table-column prop="phone" label="手机号" min-width="130" />
            <el-table-column label="位置" min-width="150">
              <template #default="scope">
                <div v-if="scope.row.latitude && scope.row.longitude">
                  {{ scope.row.latitude.toFixed(6) }}, {{ scope.row.longitude.toFixed(6) }}
                </div>
                <span v-else class="text-muted">暂无位置</span>
              </template>
            </el-table-column>
            <el-table-column label="更新时间" min-width="180">
              <template #default="scope">
                {{ formatTime(scope.row.updated_at) }}
              </template>
            </el-table-column>
            <el-table-column label="状态" min-width="100">
              <template #default="scope">
                <el-tag :type="isLocationRecent(scope.row.updated_at) ? 'success' : 'danger'">
                  {{ isLocationRecent(scope.row.updated_at) ? '在线' : '离线' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <!-- 右侧地图 -->
        <div class="locations-map">
          <div id="map-container" style="width: 100%; height: 100%;"></div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { getEmployeeLocations } from '../api/employees'
import { getMapSettings, getWebSocketConfig } from '../api/settings'
import { ElMessage } from 'element-plus'

const loading = ref(false)
const locations = ref([])
const wsConnected = ref(false)
const amapKey = ref('')
const websocketUrl = ref('')
const mapReady = ref(false) // 地图是否已完全加载
let ws = null
let map = null
let markers = [] // 存储 { marker, location } 对象数组

// 初始化地图
const initMap = () => {
  // 使用高德地图API
  if (typeof AMap !== 'undefined') {
    try {
      // 检查容器元素是否存在
      const container = document.getElementById('map-container')
      if (!container) {
        console.error('地图容器元素不存在')
        return
      }

      map = new AMap.Map('map-container', {
        zoom: 10, // 缩小地图比例，显示更大范围
        center: [102.753865,24.974177], // 默认北京
        viewMode: '3D',
        pitch: 30,
      })

      // 等待地图加载完成
      map.on('complete', function() {
        console.log('地图加载完成')
        mapReady.value = true // 标记地图已就绪
        
        // 异步加载地图控件插件
        AMap.plugin(['AMap.Scale', 'AMap.ToolBar'], function() {
          try {
            // 添加比例尺控件
            if (map && typeof map.addControl === 'function') {
              map.addControl(new AMap.Scale())
              // 添加工具条控件
              map.addControl(new AMap.ToolBar({
                position: 'RT' // 右上角
              }))
            }
          } catch (e) {
            console.error('添加地图控件失败:', e)
          }
        })
      })
    } catch (e) {
      console.error('初始化地图失败:', e)
      map = null
      mapReady.value = false
    }
  } else {
    console.error('高德地图API未加载')
    map = null
    mapReady.value = false
  }
}

// 加载高德地图API
const loadAMapScript = () => {
  return new Promise((resolve, reject) => {
    if (typeof AMap !== 'undefined') {
      resolve()
      return
    }

    if (!amapKey.value) {
      reject(new Error('高德地图Key未配置'))
      return
    }

    const script = document.createElement('script')
    script.src = `https://webapi.amap.com/maps?v=2.0&key=${amapKey.value}`
    script.onload = resolve
    script.onerror = reject
    document.head.appendChild(script)
  })
}

// 连接WebSocket
const connectWebSocket = () => {
  const token = localStorage.getItem('token')
  if (!token) {
    ElMessage.error('未登录，无法连接WebSocket')
    return
  }

  if (!websocketUrl.value) {
    ElMessage.error('WebSocket URL未配置')
    return
  }

  // 构建WebSocket URL（从配置中获取，并添加token参数）
  const wsUrl = `${websocketUrl.value}?token=${encodeURIComponent(token)}`

  try {
    ws = new WebSocket(wsUrl)

    ws.onopen = () => {
      console.log('WebSocket连接已建立')
      wsConnected.value = true
    }

    ws.onmessage = (event) => {
      // 检查是否是二进制消息（Pong响应）
      if (event.data instanceof Blob || event.data instanceof ArrayBuffer) {
        // 这是Pong响应，忽略
        return
      }

      try {
        const data = JSON.parse(event.data)
        if (data.type === 'initial_locations') {
          locations.value = data.locations || []
          // 延迟更新地图标记，确保地图已初始化
          setTimeout(() => {
            updateMapMarkers()
          }, 100)
        } else if (data.type === 'location_update') {
          // 更新位置
          const index = locations.value.findIndex(
            loc => loc.employee_id === data.location.employee_id
          )
          if (index >= 0) {
            locations.value[index] = data.location
          } else {
            locations.value.push(data.location)
          }
          // 延迟更新地图标记，确保地图已初始化
          setTimeout(() => {
            updateMapMarkers()
          }, 100)
        }
      } catch (e) {
        // 如果不是JSON，可能是Pong响应，忽略
        console.log('收到非JSON消息，可能是心跳响应')
      }
    }

    ws.onerror = (error) => {
      console.error('WebSocket错误:', error)
      console.error('WebSocket URL:', wsUrl)
      wsConnected.value = false
      // 不显示错误提示，因为onclose会处理重连
    }

    ws.onclose = (event) => {
      console.log('WebSocket连接已关闭', event.code, event.reason)
      wsConnected.value = false
      
      // 如果不是正常关闭（代码1000），尝试重连
      if (event.code !== 1000) {
        console.log('5秒后尝试重连WebSocket...')
        setTimeout(() => {
          if (wsConnected.value === false) {
            connectWebSocket()
          }
        }, 5000)
      }
    }
  } catch (e) {
    console.error('建立WebSocket连接失败:', e)
    wsConnected.value = false
    ElMessage.error('WebSocket连接失败: ' + e.message)
  }
}

// 更新地图标记
const updateMapMarkers = () => {
  // 检查地图是否已完全加载
  if (!mapReady.value) {
    console.log('地图未就绪，跳过标记更新')
    return
  }

  // 检查地图是否已初始化且可用
  if (!map) {
    console.log('地图对象不存在，跳过标记更新')
    return
  }

  // 检查map是否是有效的AMap.Map实例
  if (typeof map.add !== 'function' || typeof map.remove !== 'function') {
    console.log('地图对象无效，跳过标记更新', map)
    mapReady.value = false // 重置就绪状态
    return
  }

  // 检查AMap是否可用
  if (typeof AMap === 'undefined' || typeof AMap.Marker !== 'function') {
    console.log('高德地图API未加载，跳过标记更新')
    return
  }

  // 创建标记图标的辅助函数（使用Canvas，确保绘制在中心）
  const createMarkerIcon = (color, isOnline) => {
    const size = 40 // Canvas尺寸
    const canvas = document.createElement('canvas')
    canvas.width = size
    canvas.height = size
    const ctx = canvas.getContext('2d')
    
    // 确保绘制在Canvas正中心
    const centerX = size / 2
    const centerY = size / 2
    
    // 半径设置（留出足够边距）
    const outerRadius = 15
    const innerRadius = 13
    const centerRadius = 5
    
    // 启用抗锯齿
    ctx.imageSmoothingEnabled = true
    ctx.imageSmoothingQuality = 'high'
    
    // 清除Canvas（确保透明背景）
    ctx.clearRect(0, 0, size, size)
    
    // 绘制外圈（白色边框）
    ctx.beginPath()
    ctx.arc(centerX, centerY, outerRadius, 0, Math.PI * 2)
    ctx.fillStyle = '#ffffff'
    ctx.fill()
    
    // 绘制内圈（颜色）
    ctx.beginPath()
    ctx.arc(centerX, centerY, innerRadius, 0, Math.PI * 2)
    ctx.fillStyle = color
    ctx.fill()
    
    // 在线状态显示白色小点
    if (isOnline) {
      ctx.beginPath()
      ctx.arc(centerX, centerY, centerRadius, 0, Math.PI * 2)
      ctx.fillStyle = '#ffffff'
      ctx.fill()
    }
    
    return canvas.toDataURL('image/png')
  }

  try {
    // 清除旧标记
    markers.forEach(item => {
      try {
        if (item.marker && map && typeof map.remove === 'function') {
          map.remove(item.marker)
        }
      } catch (e) {
        console.error('移除标记失败:', e)
      }
    })
    markers = []

    // 添加新标记
    locations.value.forEach(location => {
      if (location.latitude && location.longitude) {
        try {
          // 再次检查map是否有效
          if (!map || typeof map.add !== 'function') {
            console.log('地图对象在添加标记时无效，跳过')
            return
          }

          // 判断是否在线（5分钟内）
          const isOnline = isLocationRecent(location.updated_at)
          
          // 根据在线状态设置颜色
          const markerColor = isOnline ? '#52c41a' : '#ff4d4f' // 绿色在线，红色离线
          
          // 创建自定义图标
          const iconSize = 40 // 与Canvas尺寸一致
          const iconUrl = createMarkerIcon(markerColor, isOnline)
          
          const icon = new AMap.Icon({
            size: new AMap.Size(40, 40),
            image: iconUrl,
            imageSize: new AMap.Size(iconSize, iconSize),
            // 偏移量：负值表示向左上偏移，使图标中心对准坐标点
            // imageOffset: new AMap.Pixel(-iconSize / 2, -iconSize / 2)
          })

          // 创建标签内容（优化样式，移除蓝色边框）
          const statusText = isOnline ? '在线' : '离线'
          const statusIcon = isOnline ? '●' : '●'
          const bgColor = isOnline ? '#52c41a' : '#ff4d4f'
          
          const labelContent = `
            <div style="
              background: ${bgColor};
              color: #fff;
              padding: 6px 10px;
              border-radius: 6px;
              font-size: 12px;
              font-weight: 500;
              white-space: nowrap;
              box-shadow: 0 2px 8px rgba(0,0,0,0.15);
              border: 2px solid #fff;
              display: inline-flex;
              align-items: center;
              gap: 6px;
              line-height: 1.4;
              margin: 0;
              padding: 6px 10px;
            ">
              <span style="font-size: 10px; opacity: 0.9;">${statusIcon}</span>
              <span>${location.name || location.employee_code}</span>
              <span style="font-size: 11px; opacity: 0.85; font-weight: 400;">${statusText}</span>
            </div>
          `

          const marker = new AMap.Marker({
            position: [location.longitude, location.latitude],
            title: `${location.name || location.employee_code} (${isOnline ? '在线' : '离线'})`,
            icon: icon,
            label: {
              content: labelContent,
              direction: 'right',
              offset: new AMap.Pixel(0, 0),
            },
            zIndex: isOnline ? 100 : 50, // 在线标记在上层
          })
          
          // 等待marker添加到地图后，移除label的默认边框
          map.add(marker)
          // 使用nextTick确保DOM已更新
          setTimeout(() => {
            try {
              const labelEl = marker.getLabel()
              if (labelEl) {
                // 获取label的DOM元素
                let el = null
                if (labelEl.element) {
                  el = labelEl.element
                } else if (labelEl.getContent && typeof labelEl.getContent === 'function') {
                  const content = labelEl.getContent()
                  if (content && content.parentElement) {
                    el = content.parentElement
                  }
                }
                
                if (el) {
                  // 移除所有边框和背景样式
                  el.style.cssText = 'border: none !important; background: transparent !important; padding: 0 !important; margin: 0 !important; outline: none !important; box-shadow: none !important;'
                  // 查找所有子元素，移除可能的边框
                  const children = el.querySelectorAll('*')
                  children.forEach(child => {
                    child.style.border = 'none'
                    child.style.outline = 'none'
                  })
                }
              }
            } catch (e) {
              console.error('移除label边框失败:', e)
            }
          }, 100)
          
          // 确保map.add方法存在（已在上面调用）
          markers.push({ marker, location }) // 保存marker和location的关联
        } catch (e) {
          console.error('添加标记失败:', e, location)
        }
      }
    })

    // 如果有标记，调整地图视野（优先使用在线员工的标记）
    if (markers.length > 0 && map && typeof map.setFitView === 'function') {
      try {
        // 优先使用在线员工的标记来调整视野
        const onlineMarkers = markers
          .filter(item => isLocationRecent(item.location.updated_at))
          .map(item => item.marker)
        
        const allMarkers = markers.map(item => item.marker)
        const markersToFit = onlineMarkers.length > 0 ? onlineMarkers : allMarkers
        map.setFitView(markersToFit, false, [50, 50, 50, 50], 12) // 最大缩放级别12，避免放得太大
      } catch (e) {
        console.error('调整地图视野失败:', e)
      }
    }
  } catch (e) {
    console.error('更新地图标记失败:', e)
  }
}

// 刷新位置
const refreshLocations = async (showError = true) => {
  loading.value = true
  try {
    const response = await getEmployeeLocations()
    console.log('获取位置信息响应:', response)
    
    // 检查响应格式（request拦截器已经返回了response.data）
    if (response && response.code === 200) {
      locations.value = response.data || []
      updateMapMarkers()
      // 只有在手动刷新时才显示成功提示
      if (showError) {
        if (locations.value.length === 0) {
          ElMessage.info('暂无员工位置信息')
        } else {
          ElMessage.success(`位置信息已刷新，共${locations.value.length}个员工`)
        }
      }
    } else {
      // 只有在手动刷新或明确要求时才显示错误
      if (showError) {
        ElMessage.error(response?.message || '获取位置信息失败')
      } else {
        console.log('获取位置信息失败（静默）:', response?.message)
      }
    }
  } catch (error) {
    console.error('获取位置信息失败:', error)
    console.error('错误详情:', error.response || error)
    // 只有在手动刷新或明确要求时才显示错误
    if (showError) {
      const errorMsg = error.response?.data?.message || error.message || '获取位置信息失败'
      ElMessage.error(errorMsg)
    }
  } finally {
    loading.value = false
  }
}

// 点击行
const handleRowClick = (row) => {
  if (row.latitude && row.longitude && map) {
    map.setCenter([row.longitude, row.latitude])
    map.setZoom(12) // 降低缩放级别，显示更大范围
  }
}

// 格式化时间
const formatTime = (timeStr) => {
  if (!timeStr) return '-'
  const date = new Date(timeStr)
  const now = new Date()
  const diff = now - date
  const minutes = Math.floor(diff / 60000)
  
  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}小时前`
  return date.toLocaleString('zh-CN')
}

// 判断位置是否最近（5分钟内）
const isLocationRecent = (timeStr) => {
  if (!timeStr) return false
  const date = new Date(timeStr)
  const now = new Date()
  const diff = now - date
  return diff < 5 * 60 * 1000 // 5分钟
}

// 加载配置
const loadConfig = async () => {
  try {
    // 获取地图设置
    const mapResponse = await getMapSettings()
    if (mapResponse.code === 200 && mapResponse.data) {
      amapKey.value = mapResponse.data.amap_key || ''
      if (!amapKey.value) {
        ElMessage.warning('高德地图Key未配置，请在系统设置中配置')
      }
    }

    // 获取WebSocket配置
    const wsResponse = await getWebSocketConfig()
    if (wsResponse.code === 200 && wsResponse.data) {
      let url = wsResponse.data.admin_location_url || ''
      if (url) {
        // 确保URL是ws://或wss://格式
        if (!url.startsWith('ws://') && !url.startsWith('wss://')) {
          // 将http/https转换为ws/wss
          if (url.startsWith('http://')) {
            url = url.replace('http://', 'ws://')
          } else if (url.startsWith('https://')) {
            url = url.replace('https://', 'wss://')
          }
        }
        websocketUrl.value = url
        console.log('WebSocket URL:', websocketUrl.value)
      } else {
        ElMessage.warning('WebSocket URL未配置')
      }
    }
  } catch (error) {
    console.error('加载配置失败:', error)
    ElMessage.error('加载配置失败')
  }
}

onMounted(async () => {
  // 先加载配置
  await loadConfig()

  // 加载地图API
  try {
    if (amapKey.value) {
      await loadAMapScript()
      // 等待一小段时间确保地图API完全加载
      await new Promise(resolve => setTimeout(resolve, 200))
      initMap()
      // 等待地图初始化完成
      await new Promise(resolve => setTimeout(resolve, 300))
    } else {
      ElMessage.warning('高德地图Key未配置，无法显示地图')
    }
  } catch (e) {
    console.error('加载地图API失败:', e)
    ElMessage.warning('地图加载失败，将使用列表模式')
  }

  // 连接WebSocket（在地图初始化之后）
  if (websocketUrl.value) {
    connectWebSocket()
  }

  // 初始加载位置数据（静默模式，不显示错误提示）
  await refreshLocations(false)
})

onUnmounted(() => {
  if (ws) {
    ws.close()
    ws = null
  }
  if (map) {
    try {
      map.destroy()
    } catch (e) {
      console.error('销毁地图失败:', e)
    }
    map = null
  }
  mapReady.value = false
  markers = []
})
</script>

<style scoped>
/* 覆盖高德地图label的默认样式，移除蓝色边框 */
:deep(.amap-marker-label),
:deep([class*="amap-marker-label"]),
:deep(.amap-labels) {
  border: none !important;
  background: transparent !important;
  padding: 0 !important;
  margin: 0 !important;
  box-shadow: none !important;
  outline: none !important;
}
.employee-locations-page {
  height: calc(100vh - 120px);
}

.locations-card {
  height: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.title {
  display: flex;
  flex-direction: column;
}

.main {
  font-size: 18px;
  font-weight: 600;
  color: #20253a;
}

.sub {
  font-size: 12px;
  color: #8c92a4;
  margin-top: 4px;
}

.actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.locations-container {
  display: flex;
  height: calc(100vh - 220px);
  gap: 16px;
}

.locations-list {
  width: 800px;
  flex-shrink: 0;
}

.locations-map {
  flex: 1;
  min-width: 0;
}

.text-muted {
  color: #8c92a4;
}
</style>

