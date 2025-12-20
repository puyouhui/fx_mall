<template>
  <div ref="watermarkRef" class="watermark-container" :style="watermarkStyle"></div>
</template>

<script setup>
import { onMounted, onBeforeUnmount, ref } from 'vue'

const props = defineProps({
  text: {
    type: String,
    default: '供应商后台管理系统'
  },
  fontSize: {
    type: Number,
    default: 16
  },
  dateFontSize: {
    type: Number,
    default: 12
  },
  color: {
    type: String,
    default: 'rgba(0, 0, 0, 0.1)'
  },
  rotate: {
    type: Number,
    default: -20
  },
  gap: {
    type: [Number, Array],
    default: () => [200, 200] // [x, y]
  }
})

const currentDateTime = ref('')
const watermarkStyle = ref({})
const watermarkRef = ref(null)

// 格式化日期时间
const formatDateTime = () => {
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  const hours = String(now.getHours()).padStart(2, '0')
  const minutes = String(now.getMinutes()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}`
}

// 更新日期时间
const updateDateTime = () => {
  currentDateTime.value = formatDateTime()
}

// 生成水印
const generateWatermark = () => {
  const canvasEl = document.createElement('canvas')
  const ctx = canvasEl.getContext('2d')
  
  const gapX = Array.isArray(props.gap) ? props.gap[0] : props.gap
  const gapY = Array.isArray(props.gap) ? props.gap[1] : props.gap
  
  // 计算主文本宽度和高度
  ctx.font = `${props.fontSize}px Arial`
  const textMetrics = ctx.measureText(props.text)
  const textWidth = textMetrics.width
  const textHeight = props.fontSize
  
  // 计算日期时间文本宽度和高度
  ctx.font = `${props.dateFontSize}px Arial`
  const dateMetrics = ctx.measureText(currentDateTime.value)
  const dateWidth = dateMetrics.width
  const dateHeight = props.dateFontSize
  
  // 计算总宽度和高度（取两者最大值，加上行间距）
  const lineSpacing = 8 // 行间距
  const totalWidth = Math.max(textWidth, dateWidth)
  const totalHeight = textHeight + lineSpacing + dateHeight
  
  // 设置canvas尺寸
  const angle = (props.rotate * Math.PI) / 180
  const canvasWidth = totalWidth + gapX
  const canvasHeight = totalHeight + gapY
  
  canvasEl.width = canvasWidth
  canvasEl.height = canvasHeight
  
  // 移动到中心点
  ctx.translate(canvasWidth / 2, canvasHeight / 2)
  ctx.rotate(angle)
  
  ctx.fillStyle = props.color
  ctx.textAlign = 'center'
  
  // 绘制主文本
  ctx.font = `${props.fontSize}px Arial`
  ctx.textBaseline = 'top'
  ctx.fillText(props.text, 0, -(totalHeight / 2))
  
  // 绘制日期时间文本（在主文本下方）
  ctx.font = `${props.dateFontSize}px Arial`
  ctx.fillText(currentDateTime.value, 0, -(totalHeight / 2) + textHeight + lineSpacing)
  
  // 将canvas转换为base64图片
  const base64Url = canvasEl.toDataURL()
  
  // 设置背景样式
  watermarkStyle.value = {
    backgroundImage: `url(${base64Url})`,
    backgroundRepeat: 'repeat',
    backgroundPosition: '0 0',
    pointerEvents: 'none',
    position: 'fixed',
    top: 0,
    left: 0,
    width: '100%',
    height: '100%',
    zIndex: 9999
  }
}

// 简单的防篡改：监听DOM变化
let observer = null
let timer = null

const observeWatermark = () => {
  const targetNode = document.body
  const config = { 
    childList: true, 
    subtree: true,
    attributes: true,
    attributeFilter: ['style', 'class']
  }
  
  observer = new MutationObserver((mutations) => {
    mutations.forEach((mutation) => {
      if (mutation.type === 'childList') {
        mutation.removedNodes.forEach((node) => {
          if (node.nodeType === 1 && 
              node.classList && 
              node.classList.contains('watermark-container') &&
              watermarkRef.value === node) {
            // 水印被删除，重新生成
            setTimeout(() => {
              if (watermarkRef.value && watermarkRef.value.parentNode) {
                // 如果还在DOM中，只更新样式
                Object.assign(watermarkRef.value.style, watermarkStyle.value)
              }
            }, 100)
          }
        })
      }
      if (mutation.type === 'attributes') {
        if (mutation.target === watermarkRef.value) {
          // 水印样式被修改，恢复样式
          setTimeout(() => {
            if (watermarkRef.value) {
              Object.assign(watermarkRef.value.style, watermarkStyle.value)
            }
          }, 100)
        }
      }
    })
  })
  
  observer.observe(targetNode, config)
}

onMounted(() => {
  // 初始化日期时间
  updateDateTime()
  generateWatermark()
  
  // 启动防篡改监听
  observeWatermark()
  
  // 每分钟更新一次日期时间并重新生成水印
  timer = setInterval(() => {
    updateDateTime()
    generateWatermark()
    // 更新现有元素的样式
    if (watermarkRef.value) {
      Object.assign(watermarkRef.value.style, watermarkStyle.value)
    }
  }, 60000) // 每分钟更新一次
})

onBeforeUnmount(() => {
  if (observer) {
    observer.disconnect()
  }
  if (timer) {
    clearInterval(timer)
  }
})
</script>

<style scoped>
.watermark-container {
  pointer-events: none;
  user-select: none;
}
</style>
