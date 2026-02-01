<template>
  <div class="settings-container">
    <h1 class="page-title">系统设置</h1>
    
    <el-card class="settings-card">
      <el-tabs v-model="activeTab" class="settings-tabs">
        <!-- 修改密码 Tab -->
        <el-tab-pane label="修改密码" name="password">
          <el-form
            ref="passwordFormRef"
            :model="passwordForm"
            :rules="passwordRules"
            label-width="120px"
            class="password-form"
          >
            <el-form-item label="原密码" prop="old_password">
              <el-input
                v-model="passwordForm.old_password"
                type="password"
                placeholder="请输入原密码"
                show-password
                :prefix-icon="Lock"
                style="width: 400px"
              />
            </el-form-item>
            
            <el-form-item label="新密码" prop="new_password">
              <el-input
                v-model="passwordForm.new_password"
                type="password"
                placeholder="请输入新密码（至少6位）"
                show-password
                :prefix-icon="Lock"
                style="width: 400px"
              />
            </el-form-item>
            
            <el-form-item label="确认新密码" prop="confirm_password">
              <el-input
                v-model="passwordForm.confirm_password"
                type="password"
                placeholder="请再次输入新密码"
                show-password
                :prefix-icon="Lock"
                style="width: 400px"
              />
            </el-form-item>
            
            <el-form-item>
              <el-button
                type="primary"
                :loading="passwordLoading"
                @click="handleChangePassword"
              >
                确认修改
              </el-button>
              <el-button @click="handleResetPassword">重置</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 订单配置 Tab -->
        <el-tab-pane label="订单配置" name="order">
          <el-form
            ref="orderFormRef"
            :model="orderForm"
            label-width="150px"
            class="order-form"
          >
            <el-alert
              title="订单配置说明"
              type="info"
              :closable="false"
              style="margin-bottom: 20px"
            >
              <template #default>
                <div style="line-height: 1.8;">
                  <p>• 加急订单费用：用户选择加急订单时需要额外支付的费用（元）</p>
                  <p>• 加急费用不计入商品利润，属于服务费用</p>
                </div>
              </template>
            </el-alert>

            <el-form-item label="加急订单费用（元）">
              <el-input-number
                v-model="orderForm.urgent_fee"
                :min="0"
                :precision="2"
                :step="1"
                placeholder="请输入加急订单费用"
                style="width: 300px"
              />
              <div style="margin-top: 8px; color: #909399; font-size: 12px;">
                用户选择加急订单时需要额外支付的费用
              </div>
            </el-form-item>

            <el-form-item>
              <el-button
                type="primary"
                :loading="orderLoading"
                @click="handleSaveOrderSettings"
              >
                保存配置
              </el-button>
              <el-button @click="handleResetOrderForm">重置</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 配送费计算配置 Tab -->
        <el-tab-pane label="配送费计算" name="delivery-fee">
          <el-form
            ref="deliveryFeeFormRef"
            :model="deliveryFeeForm"
            label-width="200px"
            class="delivery-fee-form"
          >
            <el-alert
              title="配送费计算配置说明"
              type="info"
              :closable="false"
              style="margin-bottom: 20px"
            >
              <template #default>
                <div style="line-height: 1.8;">
                  <p>• 这些配置用于计算订单的预估配送费</p>
                  <p>• 配送费 = 基础配送费 + 各项补贴 - 利润分成</p>
                  <p>• 利润分成仅在订单利润超过阈值时计算，且仅管理员可见</p>
                </div>
              </template>
            </el-alert>

            <el-divider content-position="left">基础配置</el-divider>
            
            <el-form-item label="基础配送费（元）">
              <el-input-number
                v-model="deliveryFeeForm.base_fee"
                :min="0"
                :precision="2"
                :step="0.5"
                placeholder="基础配送费"
                style="width: 300px"
              />
              <div style="margin-top: 8px; color: #909399; font-size: 12px;">
                每单的基础配送费用
              </div>
            </el-form-item>

            <el-divider content-position="left">孤立订单补贴</el-divider>

            <el-form-item label="孤立订单判断距离（公里）">
              <el-input-number
                v-model="deliveryFeeForm.isolated_distance"
                :min="0"
                :precision="1"
                :step="1"
                placeholder="判断距离"
                style="width: 300px"
              />
              <div style="margin-top: 8px; color: #909399; font-size: 12px;">
                在此距离内无其他订单则视为孤立订单
              </div>
            </el-form-item>

            <el-form-item label="孤立订单补贴（元）">
              <el-input-number
                v-model="deliveryFeeForm.isolated_subsidy"
                :min="0"
                :precision="2"
                :step="0.5"
                placeholder="孤立订单补贴"
                style="width: 300px"
              />
              <div style="margin-top: 8px; color: #909399; font-size: 12px;">
                孤立订单的额外补贴金额
              </div>
            </el-form-item>

            <el-divider content-position="left">件数补贴</el-divider>

            <el-form-item label="件数补贴低阈值（件）">
              <el-input-number
                v-model="deliveryFeeForm.item_threshold_low"
                :min="0"
                :precision="0"
                :step="1"
                placeholder="低阈值"
                style="width: 300px"
              />
              <div style="margin-top: 8px; color: #909399; font-size: 12px;">
                达到此件数开始计算件数补贴（低档费率）
              </div>
            </el-form-item>

            <el-form-item label="件数补贴低档费率（元/件）">
              <el-input-number
                v-model="deliveryFeeForm.item_rate_low"
                :min="0"
                :precision="2"
                :step="0.1"
                placeholder="低档费率"
                style="width: 300px"
              />
              <div style="margin-top: 8px; color: #909399; font-size: 12px;">
                低阈值到高阈值之间的每件补贴
              </div>
            </el-form-item>

            <el-form-item label="件数补贴高阈值（件）">
              <el-input-number
                v-model="deliveryFeeForm.item_threshold_high"
                :min="0"
                :precision="0"
                :step="1"
                placeholder="高阈值"
                style="width: 300px"
              />
              <div style="margin-top: 8px; color: #909399; font-size: 12px;">
                达到此件数使用高档费率
              </div>
            </el-form-item>

            <el-form-item label="件数补贴高档费率（元/件）">
              <el-input-number
                v-model="deliveryFeeForm.item_rate_high"
                :min="0"
                :precision="2"
                :step="0.1"
                placeholder="高档费率"
                style="width: 300px"
              />
              <div style="margin-top: 8px; color: #909399; font-size: 12px;">
                高阈值以上的每件补贴
              </div>
            </el-form-item>

            <el-form-item label="件数补贴最大计件数">
              <el-input-number
                v-model="deliveryFeeForm.item_max_count"
                :min="0"
                :precision="0"
                :step="1"
                placeholder="最大计件数"
                style="width: 300px"
              />
              <div style="margin-top: 8px; color: #909399; font-size: 12px;">
                超过此数量的件数不再计算补贴
              </div>
            </el-form-item>

            <el-divider content-position="left">加急订单补贴</el-divider>

            <el-form-item label="加急订单补贴（元）">
              <el-input-number
                v-model="deliveryFeeForm.urgent_subsidy"
                :min="0"
                :precision="2"
                :step="1"
                placeholder="加急订单补贴"
                style="width: 300px"
              />
              <div style="margin-top: 8px; color: #909399; font-size: 12px;">
                加急订单的额外补贴金额
              </div>
            </el-form-item>

            <el-divider content-position="left">极端天气补贴</el-divider>

            <el-form-item label="极端天气补贴（元）">
              <el-input-number
                v-model="deliveryFeeForm.weather_subsidy"
                :min="0"
                :precision="2"
                :step="0.5"
                placeholder="极端天气补贴"
                style="width: 300px"
              />
              <div style="margin-top: 8px; color: #909399; font-size: 12px;">
                雨雪或高温天气的额外补贴
              </div>
            </el-form-item>

            <el-form-item label="极端高温阈值（摄氏度）">
              <el-input-number
                v-model="deliveryFeeForm.extreme_temp"
                :min="0"
                :precision="1"
                :step="1"
                placeholder="极端高温阈值"
                style="width: 300px"
              />
              <div style="margin-top: 8px; color: #909399; font-size: 12px;">
                超过此温度视为极端天气
              </div>
            </el-form-item>

            <el-divider content-position="left">利润分成（仅管理员可见）</el-divider>

            <el-form-item label="利润分成阈值（元）">
              <el-input-number
                v-model="deliveryFeeForm.profit_threshold"
                :min="0"
                :precision="2"
                :step="1"
                placeholder="利润分成阈值"
                style="width: 300px"
              />
              <div style="margin-top: 8px; color: #909399; font-size: 12px;">
                订单利润超过此值才开始计算分成
              </div>
            </el-form-item>

            <el-form-item label="利润分成比例">
              <el-input-number
                v-model="deliveryFeeForm.profit_share_rate"
                :min="0"
                :max="1"
                :precision="4"
                :step="0.0001"
                placeholder="利润分成比例"
                style="width: 300px"
              />
              <div style="margin-top: 8px; color: #909399; font-size: 12px;">
                例如：0.08 表示 8%，即 (订单利润-配送成本) × 8%
              </div>
            </el-form-item>

            <el-form-item label="利润分成上限（元）">
              <el-input-number
                v-model="deliveryFeeForm.max_profit_share"
                :min="0"
                :precision="2"
                :step="1"
                placeholder="利润分成上限"
                style="width: 300px"
              />
              <div style="margin-top: 8px; color: #909399; font-size: 12px;">
                利润分成的最大金额，超过此值不再增加
              </div>
            </el-form-item>

            <el-form-item>
              <el-button
                type="primary"
                :loading="deliveryFeeLoading"
                @click="handleSaveDeliveryFeeSettings"
              >
                保存配置
              </el-button>
              <el-button @click="handleResetDeliveryFeeForm">重置</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 地图配置 Tab -->
        <el-tab-pane label="地图配置" name="map">
          <el-form
            ref="mapFormRef"
            :model="mapForm"
            label-width="150px"
            class="map-form"
          >
            <el-alert
              title="地图API配置说明"
              type="info"
              :closable="false"
              style="margin-bottom: 20px"
            >
              <template #default>
                <div style="line-height: 1.8;">
                  <p>• 高德地图和腾讯地图至少需要配置一个，用于地址解析功能</p>
                  <p>• 如果同时配置了两个，系统会优先使用高德地图</p>
                  <p>• 地址解析功能用于将地址文本转换为经纬度坐标，确保配送功能正常</p>
                  <p>• 获取API Key：<a href="https://lbs.amap.com/" target="_blank">高德地图开放平台</a> | <a href="https://lbs.qq.com/" target="_blank">腾讯位置服务</a></p>
                </div>
              </template>
            </el-alert>

            <el-form-item label="高德地图API Key">
              <el-input
                v-model="mapForm.amap_key"
                placeholder="请输入高德地图API Key"
                style="width: 500px"
                show-password
                clearable
              >
                <template #prefix>
                  <el-icon><Location /></el-icon>
                </template>
              </el-input>
              <div style="margin-top: 8px; color: #909399; font-size: 12px;">
                用于地址解析，优先使用高德地图
              </div>
            </el-form-item>

            <el-form-item label="腾讯地图API Key">
              <el-input
                v-model="mapForm.tencent_key"
                placeholder="请输入腾讯地图API Key"
                style="width: 500px"
                show-password
                clearable
              >
                <template #prefix>
                  <el-icon><Location /></el-icon>
                </template>
              </el-input>
              <div style="margin-top: 8px; color: #909399; font-size: 12px;">
                用于地址解析，当高德地图未配置时使用
              </div>
            </el-form-item>

            <el-form-item>
              <el-button
                type="primary"
                :loading="mapLoading"
                @click="handleSaveMapSettings"
              >
                保存配置
              </el-button>
              <el-button @click="handleResetMapForm">重置</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 打印机配置 Tab -->
        <el-tab-pane label="打印机配置" name="printer">
          <el-form
            ref="printerFormRef"
            :model="printerForm"
            label-width="150px"
            class="printer-form"
          >
            <el-alert
              title="打印机配置说明"
              type="info"
              :closable="false"
              style="margin-bottom: 20px"
            >
              <template #default>
                <div style="line-height: 1.8;">
                  <p><strong>方式一：直接连接本地打印机客户端（适用于 HTTP 页面）</strong></p>
                  <p>• 格式：http://IP地址:端口号，例如：http://198.18.0.1:17521</p>
                  <p>• 请确保打印客户端正在运行，并且地址配置正确</p>
                  <p style="margin-top: 12px;"><strong>方式二：通过中转服务连接（推荐，适用于 HTTPS 页面）</strong></p>
                  <p>• 格式：https://域名:端口号，例如：https://mall.sscchh.com:17521</p>
                  <p>• 需要在服务器上部署 node-hiprint-transit 中转服务</p>
                  <p>• 本地打印客户端需要连接到中转服务（配置相同的 token）</p>
                  <p>• 这样可以解决 HTTPS 页面的混合内容问题</p>
                  <p style="margin-top: 12px;">• 配置将保存到本地存储，刷新页面后自动生效</p>
                </div>
              </template>
            </el-alert>

            <el-form-item label="打印机地址">
              <el-input
                v-model="printerForm.address"
                placeholder="例如：http://198.18.0.1:17521 或 https://mall.sscchh.com:17521"
                style="width: 500px"
                clearable
              >
                <template #prefix>
                  <el-icon><Printer /></el-icon>
                </template>
              </el-input>
              <div style="margin-top: 8px; color: #909399; font-size: 12px;">
                打印客户端的WebSocket连接地址。HTTPS页面建议使用中转服务（https://）
              </div>
            </el-form-item>

            <el-form-item>
              <el-button
                type="primary"
                :loading="printerLoading"
                @click="handleSavePrinterSettings"
              >
                保存配置
              </el-button>
              <el-button @click="handleResetPrinterForm">重置为默认</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 微信支付配置 Tab -->
        <el-tab-pane label="微信支付" name="wechat-pay">
          <el-form
            ref="wechatPayFormRef"
            :model="wechatPayForm"
            label-width="180px"
            class="wechat-pay-form"
          >
            <el-alert
              title="微信支付配置说明"
              type="info"
              :closable="false"
              style="margin-bottom: 20px"
            >
              <template #default>
                <div style="line-height: 1.8;">
                  <p>• 需在 <a href="https://pay.weixin.qq.com/" target="_blank">微信支付商户平台</a> 申请并获取商户号、APIv3 密钥、商户证书</p>
                  <p>• 商户证书私钥：apiclient_key.pem 文件的完整内容（包含 -----BEGIN/END----- 行）</p>
                  <p>• 证书序列号：在商户平台【API安全】-【API证书】中查看</p>
                  <p>• 回调地址需为公网可访问的 HTTPS 地址，格式如：https://您的域名/api/mini/wechat-pay/notify</p>
                  <p>• 需在商户平台【产品中心】-【开发配置】中配置支付授权目录和回调地址</p>
                </div>
              </template>
            </el-alert>

            <el-form-item label="商户号">
              <el-input
                v-model="wechatPayForm.mch_id"
                placeholder="微信支付商户号"
                style="width: 400px"
                clearable
              />
            </el-form-item>

            <el-form-item label="小程序 AppID">
              <el-input
                v-model="wechatPayForm.app_id"
                placeholder="与商户号绑定的小程序 AppID"
                style="width: 400px"
                clearable
              />
            </el-form-item>

            <el-form-item label="APIv3 密钥">
              <el-input
                v-model="wechatPayForm.api_v3_key"
                type="password"
                placeholder="32位 APIv3 密钥"
                style="width: 400px"
                show-password
                clearable
              />
              <div style="margin-top: 8px; color: #909399; font-size: 12px;">
                商户平台【API安全】-【APIv3密钥】中设置
              </div>
            </el-form-item>

            <el-form-item label="商户证书序列号">
              <el-input
                v-model="wechatPayForm.serial_no"
                placeholder="API 证书序列号"
                style="width: 400px"
                clearable
              />
            </el-form-item>

            <el-form-item label="商户私钥（PEM）">
              <el-input
                v-model="wechatPayForm.private_key"
                type="textarea"
                placeholder="apiclient_key.pem 文件完整内容，包含 -----BEGIN PRIVATE KEY----- 等行"
                :rows="8"
                style="width: 500px"
              />
            </el-form-item>

            <el-form-item label="支付回调地址">
              <el-input
                v-model="wechatPayForm.notify_url"
                placeholder="https://您的域名/api/mini/wechat-pay/notify"
                style="width: 500px"
                clearable
              />
              <div style="margin-top: 8px; color: #909399; font-size: 12px;">
                需公网 HTTPS，微信支付成功后会调用此地址
              </div>
            </el-form-item>

            <el-form-item>
              <el-button
                type="primary"
                :loading="wechatPayLoading"
                @click="handleSaveWechatPaySettings"
              >
                保存配置
              </el-button>
              <el-button @click="handleResetWechatPayForm">重置</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Lock, Location, Printer } from '@element-plus/icons-vue'
import { changePassword } from '../api/auth'
import { getMapSettings, updateMapSettings, getSystemSettings, updateSystemSettings } from '../api/settings'
import { getPrinterAddress, setPrinterAddress, getDefaultPrinterAddress } from '../utils/printer'

const activeTab = ref('password')
const passwordFormRef = ref(null)
const mapFormRef = ref(null)
const orderFormRef = ref(null)
const deliveryFeeFormRef = ref(null)
const printerFormRef = ref(null)
const wechatPayFormRef = ref(null)
const passwordLoading = ref(false)
const mapLoading = ref(false)
const orderLoading = ref(false)
const deliveryFeeLoading = ref(false)
const printerLoading = ref(false)
const wechatPayLoading = ref(false)

const passwordForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

const mapForm = reactive({
  amap_key: '',
  tencent_key: ''
})

const printerForm = reactive({
  address: ''
})

const wechatPayForm = reactive({
  mch_id: '',
  app_id: '',
  api_v3_key: '',
  serial_no: '',
  private_key: '',
  notify_url: ''
})

const orderForm = reactive({
  urgent_fee: 0
})

const deliveryFeeForm = reactive({
  base_fee: 4.0,
  isolated_distance: 8.0,
  isolated_subsidy: 3.0,
  item_threshold_low: 5,
  item_rate_low: 0.5,
  item_threshold_high: 10,
  item_rate_high: 0.6,
  item_max_count: 50,
  urgent_subsidy: 10.0,
  weather_subsidy: 1.0,
  extreme_temp: 37.0,
  profit_threshold: 25.0,
  profit_share_rate: 0.08,
  max_profit_share: 50.0
})

// 自定义验证规则：确认密码
const validateConfirmPassword = (rule, value, callback) => {
  if (value === '') {
    callback(new Error('请再次输入新密码'))
  } else if (value !== passwordForm.new_password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const passwordRules = {
  old_password: [
    { required: true, message: '请输入原密码', trigger: 'blur' }
  ],
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少为6位', trigger: 'blur' }
  ],
  confirm_password: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' }
  ]
}

// 处理修改密码
const handleChangePassword = async () => {
  try {
    // 验证表单
    await passwordFormRef.value.validate()
    
    passwordLoading.value = true
    
    // 调用API修改密码
    const response = await changePassword({
      old_password: passwordForm.old_password,
      new_password: passwordForm.new_password
    })
    
    if (response.code === 200) {
      ElMessage.success('密码修改成功，请重新登录')
      
      // 询问是否立即退出登录
      ElMessageBox.confirm(
        '密码已修改成功，为了安全起见，建议您重新登录。是否立即退出登录？',
        '提示',
        {
          confirmButtonText: '立即退出',
          cancelButtonText: '稍后退出',
          type: 'success'
        }
      ).then(() => {
        // 退出登录
        localStorage.removeItem('token')
        localStorage.removeItem('username')
        window.location.href = '/login'
      }).catch(() => {
        // 用户选择稍后退出，重置表单
        handleResetPassword()
      })
    } else {
      ElMessage.error(response.message || '密码修改失败')
    }
  } catch (error) {
    console.error('修改密码失败:', error)
    if (error.response && error.response.data) {
      ElMessage.error(error.response.data.message || '密码修改失败')
    } else if (error.message) {
      ElMessage.error(error.message)
    } else {
      ElMessage.error('密码修改失败，请稍后再试')
    }
  } finally {
    passwordLoading.value = false
  }
}

// 重置密码表单
const handleResetPassword = () => {
  passwordFormRef.value?.resetFields()
  passwordForm.old_password = ''
  passwordForm.new_password = ''
  passwordForm.confirm_password = ''
}

// 加载地图设置
const loadMapSettings = async () => {
  try {
    const response = await getMapSettings()
    if (response.code === 200 && response.data) {
      mapForm.amap_key = response.data.amap_key || ''
      mapForm.tencent_key = response.data.tencent_key || ''
    }
  } catch (error) {
    console.error('加载地图设置失败:', error)
  }
}

// 保存地图设置
const handleSaveMapSettings = async () => {
  try {
    mapLoading.value = true
    
    const response = await updateMapSettings({
      amap_key: mapForm.amap_key.trim(),
      tencent_key: mapForm.tencent_key.trim()
    })
    
    if (response.code === 200) {
      ElMessage.success('地图配置保存成功')
    } else {
      ElMessage.error(response.message || '保存失败')
    }
  } catch (error) {
    console.error('保存地图设置失败:', error)
    ElMessage.error('保存失败，请稍后再试')
  } finally {
    mapLoading.value = false
  }
}

// 重置地图表单
const handleResetMapForm = () => {
  loadMapSettings()
}

// 加载订单设置
const loadOrderSettings = async () => {
  try {
    const response = await getSystemSettings()
    if (response.code === 200 && response.data) {
      const urgentFee = response.data.order_urgent_fee || '0'
      orderForm.urgent_fee = parseFloat(urgentFee) || 0
    }
  } catch (error) {
    console.error('加载订单设置失败:', error)
  }
}

// 保存订单设置
const handleSaveOrderSettings = async () => {
  try {
    orderLoading.value = true
    
    // 确保 urgent_fee 是有效的数字
    const urgentFeeValue = orderForm.urgent_fee || 0
    if (urgentFeeValue < 0) {
      ElMessage.error('加急费用不能为负数')
      orderLoading.value = false
      return
    }
    
    console.log('保存订单设置，加急费用:', urgentFeeValue)
    
    const response = await updateSystemSettings({
      order_urgent_fee: urgentFeeValue.toString()
    })
    
    console.log('保存订单设置响应:', response)
    
    if (response && response.code === 200) {
      ElMessage.success('订单配置保存成功')
    } else {
      ElMessage.error(response?.message || '保存失败')
    }
  } catch (error) {
    console.error('保存订单设置失败:', error)
    if (error.response && error.response.data) {
      ElMessage.error(error.response.data.message || '保存失败，请稍后再试')
    } else if (error.message) {
      ElMessage.error(error.message)
    } else {
      ElMessage.error('保存失败，请稍后再试')
    }
  } finally {
    orderLoading.value = false
  }
}

// 重置订单表单
const handleResetOrderForm = () => {
  loadOrderSettings()
}

// 加载配送费计算设置
const loadDeliveryFeeSettings = async () => {
  try {
    const response = await getSystemSettings()
    if (response.code === 200 && response.data) {
      const data = response.data
      deliveryFeeForm.base_fee = parseFloat(data.delivery_base_fee || '4.0') || 4.0
      deliveryFeeForm.isolated_distance = parseFloat(data.delivery_isolated_distance || '8.0') || 8.0
      deliveryFeeForm.isolated_subsidy = parseFloat(data.delivery_isolated_subsidy || '3.0') || 3.0
      deliveryFeeForm.item_threshold_low = parseInt(data.delivery_item_threshold_low || '5') || 5
      deliveryFeeForm.item_rate_low = parseFloat(data.delivery_item_rate_low || '0.5') || 0.5
      deliveryFeeForm.item_threshold_high = parseInt(data.delivery_item_threshold_high || '10') || 10
      deliveryFeeForm.item_rate_high = parseFloat(data.delivery_item_rate_high || '0.6') || 0.6
      deliveryFeeForm.item_max_count = parseInt(data.delivery_item_max_count || '50') || 50
      deliveryFeeForm.urgent_subsidy = parseFloat(data.delivery_urgent_subsidy || '10.0') || 10.0
      deliveryFeeForm.weather_subsidy = parseFloat(data.delivery_weather_subsidy || '1.0') || 1.0
      deliveryFeeForm.extreme_temp = parseFloat(data.delivery_extreme_temp || '37.0') || 37.0
      deliveryFeeForm.profit_threshold = parseFloat(data.delivery_profit_threshold || '25.0') || 25.0
      deliveryFeeForm.profit_share_rate = parseFloat(data.delivery_profit_share_rate || '0.08') || 0.08
      deliveryFeeForm.max_profit_share = parseFloat(data.delivery_max_profit_share || '50.0') || 50.0
    }
  } catch (error) {
    console.error('加载配送费计算设置失败:', error)
  }
}

// 保存配送费计算设置
const handleSaveDeliveryFeeSettings = async () => {
  try {
    deliveryFeeLoading.value = true
    
    // 验证数据
    if (deliveryFeeForm.base_fee < 0) {
      ElMessage.error('基础配送费不能为负数')
      deliveryFeeLoading.value = false
      return
    }
    if (deliveryFeeForm.item_threshold_low >= deliveryFeeForm.item_threshold_high) {
      ElMessage.error('件数补贴低阈值必须小于高阈值')
      deliveryFeeLoading.value = false
      return
    }
    if (deliveryFeeForm.profit_share_rate < 0 || deliveryFeeForm.profit_share_rate > 1) {
      ElMessage.error('利润分成比例必须在 0 到 1 之间')
      deliveryFeeLoading.value = false
      return
    }
    
    const settings = {
      delivery_base_fee: deliveryFeeForm.base_fee.toString(),
      delivery_isolated_distance: deliveryFeeForm.isolated_distance.toString(),
      delivery_isolated_subsidy: deliveryFeeForm.isolated_subsidy.toString(),
      delivery_item_threshold_low: deliveryFeeForm.item_threshold_low.toString(),
      delivery_item_rate_low: deliveryFeeForm.item_rate_low.toString(),
      delivery_item_threshold_high: deliveryFeeForm.item_threshold_high.toString(),
      delivery_item_rate_high: deliveryFeeForm.item_rate_high.toString(),
      delivery_item_max_count: deliveryFeeForm.item_max_count.toString(),
      delivery_urgent_subsidy: deliveryFeeForm.urgent_subsidy.toString(),
      delivery_weather_subsidy: deliveryFeeForm.weather_subsidy.toString(),
      delivery_extreme_temp: deliveryFeeForm.extreme_temp.toString(),
      delivery_profit_threshold: deliveryFeeForm.profit_threshold.toString(),
      delivery_profit_share_rate: deliveryFeeForm.profit_share_rate.toString(),
      delivery_max_profit_share: deliveryFeeForm.max_profit_share.toString()
    }
    
    const response = await updateSystemSettings(settings)
    
    if (response && response.code === 200) {
      ElMessage.success('配送费计算配置保存成功')
    } else {
      ElMessage.error(response?.message || '保存失败')
    }
  } catch (error) {
    console.error('保存配送费计算设置失败:', error)
    if (error.response && error.response.data) {
      ElMessage.error(error.response.data.message || '保存失败，请稍后再试')
    } else if (error.message) {
      ElMessage.error(error.message)
    } else {
      ElMessage.error('保存失败，请稍后再试')
    }
  } finally {
    deliveryFeeLoading.value = false
  }
}

// 重置配送费计算表单
const handleResetDeliveryFeeForm = () => {
  loadDeliveryFeeSettings()
}

// 加载打印机设置
const loadPrinterSettings = () => {
  printerForm.address = getPrinterAddress()
}

// 保存打印机设置
const handleSavePrinterSettings = () => {
  try {
    if (!printerForm.address || !printerForm.address.trim()) {
      ElMessage.warning('请输入打印机地址')
      return
    }

    // 简单的URL格式验证
    const urlPattern = /^https?:\/\/.+/i
    if (!urlPattern.test(printerForm.address.trim())) {
      ElMessage.error('打印机地址格式不正确，请使用 http:// 或 https:// 开头')
      return
    }

    printerLoading.value = true
    setPrinterAddress(printerForm.address.trim())
    ElMessage.success('打印机配置保存成功，将在下次连接时生效')
  } catch (error) {
    console.error('保存打印机设置失败:', error)
    ElMessage.error('保存失败，请稍后再试')
  } finally {
    printerLoading.value = false
  }
}

// 重置打印机表单
const handleResetPrinterForm = () => {
  printerForm.address = getDefaultPrinterAddress()
  setPrinterAddress(printerForm.address)
  ElMessage.success('已重置为默认地址')
}

// 加载微信支付设置
const loadWechatPaySettings = async () => {
  try {
    const response = await getSystemSettings()
    if (response.code === 200 && response.data) {
      const data = response.data
      wechatPayForm.mch_id = data.wechat_pay_mch_id || ''
      wechatPayForm.app_id = data.wechat_pay_app_id || ''
      wechatPayForm.api_v3_key = data.wechat_pay_api_v3_key || ''
      wechatPayForm.serial_no = data.wechat_pay_serial_no || ''
      wechatPayForm.private_key = data.wechat_pay_private_key || ''
      wechatPayForm.notify_url = data.wechat_pay_notify_url || ''
    }
  } catch (error) {
    console.error('加载微信支付设置失败:', error)
  }
}

// 保存微信支付设置
const handleSaveWechatPaySettings = async () => {
  try {
    wechatPayLoading.value = true
    const settings = {
      wechat_pay_mch_id: (wechatPayForm.mch_id || '').trim(),
      wechat_pay_app_id: (wechatPayForm.app_id || '').trim(),
      wechat_pay_api_v3_key: (wechatPayForm.api_v3_key || '').trim(),
      wechat_pay_serial_no: (wechatPayForm.serial_no || '').trim(),
      wechat_pay_private_key: (wechatPayForm.private_key || '').trim(),
      wechat_pay_notify_url: (wechatPayForm.notify_url || '').trim()
    }
    const response = await updateSystemSettings(settings)
    if (response && response.code === 200) {
      ElMessage.success('微信支付配置保存成功')
    } else {
      ElMessage.error(response?.message || '保存失败')
    }
  } catch (error) {
    console.error('保存微信支付设置失败:', error)
    ElMessage.error('保存失败，请稍后再试')
  } finally {
    wechatPayLoading.value = false
  }
}

// 重置微信支付表单
const handleResetWechatPayForm = () => {
  loadWechatPaySettings()
}

// 页面加载时获取地图设置、订单设置、配送费计算设置、打印机设置和微信支付设置
onMounted(() => {
  loadMapSettings()
  loadOrderSettings()
  loadDeliveryFeeSettings()
  loadPrinterSettings()
  loadWechatPaySettings()
})
</script>

<style scoped>
.settings-container {
  padding: 20px;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 20px;
}

.settings-card {
  max-width: 800px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.password-form {
  padding: 20px 0;
}

:deep(.el-form-item__label) {
  font-weight: 500;
}

:deep(.el-input__wrapper) {
  border-radius: 4px;
}

.settings-tabs {
  margin-top: 20px;
}

.password-form,
.map-form,
.order-form,
.delivery-fee-form,
.printer-form {
  padding: 20px 0;
}

.map-form :deep(.el-form-item__label) {
  font-weight: 500;
}

.map-form :deep(.el-alert) {
  margin-bottom: 20px;
}

.map-form a {
  color: #409eff;
  text-decoration: none;
}

.map-form a:hover {
  text-decoration: underline;
}

.delivery-fee-form :deep(.el-divider__text) {
  font-weight: 600;
  color: #303133;
}

.delivery-fee-form :deep(.el-form-item__label) {
  font-weight: 500;
}

.printer-form :deep(.el-form-item__label) {
  font-weight: 500;
}
</style>

