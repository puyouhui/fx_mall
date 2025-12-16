<template>
  <div class="orders-page">
    <el-card class="orders-card" shadow="never">
      <div class="card-header">
        <div class="title">
          <span class="main">订单管理</span>
          <span class="sub">查看和管理所有订单</span>
        </div>
        <div class="actions">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索订单ID / 用户ID"
            clearable
            @keyup.enter="handleSearch"
            style="width: 200px; margin-right: 10px;"
          />
          <el-select
            v-model="statusFilter"
            placeholder="订单状态"
            clearable
            style="width: 150px; margin-right: 10px;"
            @change="handleSearch"
          >
            <el-option label="待配送" value="pending_delivery" />
            <el-option label="待取货" value="pending_pickup" />
            <el-option label="配送中" value="delivering" />
            <el-option label="已送达" value="delivered" />
            <el-option label="已收款" value="paid" />
            <el-option label="已取消" value="cancelled" />
          </el-select>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
        </div>
      </div>

      <el-table
        v-loading="loading"
        :data="orders"
        border
        stripe
        class="orders-table"
        empty-text="暂无订单数据"
        row-key="id"
      >
        <!-- <el-table-column prop="id" label="订单ID" width="100" /> -->
        <el-table-column prop="order_number" label="订单编号" width="180" />
        <el-table-column label="用户信息" min-width="180">
          <template #default="scope">
            <div v-if="scope.row.user">
              <div>{{ scope.row.user.name || '未命名' }}</div>
              <div style="color: #909399; font-size: 12px;">用户{{ scope.row.user.user_code || scope.row.user_id }}</div>
            </div>
            <span v-else>用户ID: {{ scope.row.user_id }}</span>
          </template>
        </el-table-column>
        <el-table-column label="销售员" width="120">
          <template #default="scope">
            <div v-if="scope.row.user && scope.row.user.sales_employee">
              <el-tag size="small" type="info">
                {{ scope.row.user.sales_employee.name || scope.row.user.sales_employee.employee_code }}
              </el-tag>
              <div v-if="scope.row.user.sales_employee.employee_code" style="color: #909399; font-size: 11px; margin-top: 2px;">
                {{ scope.row.user.sales_employee.employee_code }}
              </div>
            </div>
            <span v-else style="color: #c0c4cc;">-</span>
          </template>
        </el-table-column>
        <el-table-column label="收货地址" min-width="200">
          <template #default="scope">
            <div v-if="scope.row.address">
              <div>{{ scope.row.address.name || '-' }}</div>
              <div style="color: #909399; font-size: 12px;">{{ scope.row.address.address || '-' }}</div>
            </div>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="订单状态" width="120">
          <template #default="scope">
            <el-tag :type="getStatusType(scope.row.status)">
              {{ formatStatus(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="商品件数" width="120" align="center">
          <template #default="scope">
            <el-button 
              type="primary" 
              link 
              @click="handleViewOrderItems(scope.row.id)"
              :disabled="!scope.row.item_count || scope.row.item_count === 0"
            >
              {{ scope.row.item_count || 0 }} 件
            </el-button>
          </template>
        </el-table-column>
        <el-table-column label="金额信息" min-width="150">
          <template #default="scope">
            <div style="color: #ff4d4f; font-weight: 600; font-size: 14px;">
              实付: ¥{{ formatMoney(scope.row.total_amount) }}
            </div>
            <div style="color: #909399; font-size: 12px; margin-top: 4px;">
              <span v-if="scope.row.delivery_fee > 0">
                配送费: ¥{{ formatMoney(scope.row.delivery_fee) }}
              </span>
              <span v-else style="color: #67c23a;">
                配送费: 免费配送
              </span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="销售分成" width="150" align="center">
          <template #default="scope">
            <!-- 已收款订单：显示已计入的分成 -->
            <div v-if="scope.row.status === 'paid' && scope.row.sales_commission">
              <div style="color: #409eff; font-weight: 600; font-size: 14px;">
                已计入: ¥{{ formatMoney(scope.row.sales_commission.total_commission) }}
              </div>
              <div v-if="!scope.row.sales_commission.is_valid_order" style="color: #909399; font-size: 11px; margin-top: 2px;">
                无效订单
              </div>
              <!-- 如果预览和已计入不一致，显示预览值 -->
              <div v-if="scope.row.sales_commission_preview && 
                         Math.abs(scope.row.sales_commission_preview.total_commission - scope.row.sales_commission.total_commission) > 0.01" 
                   style="color: #909399; font-size: 11px; margin-top: 2px;">
                预览: ¥{{ formatMoney(scope.row.sales_commission_preview.total_commission) }}
              </div>
            </div>
            <!-- 未收款订单：显示预览分成 -->
            <div v-else-if="scope.row.sales_commission_preview">
              <div style="color: #e6a23c; font-weight: 600; font-size: 14px;">
                预计: ¥{{ formatMoney(scope.row.sales_commission_preview.total_commission) }}
              </div>
              <div style="color: #909399; font-size: 11px; margin-top: 2px;">
                预览（收款后计入）
              </div>
              <div v-if="!scope.row.sales_commission_preview.is_valid_order" style="color: #f56c6c; font-size: 11px; margin-top: 2px;">
                无效订单
              </div>
            </div>
            <span v-else style="color: #c0c4cc;">-</span>
          </template>
        </el-table-column>
        <el-table-column label="配送员和配送费" width="180" align="center">
          <template #default="scope">
            <div v-if="scope.row.delivery_employee">
              <div style="font-weight: 600; font-size: 14px; color: #606266;">
                {{ scope.row.delivery_employee.name || scope.row.delivery_employee.employee_code }}
              </div>
              <div v-if="scope.row.delivery_employee.employee_code" style="color: #909399; font-size: 11px; margin-top: 2px;">
                {{ scope.row.delivery_employee.employee_code }}
              </div>
              <div v-if="scope.row.rider_payable_fee && scope.row.rider_payable_fee > 0" style="color: #67c23a; font-weight: 600; font-size: 13px; margin-top: 4px;">
                配送费: ¥{{ formatMoney(scope.row.rider_payable_fee) }}
              </div>
            </div>
            <span v-else style="color: #c0c4cc;">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="下单时间" min-width="160">
          <template #default="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right" align="left">
          <template #default="scope">
            <div class="action-buttons">
              <el-button type="primary" link @click="handleViewDetail(scope.row.id)">
                详情
              </el-button>
              <el-dropdown 
                v-if="canShowStatusActions(scope.row.status)"
                @command="(cmd) => handleStatusChange(scope.row.id, scope.row.status, cmd)"
                trigger="click"
                placement="bottom-end"
              >
                <el-button type="primary" link>
                  状态操作
                  <el-icon style="margin-left: 4px;"><ArrowDown /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item 
                      v-if="isPendingDelivery(scope.row.status)" 
                      command="delivering"
                    >
                      开始配送
                    </el-dropdown-item>
                    <el-dropdown-item 
                      v-if="scope.row.status === 'delivering'" 
                      command="delivered"
                    >
                      标记已送达
                    </el-dropdown-item>
                    <el-dropdown-item 
                      v-if="scope.row.status === 'delivered' || scope.row.status === 'shipped'" 
                      command="paid"
                    >
                      标记已收款
                    </el-dropdown-item>
                    <el-dropdown-item 
                      v-if="isPendingDelivery(scope.row.status)" 
                      command="cancelled"
                      divided
                    >
                      取消订单
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          background
          layout="total, prev, pager, next, jumper"
          :page-size="pagination.pageSize"
          :current-page="pagination.pageNum"
          :total="pagination.total"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>

    <!-- 订单详情对话框 -->
    <el-dialog
      v-model="detailDialogVisible"
      title="订单详情"
      width="900px"
      destroy-on-close
    >
      <div v-loading="detailLoading" v-if="orderDetail">
        <el-tabs v-model="activeTab" type="border-card">
          <!-- 基本信息标签页 -->
          <el-tab-pane label="基本信息" name="basic">
        <!-- 订单基本信息 -->
            <el-descriptions :column="2" border style="margin-bottom: 20px;">
          <el-descriptions-item label="订单ID">{{ orderDetail.order?.id }}</el-descriptions-item>
          <el-descriptions-item label="订单编号">{{ orderDetail.order?.order_number || '-' }}</el-descriptions-item>
          <el-descriptions-item label="订单状态">
            <el-tag :type="getStatusType(orderDetail.order?.status)">
              {{ formatStatus(orderDetail.order?.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="下单时间">{{ formatDate(orderDetail.order?.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="更新时间">{{ formatDate(orderDetail.order?.updated_at) }}</el-descriptions-item>
        </el-descriptions>

        <!-- 用户信息 -->
        <el-divider content-position="left">用户信息</el-divider>
            <el-descriptions :column="2" border style="margin-bottom: 20px;" v-if="orderDetail.user">
          <el-descriptions-item label="用户ID">{{ orderDetail.user.id }}</el-descriptions-item>
          <el-descriptions-item label="用户编号">用户{{ orderDetail.user.user_code || '-' }}</el-descriptions-item>
          <el-descriptions-item label="姓名">{{ orderDetail.user.name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="手机号">{{ orderDetail.user.phone || '-' }}</el-descriptions-item>
          <el-descriptions-item label="用户类型">
            <el-tag :type="orderDetail.user.user_type === 'wholesale' ? 'warning' : 'success'">
              {{ orderDetail.user.user_type === 'wholesale' ? '批发用户' : '零售用户' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="销售员" v-if="orderDetail.user.sales_employee">
            <el-tag type="info">
              {{ orderDetail.user.sales_employee.name || orderDetail.user.sales_employee.employee_code }}
              <span v-if="orderDetail.user.sales_employee.employee_code" style="margin-left: 4px;">
                ({{ orderDetail.user.sales_employee.employee_code }})
              </span>
            </el-tag>
          </el-descriptions-item>
        </el-descriptions>

        <!-- 收货地址 -->
        <el-divider content-position="left">收货地址</el-divider>
            <el-descriptions :column="2" border style="margin-bottom: 20px;" v-if="orderDetail.address">
          <el-descriptions-item label="地址名称">{{ orderDetail.address.name || '-' }}</el-descriptions-item>
          <el-descriptions-item label="联系人">{{ orderDetail.address.contact || '-' }}</el-descriptions-item>
          <el-descriptions-item label="手机号">{{ orderDetail.address.phone || '-' }}</el-descriptions-item>
          <el-descriptions-item label="详细地址" :span="2">{{ orderDetail.address.address || '-' }}</el-descriptions-item>
        </el-descriptions>

        <!-- 订单明细 -->
        <el-divider content-position="left">订单明细</el-divider>
            <el-table :data="orderDetail.order_items" border stripe style="margin-bottom: 20px;">
          <el-table-column prop="product_name" label="商品名称" min-width="150" />
          <el-table-column prop="spec_name" label="规格" width="120" />
          <el-table-column prop="quantity" label="数量" width="80" align="center" />
          <el-table-column prop="unit_price" label="单价" width="100" align="right">
            <template #default="scope">
              ¥{{ scope.row.unit_price?.toFixed(2) || '0.00' }}
            </template>
          </el-table-column>
          <el-table-column prop="subtotal" label="小计" width="100" align="right">
            <template #default="scope">
              ¥{{ scope.row.subtotal?.toFixed(2) || '0.00' }}
            </template>
          </el-table-column>
        </el-table>

            <!-- 其他信息 -->
            <el-divider content-position="left">其他信息</el-divider>
            <el-descriptions :column="2" border>
              <el-descriptions-item label="备注">{{ orderDetail.order?.remark || '-' }}</el-descriptions-item>
              <el-descriptions-item label="缺货处理">
                {{ formatOutOfStockStrategy(orderDetail.order?.out_of_stock_strategy) }}
              </el-descriptions-item>
              <el-descriptions-item label="信任签收">
                <el-tag :type="orderDetail.order?.trust_receipt ? 'success' : 'info'">
                  {{ orderDetail.order?.trust_receipt ? '是' : '否' }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="隐藏价格">
                <el-tag :type="orderDetail.order?.hide_price ? 'warning' : 'info'">
                  {{ orderDetail.order?.hide_price ? '是' : '否' }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="要求电话联系">
                <el-tag :type="orderDetail.order?.require_phone_contact ? 'success' : 'info'">
                  {{ orderDetail.order?.require_phone_contact ? '是' : '否' }}
                </el-tag>
              </el-descriptions-item>
            </el-descriptions>
          </el-tab-pane>

          <!-- 金额信息标签页 -->
          <el-tab-pane label="金额信息" name="amount">
        <!-- 金额汇总 -->
        <el-divider content-position="left">金额汇总</el-divider>
            <el-descriptions :column="1" border style="margin-bottom: 20px;">
          <el-descriptions-item label="商品金额">
            ¥{{ orderDetail.order?.goods_amount?.toFixed(2) || '0.00' }}
          </el-descriptions-item>
          <el-descriptions-item label="配送费">
            ¥{{ orderDetail.order?.delivery_fee?.toFixed(2) || '0.00' }}
          </el-descriptions-item>
              <el-descriptions-item label="加急费" v-if="orderDetail.order?.is_urgent && (orderDetail.order?.urgent_fee || 0) > 0">
                <el-tag type="danger" size="small">¥{{ orderDetail.order?.urgent_fee?.toFixed(2) || '0.00' }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="积分抵扣" v-if="(orderDetail.order?.points_discount || 0) > 0">
                <span style="color: #f56c6c;">-¥{{ orderDetail.order?.points_discount?.toFixed(2) || '0.00' }}</span>
          </el-descriptions-item>
              <el-descriptions-item label="优惠券抵扣" v-if="(orderDetail.order?.coupon_discount || 0) > 0">
                <span style="color: #f56c6c;">-¥{{ orderDetail.order?.coupon_discount?.toFixed(2) || '0.00' }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="实付金额" label-class-name="total-amount-label">
            <span class="total-amount">¥{{ orderDetail.order?.total_amount?.toFixed(2) || '0.00' }}</span>
          </el-descriptions-item>
        </el-descriptions>

            <!-- 利润信息（简化版） -->
        <el-divider content-position="left">利润信息</el-divider>
            <el-descriptions :column="1" border v-if="orderDetail.simplified_profit && Object.keys(orderDetail.simplified_profit).length > 0">
              <el-descriptions-item label="平台总收入（实付金额）" label-class-name="revenue-label">
                <span class="revenue-amount">¥{{ (orderDetail.simplified_profit.platform_revenue || 0).toFixed(2) }}</span>
                <div style="margin-top: 4px; font-size: 12px; color: #909399;">
                  商品金额(¥{{ (orderDetail.order?.goods_amount || 0).toFixed(2) }}) 
                  + 配送费(¥{{ (orderDetail.order?.delivery_fee || 0).toFixed(2) }}) 
                  <span v-if="(orderDetail.order?.urgent_fee || 0) > 0">+ 加急费(¥{{ (orderDetail.order?.urgent_fee || 0).toFixed(2) }})</span>
                  <span v-if="(orderDetail.order?.coupon_discount || 0) > 0">- 优惠券(¥{{ (orderDetail.order?.coupon_discount || 0).toFixed(2) }})</span>
                  <span v-if="(orderDetail.order?.points_discount || 0) > 0">- 积分(¥{{ (orderDetail.order?.points_discount || 0).toFixed(2) }})</span>
                </div>
              </el-descriptions-item>
              <el-descriptions-item label="商品总成本" label-class-name="cost-label">
                <span class="cost-amount">¥{{ (orderDetail.simplified_profit.goods_cost || 0).toFixed(2) }}</span>
              </el-descriptions-item>
              <el-descriptions-item label="毛利润（平台总收入 - 商品总成本）" label-class-name="gross-profit-label">
                <span class="gross-profit-amount">¥{{ (orderDetail.simplified_profit.gross_profit || 0).toFixed(2) }}</span>
                <div style="margin-top: 4px; font-size: 12px; color: #909399;">
                  = 平台总收入(¥{{ (orderDetail.simplified_profit.platform_revenue || 0).toFixed(2) }}) 
                  - 商品总成本(¥{{ (orderDetail.simplified_profit.goods_cost || 0).toFixed(2) }})
                </div>
          </el-descriptions-item>
              <el-descriptions-item label="配送成本" label-class-name="delivery-cost-label">
                <span class="delivery-cost-amount">¥{{ (orderDetail.simplified_profit.delivery_cost || 0).toFixed(2) }}</span>
          </el-descriptions-item>
              <el-descriptions-item label="净利润（平台总收入 - 商品总成本 - 配送成本）" label-class-name="net-profit-label">
                <span class="net-profit-amount" :class="{ 'profit-positive': (orderDetail.simplified_profit.net_profit || 0) >= 0, 'profit-negative': (orderDetail.simplified_profit.net_profit || 0) < 0 }">
                  ¥{{ (orderDetail.simplified_profit.net_profit || 0).toFixed(2) }}
                </span>
                <div style="margin-top: 4px; font-size: 12px; color: #909399;">
                  = 平台总收入(¥{{ (orderDetail.simplified_profit.platform_revenue || 0).toFixed(2) }}) 
                  - 商品总成本(¥{{ (orderDetail.simplified_profit.goods_cost || 0).toFixed(2) }}) 
                  - 配送成本(¥{{ (orderDetail.simplified_profit.delivery_cost || 0).toFixed(2) }})
                </div>
                <div style="margin-top: 4px; font-size: 12px; font-weight: 600;" :style="{ color: (orderDetail.simplified_profit.net_profit || 0) >= 0 ? '#67c23a' : '#f56c6c' }">
                  {{ (orderDetail.simplified_profit.net_profit || 0) >= 0 ? '✓ 平台盈利' : '✗ 平台亏损' }}
                </div>
          </el-descriptions-item>
        </el-descriptions>
        <el-empty v-else description="利润信息暂不可用" :image-size="80" />
          </el-tab-pane>

          <!-- 配送费详情标签页 -->
          <el-tab-pane label="配送费详情" name="delivery">
        <el-descriptions :column="1" border v-if="orderDetail.delivery_fee_calculation && Object.keys(orderDetail.delivery_fee_calculation).length > 0">
          <el-descriptions-item label="基础配送费">
            ¥{{ (orderDetail.delivery_fee_calculation.base_fee || 0).toFixed(2) }}
          </el-descriptions-item>
          <el-descriptions-item label="孤立订单补贴" v-if="orderDetail.delivery_fee_calculation.isolated_fee > 0">
            <el-tag type="warning" size="small">+¥{{ (orderDetail.delivery_fee_calculation.isolated_fee || 0).toFixed(2) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="件数补贴" v-if="orderDetail.delivery_fee_calculation.item_fee > 0">
            <el-tag type="info" size="small">+¥{{ (orderDetail.delivery_fee_calculation.item_fee || 0).toFixed(2) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="加急订单补贴" v-if="orderDetail.delivery_fee_calculation.urgent_fee > 0">
            <el-tag type="danger" size="small">+¥{{ (orderDetail.delivery_fee_calculation.urgent_fee || 0).toFixed(2) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="极端天气补贴" v-if="orderDetail.delivery_fee_calculation.weather_fee > 0">
            <el-tag type="warning" size="small">+¥{{ (orderDetail.delivery_fee_calculation.weather_fee || 0).toFixed(2) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="配送员实际所得（预估配送费）" label-class-name="rider-fee-label">
            <span class="rider-fee">
              ¥{{ (orderDetail.delivery_fee_calculation.rider_payable_fee || 0).toFixed(2) }}
              <span v-if="orderDetail.delivery_fee_calculation.profit_share > 0" style="color: #67c23a; margin-left: 8px; font-size: 14px;">
                （包含利润分成¥{{ (orderDetail.delivery_fee_calculation.profit_share || 0).toFixed(2) }}）
              </span>
            </span>
          </el-descriptions-item>
          <el-descriptions-item label="利润分成明细" v-if="orderDetail.delivery_fee_calculation.profit_share > 0">
            <el-tag type="success" size="small">+¥{{ (orderDetail.delivery_fee_calculation.profit_share || 0).toFixed(2) }}</el-tag>
            <span style="margin-left: 8px; color: #909399; font-size: 12px;">(已包含在预估配送费中，仅管理员可见)</span>
          </el-descriptions-item>
          <el-descriptions-item label="平台总成本" label-class-name="platform-cost-label">
            <span class="platform-cost">¥{{ (orderDetail.delivery_fee_calculation.total_platform_cost || 0).toFixed(2) }}</span>
          </el-descriptions-item>
        </el-descriptions>
        <el-empty v-else description="配送费计算信息暂不可用" :image-size="80" />
          </el-tab-pane>

          <!-- 销售分成标签页 -->
          <el-tab-pane label="销售分成" name="sales_commission">
            <!-- 已收款订单：显示已计入的分成 -->
            <div v-if="orderDetail.order?.status === 'paid' && orderDetail.sales_commission">
              <el-descriptions :column="1" border>
                <el-descriptions-item label="销售员" v-if="orderDetail.user?.sales_employee">
                  <el-tag type="info">
                    {{ orderDetail.user.sales_employee.name || orderDetail.user.sales_employee.employee_code }}
                    <span v-if="orderDetail.user.sales_employee.employee_code" style="margin-left: 4px;">
                      ({{ orderDetail.user.sales_employee.employee_code }})
                    </span>
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="订单状态">
                  <el-tag :type="orderDetail.sales_commission.is_valid_order ? 'success' : 'info'">
                    {{ orderDetail.sales_commission.is_valid_order ? '有效订单' : '无效订单' }}
                  </el-tag>
                  <span v-if="!orderDetail.sales_commission.is_valid_order" style="color: #909399; margin-left: 8px; font-size: 12px;">
                    (订单利润不满足最小阈值)
                  </span>
                </el-descriptions-item>
                <el-descriptions-item label="是否新客户首单">
                  <el-tag :type="orderDetail.sales_commission.is_new_customer_order ? 'warning' : 'info'">
                    {{ orderDetail.sales_commission.is_new_customer_order ? '是' : '否' }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="结算状态">
                  <el-tag :type="orderDetail.sales_commission.is_settled ? 'success' : 'info'">
                    {{ orderDetail.sales_commission.is_settled ? '已结算' : '未结算' }}
                  </el-tag>
                  <span v-if="orderDetail.sales_commission.settlement_date" style="color: #909399; margin-left: 8px; font-size: 12px;">
                    {{ formatDate(orderDetail.sales_commission.settlement_date) }}
                  </span>
                </el-descriptions-item>
                <el-descriptions-item label="计算月份">
                  {{ orderDetail.sales_commission.calculation_month || '-' }}
                </el-descriptions-item>
                <el-descriptions-item label="订单金额（平台总收入）">
                  ¥{{ (orderDetail.sales_commission.order_amount || 0).toFixed(2) }}
                </el-descriptions-item>
                <el-descriptions-item label="商品总成本">
                  ¥{{ (orderDetail.sales_commission.goods_cost || 0).toFixed(2) }}
                </el-descriptions-item>
                <el-descriptions-item label="配送成本">
                  ¥{{ (orderDetail.sales_commission.delivery_cost || 0).toFixed(2) }}
                </el-descriptions-item>
                <el-descriptions-item label="订单利润">
                  <span style="font-weight: 600; color: #409eff;">
                    ¥{{ (orderDetail.sales_commission.order_profit || 0).toFixed(2) }}
                  </span>
                  <div style="margin-top: 4px; font-size: 12px; color: #909399;">
                    = 订单金额(¥{{ (orderDetail.sales_commission.order_amount || 0).toFixed(2) }}) 
                    - 商品总成本(¥{{ (orderDetail.sales_commission.goods_cost || 0).toFixed(2) }}) 
                    - 配送成本(¥{{ (orderDetail.sales_commission.delivery_cost || 0).toFixed(2) }})
                  </div>
                </el-descriptions-item>
                <el-descriptions-item label="基础提成（45%）">
                  <span style="color: #606266;">
                    ¥{{ (orderDetail.sales_commission.base_commission || 0).toFixed(2) }}
                  </span>
                  <div style="margin-top: 4px; font-size: 12px; color: #909399;">
                    = 订单利润(¥{{ (orderDetail.sales_commission.order_profit || 0).toFixed(2) }}) × 45%
                  </div>
                </el-descriptions-item>
                <el-descriptions-item label="新客开发激励（20%）" v-if="orderDetail.sales_commission.is_new_customer_order && (orderDetail.sales_commission.new_customer_bonus || 0) > 0">
                  <el-tag type="warning" size="small">
                    +¥{{ (orderDetail.sales_commission.new_customer_bonus || 0).toFixed(2) }}
                  </el-tag>
                  <div style="margin-top: 4px; font-size: 12px; color: #909399;">
                    = 订单利润(¥{{ (orderDetail.sales_commission.order_profit || 0).toFixed(2) }}) × 20%
                  </div>
                </el-descriptions-item>
                <el-descriptions-item label="阶梯提成" v-if="orderDetail.sales_commission.tier_level > 0 && (orderDetail.sales_commission.tier_commission || 0) > 0">
                  <el-tag :type="orderDetail.sales_commission.tier_level >= 3 ? 'danger' : orderDetail.sales_commission.tier_level >= 2 ? 'warning' : 'success'" size="small">
                    阶梯{{ orderDetail.sales_commission.tier_level }}: +¥{{ (orderDetail.sales_commission.tier_commission || 0).toFixed(2) }}
                  </el-tag>
                  <div style="margin-top: 4px; font-size: 12px; color: #909399;">
                    基于当月总销售额达到阶梯{{ orderDetail.sales_commission.tier_level }}阈值
                  </div>
                </el-descriptions-item>
                <el-descriptions-item label="总分成" label-class-name="total-commission-label">
                  <span class="total-commission" style="font-size: 18px; font-weight: 700; color: #409eff;">
                    ¥{{ (orderDetail.sales_commission.total_commission || 0).toFixed(2) }}
                  </span>
                  <div style="margin-top: 4px; font-size: 12px; color: #909399;">
                    = 基础提成(¥{{ (orderDetail.sales_commission.base_commission || 0).toFixed(2) }})
                    <span v-if="(orderDetail.sales_commission.new_customer_bonus || 0) > 0">
                      + 新客激励(¥{{ (orderDetail.sales_commission.new_customer_bonus || 0).toFixed(2) }})
                    </span>
                    <span v-if="(orderDetail.sales_commission.tier_commission || 0) > 0">
                      + 阶梯提成(¥{{ (orderDetail.sales_commission.tier_commission || 0).toFixed(2) }})
                    </span>
                  </div>
                </el-descriptions-item>
              </el-descriptions>
            </div>
            <!-- 未收款订单：显示预览分成 -->
            <div v-else-if="orderDetail.sales_commission_preview">
              <el-descriptions :column="1" border>
                <el-descriptions-item label="销售员" v-if="orderDetail.user?.sales_employee">
                  <el-tag type="info">
                    {{ orderDetail.user.sales_employee.name || orderDetail.user.sales_employee.employee_code }}
                    <span v-if="orderDetail.user.sales_employee.employee_code" style="margin-left: 4px;">
                      ({{ orderDetail.user.sales_employee.employee_code }})
                    </span>
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="订单状态">
                  <el-tag :type="orderDetail.sales_commission_preview.is_valid_order ? 'success' : 'info'">
                    {{ orderDetail.sales_commission_preview.is_valid_order ? '有效订单（预计）' : '无效订单（预计）' }}
                  </el-tag>
                  <span v-if="!orderDetail.sales_commission_preview.is_valid_order" style="color: #909399; margin-left: 8px; font-size: 12px;">
                    (订单利润不满足最小阈值)
                  </span>
                </el-descriptions-item>
                <el-descriptions-item label="提示">
                  <el-alert
                    type="warning"
                    :closable="false"
                    show-icon
                  >
                    <template #title>
                      <span style="font-size: 13px;">此订单尚未收款，以下为预计分成。订单收款后才会正式计入销售员的有效分成。</span>
                    </template>
                  </el-alert>
                </el-descriptions-item>
                <el-descriptions-item label="预计基础提成（45%）">
                  <span style="color: #e6a23c;">
                    ¥{{ (orderDetail.sales_commission_preview.base_commission || 0).toFixed(2) }}
                  </span>
                </el-descriptions-item>
                <el-descriptions-item label="预计新客开发激励（20%）" v-if="orderDetail.sales_commission_preview.is_new_customer_order && (orderDetail.sales_commission_preview.new_customer_bonus || 0) > 0">
                  <el-tag type="warning" size="small">
                    +¥{{ (orderDetail.sales_commission_preview.new_customer_bonus || 0).toFixed(2) }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="预计阶梯提成" v-if="orderDetail.sales_commission_preview.tier_level > 0 && (orderDetail.sales_commission_preview.tier_commission || 0) > 0">
                  <el-tag :type="orderDetail.sales_commission_preview.tier_level >= 3 ? 'danger' : orderDetail.sales_commission_preview.tier_level >= 2 ? 'warning' : 'success'" size="small">
                    阶梯{{ orderDetail.sales_commission_preview.tier_level }}: +¥{{ (orderDetail.sales_commission_preview.tier_commission || 0).toFixed(2) }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="预计总分成" label-class-name="total-commission-label">
                  <span class="total-commission" style="font-size: 18px; font-weight: 700; color: #e6a23c;">
                    ¥{{ (orderDetail.sales_commission_preview.total_commission || 0).toFixed(2) }}
                  </span>
                  <div style="margin-top: 4px; font-size: 12px; color: #909399;">
                    = 基础提成(¥{{ (orderDetail.sales_commission_preview.base_commission || 0).toFixed(2) }})
                    <span v-if="(orderDetail.sales_commission_preview.new_customer_bonus || 0) > 0">
                      + 新客激励(¥{{ (orderDetail.sales_commission_preview.new_customer_bonus || 0).toFixed(2) }})
                    </span>
                    <span v-if="(orderDetail.sales_commission_preview.tier_commission || 0) > 0">
                      + 阶梯提成(¥{{ (orderDetail.sales_commission_preview.tier_commission || 0).toFixed(2) }})
                    </span>
                  </div>
                </el-descriptions-item>
              </el-descriptions>
            </div>
            <!-- 无销售分成信息 -->
            <el-empty v-else description="暂无销售分成信息（订单可能没有关联销售员）" :image-size="80" />
          </el-tab-pane>
        </el-tabs>
      </div>
      <template #footer>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 商品列表对话框 -->
    <el-dialog
      v-model="itemsDialogVisible"
      title="订单商品列表"
      width="800px"
      destroy-on-close
    >
      <div v-loading="itemsLoading">
        <el-table :data="orderItems" border stripe v-if="orderItems.length > 0">
          <el-table-column type="index" label="序号" width="60" align="center" />
          <el-table-column label="商品图片" width="100" align="center">
            <template #default="scope">
              <el-image
                v-if="scope.row.image"
                :src="scope.row.image"
                style="width: 60px; height: 60px; border-radius: 4px;"
                fit="cover"
                :preview-src-list="[scope.row.image]"
              />
              <span v-else style="color: #909399;">无图片</span>
            </template>
          </el-table-column>
          <el-table-column prop="product_name" label="商品名称" min-width="150" />
          <el-table-column prop="spec_name" label="规格" width="120" />
          <el-table-column prop="quantity" label="数量" width="80" align="center" />
          <el-table-column prop="unit_price" label="单价" width="100" align="right">
            <template #default="scope">
              ¥{{ formatMoney(scope.row.unit_price) }}
            </template>
          </el-table-column>
          <el-table-column prop="subtotal" label="小计" width="100" align="right">
            <template #default="scope">
              ¥{{ formatMoney(scope.row.subtotal) }}
            </template>
          </el-table-column>
        </el-table>
        <el-empty v-else description="暂无商品数据" />
      </div>
      <template #footer>
        <el-button @click="itemsDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown, QuestionFilled } from '@element-plus/icons-vue'
import { getOrders, getOrderDetail, updateOrderStatus } from '../api/orders'

const loading = ref(false)
const orders = ref([])
const searchKeyword = ref('')
const statusFilter = ref('')
const pagination = reactive({
  pageNum: 1,
  pageSize: 10,
  total: 0
})

// 订单详情相关
const detailDialogVisible = ref(false)
const detailLoading = ref(false)
const orderDetail = ref(null)
const activeTab = ref('basic') // 当前激活的标签页

// 商品列表相关
const itemsDialogVisible = ref(false)
const itemsLoading = ref(false)
const orderItems = ref([])

const loadOrders = async () => {
  loading.value = true
  try {
    const res = await getOrders({
      pageNum: pagination.pageNum,
      pageSize: pagination.pageSize,
      keyword: searchKeyword.value,
      status: statusFilter.value
    })
    // 处理响应数据 - 兼容不同的响应格式
    // 情况1: 标准格式 { code: 200, data: { list: [], total: 0 }, message: "..." }
    // 情况2: 直接返回数据 { list: [], total: 0 }
    // 情况3: 直接返回数组 []
    
    let orderList = []
    let total = 0
    
    if (res) {
      // 如果有 code 字段，说明是标准格式
      if (res.code === 200 && res.data) {
        orderList = res.data.list || []
        total = res.data.total || 0
      } 
      // 如果直接有 list 字段，说明是数据格式
      else if (res.list && Array.isArray(res.list)) {
        orderList = res.list
        total = res.total || 0
      }
      // 如果直接是数组
      else if (Array.isArray(res)) {
        orderList = res
        total = res.length
      }
      // 如果 data 直接是数组（某些API可能这样返回）
      else if (res.data && Array.isArray(res.data)) {
        orderList = res.data
        total = res.total || res.data.length
      }
    }
    
    // 确保赋值的是数组
    orders.value = Array.isArray(orderList) ? [...orderList] : []
    pagination.total = Number(total) || 0
  } catch (error) {
    console.error('获取订单失败:', error)
    console.error('错误详情:', error.response || error)
    orders.value = []
    pagination.total = 0
    ElMessage.error('获取订单列表失败，请稍后再试')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.pageNum = 1
  loadOrders()
}

const handlePageChange = (page) => {
  pagination.pageNum = page
  loadOrders()
}

const formatDate = (value) => {
  if (!value) return '-'
  return new Date(value).toLocaleString('zh-CN')
}

const formatMoney = (value) => {
  if (value === null || value === undefined) return '0.00'
  const num = Number(value)
  if (isNaN(num)) return '0.00'
  return num.toFixed(2)
}

const formatStatus = (status) => {
  const statusMap = {
    'pending': '待配送',           // 兼容旧状态
    'pending_delivery': '待配送',
    'pending_pickup': '待取货',
    'delivering': '配送中',
    'delivered': '已送达',
    'paid': '已收款',
    'completed': '已收款',        // 兼容旧状态
    'cancelled': '已取消',
    'shipped': '已送达'            // 兼容旧状态
  }
  return statusMap[status] || status
}

const getStatusType = (status) => {
  const typeMap = {
    'pending': 'danger',             // 兼容旧状态 - 待配送 - 红色
    'pending_delivery': 'danger',    // 待配送 - 红色
    'pending_pickup': 'warning',     // 待取货 - 橙色
    'delivering': 'primary',         // 配送中 - 蓝色
    'delivered': 'warning',          // 已送达 - 橙色
    'shipped': 'warning',            // 兼容旧状态 - 已送达 - 橙色
    'paid': 'success',               // 已收款 - 绿色
    'completed': 'success',          // 兼容旧状态 - 已收款 - 绿色
    'cancelled': 'info'              // 已取消 - 灰色
  }
  return typeMap[status] || 'info'
}

const formatOutOfStockStrategy = (strategy) => {
  const strategyMap = {
    'cancel_item': '取消缺货商品',
    'ship_available': '先发有货商品',
    'contact_me': '联系我'
  }
  return strategyMap[strategy] || strategy
}

const handleViewDetail = async (id) => {
  detailDialogVisible.value = true
  detailLoading.value = true
  orderDetail.value = null
  activeTab.value = 'basic' // 重置为基本信息标签页

  try {
    const res = await getOrderDetail(id)
    if (res && res.code === 200) {
      orderDetail.value = res.data
    } else {
      ElMessage.error(res?.message || '获取订单详情失败')
      detailDialogVisible.value = false
    }
  } catch (error) {
    console.error('获取订单详情失败:', error)
    ElMessage.error('获取订单详情失败，请稍后再试')
    detailDialogVisible.value = false
  } finally {
    detailLoading.value = false
  }
}

// 查看订单商品列表
const handleViewOrderItems = async (orderId) => {
  itemsDialogVisible.value = true
  itemsLoading.value = true
  orderItems.value = []

  try {
    const res = await getOrderDetail(orderId)
    if (res && res.code === 200 && res.data) {
      orderItems.value = Array.isArray(res.data.order_items) ? res.data.order_items : []
    } else {
      ElMessage.error(res?.message || '获取商品列表失败')
      itemsDialogVisible.value = false
    }
  } catch (error) {
    console.error('获取商品列表失败:', error)
    ElMessage.error('获取商品列表失败，请稍后再试')
    itemsDialogVisible.value = false
  } finally {
    itemsLoading.value = false
  }
}

// 判断是否显示状态操作按钮
const canShowStatusActions = (status) => {
  // 已收款和已取消不显示操作按钮
  if (status === 'paid' || status === 'completed' || status === 'cancelled') {
    return false
  }
  return true
}

// 判断是否是待配送状态（包括旧的 pending 状态）
const isPendingDelivery = (status) => {
  return status === 'pending' || status === 'pending_delivery'
}

// 处理订单状态变更
const handleStatusChange = async (orderId, currentStatus, newStatus) => {
  const statusMap = {
    'delivering': '开始配送',
    'delivered': '标记已送达',
    'paid': '标记已收款',
    'cancelled': '取消订单'
  }
  
  const actionName = statusMap[newStatus] || '更新状态'
  
  try {
    await ElMessageBox.confirm(
      `确定要${actionName}吗？`,
      '确认操作',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const res = await updateOrderStatus(orderId, newStatus)
    if (res && res.code === 200) {
      ElMessage.success(`${actionName}成功`)
      // 重新加载订单列表
      loadOrders()
      // 如果详情对话框打开，也刷新详情
      if (detailDialogVisible.value && orderDetail.value && orderDetail.value.order?.id === orderId) {
        handleViewDetail(orderId)
      }
    } else {
      ElMessage.error(res?.message || `${actionName}失败`)
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('更新订单状态失败:', error)
      ElMessage.error(`${actionName}失败，请稍后再试`)
    }
  }
}

onMounted(() => {
  loadOrders()
})
</script>

<style scoped>
.orders-page {
  padding: 20px;
}

.orders-card {
  min-height: calc(100vh - 100px);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.title {
  display: flex;
  flex-direction: column;
}

.main {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
}

.sub {
  font-size: 14px;
  color: #909399;
  margin-top: 4px;
}

.actions {
  display: flex;
  align-items: center;
}

.orders-table {
  margin-top: 20px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.total-amount-label {
  font-weight: 600;
}

.total-amount {
  font-size: 18px;
  font-weight: 700;
  color: #ff4d4f;
}

.rider-fee-label {
  font-weight: 600;
}

.rider-fee {
  font-size: 16px;
  font-weight: 700;
  color: #409eff;
}

.platform-cost-label {
  font-weight: 600;
}

.platform-cost {
  font-size: 16px;
  font-weight: 700;
  color: #67c23a;
}

.profit-label {
  font-weight: 600;
}

.profit-amount {
  font-size: 18px;
  font-weight: 700;
  color: #67c23a;
}

.net-profit-label {
  font-weight: 600;
}

.net-profit-amount {
  font-size: 18px;
  font-weight: 700;
  color: #e6a23c;
}

.real-profit-label {
  font-weight: 600;
}

.revenue-label {
  font-weight: 600;
}

.revenue-amount {
  font-size: 18px;
  font-weight: 700;
  color: #409eff;
}

.cost-label {
  font-weight: 600;
}

.cost-amount {
  font-size: 16px;
  font-weight: 600;
  color: #909399;
}

.gross-profit-label {
  font-weight: 600;
}

.gross-profit-amount {
  font-size: 18px;
  font-weight: 700;
  color: #67c23a;
}

.delivery-cost-label {
  font-weight: 600;
}

.delivery-cost-amount {
  font-size: 16px;
  font-weight: 600;
  color: #909399;
}

.profit-positive {
  color: #67c23a;
}

.profit-negative {
  color: #f56c6c;
}

.action-buttons {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>

