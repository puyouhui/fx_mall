<template>
  <view class="rich-content-page">
    <!-- 自定义导航栏 -->
    <view class="custom-navbar" :style="{ paddingTop: statusBarHeight + 'px' }">
      <view class="navbar-content">
        <view class="navbar-left" @click="goBack">
          <uni-icons type="left" size="24" color="#fff"></uni-icons>
        </view>
        <view class="navbar-title">{{ title }}</view>
        <view class="navbar-right"></view>
      </view>
    </view>

    <!-- 加载中 -->
    <view v-if="loading" class="loading-container">
      <uni-icons type="spinner-cycle" size="40" color="#20CB6B"></uni-icons>
      <text class="loading-text">加载中...</text>
    </view>

    <!-- 内容 -->
    <view v-else-if="content" class="content-container" :style="{ paddingTop: (statusBarHeight + 44 + 16) + 'px' }">
      <view class="content-header">
        <text class="content-title">{{ content.title }}</text>
        <view class="content-meta">
          <text class="meta-item">发布时间：{{ formatDate(content.published_at) }}</text>
          <text class="meta-item">浏览次数：{{ content.view_count }}</text>
        </view>
      </view>
      
      <view class="content-body">
        <rich-text :nodes="content.content" class="rich-text"></rich-text>
      </view>
    </view>

    <!-- 错误提示 -->
    <view v-else class="error-container">
      <uni-icons type="info-filled" size="60" color="#999"></uni-icons>
      <text class="error-text">{{ errorMsg }}</text>
      <button class="retry-btn" @click="loadContent">重试</button>
    </view>
  </view>
</template>

<script>
import { getRichContentDetail } from '@/api/richContent'

export default {
  data() {
    return {
      statusBarHeight: 0,
      contentId: null,
      title: '内容详情',
      content: null,
      loading: false,
      errorMsg: ''
    }
  },
  
  onLoad(options) {
    // 获取状态栏高度
    const systemInfo = uni.getSystemInfoSync()
    this.statusBarHeight = systemInfo.statusBarHeight
    
    // 获取传入的富文本内容ID
    if (options.id) {
      this.contentId = parseInt(options.id)
      this.loadContent()
    } else {
      this.errorMsg = '缺少内容ID参数'
    }
  },
  
  methods: {
    // 加载富文本内容
    async loadContent() {
      if (!this.contentId) {
        this.errorMsg = '缺少内容ID参数'
        return
      }
      
      this.loading = true
      this.errorMsg = ''
      
      try {
        const res = await getRichContentDetail(this.contentId)
        console.log('富文本内容响应:', res)
        
        // 检查响应格式
        if (res && res.code === 200 && res.data) {
          this.content = res.data
          this.title = res.data.title || '内容详情'
          
          // 处理富文本内容，为图片添加样式（小程序兼容）
          if (this.content.content) {
            this.content.content = this.processRichTextContent(this.content.content)
          }
          
          // 设置页面标题
          uni.setNavigationBarTitle({
            title: this.title
          })
        } else if (res && res.code !== 200) {
          // 业务错误码
          this.errorMsg = res.message || res.msg || '内容不存在或未发布'
        } else if (res && !res.data) {
          // 数据为空
          this.errorMsg = '内容不存在或未发布'
        } else {
          // 未知错误
          this.errorMsg = '数据格式错误'
        }
      } catch (error) {
        console.error('加载富文本内容失败:', error)
        // 显示更详细的错误信息
        if (error && error.message) {
          this.errorMsg = error.message
        } else if (error && error.msg) {
          this.errorMsg = error.msg
        } else {
          this.errorMsg = '加载失败，请稍后重试'
        }
      } finally {
        this.loading = false
      }
    },
    
    // 返回上一页
    goBack() {
      uni.navigateBack({
        delta: 1
      })
    },
    
    // 格式化日期
    formatDate(dateStr) {
      if (!dateStr) return ''
      const date = new Date(dateStr)
      const year = date.getFullYear()
      const month = String(date.getMonth() + 1).padStart(2, '0')
      const day = String(date.getDate()).padStart(2, '0')
      const hour = String(date.getHours()).padStart(2, '0')
      const minute = String(date.getMinutes()).padStart(2, '0')
      return `${year}-${month}-${day} ${hour}:${minute}`
    },
    
    // 处理富文本内容，为图片添加内联样式（小程序兼容）
    processRichTextContent(html) {
      if (!html) return html
      
      // 使用正则表达式为所有图片添加样式
      // 匹配 <img> 标签
      const imgRegex = /<img([^>]*)>/gi
      
      return html.replace(imgRegex, (match, attributes) => {
        // 检查是否已有 style 属性
        if (attributes.includes('style=')) {
          // 如果已有 style，添加 max-width
          return match.replace(/style="([^"]*)"/i, (styleMatch, styleContent) => {
            // 检查是否已有 max-width
            if (!styleContent.includes('max-width')) {
              return `style="${styleContent}; max-width: 100%; width: 100%; height: auto; display: block; box-sizing: border-box;"`
            }
            return styleMatch
          })
        } else {
          // 如果没有 style，添加完整的 style 属性
          return `<img${attributes} style="max-width: 100%; width: 100%; height: auto; display: block; margin: 10px 0; box-sizing: border-box;">`
        }
      })
    }
  }
}
</script>

<style scoped>
.rich-content-page {
  min-height: 100vh;
  background-color: #f5f5f5;
}

/* 自定义导航栏 */
.custom-navbar {
  background-color: #20CB6B;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 1000;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.navbar-content {
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
}

.navbar-left {
  width: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.navbar-title {
  flex: 1;
  text-align: center;
  font-size: 17px;
  font-weight: 500;
  color: #fff;
}

.navbar-right {
  width: 60px;
}

/* 加载状态 */
.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 100px 0;
}

.loading-text {
  margin-top: 10px;
  font-size: 14px;
  color: #999;
}

/* 内容区域 */
.content-container {
  padding: 16px;
  padding-bottom: 20px;
}

.content-header {
  background-color: #fff;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.content-title {
  font-size: 20px;
  font-weight: bold;
  color: #333;
  line-height: 1.5;
  display: block;
  margin-bottom: 12px;
}

.content-meta {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.meta-item {
  font-size: 12px;
  color: #999;
}

.content-body {
  background-color: #fff;
  padding: 20px;
  border-radius: 8px;
  min-height: 300px;
  overflow: hidden;
  word-wrap: break-word;
  word-break: break-all;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.rich-text {
  font-size: 16px;
  line-height: 1.8;
  color: #333;
  width: 100%;
  box-sizing: border-box;
  overflow: hidden;
}

/* 富文本内容样式 - 小程序兼容 */
/* 注意：小程序 rich-text 组件对内部样式支持有限，主要通过内联样式控制 */

/* 错误状态 */
.error-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 100px 20px;
}

.error-text {
  margin-top: 16px;
  font-size: 14px;
  color: #999;
  text-align: center;
}

.retry-btn {
  margin-top: 20px;
  width: 200px;
  height: 40px;
  line-height: 40px;
  background-color: #20CB6B;
  color: #fff;
  border-radius: 20px;
  font-size: 14px;
  border: none;
}
</style>

