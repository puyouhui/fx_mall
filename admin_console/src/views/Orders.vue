<template>
  <div class="orders-page">
    <el-card class="orders-card" shadow="never">
      <div class="card-header">
        <div class="title">
          <span class="main">订单管理</span>
          <span class="sub">查看和管理所有订单</span>
        </div>
        <div class="actions">
          <el-input v-model="searchKeyword" placeholder="搜索订单ID / 用户ID" clearable @keyup.enter="handleSearch"
            style="width: 200px; margin-right: 10px;" />
          <el-select v-model="statusFilter" placeholder="订单状态" clearable style="width: 150px; margin-right: 10px;"
            @change="handleSearch">
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

      <el-table v-loading="loading" :data="orders" border stripe class="orders-table" empty-text="暂无订单数据" row-key="id">
        <!-- <el-table-column prop="id" label="订单ID" width="100" /> -->
        <el-table-column prop="order_number" label="订单编号" width="180" align="center" />
        <el-table-column label="用户信息" min-width="180" align="center">
          <template #default="scope">
            <div v-if="scope.row.user">
              <div>{{ scope.row.user.name || '未命名' }}</div>
              <div style="color: #909399; font-size: 12px;">用户{{ scope.row.user.user_code || scope.row.user_id }}</div>
            </div>
            <span v-else>用户ID: {{ scope.row.user_id }}</span>
          </template>
        </el-table-column>
        <el-table-column label="销售员" width="120" align="center">
          <template #default="scope">
            <div v-if="scope.row.user && scope.row.user.sales_employee">
              <el-tag size="small" type="info">
                {{ scope.row.user.sales_employee.name || scope.row.user.sales_employee.employee_code }}
              </el-tag>
              <div v-if="scope.row.user.sales_employee.employee_code"
                style="color: #909399; font-size: 11px; margin-top: 2px;">
                {{ scope.row.user.sales_employee.employee_code }}
              </div>
            </div>
            <span v-else style="color: #c0c4cc;">-</span>
          </template>
        </el-table-column>
        <el-table-column label="收货地址" min-width="200" align="center">
          <template #default="scope">
            <div v-if="scope.row.address">
              <div>{{ scope.row.address.name || '-' }}</div>
              <div style="color: #909399; font-size: 12px;">{{ scope.row.address.address || '-' }}</div>
            </div>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="订单状态" width="120" align="center">
          <template #default="scope">
            <el-tag :type="getStatusType(scope.row.status)">
              {{ formatStatus(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="商品件数" width="120" align="center">
          <template #default="scope">
            <el-button type="primary" link @click="handleViewOrderItems(scope.row.id)"
              :disabled="!scope.row.item_count || scope.row.item_count === 0">
              {{ scope.row.item_count || 0 }} 件
            </el-button>
          </template>
        </el-table-column>
        <el-table-column label="金额信息" min-width="150" align="center">
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
              <div v-if="!scope.row.sales_commission.is_valid_order"
                style="color: #909399; font-size: 11px; margin-top: 2px;">
                无效订单
              </div>
              <!-- 如果预览和已计入不一致，显示预览值 -->
              <div
                v-if="scope.row.sales_commission_preview &&
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
              <div v-if="!scope.row.sales_commission_preview.is_valid_order"
                style="color: #f56c6c; font-size: 11px; margin-top: 2px;">
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
              <div v-if="scope.row.delivery_employee.employee_code"
                style="color: #909399; font-size: 11px; margin-top: 2px;">
                {{ scope.row.delivery_employee.employee_code }}
              </div>
              <div v-if="scope.row.rider_payable_fee && scope.row.rider_payable_fee > 0"
                style="color: #67c23a; font-weight: 600; font-size: 13px; margin-top: 4px;">
                配送费: ¥{{ formatMoney(scope.row.rider_payable_fee) }}
              </div>
            </div>
            <span v-else style="color: #c0c4cc;">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="下单时间" min-width="160" align="center">
          <template #default="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right" align="center">
          <template #default="scope">
            <div class="action-buttons">
              <el-button type="primary" link @click="handleViewDetail(scope.row.id)">
                详情
              </el-button>
              <el-dropdown @command="(cmd) => handlePrintCommand(cmd, scope.row)" trigger="click">
                <el-button type="success" link>
                  打印
                  <el-icon style="margin-left: 4px;">
                    <ArrowDown />
                  </el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="ticket">打印小票</el-dropdown-item>
                    <el-dropdown-item command="material">打印物料</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
              <el-dropdown
                @command="(cmd) => handleOrderAction(scope.row.id, scope.row.status, cmd)" trigger="click"
                placement="bottom-end">
                <el-button type="primary" link>
                  操作
                  <el-icon style="margin-left: 4px;">
                    <ArrowDown />
                  </el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item v-if="canShowStatusActions(scope.row.status) && isPendingDelivery(scope.row.status)" command="delivering">
                      开始配送
                    </el-dropdown-item>
                    <el-dropdown-item v-if="canShowStatusActions(scope.row.status) && scope.row.status === 'delivering'" command="delivered">
                      标记已送达
                    </el-dropdown-item>
                    <el-dropdown-item v-if="canShowStatusActions(scope.row.status) && (scope.row.status === 'delivered' || scope.row.status === 'shipped')"
                      command="paid">
                      标记已收款
                    </el-dropdown-item>
                    <el-dropdown-item v-if="canShowStatusActions(scope.row.status) && isPendingDelivery(scope.row.status)" command="cancelled" divided>
                      取消订单
                    </el-dropdown-item>
                    <el-dropdown-item command="recalculate" divided>
                      强制重新计算
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination background layout="total, prev, pager, next, jumper" :page-size="pagination.pageSize"
          :current-page="pagination.pageNum" :total="pagination.total" @current-change="handlePageChange" />
      </div>
    </el-card>

    <!-- 订单详情对话框 -->
    <el-dialog v-model="detailDialogVisible" title="订单详情" width="900px" destroy-on-close>
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
              <el-descriptions-item label="详细地址" :span="2">{{ orderDetail.address.address || '-'
              }}</el-descriptions-item>
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
              <el-descriptions-item label="加急费"
                v-if="orderDetail.order?.is_urgent && (orderDetail.order?.urgent_fee || 0) > 0">
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
            <el-descriptions :column="1" border
              v-if="orderDetail.simplified_profit && Object.keys(orderDetail.simplified_profit).length > 0">
              <el-descriptions-item label="平台总收入（实付金额）" label-class-name="revenue-label">
                <span class="revenue-amount">¥{{ (orderDetail.simplified_profit.platform_revenue || 0).toFixed(2)
                }}</span>
                <div style="margin-top: 4px; font-size: 12px; color: #909399;">
                  商品金额(¥{{ (orderDetail.order?.goods_amount || 0).toFixed(2) }})
                  + 配送费(¥{{ (orderDetail.order?.delivery_fee || 0).toFixed(2) }})
                  <span v-if="(orderDetail.order?.urgent_fee || 0) > 0">+ 加急费(¥{{ (orderDetail.order?.urgent_fee ||
                    0).toFixed(2) }})</span>
                  <span v-if="(orderDetail.order?.coupon_discount || 0) > 0">- 优惠券(¥{{
                    (orderDetail.order?.coupon_discount
                      || 0).toFixed(2) }})</span>
                  <span v-if="(orderDetail.order?.points_discount || 0) > 0">- 积分(¥{{
                    (orderDetail.order?.points_discount ||
                      0).toFixed(2) }})</span>
                </div>
              </el-descriptions-item>
              <el-descriptions-item label="商品总成本" label-class-name="cost-label">
                <span class="cost-amount">¥{{ (orderDetail.simplified_profit.goods_cost || 0).toFixed(2) }}</span>
              </el-descriptions-item>
              <el-descriptions-item label="毛利润（平台总收入 - 商品总成本）" label-class-name="gross-profit-label">
                <span class="gross-profit-amount">¥{{ (orderDetail.simplified_profit.gross_profit || 0).toFixed(2)
                }}</span>
                <div style="margin-top: 4px; font-size: 12px; color: #909399;">
                  = 平台总收入(¥{{ (orderDetail.simplified_profit.platform_revenue || 0).toFixed(2) }})
                  - 商品总成本(¥{{ (orderDetail.simplified_profit.goods_cost || 0).toFixed(2) }})
                </div>
              </el-descriptions-item>
              <el-descriptions-item label="配送成本" label-class-name="delivery-cost-label">
                <span class="delivery-cost-amount">¥{{ (orderDetail.simplified_profit.delivery_cost || 0).toFixed(2)
                }}</span>
              </el-descriptions-item>
              <el-descriptions-item label="净利润（平台总收入 - 商品总成本 - 配送成本）" label-class-name="net-profit-label">
                <span class="net-profit-amount"
                  :class="{ 'profit-positive': (orderDetail.simplified_profit.net_profit || 0) >= 0, 'profit-negative': (orderDetail.simplified_profit.net_profit || 0) < 0 }">
                  ¥{{ (orderDetail.simplified_profit.net_profit || 0).toFixed(2) }}
                </span>
                <div style="margin-top: 4px; font-size: 12px; color: #909399;">
                  = 平台总收入(¥{{ (orderDetail.simplified_profit.platform_revenue || 0).toFixed(2) }})
                  - 商品总成本(¥{{ (orderDetail.simplified_profit.goods_cost || 0).toFixed(2) }})
                  - 配送成本(¥{{ (orderDetail.simplified_profit.delivery_cost || 0).toFixed(2) }})
                </div>
                <div style="margin-top: 4px; font-size: 12px; font-weight: 600;"
                  :style="{ color: (orderDetail.simplified_profit.net_profit || 0) >= 0 ? '#67c23a' : '#f56c6c' }">
                  {{ (orderDetail.simplified_profit.net_profit || 0) >= 0 ? '✓ 平台盈利' : '✗ 平台亏损' }}
                </div>
              </el-descriptions-item>
            </el-descriptions>
            <el-empty v-else description="利润信息暂不可用" :image-size="80" />
          </el-tab-pane>

          <!-- 配送详情标签页 -->
          <el-tab-pane label="配送详情" name="delivery">
            <!-- 配送费详情 -->
            <el-divider content-position="left">配送费详情</el-divider>
            <el-descriptions :column="1" border style="margin-bottom: 20px;"
              v-if="orderDetail.delivery_fee_calculation && Object.keys(orderDetail.delivery_fee_calculation).length > 0">
              <el-descriptions-item label="基础配送费">
                ¥{{ (orderDetail.delivery_fee_calculation.base_fee || 0).toFixed(2) }}
              </el-descriptions-item>
              <el-descriptions-item label="孤立订单补贴" v-if="orderDetail.delivery_fee_calculation.isolated_fee > 0">
                <el-tag type="warning" size="small">+¥{{ (orderDetail.delivery_fee_calculation.isolated_fee ||
                  0).toFixed(2)
                }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="件数补贴" v-if="orderDetail.delivery_fee_calculation.item_fee > 0">
                <el-tag type="info" size="small">+¥{{ (orderDetail.delivery_fee_calculation.item_fee || 0).toFixed(2)
                }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="加急订单补贴" v-if="orderDetail.delivery_fee_calculation.urgent_fee > 0">
                <el-tag type="danger" size="small">+¥{{ (orderDetail.delivery_fee_calculation.urgent_fee ||
                  0).toFixed(2)
                }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="极端天气补贴" v-if="orderDetail.delivery_fee_calculation.weather_fee > 0">
                <el-tag type="warning" size="small">+¥{{ (orderDetail.delivery_fee_calculation.weather_fee ||
                  0).toFixed(2)
                }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="配送员实际所得（预估配送费）" label-class-name="rider-fee-label">
                <span class="rider-fee">
                  ¥{{ (orderDetail.delivery_fee_calculation.rider_payable_fee || 0).toFixed(2) }}
                  <span v-if="orderDetail.delivery_fee_calculation.profit_share > 0"
                    style="color: #67c23a; margin-left: 8px; font-size: 14px;">
                    （包含利润分成¥{{ (orderDetail.delivery_fee_calculation.profit_share || 0).toFixed(2) }}）
                  </span>
                </span>
              </el-descriptions-item>
              <el-descriptions-item label="利润分成明细" v-if="orderDetail.delivery_fee_calculation.profit_share > 0">
                <el-tag type="success" size="small">+¥{{ (orderDetail.delivery_fee_calculation.profit_share ||
                  0).toFixed(2)
                }}</el-tag>
                <span style="margin-left: 8px; color: #909399; font-size: 12px;">(已包含在预估配送费中，仅管理员可见)</span>
              </el-descriptions-item>
              <el-descriptions-item label="平台总成本" label-class-name="platform-cost-label">
                <span class="platform-cost">¥{{ (orderDetail.delivery_fee_calculation.total_platform_cost ||
                  0).toFixed(2)
                }}</span>
              </el-descriptions-item>
            </el-descriptions>
            <el-empty v-else description="配送费计算信息暂不可用" :image-size="80" />

            <!-- 配送记录 -->
            <el-divider content-position="left">配送记录</el-divider>
            <div v-if="orderDetail.delivery_record" style="margin-bottom: 20px;">
              <el-descriptions :column="2" border>
                <el-descriptions-item label="配送员员工码">
                  {{ orderDetail.delivery_record.delivery_employee_code || '-' }}
                </el-descriptions-item>
                <el-descriptions-item label="完成时间">
                  {{ formatDate(orderDetail.delivery_record.completed_at) }}
                </el-descriptions-item>
              </el-descriptions>
            </div>
            <el-empty v-else description="暂无配送记录" :image-size="80" />

            <!-- 配送日志 -->
            <el-divider content-position="left">配送流程日志</el-divider>
            <el-timeline v-if="orderDetail.delivery_logs && orderDetail.delivery_logs.length > 0" style="margin-top: 20px;">
              <el-timeline-item
                v-for="(log, index) in orderDetail.delivery_logs"
                :key="index"
                :timestamp="formatDate(log.action_time)"
                placement="top">
                <el-card>
                  <h4>{{ formatDeliveryLogAction(log.action) }}</h4>
                  <p v-if="log.delivery_employee_code" style="color: #909399; font-size: 12px; margin-top: 4px;">
                    配送员：{{ log.delivery_employee_code }}
                  </p>
                  <p v-if="log.remark" style="color: #606266; font-size: 13px; margin-top: 4px;">
                    {{ log.remark }}
                  </p>
                </el-card>
              </el-timeline-item>
            </el-timeline>
            <el-empty v-else description="暂无配送流程日志" :image-size="80" />

            <!-- 配送完成图片 -->
            <el-divider content-position="left" v-if="orderDetail.delivery_record && (orderDetail.delivery_record.product_image_url || orderDetail.delivery_record.doorplate_image_url)">
              配送完成图片
            </el-divider>
            <div v-if="orderDetail.delivery_record && (orderDetail.delivery_record.product_image_url || orderDetail.delivery_record.doorplate_image_url)"
              style="margin-top: 20px;">
              <el-row :gutter="20">
                <el-col :span="12" v-if="orderDetail.delivery_record.product_image_url">
                  <div style="text-align: center; margin-bottom: 20px;">
                    <div style="font-weight: 600; margin-bottom: 8px; color: #606266;">货物照片</div>
                    <el-image
                      :src="orderDetail.delivery_record.product_image_url"
                      style="width: 100%; max-width: 400px; border-radius: 8px;"
                      fit="cover"
                      :preview-src-list="[orderDetail.delivery_record.product_image_url]"
                      preview-teleported>
                    </el-image>
                  </div>
                </el-col>
                <el-col :span="12" v-if="orderDetail.delivery_record.doorplate_image_url">
                  <div style="text-align: center; margin-bottom: 20px;">
                    <div style="font-weight: 600; margin-bottom: 8px; color: #606266;">门牌照片</div>
                    <el-image
                      :src="orderDetail.delivery_record.doorplate_image_url"
                      style="width: 100%; max-width: 400px; border-radius: 8px;"
                      fit="cover"
                      :preview-src-list="[orderDetail.delivery_record.doorplate_image_url]"
                      preview-teleported>
                    </el-image>
                  </div>
                </el-col>
              </el-row>
            </div>
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
                  <span v-if="!orderDetail.sales_commission.is_valid_order"
                    style="color: #909399; margin-left: 8px; font-size: 12px;">
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
                  <span v-if="orderDetail.sales_commission.settlement_date"
                    style="color: #909399; margin-left: 8px; font-size: 12px;">
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
                <el-descriptions-item label="新客开发激励（20%）"
                  v-if="orderDetail.sales_commission.is_new_customer_order && (orderDetail.sales_commission.new_customer_bonus || 0) > 0">
                  <el-tag type="warning" size="small">
                    +¥{{ (orderDetail.sales_commission.new_customer_bonus || 0).toFixed(2) }}
                  </el-tag>
                  <div style="margin-top: 4px; font-size: 12px; color: #909399;">
                    = 订单利润(¥{{ (orderDetail.sales_commission.order_profit || 0).toFixed(2) }}) × 20%
                  </div>
                </el-descriptions-item>
                <el-descriptions-item label="阶梯提成"
                  v-if="orderDetail.sales_commission.tier_level > 0 && (orderDetail.sales_commission.tier_commission || 0) > 0">
                  <el-tag
                    :type="orderDetail.sales_commission.tier_level >= 3 ? 'danger' : orderDetail.sales_commission.tier_level >= 2 ? 'warning' : 'success'"
                    size="small">
                    阶梯{{ orderDetail.sales_commission.tier_level }}: +¥{{ (orderDetail.sales_commission.tier_commission
                      ||
                      0).toFixed(2) }}
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
                  <span v-if="!orderDetail.sales_commission_preview.is_valid_order"
                    style="color: #909399; margin-left: 8px; font-size: 12px;">
                    (订单利润不满足最小阈值)
                  </span>
                </el-descriptions-item>
                <el-descriptions-item label="提示">
                  <el-alert type="warning" :closable="false" show-icon>
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
                <el-descriptions-item label="预计新客开发激励（20%）"
                  v-if="orderDetail.sales_commission_preview.is_new_customer_order && (orderDetail.sales_commission_preview.new_customer_bonus || 0) > 0">
                  <el-tag type="warning" size="small">
                    +¥{{ (orderDetail.sales_commission_preview.new_customer_bonus || 0).toFixed(2) }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="预计阶梯提成"
                  v-if="orderDetail.sales_commission_preview.tier_level > 0 && (orderDetail.sales_commission_preview.tier_commission || 0) > 0">
                  <el-tag
                    :type="orderDetail.sales_commission_preview.tier_level >= 3 ? 'danger' : orderDetail.sales_commission_preview.tier_level >= 2 ? 'warning' : 'success'"
                    size="small">
                    阶梯{{ orderDetail.sales_commission_preview.tier_level }}: +¥{{
                      (orderDetail.sales_commission_preview.tier_commission || 0).toFixed(2) }}
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
        <el-dropdown @command="(cmd) => handlePrintCommandFromDetail(cmd)" trigger="click">
          <el-button type="primary">
            打印
            <el-icon style="margin-left: 4px;">
              <ArrowDown />
            </el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="ticket">打印小票</el-dropdown-item>
              <el-dropdown-item command="material">打印物料</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </template>
    </el-dialog>

    <!-- 商品列表对话框 -->
    <el-dialog v-model="itemsDialogVisible" title="订单商品列表" width="800px" destroy-on-close>
      <div v-loading="itemsLoading">
        <el-table :data="orderItems" border stripe v-if="orderItems.length > 0">
          <el-table-column type="index" label="序号" width="60" align="center" />
          <el-table-column label="商品图片" width="100" align="center">
            <template #default="scope">
              <el-image v-if="scope.row.image" :src="scope.row.image"
                style="width: 60px; height: 60px; border-radius: 4px;" fit="cover"
                :preview-src-list="[scope.row.image]" />
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
import { getOrders, getOrderDetail, updateOrderStatus, recalculateOrderProfit } from '../api/orders'
import { hiprint } from 'vue-plugin-hiprint'
import { getPrinterAddress, getPrintOptions, isOnlineEnvironment } from '../utils/printer'

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

const formatDeliveryLogAction = (action) => {
  const actionMap = {
    'created': '订单创建',
    'accepted': '接单',
    'pickup_started': '开始取货',
    'pickup_completed': '取货完成',
    'delivering_started': '开始配送',
    'delivering_completed': '配送完成'
  }
  return actionMap[action] || action
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

// 处理订单操作（包括状态变更和其他操作）
const handleOrderAction = async (orderId, currentStatus, command) => {
  // 如果是重新计算，调用重新计算函数
  if (command === 'recalculate') {
    await handleRecalculateProfit(orderId)
    return
  }

  // 其他命令按状态变更处理
  const statusMap = {
    'delivering': '开始配送',
    'delivered': '标记已送达',
    'paid': '标记已收款',
    'cancelled': '取消订单'
  }

  const actionName = statusMap[command] || '更新状态'

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

    const res = await updateOrderStatus(orderId, command)
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

// 处理强制重新计算订单利润
const handleRecalculateProfit = async (orderId) => {
  try {
    await ElMessageBox.confirm(
      '确定要强制重新计算订单利润吗？此操作会重新计算订单的成本和利润。',
      '确认操作',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    ElMessage.info('正在重新计算，请稍候...')
    const res = await recalculateOrderProfit(orderId)
    if (res && res.code === 200) {
      ElMessage.success('重新计算成功')
      // 显示计算结果
      if (res.data) {
        const data = res.data
        const message = `计算结果：\n商品总金额：¥${data.goods_amount}\n计算总成本：¥${data.calculated_cost}\n计算利润：¥${data.calculated_profit}\n存储利润：¥${data.stored_profit || 'N/A'}\n存储净利润：¥${data.stored_net_profit || 'N/A'}`
        ElMessageBox.alert(message, '计算完成', {
          confirmButtonText: '确定',
          type: 'success'
        })
      }
      // 重新加载订单列表
      loadOrders()
      // 如果详情对话框打开，也刷新详情
      if (detailDialogVisible.value && orderDetail.value && orderDetail.value.order?.id === orderId) {
        handleViewDetail(orderId)
      }
    } else {
      ElMessage.error(res?.message || '重新计算失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('重新计算订单利润失败:', error)
      ElMessage.error('重新计算失败，请稍后再试')
    }
  }
}

// 处理打印命令（从列表）
const handlePrintCommand = async (command, order) => {
  if (command === 'ticket') {
    // 打印小票
    await handlePrintOrder(order)
  } else if (command === 'material') {
    // 打印物料
    await handlePrintMaterial(order)
  }
}

// 处理打印命令（从详情对话框）
const handlePrintCommandFromDetail = async (command) => {
  if (!orderDetail.value) {
    ElMessage.warning('订单详情未加载')
    return
  }
  
  if (command === 'ticket') {
    // 打印小票
    printOrder(orderDetail.value)
  } else if (command === 'material') {
    // 打印物料
    await handlePrintMaterial(orderDetail.value)
  }
}

// 打印订单（从列表）
const handlePrintOrder = async (order) => {
  try {
    // 如果订单没有完整信息，先获取订单详情
    let orderData = order
    if (!order.order_items || order.order_items.length === 0) {
      const res = await getOrderDetail(order.id)
      if (res && res.code === 200) {
        orderData = res.data
      } else {
        ElMessage.error('获取订单详情失败，无法打印')
        return
      }
    }

    // 调用打印函数
    printOrder(orderData)
  } catch (error) {
    console.error('打印订单失败:', error)
    ElMessage.error('打印订单失败，请稍后再试')
  }
}

// 打印订单函数（80mm纸张）
const printOrder = (orderData) => {
  // 检查 hiprint 是否初始化
  if (!hiprint) {
    ElMessage.error('打印功能未初始化，请刷新页面重试')
    console.error('hiprint 未初始化')
    return
  }

  // 尝试重新连接（如果未连接）
  if (!hiprint.hiwebSocket || !hiprint.hiwebSocket.opened) {
    console.warn('打印机未连接，尝试重新连接...')
    try {
      // 重新初始化连接
      const printerAddress = getPrinterAddress()
      hiprint.init({
        host: printerAddress,
        token: "vue-plugin-hiprint",
      })

      // 等待一下让连接建立
      setTimeout(() => {
        checkAndPrint(orderData)
      }, 500)
      return
    } catch (error) {
      console.error('重新连接失败:', error)
      const printerAddress = getPrinterAddress()
      ElMessage.error(`打印机连接失败，请检查打印客户端是否运行（地址: ${printerAddress}）`)
      return
    }
  }

  // 执行打印
  executePrint(orderData)
}

// 检查连接并打印
const checkAndPrint = (orderData) => {
  const isConnected = hiprint.hiwebSocket?.opened || false
  console.log('连接检查结果:', {
    hasHiwebSocket: !!hiprint.hiwebSocket,
    opened: hiprint.hiwebSocket?.opened,
    isConnected: isConnected
  })

  if (!isConnected) {
    const printerAddress = getPrinterAddress()
    ElMessage.error(`打印机未连接，请检查打印客户端是否运行（地址: ${printerAddress}）`)
    console.error('打印机未连接，详细信息:', {
      hiwebSocket: hiprint.hiwebSocket,
      opened: hiprint.hiwebSocket?.opened,
      printerAddress: printerAddress
    })
    return
  }

  executePrint(orderData)
}

// 将网络图片转换为 base64 格式
const convertImageToBase64 = (imageUrl) => {
  return new Promise((resolve, reject) => {
    const img = new Image()
    img.crossOrigin = 'anonymous' // 允许跨域

    img.onload = () => {
      try {
        const canvas = document.createElement('canvas')
        canvas.width = img.width
        canvas.height = img.height
        const ctx = canvas.getContext('2d')
        ctx.drawImage(img, 0, 0)
        const base64 = canvas.toDataURL('image/png')
        resolve(base64)
      } catch (error) {
        reject(error)
      }
    }

    img.onerror = (error) => {
      reject(new Error('图片加载失败: ' + error))
    }

    img.src = imageUrl
  })
}

// 执行打印（使用文本元素方式）
const executePrint = async (orderData) => {
  try {
    // 创建打印模板
    const hiprintTemplate = new hiprint.PrintTemplate()

    // 添加打印面板（80mm宽度）
    const panel = hiprintTemplate.addPrintPanel({
      width: 80, // 80mm纸张宽度
      height: 350, // 初始高度，会根据内容自动调整
      paperFooter: 0,
      paperHeader: 0,
      paperNumberLeft: 0,
      paperNumberRight: 0,
      paperNumberFormat: ' ',
    })

    // 使用文本元素方式添加内容（基于HTML结构）
    let currentTop = 5 // 从顶部开始，减少上方空白

    const order = orderData.order || orderData
    const user = orderData.user || orderData
    const address = orderData.address || orderData
    const orderItems = orderData.order_items || []
    const orderNumber = order.order_number || orderData.order_number || '-'
    const orderTime = order.created_at || orderData.created_at
    const timeStr = orderTime ? formatDate(orderTime) : '-'
    const goodsAmount = order.goods_amount || 0
    const deliveryFee = order.delivery_fee || 0
    const urgentFee = order.urgent_fee || 0
    const couponDiscount = order.coupon_discount || 0 // 优惠券抵扣金额
    const totalAmount = order.total_amount || 0
    const status = order.status || orderData.status
    const hidePrice = order.hide_price || orderData.hide_price || false // 是否环保小票（隐藏价格）
    const remark = order.remark || orderData.remark || '' // 订单备注
    const trustReceipt = order.trust_receipt || orderData.trust_receipt || false // 信任签收
    const requirePhoneContact = order.require_phone_contact || orderData.require_phone_contact || false // 要求电话联系

    // 格式化价格：如果是环保小票，返回**，否则返回格式化的金额
    const formatPrice = (amount) => {
      return hidePrice ? '**' : formatMoney(amount)
    }

    // 格式化电话号码：显示前三位和后三位，中间用*代替
    const formatPhone = (phone) => {
      if (!phone) return '-'
      const phoneStr = String(phone)
      if (phoneStr.length <= 6) return phoneStr
      const firstThree = phoneStr.slice(0, 3)
      const lastThree = phoneStr.slice(-3)
      return firstThree + '****' + lastThree
    }

    // 订单标题：根据是否环保小票显示不同标题
    const title = hidePrice ? "橙心选（环保票）" : "橙心选"
    panel.addPrintText({
      options: {
        width: 220, // 尝试更大的值以占满 80mm 宽度
        height: 20,
        top: currentTop, // 从顶部开始
        left: 0,
        title: title,
        textAlign: "center",
        fontSize: 14,
        fontWeight: "bold"
      },
    })
    currentTop += 30 // 减少标题后的间距

    // 订单编号
    panel.addPrintText({
      options: {
        width: 300, // 尝试更大的值以占满 80mm 宽度
        top: currentTop,
        left: 0,
        title: `订单号：${orderNumber}`,
        textAlign: "left",
        fontSize: 10
      },
    })
    currentTop += 15

    // 下单时间
    panel.addPrintText({
      options: {
        width: 300, // 尝试更大的值以占满 80mm 宽度
        top: currentTop,
        left: 0,
        title: `下单时间：${timeStr}`,
        textAlign: "left",
        fontSize: 9
      },
    })
    currentTop += 15

    // 分隔线
    panel.addPrintText({
      options: {
        width: 230, // 尝试更大的值以占满 80mm 宽度
        top: currentTop,
        left: 0,
        title: "-------------------------------------------",
        textAlign: "center",
        fontSize: 9
      },
    })
    currentTop += 15

    // 地址信息
    if (address) {
      // 地址名称
      const addressName = address.name || '-'
      
      panel.addPrintText({
        options: {
          width: 300, // 尝试更大的值以占满 80mm 宽度
          top: currentTop,
          left: 0,
          title: `名称：${addressName}`,
          textAlign: "left",
          fontSize: 11
        },
      })
      currentTop += 20

      // 地址电话
      if (address.phone) {
        panel.addPrintText({
          options: {
            width: 300, // 尝试更大的值以占满 80mm 宽度
            top: currentTop,
            left: 0,
            title: `电话：${formatPhone(address.phone)}`,
            textAlign: "left",
            fontSize: 11
          },
        })
        currentTop += 15
      }
    }

    // 收货地址
    if (address && address.address) {
      const addressText = `地址：${address.address}`
      // 估算文本行数：每行约20个字符（根据宽度230和字体大小9估算）
      const estimatedLines = Math.ceil(addressText.length / 20)
      const textHeight = Math.max(15, estimatedLines * 15) // 最小15px，每行15px
      
      panel.addPrintText({
        options: {
          width: 230, // 尝试更大的值以占满 80mm 宽度
          height: textHeight, // 设置高度以容纳多行文本
          top: currentTop,
          left: 0,
          title: addressText,
          textAlign: "left",
          fontSize: 9,
          lineHeight: 15 // 设置行高，确保换行时有足够间距
        },
      })
      currentTop += textHeight + 5 // 根据实际文本高度调整间距
    }

    // 信任签收和电话联系提示
    // 只有当有信任签收或备注时，才显示分割线（如果只有电话联系，不显示分割线）
    if (trustReceipt || (remark && remark.trim())) {
      // 添加分割线，和客户信息分割开
      currentTop += 5
      panel.addPrintText({
        options: {
          width: 230,
          top: currentTop,
          left: 0,
          title: "-------------------------------------------",
          textAlign: "center",
          fontSize: 9
        },
      })
      currentTop += 15
    }

    if (trustReceipt || requirePhoneContact) {
      if (trustReceipt) {
        panel.addPrintText({
          options: {
            width: 230,
            top: currentTop,
            left: 0,
            title: '注意：客户已开启信任签收',
            textAlign: "left",
            fontSize: 10,
            fontWeight: "bold"
          },
        })
        currentTop += 18
      }
      if (requirePhoneContact) {
        panel.addPrintText({
          options: {
            width: 230,
            top: currentTop,
            left: 0,
            title: '注意：配送前需电话联系',
            textAlign: "left",
            fontSize: 10,
            fontWeight: "bold"
          },
        })
        currentTop += 18
      }
    }

    // 订单备注（如果有备注，字体要大一点，因为备注很重要）
    if (remark && remark.trim()) {
      currentTop += 5
      const remarkText = `订单备注：${remark.trim()}`
      // 估算文本行数：每行约18个字符（根据宽度230和字体大小12估算）
      const estimatedLines = Math.ceil(remarkText.length / 18)
      const textHeight = Math.max(20, estimatedLines * 20) // 最小20px，每行20px
      
      panel.addPrintText({
        options: {
          width: 230,
          height: textHeight,
          top: currentTop,
          left: 0,
          title: remarkText,
          textAlign: "left",
          fontSize: 12, // 字体大一点，因为备注很重要
          fontWeight: "bold",
          lineHeight: 20 // 设置行高
        },
      })
      currentTop += textHeight + 5
    }

    // 分隔线
    currentTop += 3
    panel.addPrintText({
      options: {
        width: 230, // 尝试更大的值以占满 80mm 宽度
        top: currentTop,
        left: 0,
        title: "-------------------------------------------",
        textAlign: "center",
        fontSize: 9
      },
    })
    currentTop += 15

    // 商品列表
    if (orderItems.length > 0) {
      orderItems.forEach((item) => {

        const quantity = item.quantity || 0
        const unitPrice = item.unit_price || 0
        const subtotal = item.subtotal || (quantity * unitPrice)

        const productName = `${item.product_name || ''} ${item.spec_name || ''}`.trim()
        const productNameText = productName + ' ' + ' X ' + quantity
        // 估算文本行数：每行约24个字符（根据宽度230和字体大小11估算，80mm纸张实际可以容纳更多字符）
        const estimatedLines = Math.ceil(productNameText.length / 24)
        const textHeight = Math.max(18, estimatedLines * 18) // 最小18px，每行18px
        
        panel.addPrintText({
          options: {
            width: 230, // 尝试更大的值以占满 80mm 宽度
            height: textHeight, // 设置高度以容纳多行文本
            top: currentTop,
            left: 0,
            title: productNameText,
            textAlign: "left",
            fontSize: 11,
            fontWeight: "bold",
            lineHeight: 18 // 设置行高，确保换行时有足够间距
          },
        })
        
        // 商品名称到价格的间距：使用固定的紧凑间距
        // 无论单行还是多行，都使用相同的间距（一行高度+小间距），让价格紧跟在名称下方
        currentTop += 18 + 3 // 固定使用一行高度(18px) + 小间距(3px) = 21px

        panel.addPrintText({
          options: {
            width: 230, // 尝试更大的值以占满 80mm 宽度
            top: currentTop,
            left: 0,
            title: `  ${quantity} × ¥${formatPrice(unitPrice)} = ¥${formatPrice(subtotal)}`,
            textAlign: "left",
            fontSize: 10
          },
        })
        
        // 价格行之后到下个商品的间距：固定间距
        currentTop += 18 // 固定间距，让每个商品之间的间距一致
      })
    }

    // 分隔线
    currentTop += 5
    panel.addPrintText({
      options: {
        width: 230, // 尝试更大的值以占满 80mm 宽度
        top: currentTop,
        left: 0,
        title: "-------------------------------------------",
        textAlign: "center",
        fontSize: 9
      },
    })
    currentTop += 22

    // 金额汇总
    panel.addPrintText({
      options: {
        width: 220, // 尝试更大的值以占满 80mm 宽度
        top: currentTop,
        left: 0,
        title: `商品金额：¥${formatPrice(goodsAmount)}`,
        textAlign: "right",
        fontSize: 10
      },
    })
    currentTop += 20

    // 配送费：有配送费显示金额，没有配送费显示"免配送费"
    panel.addPrintText({
      options: {
        width: 220, // 尝试更大的值以占满 80mm 宽度
        top: currentTop,
        left: 0,
        title: deliveryFee > 0 ? `配送费：¥${formatPrice(deliveryFee)}` : '免配送费',
        textAlign: "right",
        fontSize: 10
      },
    })
    currentTop += 20

    if (urgentFee > 0) {
      panel.addPrintText({
        options: {
          width: 220, // 尝试更大的值以占满 80mm 宽度
          top: currentTop,
          left: 0,
          title: `加急费：¥${formatPrice(urgentFee)}`,
          textAlign: "right",
          fontSize: 10
        },
      })
      currentTop += 20
    }

    // 共计优惠（如果使用了优惠券）
    if (couponDiscount > 0) {
      panel.addPrintText({
        options: {
          width: 220, // 尝试更大的值以占满 80mm 宽度
          top: currentTop,
          left: 0,
          title: `共计优惠：-¥${formatPrice(couponDiscount)}`,
          textAlign: "right",
          fontSize: 10
        },
      })
      currentTop += 20
    }

    panel.addPrintText({
      options: {
        width: 220, // 尝试更大的值以占满 80mm 宽度
        top: currentTop,
        left: 0,
        title: `实付金额：¥${formatPrice(totalAmount)}`,
        textAlign: "right",
        fontSize: 12,
        fontWeight: "bold"
      },
    })
    currentTop += 30

    // 订单编号条形码（放在底部，居中）
    if (orderNumber && orderNumber !== '-') {
      panel.addPrintText({
        options: {
          width: 200, // 条形码宽度
          height: 45, // 条形码高度
          top: currentTop,
          left: 15, // 居中位置调整
          title: orderNumber,
          textType: "barcode", // 改为条形码
        },
      })
      currentTop += 60 // 条形码高度 + 间距
    }

    // 底部感谢文字
    currentTop += 10 // 增加间距
    panel.addPrintText({
      options: {
        width: 220, // 尝试更大的值以占满 80mm 宽度
        top: currentTop,
        left: 0,
        title: "微信搜索“橙心选”小程序，",
        textAlign: "center", // 居中对齐
        fontSize: 11
      },
    })
    currentTop += 20

    panel.addPrintText({
      options: {
        width: 220, // 尝试更大的值以占满 80mm 宽度
        top: currentTop,
        left: 0,
        title: "了解更多优惠产品！",
        textAlign: "center", // 居中对齐
        fontSize: 11
      },
    })

    // 使用 print2 方法进行静默打印（通过 WebSocket 发送到打印客户端）
    // 根据环境（本地/线上）自动调整打印选项
    const printOptions = await getPrintOptions({}, hiprint)
    
    // 检查连接状态
    if (!hiprint.hiwebSocket || !hiprint.hiwebSocket.opened) {
      ElMessage.error('打印机未连接，请检查连接状态')
      console.error('打印失败：WebSocket 未连接')
      return
    }
    
    console.log('开始打印，选项:', printOptions)
    hiprintTemplate.print2(panel, printOptions)
    
    // 线上环境可能需要更长的等待时间
    if (isOnlineEnvironment()) {
      console.log('✅ 线上环境打印，使用中转服务')
    } else {
      console.log('✅ 本地环境打印，直接连接')
    }
    
    ElMessage.success('打印任务已发送')
  } catch (error) {
    console.error('打印失败:', error)
    console.error('错误详情:', error.stack)
    ElMessage.error('打印失败：' + (error.message || '未知错误'))
  }
}

// 物料打印功能
const handlePrintMaterial = async (orderData) => {
  // 检查 hiprint 是否初始化
  if (!hiprint) {
    ElMessage.error('打印功能未初始化，请刷新页面重试')
    return
  }

  try {
    // 检查连接状态
    if (!hiprint.hiwebSocket || !hiprint.hiwebSocket.opened) {
      ElMessage.warning('打印机未连接，请检查打印客户端是否运行')
      return
    }

    // 如果传入的是订单对象（从列表调用），可能需要先获取订单详情
    let finalOrderData = orderData
    if (orderData.id && (!orderData.order_items || orderData.order_items.length === 0)) {
      // 从列表调用，需要先获取订单详情
      const res = await getOrderDetail(orderData.id)
      if (res && res.code === 200) {
        finalOrderData = res.data
      } else {
        ElMessage.error('获取订单详情失败，无法打印物料标签')
        return
      }
    }

    // 获取订单信息
    const order = finalOrderData.order || finalOrderData
    const user = finalOrderData.user || finalOrderData
    const orderItems = finalOrderData.order_items || []
    const orderTime = order.created_at || finalOrderData.created_at
    // 格式化日期，去掉年份
    const formatDateWithoutYear = (value) => {
      if (!value) return '-'
      const date = new Date(value)
      const month = String(date.getMonth() + 1).padStart(2, '0')
      const day = String(date.getDate()).padStart(2, '0')
      const hours = String(date.getHours()).padStart(2, '0')
      const minutes = String(date.getMinutes()).padStart(2, '0')
      return `${month}-${day} ${hours}:${minutes}`
    }
    const timeStr = orderTime ? formatDateWithoutYear(orderTime) : '-'
    
    // 调试信息：打印数据结构
    console.log('物料打印 - 订单数据:', {
      orderData: finalOrderData,
      order: order,
      user: user,
      orderItems: orderItems,
      orderItemsLength: orderItems.length
    })
    
    // 格式化客户名称（只显示后几位，其他用*代替）
    const formatCustomerName = (name) => {
      if (!name) return '***'
      const nameStr = String(name)
      if (nameStr.length <= 3) return nameStr
      const lastThree = nameStr.slice(-3)
      const stars = '*'.repeat(nameStr.length - 3)
      return stars + lastThree
    }
    
    // 计算总件数（所有商品的quantity之和）
    const totalItems = orderItems.reduce((sum, item) => sum + (item.quantity || 0), 0)
    
    console.log('物料打印 - 总件数:', totalItems, '商品列表:', orderItems)
    
    if (totalItems === 0) {
      ElMessage.warning('订单没有商品，无法打印物料标签')
      console.warn('订单没有商品，订单数据:', finalOrderData)
      return
    }

    // 循环打印，每件商品打印一张标签
    for (let currentItem = 1; currentItem <= totalItems; currentItem++) {
      // 创建打印模板
      const hiprintTemplate = new hiprint.PrintTemplate()

      // 添加打印面板（60mm*40mm）
      const panel = hiprintTemplate.addPrintPanel({
        width: 60, // 60mm宽度
        height: 40, // 40mm高度
        paperFooter: 0,
        paperHeader: 0,
        paperNumberLeft: 0,
        paperNumberRight: 0,
        paperNumberFormat: ' ',
      })

      let currentTop = 5

      // 标题：买一次性用品,橙心选更方便!
      panel.addPrintText({
        options: {
          width: 170,
          height: 20,
          top: currentTop,
          left: 0,
          title: '橙心选，进货更方便, 生意大“橙”功!',
          textAlign: 'center',
          fontSize: 12,
          fontWeight: 'bold'
        }
      })
      currentTop += 20

      // 客户信息
      if (user) {
        // 判断用户名，如果没有用户名使用用户编号
        let customerName = user.name 
          ? formatCustomerName(user.name)
          : (user.user_code ? `用户${user.user_code}` : (user.user_id ? `用户${user.user_id}` : '***'))
        
        panel.addPrintText({
          options: {
            width: 120, // 左侧区域宽度
            top: currentTop,
            left: 5,
            title: `客户: ${customerName}`,
            textAlign: 'left',
            fontSize: 9
          }
        })
      }
      currentTop += 15

      // 日期
      panel.addPrintText({
        options: {
          width: 120,
          top: currentTop,
          left: 5,
          title: `日期: ${timeStr}`,
          textAlign: 'left',
          fontSize: 8
        }
      })
      currentTop += 20

      // 件数信息：当前序号/总数 和 共总数件 分开显示，但在同一行
      const itemCountPart = `${currentItem}/${totalItems}` // 件数部分，如 1/20, 2/20, ..., 20/20
      const totalCountPart = `共${totalItems}件` // 总数部分
      
      // 件数部分：1/20, 2/20, ...
      panel.addPrintText({
        options: {
          width: 50,
          top: currentTop,
          left: 10,
          title: itemCountPart,
          textAlign: 'left',
          fontSize: 12,
          fontWeight: 'bold'
        }
      })
      // 总数部分：共20件（在同一行，右侧位置）
      panel.addPrintText({
        options: {
          width: 40,
          top: currentTop,
          left: 50, // 放在件数右侧
          title: totalCountPart,
          textAlign: 'left',
          fontSize: 8,
        }
      })
      currentTop += 20

      // 推荐信息
      panel.addPrintText({
        options: {
          width: 120,
          top: currentTop,
          left: 10,
          title: '推荐小程序下单',
          textAlign: 'left',
          fontSize: 8
        }
      })
      currentTop += 13

      panel.addPrintText({
        options: {
          width: 120,
          top: currentTop,
          left: 10,
          title: '享更多优惠~',
          textAlign: 'left',
          fontSize: 8
        }
      })

      // 右侧二维码图片
      // 注意：需要将图片转换为 base64 格式
      // 这里使用一个占位符，实际使用时需要替换为真实的二维码图片 base64 或 URL
      const qrCodeImageUrl = 'https://www.sscchh.com/minio/sch/product_1766382995.png' // TODO: 替换为实际的二维码图片 URL 或 base64
      
      if (qrCodeImageUrl) {
        try {
          // 如果是网络图片，需要转换为 base64
          const imageBase64 = qrCodeImageUrl.startsWith('http') 
            ? await convertImageToBase64(qrCodeImageUrl)
            : qrCodeImageUrl
          
          panel.addPrintImage({
            options: {
              width: 80,
              height: 80,
              top: 27,
              left: 83,
              src: imageBase64
            }
          })
        } catch (error) {
          console.error('加载二维码图片失败:', error)
        }
      }

      // 执行打印，指定打印机为 Deli DL-720C
      // 根据环境（本地/线上）自动调整打印选项
      const printOptions = await getPrintOptions({
        printer: 'Deli DL-720C'
      }, hiprint)
      
      // 使用 Promise 包装，确保每次打印任务按顺序执行
      await new Promise((resolve, reject) => {
        try {
          hiprintTemplate.print2(panel, printOptions)
          
          // 线上环境需要更长的等待时间，确保通过中转服务发送成功
          const waitTime = isOnlineEnvironment() ? 1200 : 800
          setTimeout(() => {
            resolve()
          }, waitTime)
        } catch (error) {
          reject(error)
        }
      })

      // 每张标签之间增加延迟，确保打印机有足够时间处理每张标签
      // 延迟时间根据打印机处理速度调整，这里设置为 1.5 秒
      if (currentItem < totalItems) {
        await new Promise(resolve => setTimeout(resolve, 1500))
      }
    }

    ElMessage.success(`物料打印任务已发送，共打印 ${totalItems} 张标签`)
  } catch (error) {
    console.error('物料打印失败:', error)
    ElMessage.error('物料打印失败：' + (error.message || '未知错误'))
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
