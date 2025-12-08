<template>
  <div class="suppliers-container">
    <el-card>
      <h2 class="page-title">供应商管理</h2>

      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="primary" @click="handleAddSupplier">
            <el-icon>
              <Plus />
            </el-icon>
            新增供应商
          </el-button>
        </div>
      </div>

      <!-- 供应商列表 -->
      <el-card class="suppliers-card">
        <el-table :data="suppliersData" stripe>
          <el-table-column prop="id" label="ID" align="center" width="80" />
          <el-table-column prop="name" label="供应商名称" align="center" />
          <el-table-column prop="contact" label="联系人" align="center" />
          <el-table-column prop="phone" label="联系电话" align="center" />
          <el-table-column prop="email" label="邮箱" align="center" />
          <el-table-column prop="address" label="地址" align="center" show-overflow-tooltip />
          <el-table-column prop="username" label="登录账号" align="center" />
          <el-table-column prop="status" label="状态" align="center" width="100">
            <template #default="scope">
              <el-tag :type="scope.row.status === 1 ? 'success' : 'danger'">
                {{ scope.row.status === 1 ? '启用' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" align="center" width="180">
            <template #default="scope">
              {{ formatDate(scope.row.created_at) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" align="center" fixed="right" width="180">
            <template #default="scope">
              <el-button type="primary" size="small" @click="handleEditSupplier(scope.row)">
                编辑
              </el-button>
              <el-button 
                type="danger" 
                size="small" 
                :disabled="scope.row.username === 'self_operated'"
                @click="handleDeleteSupplier(scope.row.id)">
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </el-card>

    <!-- 新增/编辑供应商弹窗 -->
    <el-dialog v-model="dialogVisible" :title="dialogType === 'add' ? '新增供应商' : '编辑供应商'" width="600px">
      <el-form ref="supplierFormRef" :model="supplierForm" :rules="supplierRules" label-width="100px">
        <el-form-item label="供应商名称" prop="name">
          <el-input v-model="supplierForm.name" placeholder="请输入供应商名称" />
        </el-form-item>
        <el-form-item label="联系人" prop="contact">
          <el-input v-model="supplierForm.contact" placeholder="请输入联系人" />
        </el-form-item>
        <el-form-item label="联系电话" prop="phone">
          <el-input v-model="supplierForm.phone" placeholder="请输入联系电话" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="supplierForm.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="地址" prop="address">
          <div style="display: flex; gap: 8px;">
            <el-input v-model="supplierForm.address" type="textarea" :rows="2" placeholder="请通过地图选择地址" readonly style="flex: 1;" />
            <el-button type="primary" @click="handleOpenAddressPicker">选择地址</el-button>
          </div>
          <div v-if="supplierForm.latitude && supplierForm.longitude" style="margin-top: 8px; color: #909399; font-size: 12px;">
            坐标：{{ supplierForm.latitude.toFixed(6) }}, {{ supplierForm.longitude.toFixed(6) }}
          </div>
        </el-form-item>
        <el-form-item label="登录账号" prop="username">
          <el-input v-model="supplierForm.username" placeholder="请输入登录账号" />
        </el-form-item>
        <el-form-item v-if="dialogType === 'add'" label="登录密码" prop="password">
          <el-input v-model="supplierForm.password" type="password" show-password placeholder="请输入登录密码（至少6位）" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-switch v-model="supplierForm.status" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 地址选择弹窗 -->
    <el-dialog
      v-model="showAddressPicker"
      title="选择地址"
      width="90%"
      :close-on-click-modal="false"
      destroy-on-close
      @opened="initAddressPicker"
    >
      <div id="supplierAddressPickerMap" style="width: 100%; height: 500px;"></div>
      <div style="margin-top: 16px;">
        <el-input
          v-model="selectedAddressText"
          placeholder="已选择的地址"
          readonly
          style="margin-bottom: 8px;"
        />
        <div style="display: flex; gap: 8px;">
          <el-button @click="showAddressPicker = false">取消</el-button>
          <el-button type="primary" @click="handleConfirmAddress">确认选择</el-button>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { getAllSuppliers, createSupplier, updateSupplier, deleteSupplier, reverseGeocode } from '../api/suppliers'
import { getMapSettings } from '../api/settings'
import { formatDate } from '../utils/time-format'

const suppliersData = ref([])

// 弹窗相关
const dialogVisible = ref(false)
const dialogType = ref('add')
const supplierFormRef = ref(null)
const supplierForm = reactive({
  id: '',
  name: '',
  contact: '',
  phone: '',
  email: '',
  address: '',
  latitude: null,
  longitude: null,
  username: '',
  password: '',
  status: true
})

// 表单验证规则
const supplierRules = {
  name: [
    { required: true, message: '请输入供应商名称', trigger: 'blur' }
  ],
  address: [
    { required: true, message: '请选择地址', trigger: 'blur' }
  ],
  username: [
    { required: true, message: '请输入登录账号', trigger: 'blur' },
    { min: 3, max: 20, message: '账号长度在 3 到 20 个字符', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入登录密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度在 6 到 20 个字符', trigger: 'blur' }
  ]
}

// 地址选择器相关
const showAddressPicker = ref(false)
const selectedAddressText = ref('')
const selectedLatitude = ref(null)
const selectedLongitude = ref(null)
let addressPickerMap = null
let addressPickerMarker = null
let addressPickerGeocoder = null

// 初始化数据
const initData = async () => {
  try {
    const response = await getAllSuppliers()
    if (response.code === 200 && response.data) {
      suppliersData.value = response.data
    }
  } catch (error) {
    console.error('加载数据失败:', error)
    ElMessage.error('加载数据失败')
  }
}

// 打开新增供应商弹窗
const handleAddSupplier = () => {
  dialogType.value = 'add'
  if (supplierFormRef.value) {
    supplierFormRef.value.resetFields()
  }
  Object.assign(supplierForm, {
    id: '',
    name: '',
    contact: '',
    phone: '',
    email: '',
    address: '',
    latitude: null,
    longitude: null,
    username: '',
    password: '',
    status: true
  })
  dialogVisible.value = true
}

// 打开编辑供应商弹窗
const handleEditSupplier = (row) => {
  dialogType.value = 'edit'
  Object.assign(supplierForm, {
    id: row.id,
    name: row.name,
    contact: row.contact || '',
    phone: row.phone || '',
    email: row.email || '',
    address: row.address || '',
    latitude: row.latitude || null,
    longitude: row.longitude || null,
    username: row.username,
    password: '', // 编辑时不显示密码
    status: row.status === 1 // 转换为布尔值供el-switch使用
  })
  dialogVisible.value = true
}

// 删除供应商
const handleDeleteSupplier = async (id) => {
  try {
    // 检查是否是自营供应商
    const supplier = suppliersData.value.find(s => s.id === id)
    if (supplier && supplier.username === 'self_operated') {
      ElMessage.warning('不能删除系统默认的"自营"供应商')
      return
    }

    await ElMessageBox.confirm('确定要删除这个供应商吗？', '确认删除', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    const response = await deleteSupplier(id)
    if (response.code === 200) {
      await initData()
      ElMessage.success('删除成功')
    } else {
      ElMessage.error(response.message || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

// 提交表单
const handleSubmit = async () => {
  try {
    await supplierFormRef.value.validate()

    const formData = {
      name: supplierForm.name,
      contact: supplierForm.contact,
      phone: supplierForm.phone,
      email: supplierForm.email,
      address: supplierForm.address,
      latitude: supplierForm.latitude,
      longitude: supplierForm.longitude,
      username: supplierForm.username,
      status: supplierForm.status ? 1 : 0
    }

    if (dialogType.value === 'add') {
      formData.password = supplierForm.password
      const response = await createSupplier(formData)
      if (response.code === 200) {
        ElMessage.success('创建成功')
        dialogVisible.value = false
        await initData()
      } else {
        ElMessage.error(response.message || '创建失败')
      }
    } else {
      const response = await updateSupplier(supplierForm.id, formData)
      if (response.code === 200) {
        ElMessage.success('更新成功')
        dialogVisible.value = false
        await initData()
      } else {
        ElMessage.error(response.message || '更新失败')
      }
    }
  } catch (error) {
    if (error !== false) {
      console.error('提交失败:', error)
    }
  }
}

// 打开地址选择器
const handleOpenAddressPicker = () => {
  selectedAddressText.value = supplierForm.address || ''
  selectedLatitude.value = supplierForm.latitude
  selectedLongitude.value = supplierForm.longitude
  showAddressPicker.value = true
}

// 动态加载高德地图API
const loadAmapScript = (key) => {
  return new Promise((resolve, reject) => {
    if (window.AMap) {
      // 如果AMap已加载，检查插件是否已加载
      if (window.AMap.plugin) {
        resolve()
        return
      }
    }
    const script = document.createElement('script')
    script.src = `https://webapi.amap.com/maps?v=2.0&key=${key}`
    script.onload = () => {
      // 等待AMap完全初始化
      if (window.AMap && window.AMap.plugin) {
        resolve()
      } else {
        setTimeout(() => resolve(), 100)
      }
    }
    script.onerror = () => reject(new Error('地图加载失败'))
    document.head.appendChild(script)
  })
}

// 加载高德地图插件
const loadAmapPlugin = (pluginName) => {
  return new Promise((resolve, reject) => {
    if (!window.AMap || !window.AMap.plugin) {
      reject(new Error('AMap未加载'))
      return
    }
    window.AMap.plugin(pluginName, () => {
      resolve()
    })
  })
}

// 初始化地址选择器地图
const initAddressPicker = async () => {
  await nextTick()
  
  // 从系统设置获取高德地图API Key
  let amapKey = ''
  try {
    const res = await getMapSettings()
    if (res.code === 200 && res.data) {
      amapKey = res.data.amap_key || ''
    }
  } catch (error) {
    console.error('获取地图设置失败:', error)
  }

  if (!amapKey) {
    ElMessage.warning('未配置高德地图API Key，请在系统设置中配置')
    return
  }

  // 动态加载地图API
  try {
    await loadAmapScript(amapKey)
  } catch (error) {
    ElMessage.error('地图加载失败，请检查API Key是否正确')
    return
  }

  // 加载地理编码插件
  try {
    await loadAmapPlugin('AMap.Geocoder')
  } catch (error) {
    ElMessage.error('地理编码插件加载失败')
    return
  }

  // 创建地图实例（默认中心为昆明市）
  const defaultCenter = selectedLongitude.value && selectedLatitude.value
    ? [selectedLongitude.value, selectedLatitude.value]
    : [102.712251, 25.040609] // 默认昆明市

  addressPickerMap = new AMap.Map('supplierAddressPickerMap', {
    zoom: 15,
    center: defaultCenter
  })

  // 创建地理编码实例
  addressPickerGeocoder = new AMap.Geocoder({
    city: '全国'
  })

  // 如果已有地址，创建并显示标记
  if (selectedLongitude.value && selectedLatitude.value) {
    addressPickerMarker = new AMap.Marker({
      position: [selectedLongitude.value, selectedLatitude.value],
      draggable: false
    })
    addressPickerMap.add(addressPickerMarker)
    updateAddressFromCoordinates(selectedLongitude.value, selectedLatitude.value)
  }

  // 地图点击事件（点击后再添加marker）
  addressPickerMap.on('click', (e) => {
    const { lng, lat } = e.lnglat
    
    // 如果marker不存在，创建它
    if (!addressPickerMarker) {
      addressPickerMarker = new AMap.Marker({
        position: [lng, lat],
        draggable: false
      })
      addressPickerMap.add(addressPickerMarker)
    } else {
      // 如果marker已存在，更新位置
      addressPickerMarker.setPosition([lng, lat])
    }
    
    updateAddressFromCoordinates(lng, lat)
  })
}

// 根据坐标更新地址
const updateAddressFromCoordinates = (lng, lat) => {
  selectedLongitude.value = lng
  selectedLatitude.value = lat

  addressPickerGeocoder.getAddress([lng, lat], (status, result) => {
    if (status === 'complete' && result.info === 'OK') {
      selectedAddressText.value = result.regeocode.formattedAddress
    } else {
      selectedAddressText.value = `${lat.toFixed(6)}, ${lng.toFixed(6)}`
    }
  })
}

// 确认选择地址
const handleConfirmAddress = async () => {
  if (!selectedLongitude.value || !selectedLatitude.value) {
    ElMessage.warning('请先选择地址位置')
    return
  }

  // 先填充经纬度
  supplierForm.longitude = selectedLongitude.value
  supplierForm.latitude = selectedLatitude.value

  // 调用接口进行逆地理编码，获取地址
  try {
    const res = await reverseGeocode(selectedLongitude.value, selectedLatitude.value)
    if (res.code === 200 && res.data && res.data.success && res.data.address) {
      supplierForm.address = res.data.address
      ElMessage.success('地址解析成功')
    } else {
      // 如果接口解析失败，使用地图反解析的地址作为备用
      supplierForm.address = selectedAddressText.value || ''
      ElMessage.warning(res.message || '接口解析失败，使用地图解析的地址')
    }
  } catch (error) {
    console.error('逆地理编码失败:', error)
    // 如果接口解析失败，使用地图反解析的地址作为备用
    supplierForm.address = selectedAddressText.value || ''
    ElMessage.warning('接口解析失败，使用地图解析的地址')
  }

  showAddressPicker.value = false
}

onMounted(() => {
  initData()
})
</script>

<style scoped>
.suppliers-container {
  padding: 20px;
}

.page-title {
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 20px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  margin-bottom: 20px;
}

.suppliers-card {
  margin-top: 20px;
}
</style>

