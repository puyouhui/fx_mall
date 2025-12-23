<template>
  <div class="history-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <div class="header-title">
            <span>供应商备货记录</span>
            <el-tag type="info" size="small" style="margin-left: 10px;">
              共 {{ pagination.total }} 天
            </el-tag>
          </div>
          <div class="header-actions">
            <el-date-picker
              v-model="selectedDate"
              type="date"
              placeholder="选择日期"
              format="YYYY-MM-DD"
              value-format="YYYY-MM-DD"
              @change="handleDateChange"
              style="margin-right: 10px;"
            />
            <el-button 
              type="success" 
              :disabled="selectedRows.length === 0"
              @click="handleDownloadSelected"
              style="margin-right: 10px;"
            >
              <el-icon><Download /></el-icon>
              下载选中报表 ({{ selectedRows.length }})
            </el-button>
            <el-button type="primary" @click="handleRefresh">刷新</el-button>
          </div>
        </div>
      </template>

      <el-table 
        :data="historyList" 
        v-loading="loading" 
        stripe
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" align="center" />
        <el-table-column prop="date" label="日期" align="center">
          <template #default="scope">
            {{ formatDate(scope.row.date) }}
          </template>
        </el-table-column>
        <el-table-column label="待备货" align="center">
          <template #default="scope">
            <el-tag type="warning" size="small">
              {{ scope.row.pending_item_count || 0 }} 件
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="已取货" align="center">
          <template #default="scope">
            <el-tag type="success" size="small">
              {{ scope.row.picked_item_count || 0 }} 件
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="货物总件数" align="center">
          <template #default="scope">
            {{ scope.row.total_item_count || 0 }} 件
          </template>
        </el-table-column>
        <el-table-column label="货物种类数" align="center">
          <template #default="scope">
            {{ scope.row.total_goods_count || 0 }} 种
          </template>
        </el-table-column>
        <el-table-column label="总金额" align="center">
          <template #default="scope">
            <span class="cost-price">¥{{ formatPrice(scope.row.total_amount) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="已取货金额" align="center">
          <template #default="scope">
            <span 
              :class="scope.row.picked_amount !== scope.row.total_amount ? 'cost-price-warning' : 'cost-price'"
            >
              ¥{{ formatPrice(scope.row.picked_amount) }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="操作" align="center" fixed="right" width="180">
          <template #default="scope">
            <el-button type="primary" size="small" link @click="handleViewDetail(scope.row)">
              查看详情
            </el-button>
            <el-button type="success" size="small" link @click="handleDownloadSingle(scope.row)">
              下载报表
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination 
          v-model:current-page="pagination.page" 
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]" 
          :total="pagination.total" 
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange" 
          @current-change="handlePageChange" 
        />
      </div>
    </el-card>

    <!-- 历史详情抽屉 -->
    <el-drawer v-model="detailDrawerVisible" title="历史详情" :size="800" direction="rtl">
      <div v-if="currentHistory" class="history-detail">
        <!-- 日期信息 -->
        <div class="detail-section">
          <h3>日期信息</h3>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="日期">
              {{ formatDate(currentHistory.date) }}
            </el-descriptions-item>
            <el-descriptions-item label="待备货件数">
              <el-tag type="warning">{{ currentHistory.pending_item_count || 0 }} 件</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="已取货件数">
              <el-tag type="success">{{ currentHistory.picked_item_count || 0 }} 件</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="货物总件数">
              {{ currentHistory.total_item_count || 0 }} 件
            </el-descriptions-item>
            <el-descriptions-item label="货物种类数">
              {{ currentHistory.total_goods_count || 0 }} 种
            </el-descriptions-item>
            <el-descriptions-item label="总金额">
              <span class="cost-price">¥{{ formatPrice(currentHistory.total_amount) }}</span>
            </el-descriptions-item>
          </el-descriptions>
        </div>

        <!-- 待备货货物明细 -->
        <div v-if="currentHistory.pending_goods && currentHistory.pending_goods.length > 0" class="detail-section">
          <h3>待备货货物明细</h3>
          <el-table :data="currentHistory.pending_goods" border style="width: 100%">
            <el-table-column label="商品图片" width="100" align="center">
              <template #default="scope">
                <el-image 
                  v-if="scope.row.image" 
                  :src="scope.row.image" 
                  fit="cover"
                  style="width: 60px; height: 60px; border-radius: 4px;" 
                  :preview-src-list="[scope.row.image]"
                  :preview-teleported="true" 
                />
                <span v-else class="no-image">暂无图片</span>
              </template>
            </el-table-column>
            <el-table-column prop="product_name" label="商品名称" min-width="150" align="center" show-overflow-tooltip />
            <el-table-column prop="spec_name" label="规格" width="120" align="center" />
            <el-table-column label="数量" width="80" align="center">
              <template #default="scope">
                {{ scope.row.quantity }} 件
              </template>
            </el-table-column>
            <el-table-column label="价格" width="120" align="center">
              <template #default="scope">
                <span class="cost-price">¥{{ formatPrice(scope.row.cost_price) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="小计" width="120" align="center">
              <template #default="scope">
                <span class="cost-price">¥{{ formatPrice(scope.row.total_cost) }}</span>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <!-- 已取货货物明细 -->
        <div v-if="currentHistory.picked_goods && currentHistory.picked_goods.length > 0" class="detail-section">
          <h3>已取货货物明细</h3>
          <el-table :data="currentHistory.picked_goods" border style="width: 100%">
            <el-table-column label="商品图片" width="100" align="center">
              <template #default="scope">
                <el-image 
                  v-if="scope.row.image" 
                  :src="scope.row.image" 
                  fit="cover"
                  style="width: 60px; height: 60px; border-radius: 4px;" 
                  :preview-src-list="[scope.row.image]"
                  :preview-teleported="true" 
                />
                <span v-else class="no-image">暂无图片</span>
              </template>
            </el-table-column>
            <el-table-column prop="product_name" label="商品名称" min-width="150" align="center" show-overflow-tooltip />
            <el-table-column prop="spec_name" label="规格" width="120" align="center" />
            <el-table-column label="数量" width="80" align="center">
              <template #default="scope">
                {{ scope.row.quantity }} 件
              </template>
            </el-table-column>
            <el-table-column label="价格" width="120" align="center">
              <template #default="scope">
                <span class="cost-price">¥{{ formatPrice(scope.row.cost_price) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="小计" width="120" align="center">
              <template #default="scope">
                <span class="cost-price">¥{{ formatPrice(scope.row.total_cost) }}</span>
              </template>
            </el-table-column>
          </el-table>
        </div>
        <div v-if="(!currentHistory.pending_goods || currentHistory.pending_goods.length === 0) && 
                    (!currentHistory.picked_goods || currentHistory.picked_goods.length === 0)" 
             class="empty-data">
          暂无货物明细
        </div>
      </div>
      <div v-else class="loading-container">
        <el-skeleton :rows="5" animated />
      </div>
    </el-drawer>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Download } from '@element-plus/icons-vue'
import { getHistoryByDate, getHistoryDetail } from '../api/history'
import * as XLSX from 'xlsx-js-style'

const loading = ref(false)
const historyList = ref([])
const selectedDate = ref('')
const detailDrawerVisible = ref(false)
const currentHistory = ref(null)
const selectedRows = ref([])

const pagination = ref({
  page: 1,
  pageSize: 20,
  total: 0
})

// 加载历史记录列表
const loadHistory = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.value.page,
      page_size: pagination.value.pageSize
    }

    if (selectedDate.value) {
      params.date = selectedDate.value
    }

    const response = await getHistoryByDate(params)

    if (response.code === 200 && response.data) {
      historyList.value = response.data.list || []
      pagination.value.total = response.data.total || 0
    } else {
      ElMessage.error(response.message || '获取历史记录失败')
    }
  } catch (error) {
    console.error('获取历史记录失败:', error)
    ElMessage.error(error.message || '获取历史记录失败，请稍后再试')
  } finally {
    loading.value = false
  }
}

// 日期改变
const handleDateChange = () => {
  pagination.value.page = 1
  loadHistory()
}

// 刷新
const handleRefresh = () => {
  selectedDate.value = ''
  pagination.value.page = 1
  loadHistory()
}

// 查看详情
const handleViewDetail = async (history) => {
  detailDrawerVisible.value = true
  currentHistory.value = null

  try {
    const response = await getHistoryDetail(history.date)
    if (response.code === 200 && response.data) {
      currentHistory.value = response.data
    } else {
      ElMessage.error(response.message || '获取历史详情失败')
      // 如果获取详情失败，使用列表中的基本信息
      currentHistory.value = {
        ...history,
        pending_goods: [],
        picked_goods: []
      }
    }
  } catch (error) {
    console.error('获取历史详情失败:', error)
    ElMessage.error(error.message || '获取历史详情失败，请稍后再试')
    // 如果获取详情失败，使用列表中的基本信息
    currentHistory.value = {
      ...history,
      pending_goods: [],
      picked_goods: []
    }
  }
}

// 分页大小改变
const handleSizeChange = (size) => {
  pagination.value.pageSize = size
  pagination.value.page = 1
  loadHistory()
}

// 页码改变
const handlePageChange = (page) => {
  pagination.value.page = page
  loadHistory()
}

// 格式化价格
const formatPrice = (price) => {
  if (price === undefined || price === null) return '0.00'
  return Number(price).toFixed(2)
}

// 格式化日期
const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  // 如果已经是格式化的日期字符串，直接返回
  if (dateStr.includes('-')) {
    return dateStr
  }
  // 如果是日期对象，格式化
  const date = new Date(dateStr)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

// 选择改变
const handleSelectionChange = (selection) => {
  selectedRows.value = selection
}

// 下载单个报表
const handleDownloadSingle = async (row) => {
  try {
    // 获取该日期的详细数据
    const response = await getHistoryDetail(row.date)
    if (response.code === 200 && response.data) {
      const historyData = response.data
      generateExcel([historyData], `历史记录_${row.date}`)
      ElMessage.success('报表下载成功')
    } else {
      ElMessage.error('获取数据失败')
    }
  } catch (error) {
    console.error('下载报表失败:', error)
    ElMessage.error('下载报表失败，请稍后再试')
  }
}

// 下载选中的报表
const handleDownloadSelected = async () => {
  if (selectedRows.value.length === 0) {
    ElMessage.warning('请先选择要下载的历史记录')
    return
  }

  try {
    // 获取所有选中日期的详细数据，使用并行请求提高效率
    const dateToRowMap = new Map()
    selectedRows.value.forEach(row => {
      dateToRowMap.set(row.date, row)
    })
    
    const promises = selectedRows.value.map(row => 
      getHistoryDetail(row.date).then(response => ({
        date: row.date,
        response
      })).catch(error => ({
        date: row.date,
        error
      }))
    )
    
    const results = await Promise.all(promises)
    
    const historyDataList = []
    for (const result of results) {
      if (result.error) {
        console.error(`获取日期 ${result.date} 的详情失败:`, result.error)
        continue
      }
      
      const { response } = result
      if (response.code === 200 && response.data) {
        // 确保每个日期的数据都包含正确的日期和统计信息
        historyDataList.push({
          ...response.data,
          date: result.date // 确保使用正确的日期
        })
      } else {
        console.warn(`获取日期 ${result.date} 的详情失败:`, response.message)
      }
    }

    if (historyDataList.length === 0) {
      ElMessage.error('获取数据失败，请检查选中的日期是否有数据')
      return
    }

    // 按日期排序，确保顺序正确
    historyDataList.sort((a, b) => {
      return new Date(a.date) - new Date(b.date)
    })

    const dateRange = selectedRows.value.length === 1
      ? selectedRows.value[0].date
      : `${selectedRows.value[selectedRows.value.length - 1].date}_至_${selectedRows.value[0].date}`
    
    generateExcel(historyDataList, `供应商备货记录_${dateRange}`)
    ElMessage.success(`成功下载 ${historyDataList.length} 条供应商备货记录报表`)
  } catch (error) {
    console.error('下载报表失败:', error)
    ElMessage.error('下载报表失败，请稍后再试')
  }
}

// 生成Excel文件
const generateExcel = (historyDataList, filename) => {
  // 创建工作簿
  const wb = XLSX.utils.book_new()

  // 创建一个工作表，包含所有日期的数据
  const allData = []
  
  // 总标题行
  allData.push(['供应商备货记录报表'])
  allData.push([])

  // 循环每个日期，添加数据
  historyDataList.forEach((historyData, index) => {
    // 如果不是第一个日期，添加分隔行（使用特殊标记，后续会用黄色背景显示）
    if (index > 0) {
      allData.push(['__SEPARATOR__']) // 使用特殊标记标识分隔行
    }
    
    // 汇总信息区域（表格形式，日期包含在标题中）
    allData.push([`汇总信息 - ${historyData.date}`])
    allData.push(['项目', '数值'])
    allData.push(['待备货件数', `${historyData.pending_item_count || 0} 件`])
    allData.push(['已取货件数', `${historyData.picked_item_count || 0} 件`])
    allData.push(['货物总件数', `${historyData.total_item_count || 0} 件`])
    allData.push(['货物种类数', `${historyData.total_goods_count || 0} 种`])
    allData.push(['总金额', `¥${formatPrice(historyData.total_amount)}`])
    allData.push(['已取货金额', `¥${formatPrice(historyData.picked_amount)}`])
    allData.push([])
    
    // 待备货货物明细
    allData.push(['待备货明细'])
    allData.push(['序号', '商品名称', '规格', '数量', '价格', '小计'])
    
    if (historyData.pending_goods && historyData.pending_goods.length > 0) {
      let pendingTotal = 0
      historyData.pending_goods.forEach((goods, idx) => {
        const quantity = goods.quantity || 0
        const costPrice = goods.cost_price || 0
        const totalCost = goods.total_cost || 0
        pendingTotal += totalCost
        
        allData.push([
          idx + 1,
          goods.product_name || '',
          goods.spec_name || '',
          `${quantity} 件`,
          `¥${formatPrice(costPrice)}`,
          `¥${formatPrice(totalCost)}`
        ])
      })
      // 待备货小计
      allData.push(['', '', '', '合计', '', `¥${formatPrice(pendingTotal)}`])
    } else {
      allData.push(['暂无数据', '', '', '', '', ''])
    }
    
    allData.push([])
    
    // 已取货货物明细
    allData.push(['已取货明细'])
    allData.push(['序号', '商品名称', '规格', '数量', '成本价', '小计'])
    
    if (historyData.picked_goods && historyData.picked_goods.length > 0) {
      let pickedTotal = 0
      historyData.picked_goods.forEach((goods, idx) => {
        const quantity = goods.quantity || 0
        const costPrice = goods.cost_price || 0
        const totalCost = goods.total_cost || 0
        pickedTotal += totalCost
        
        allData.push([
          idx + 1,
          goods.product_name || '',
          goods.spec_name || '',
          `${quantity} 件`,
          `¥${formatPrice(costPrice)}`,
          `¥${formatPrice(totalCost)}`
        ])
      })
      // 已取货小计
      allData.push(['', '', '', '合计', '', `¥${formatPrice(pickedTotal)}`])
    } else {
      allData.push(['暂无数据', '', '', '', '', ''])
    }

  })

  // 创建工作表
  const ws = XLSX.utils.aoa_to_sheet(allData)

    // 设置列宽（优化宽度，第一列序号更宽）
    ws['!cols'] = [
      { wch: 15 }, // 序号（进一步增加宽度）
      { wch: 35 }, // 商品名称（进一步增加宽度，便于显示完整名称）
      { wch: 22 }, // 规格（增加宽度）
      { wch: 15 }, // 数量（增加宽度）
      { wch: 18 }, // 成本价/价格（增加宽度）
      { wch: 18 }  // 小计（增加宽度）
    ]

  // 设置行高
  ws['!rows'] = []
  const totalRows = allData.length
  for (let i = 0; i < totalRows; i++) {
    const rowData = allData[i]
    if (i === 0) {
      ws['!rows'][i] = { hpt: 30 } // 总标题行高度
    } else if (rowData && rowData[0] !== undefined) {
      const firstCell = String(rowData[0] || '')
      if (firstCell === '__SEPARATOR__') {
        ws['!rows'][i] = { hpt: 15 } // 分隔行高度
      } else if ((firstCell && firstCell.startsWith('汇总信息')) || firstCell === '待备货明细' || firstCell === '已取货明细') {
        ws['!rows'][i] = { hpt: 25 } // 区域标题行高度
      } else if (firstCell === '项目' || firstCell === '序号' || (firstCell === '商品名称' && rowData[1] === '规格')) {
        ws['!rows'][i] = { hpt: 22 } // 表头行高度
      } else {
        ws['!rows'][i] = { hpt: 20 } // 普通行高度
      }
    } else {
      ws['!rows'][i] = { hpt: 20 } // 普通行高度
    }
  }

    // 定义边框样式
    const borderStyle = {
      top: { style: 'thin', color: { rgb: '000000' } },
      bottom: { style: 'thin', color: { rgb: '000000' } },
      left: { style: 'thin', color: { rgb: '000000' } },
      right: { style: 'thin', color: { rgb: '000000' } }
    }

    // 定义标题样式
    const titleStyle = {
      font: { bold: true, sz: 16, color: { rgb: '000000' } },
      alignment: { horizontal: 'center', vertical: 'center' },
      fill: { fgColor: { rgb: 'E6E6FA' } },
      border: borderStyle
    }

    // 定义表头样式
    const headerStyle = {
      font: { bold: true, sz: 11, color: { rgb: 'FFFFFF' } },
      alignment: { horizontal: 'center', vertical: 'center' },
      fill: { fgColor: { rgb: '4472C4' } },
      border: borderStyle
    }

    // 定义区域标题样式
    const sectionStyle = {
      font: { bold: true, sz: 12, color: { rgb: '000000' } },
      alignment: { horizontal: 'center', vertical: 'center' },
      fill: { fgColor: { rgb: 'D9E1F2' } },
      border: borderStyle
    }

    // 定义数据样式（全部居中）
    const dataStyle = {
      alignment: { horizontal: 'center', vertical: 'center' },
      border: borderStyle
    }

    // 定义数字样式（居中）
    const numberStyle = {
      alignment: { horizontal: 'center', vertical: 'center' },
      border: borderStyle
    }

    // 定义汇总标签样式（加粗）
    const summaryLabelStyle = {
      font: { bold: true, sz: 11, color: { rgb: '000000' } },
      alignment: { horizontal: 'center', vertical: 'center' },
      fill: { fgColor: { rgb: 'F0F0F0' } },
      border: borderStyle
    }

    // 定义汇总数值样式（居中）
    const summaryValueStyle = {
      font: { bold: false, sz: 11 },
      alignment: { horizontal: 'center', vertical: 'center' },
      border: borderStyle
    }

    // 定义汇总表头样式
    const summaryHeaderStyle = {
      font: { bold: true, sz: 11, color: { rgb: 'FFFFFF' } },
      alignment: { horizontal: 'center', vertical: 'center' },
      fill: { fgColor: { rgb: '4472C4' } },
      border: borderStyle
    }

    // 定义小计样式（居中）
    const totalStyle = {
      font: { bold: true, sz: 11, color: { rgb: '000000' } },
      alignment: { horizontal: 'center', vertical: 'center' },
      fill: { fgColor: { rgb: 'F2F2F2' } },
      border: borderStyle
    }

    // 定义分隔行样式（黄色背景）
    const separatorStyle = {
      alignment: { horizontal: 'center', vertical: 'center' },
      fill: { fgColor: { rgb: 'FFFF00' } }, // 黄色背景
      border: borderStyle
    }

  // 应用样式到所有单元格
  const range = XLSX.utils.decode_range(ws['!ref'])
  for (let R = range.s.r; R <= range.e.r; ++R) {
    for (let C = range.s.c; C <= range.e.c; ++C) {
      const cellAddress = XLSX.utils.encode_cell({ r: R, c: C })
      if (!ws[cellAddress]) ws[cellAddress] = { t: 's', v: '' }
      
      const rowData = allData[R]
      const cellValue = rowData && rowData[C] !== undefined ? String(rowData[C] || '') : ''
      const firstCellValue = rowData && rowData[0] !== undefined ? String(rowData[0] || '') : ''
      
      // 总标题行（第0行）
      if (R === 0) {
        ws[cellAddress].s = titleStyle
      }
      // 分隔行（使用黄色背景）
      else if (firstCellValue === '__SEPARATOR__') {
        // 清空单元格内容，只显示黄色背景
        ws[cellAddress].v = ''
        ws[cellAddress].t = 's'
        ws[cellAddress].s = separatorStyle
      }
      // 汇总信息标题（包含日期）
      else if (firstCellValue && firstCellValue.startsWith('汇总信息')) {
        ws[cellAddress].s = sectionStyle
      }
      // 汇总信息表头（项目、数值）
      else if (cellValue === '项目' || cellValue === '数值') {
        ws[cellAddress].s = summaryHeaderStyle
      }
      // 汇总信息数据行（待备货件数等，不再包含日期行）
      else if (firstCellValue && ['待备货件数', '已取货件数', '货物总件数', '货物种类数', '总金额', '已取货金额'].includes(firstCellValue)) {
        if (C === 0) {
          ws[cellAddress].s = summaryLabelStyle
        } else if (C === 1) {
          ws[cellAddress].s = summaryValueStyle
        } else {
          ws[cellAddress].s = dataStyle
        }
      }
      // 待备货明细标题
      else if (cellValue === '待备货明细') {
        ws[cellAddress].s = sectionStyle
      }
      // 已取货明细标题
      else if (cellValue === '已取货明细') {
        ws[cellAddress].s = sectionStyle
      }
      // 表头行（序号、商品名称等）
      else if (cellValue === '序号' || cellValue === '商品名称' || cellValue === '规格' || cellValue === '数量' || cellValue === '价格' || cellValue === '成本价' || cellValue === '小计') {
        ws[cellAddress].s = headerStyle
      }
      // 合计行
      else if (cellValue === '合计') {
        ws[cellAddress].s = totalStyle
      }
      // 暂无数据
      else if (cellValue === '暂无数据') {
        ws[cellAddress].s = dataStyle
      }
      // 其他数据行
      else {
        ws[cellAddress].s = dataStyle
      }
    }
  }

  // 合并单元格
  if (!ws['!merges']) ws['!merges'] = []
  
  // 总标题行合并
  ws['!merges'].push({ s: { r: 0, c: 0 }, e: { r: 0, c: 5 } })
  
  // 遍历所有行，合并相应的单元格
  for (let R = 0; R < allData.length; R++) {
    const rowData = allData[R]
    if (!rowData || rowData.length === 0) continue
    
    const firstCell = String(rowData[0] || '')
    
    // 汇总信息标题合并（包含日期）
    if (firstCell && firstCell.startsWith('汇总信息')) {
      ws['!merges'].push({ s: { r: R, c: 0 }, e: { r: R, c: 5 } })
    }
    // 待备货明细标题合并
    else if (firstCell === '待备货明细') {
      ws['!merges'].push({ s: { r: R, c: 0 }, e: { r: R, c: 5 } })
    }
    // 已取货明细标题合并
    else if (firstCell === '已取货明细') {
      ws['!merges'].push({ s: { r: R, c: 0 }, e: { r: R, c: 5 } })
    }
    // 分隔行合并（黄色背景）
    else if (firstCell === '__SEPARATOR__') {
      ws['!merges'].push({ s: { r: R, c: 0 }, e: { r: R, c: 5 } })
    }
    // 汇总信息数据行：合并第2-5列（不再包含日期行）
    else if (firstCell && typeof firstCell === 'string' && ['待备货件数', '已取货件数', '货物总件数', '货物种类数', '总金额', '已取货金额'].includes(firstCell)) {
      ws['!merges'].push({ s: { r: R, c: 2 }, e: { r: R, c: 5 } })
    }
  }

  // 添加工作表到工作簿
  const sheetName = '供应商备货记录'
  XLSX.utils.book_append_sheet(wb, ws, sheetName)

  // 下载文件
  XLSX.writeFile(wb, `${filename}.xlsx`)
}

onMounted(() => {
  loadHistory()
})
</script>

<style scoped>
.history-page {
  padding: 0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-title {
  display: flex;
  align-items: center;
}

.header-actions {
  display: flex;
  align-items: center;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.cost-price {
  color: #409eff;
  font-weight: 500;
}

.cost-price-warning {
  color: #f56c6c;
  font-weight: 500;
}

.history-detail {
  padding: 0;
}

.detail-section {
  margin-bottom: 30px;
}

.detail-section h3 {
  margin-bottom: 15px;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  border-left: 4px solid #409eff;
  padding-left: 10px;
}

.no-image {
  color: #909399;
  font-size: 12px;
}

.empty-data {
  text-align: center;
  padding: 40px;
  color: #909399;
  font-size: 16px;
}

.loading-container {
  padding: 20px;
}
</style>
