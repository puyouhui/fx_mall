<template>
  <div class="mini-users-page">
    <el-card class="mini-users-card" shadow="never">
      <div class="card-header">
        <div class="title">
          <span class="main">小程序用户</span>
          <span class="sub">查看登录用户唯一ID与基础信息</span>
        </div>
        <div class="actions">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索唯一ID / 姓名 / 电话"
            clearable
            @keyup.enter="handleSearch"
          />
          <el-button type="primary" @click="handleSearch">搜索</el-button>
        </div>
      </div>

      <el-table
        v-loading="loading"
        :data="users"
        border
        stripe
        class="mini-users-table"
      >
        <!-- <el-table-column prop="unique_id" label="唯一ID" min-width="220" /> -->
        <el-table-column prop="id" label="ID" min-width="60" />
        <el-table-column prop="user_code" label="用户编号" min-width="120">
          <template #default="scope">
            <span v-if="scope.row.user_code" class="user-code-text">用户{{ scope.row.user_code }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="用户姓名" min-width="120">
          <template #default="scope">
            {{ scope.row.name || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="phone" label="手机号" min-width="130">
          <template #default="scope">
            {{ scope.row.phone || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="coupon_count" label="优惠券数量" min-width="100" align="center">
          <template #default="scope">
            <el-tag 
              type="info" 
              style="cursor: pointer;"
              @click="handleViewUserCoupons(scope.row)"
            >
              {{ scope.row.coupon_count || 0 }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="user_type" label="用户类型" min-width="120">
          <template #default="scope">
            <el-tag :type="scope.row.user_type === 'wholesale' ? 'warning' : 'success'">
              {{ formatUserType(scope.row.user_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="profile_completed" label="资料完善" min-width="110">
          <template #default="scope">
            <el-tag :type="scope.row.profile_completed ? 'success' : 'info'">
              {{ scope.row.profile_completed ? '已完善' : '未完善' }}
            </el-tag>
          </template>
        </el-table-column>
        <!-- <el-table-column prop="is_sales_employee" label="是否为内部销售员" min-width="140" align="center">
          <template #default="scope">
            <el-tag :type="scope.row.is_sales_employee ? 'success' : 'info'">
              {{ scope.row.is_sales_employee ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column> -->
        <el-table-column prop="sales_employee" label="绑定销售员" min-width="180">
          <template #default="scope">
            <div v-if="scope.row.sales_employee" class="sales-employee-info">
              <span class="sales-employee-name">
                {{ scope.row.sales_employee.name || '未命名' }}
              </span>
              <el-tag size="small" type="info" style="margin-left: 8px;">
                {{ scope.row.sales_employee.employee_code }}
              </el-tag>
            </div>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="store_type" label="店铺类型" min-width="120">
          <template #default="scope">
            {{ scope.row.store_type || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="首次登录时间" min-width="160">
          <template #default="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" min-width="180" fixed="right">
          <template #default="scope">
            <el-button type="success" link @click="handleIssueCoupon(scope.row)">
              发放优惠券
            </el-button>
            <el-button type="primary" link @click="handleViewDetail(scope.row.id)">
              查看详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 用户详情对话框 -->
      <el-dialog
        v-model="detailDialogVisible"
        title="用户详情"
        width="900px"
        :close-on-click-modal="false"
        :lock-scroll="true"
        :modal="true"
        class="user-detail-dialog"
      >
        <div v-loading="detailLoading" class="user-detail">
          <div v-if="userDetail" class="detail-content">
            <el-tabs v-model="activeTab" class="user-detail-tabs">
              <!-- 基本信息 -->
              <el-tab-pane label="基本信息" name="basic">
                <div class="tab-content">
                  <div class="basic-info-wrapper">
                    <!-- 头像 -->
                    <div class="avatar-container">
                      <el-image
                        v-if="userDetail.avatar"
                        :src="userDetail.avatar"
                        class="avatar-image-small"
                        fit="cover"
                        :preview-src-list="[userDetail.avatar]"
                      />
                      <div v-else class="no-avatar-small">
                        <el-icon :size="20"><Picture /></el-icon>
                      </div>
                    </div>
                    <!-- 信息列表 -->
                    <div class="info-list">
                      <el-descriptions :column="2" border class="custom-descriptions">
                        <el-descriptions-item label="用户ID" label-class-name="desc-label">
                          <span class="desc-value">{{ userDetail.id }}</span>
                        </el-descriptions-item>
                        <el-descriptions-item label="用户编号" label-class-name="desc-label">
                          <span v-if="userDetail.user_code" class="desc-value user-code-text">用户{{ userDetail.user_code }}</span>
                          <span v-else class="desc-value">-</span>
                        </el-descriptions-item>
                        <el-descriptions-item label="用户姓名" label-class-name="desc-label">
                          <span class="desc-value">{{ userDetail.name || '-' }}</span>
                        </el-descriptions-item>
                        <el-descriptions-item label="唯一ID" label-class-name="desc-label">
                          <span class="desc-value unique-id">{{ userDetail.unique_id }}</span>
                        </el-descriptions-item>
                        <el-descriptions-item label="手机号" label-class-name="desc-label">
                          <span class="desc-value phone-number">{{ userDetail.phone || '-' }}</span>
                        </el-descriptions-item>
                        <el-descriptions-item label="用户类型" label-class-name="desc-label">
                          <el-tag 
                            :type="userDetail.user_type === 'wholesale' ? 'warning' : (userDetail.user_type === 'retail' ? 'success' : 'info')"
                            class="user-type-tag"
                          >
                            {{ formatUserType(userDetail.user_type) }}
                          </el-tag>
                        </el-descriptions-item>
                        <el-descriptions-item label="资料完善" label-class-name="desc-label">
                          <el-tag 
                            :type="userDetail.profile_completed ? 'success' : 'info'"
                            class="profile-tag"
                          >
                            {{ userDetail.profile_completed ? '已完善' : '未完善' }}
                          </el-tag>
                        </el-descriptions-item>
                        <el-descriptions-item label="是否为内部销售员" label-class-name="desc-label">
                          <el-tag 
                            :type="userDetail.is_sales_employee ? 'success' : 'info'"
                            class="sales-employee-tag"
                          >
                            {{ userDetail.is_sales_employee ? '是' : '否' }}
                          </el-tag>
                        </el-descriptions-item>
                        <el-descriptions-item label="积分" label-class-name="desc-label">
                          <span class="desc-value points-value">{{ userDetail.points !== undefined ? userDetail.points : 0 }}</span>
                        </el-descriptions-item>
                        <el-descriptions-item label="店铺类型" label-class-name="desc-label">
                          <span class="desc-value">{{ userDetail.store_type || '-' }}</span>
                        </el-descriptions-item>
                        <el-descriptions-item label="销售员代码" label-class-name="desc-label">
                          <span class="desc-value">{{ userDetail.sales_code || '-' }}</span>
                        </el-descriptions-item>
                        <el-descriptions-item label="首次登录时间" label-class-name="desc-label">
                          <span class="desc-value time-text">{{ formatDate(userDetail.created_at) }}</span>
                        </el-descriptions-item>
                        <el-descriptions-item label="最近更新时间" label-class-name="desc-label">
                          <span class="desc-value time-text">{{ formatDate(userDetail.updated_at) }}</span>
                        </el-descriptions-item>
                      </el-descriptions>
                    </div>
                  </div>
                </div>
              </el-tab-pane>

              <!-- 收货地址 -->
              <el-tab-pane label="收货地址" name="addresses">
                <div class="tab-content">
                  <div class="section-title">
                    <span>收货地址</span>
                    <span class="address-count" v-if="userDetail.addresses && userDetail.addresses.length > 0">
                      (共{{ userDetail.addresses.length }}个)
                    </span>
                  </div>
                  <div class="section-content">
                    <div v-if="userDetail.addresses && userDetail.addresses.length > 0" class="addresses-list">
                      <div 
                        v-for="address in userDetail.addresses" 
                        :key="address.id" 
                        class="address-item"
                        :class="{ 'is-default': address.is_default }"
                      >
                        <div class="address-header">
                          <el-tag v-if="address.is_default" type="success" size="small">默认地址</el-tag>
                          <span style="margin-left: auto;"></span>
                          <el-button 
                            type="primary" 
                            link 
                            size="small" 
                            @click="handleEditAddress(address)"
                          >
                            编辑
                          </el-button>
                          <el-button
                            type="danger"
                            link
                            size="small"
                            @click="handleDeleteAddress(address)"
                          >
                            删除
                          </el-button>
                        </div>
                        <el-descriptions :column="2" border class="address-descriptions">
                          <el-descriptions-item label="地址名称" label-class-name="desc-label">
                            <span class="desc-value">{{ address.name || '-' }}</span>
                          </el-descriptions-item>
                          <el-descriptions-item label="联系人" label-class-name="desc-label">
                            <span class="desc-value">{{ address.contact || '-' }}</span>
                          </el-descriptions-item>
                          <el-descriptions-item label="手机号" label-class-name="desc-label">
                            <span class="desc-value phone-number">{{ address.phone || '-' }}</span>
                          </el-descriptions-item>
                          <el-descriptions-item label="店铺类型" label-class-name="desc-label">
                            <span class="desc-value">{{ address.store_type || '-' }}</span>
                          </el-descriptions-item>
                          <el-descriptions-item label="详细地址" label-class-name="desc-label" :span="2">
                            <span class="desc-value address-text">{{ address.address || '-' }}</span>
                          </el-descriptions-item>
                          <el-descriptions-item label="经纬度" label-class-name="desc-label">
                            <span v-if="address.latitude && address.longitude" class="desc-value coordinates">
                              {{ address.latitude }}, {{ address.longitude }}
                            </span>
                            <span v-else class="desc-value">-</span>
                          </el-descriptions-item>
                          <el-descriptions-item label="门头照片" label-class-name="desc-label" v-if="address.avatar" :span="2">
                            <el-image
                              :src="address.avatar"
                              class="address-avatar-image"
                              fit="cover"
                              :preview-src-list="[address.avatar]"
                            />
                          </el-descriptions-item>
                        </el-descriptions>
                      </div>
                    </div>
                    <el-empty v-else description="暂无地址" :image-size="80" />
                  </div>
                </div>
              </el-tab-pane>

              <!-- 发票信息 -->
              <el-tab-pane label="发票信息" name="invoice">
                <div class="tab-content" v-if="invoiceForm">
                  <el-form :model="invoiceForm" label-width="140px" class="invoice-form">
                    <el-row :gutter="20">
                      <el-col :span="12">
                        <el-form-item label="发票类型">
                          <el-radio-group v-model="invoiceForm.invoice_type">
                            <el-radio label="company">企业</el-radio>
                            <el-radio label="personal">个人</el-radio>
                          </el-radio-group>
                        </el-form-item>
                      </el-col>
                      <el-col :span="12">
                        <el-form-item label="发票抬头" required>
                          <el-input v-model="invoiceForm.title" placeholder="请输入发票抬头" />
                        </el-form-item>
                      </el-col>
                    </el-row>
                    <el-row :gutter="20" v-if="invoiceForm && invoiceForm.invoice_type === 'company'">
                      <el-col :span="12">
                        <el-form-item label="纳税人识别号" required>
                          <el-input v-model="invoiceForm.tax_number" placeholder="请输入纳税人识别号" />
                        </el-form-item>
                      </el-col>
                      <el-col :span="12">
                        <el-form-item label="公司地址">
                          <el-input v-model="invoiceForm.company_address" placeholder="请输入公司地址" />
                        </el-form-item>
                      </el-col>
                    </el-row>
                    <el-row :gutter="20" v-if="invoiceForm && invoiceForm.invoice_type === 'company'">
                      <el-col :span="12">
                        <el-form-item label="公司电话">
                          <el-input v-model="invoiceForm.company_phone" placeholder="请输入公司电话" />
                        </el-form-item>
                      </el-col>
                      <el-col :span="12">
                        <el-form-item label="开户银行">
                          <el-input v-model="invoiceForm.bank_name" placeholder="请输入开户银行" />
                        </el-form-item>
                      </el-col>
                    </el-row>
                    <el-row :gutter="20" v-if="invoiceForm && invoiceForm.invoice_type === 'company'">
                      <el-col :span="12">
                        <el-form-item label="银行账号">
                          <el-input v-model="invoiceForm.bank_account" placeholder="请输入银行账号" />
                        </el-form-item>
                      </el-col>
                    </el-row>
                  </el-form>
                </div>
              </el-tab-pane>

            </el-tabs>
          </div>
        </div>
        <template #footer>
          <div class="dialog-footer">
            <el-button @click="detailDialogVisible = false">关闭</el-button>
            <el-button 
              v-if="activeTab === 'invoice'" 
              type="primary" 
              :loading="invoiceSaving" 
              @click="handleSaveInvoice"
            >
              保存发票
            </el-button>
            <el-button 
              v-if="activeTab === 'basic'" 
              type="primary" 
              @click="handleEdit"
            >
              编辑
            </el-button>
          </div>
        </template>
      </el-dialog>

      <!-- 编辑用户对话框 -->
      <el-dialog
        v-model="editDialogVisible"
        title="编辑用户信息"
        width="700px"
        :close-on-click-modal="false"
      >
        <el-form
          ref="editFormRef"
          :model="editForm"
          :rules="editFormRules"
          label-width="120px"
        >
          <el-form-item label="用户姓名" prop="name">
            <el-input v-model="editForm.name" placeholder="请输入用户姓名" />
          </el-form-item>
          <el-form-item label="手机号" prop="phone">
            <el-input v-model="editForm.phone" placeholder="请输入手机号" />
          </el-form-item>
          <el-form-item label="店铺类型" prop="storeType">
            <el-input v-model="editForm.storeType" placeholder="请输入店铺类型" />
          </el-form-item>
          <el-form-item label="绑定销售员" prop="salesEmployeeId">
            <el-select 
              v-model="editForm.salesEmployeeId" 
              placeholder="请选择销售员" 
              clearable
              style="width: 100%"
              @change="handleSalesEmployeeChange"
            >
              <el-option
                v-for="emp in salesEmployees"
                :key="emp.id"
                :label="`${emp.name || emp.employee_code} (${emp.employee_code})`"
                :value="emp.id"
              >
                <span>{{ emp.name || '未命名' }}</span>
                <span style="color: #8492a6; font-size: 13px; margin-left: 8px;">{{ emp.employee_code }}</span>
              </el-option>
            </el-select>
          </el-form-item>
          <el-form-item label="用户头像" prop="avatar">
            <el-upload
              class="avatar-uploader"
              :action="uploadAvatarUrl"
              :show-file-list="false"
              :on-success="handleAvatarSuccess"
              :before-upload="beforeAvatarUpload"
              :headers="uploadHeaders"
            >
              <el-image
                v-if="editForm.avatar"
                :src="editForm.avatar"
                class="avatar"
                fit="cover"
              />
              <el-icon v-else class="avatar-uploader-icon"><Plus /></el-icon>
            </el-upload>
            <div class="upload-tip">支持 JPG、PNG 格式，大小不超过 5MB</div>
          </el-form-item>
          <el-form-item label="用户类型" prop="userType">
            <el-select v-model="editForm.userType" placeholder="请选择用户类型" style="width: 100%">
              <el-option label="未选择" value="unknown" />
              <el-option label="零售用户" value="retail" />
              <el-option label="批发用户" value="wholesale" />
            </el-select>
          </el-form-item>
          <el-form-item label="资料完善" prop="profileCompleted">
            <el-switch
              v-model="editForm.profileCompleted"
              active-text="已完善"
              inactive-text="未完善"
            />
          </el-form-item>
          <el-form-item label="设置为销售员" prop="isSalesEmployee">
            <el-switch
              v-model="editForm.isSalesEmployee"
              active-text="是"
              inactive-text="否"
              @change="handleSalesEmployeeSwitchChange"
            />
            <div style="font-size: 12px; color: #909399; margin-top: 5px;">
              设置为销售员后，该用户可以通过小程序查看其负责客户的订单详情
            </div>
          </el-form-item>
        </el-form>
        <template #footer>
          <span class="dialog-footer">
            <el-button @click="editDialogVisible = false">取消</el-button>
            <el-button type="primary" :loading="editSubmitting" @click="handleSaveEdit">
              保存
            </el-button>
          </span>
        </template>
      </el-dialog>

      <!-- 编辑地址对话框 -->
      <el-dialog
        v-model="addressEditDialogVisible"
        title="编辑地址"
        width="800px"
        :close-on-click-modal="false"
        destroy-on-close
        class="address-edit-dialog"
      >
        <el-form
          ref="addressEditFormRef"
          :model="addressEditForm"
          :rules="addressEditFormRules"
          label-width="100px"
          class="address-edit-form"
        >
          <el-form-item label="地址名称" prop="name">
            <el-input v-model="addressEditForm.name" placeholder="请输入地址名称" />
          </el-form-item>
          <el-form-item label="联系人" prop="contact">
            <el-input v-model="addressEditForm.contact" placeholder="请输入联系人" />
          </el-form-item>
          <el-form-item label="手机号" prop="phone">
            <el-input v-model="addressEditForm.phone" placeholder="请输入手机号" />
          </el-form-item>
          <el-form-item label="详细地址" prop="address">
            <div style="display: flex; gap: 8px; align-items: flex-start;">
              <el-input 
                v-model="addressEditForm.address" 
                type="textarea" 
                :rows="3"
                placeholder="请输入详细地址或点击选择地址按钮"
                style="flex: 1"
              />
              <div style="display: flex; flex-direction: column; gap: 8px;">
                <el-button 
                  type="primary" 
                  @click="showAddressPicker = true"
                  style="white-space: nowrap"
                >
                  选择地址
                </el-button>
                <el-button 
                  type="default" 
                  :loading="geocoding"
                  @click="handleGeocodeAddress"
                  style="white-space: nowrap"
                >
                  解析地址
                </el-button>
              </div>
            </div>
          </el-form-item>
          <el-form-item label="店铺类型">
            <el-input v-model="addressEditForm.storeType" placeholder="请输入店铺类型" />
          </el-form-item>
          <el-form-item label="门头照片">
            <div class="avatar-upload-wrapper">
              <el-upload
                class="avatar-uploader"
                :action="uploadAddressAvatarUrl"
                :show-file-list="false"
                :on-success="handleAddressAvatarSuccess"
                :before-upload="beforeAddressAvatarUpload"
                :headers="uploadHeaders"
                :on-error="handleAddressAvatarError"
              >
                <el-image
                  v-if="addressEditForm.avatar"
                  :src="addressEditForm.avatar"
                  class="avatar-image"
                  fit="cover"
                  :preview-src-list="[addressEditForm.avatar]"
                />
                <div v-else class="avatar-upload-placeholder">
                  <el-icon class="avatar-upload-icon"><Plus /></el-icon>
                  <div class="avatar-upload-text">上传门头照片</div>
                </div>
              </el-upload>
              <div class="upload-tip">支持 JPG、PNG 格式，大小不超过 5MB</div>
            </div>
          </el-form-item>
          <el-form-item label="经纬度">
            <div style="display: flex; gap: 12px;">
              <el-input 
                v-model="addressEditForm.longitude" 
                placeholder="经度"
                style="flex: 1"
                type="number"
              />
              <el-input 
                v-model="addressEditForm.latitude" 
                placeholder="纬度"
                style="flex: 1"
                type="number"
              />
            </div>
          </el-form-item>
          <el-form-item label="设为默认地址">
            <el-switch
              v-model="addressEditForm.isDefault"
              active-text="是"
              inactive-text="否"
            />
          </el-form-item>
        </el-form>
        <template #footer>
          <span class="dialog-footer">
            <el-button @click="addressEditDialogVisible = false">取消</el-button>
            <el-button type="primary" :loading="addressEditSubmitting" @click="handleSaveAddressEdit">
              保存
            </el-button>
          </span>
        </template>
      </el-dialog>

      <!-- 地址选择弹窗 -->
      <el-dialog
        v-model="showAddressPicker"
        title="选择地址"
        width="90%"
        :close-on-click-modal="false"
        destroy-on-close
      >
        <div id="addressPickerMap" style="width: 100%; height: 500px;"></div>
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

      <!-- 发放优惠券弹窗 -->
      <el-dialog
        v-model="issueCouponDialogVisible"
        title="发放优惠券"
        width="600px"
        destroy-on-close
      >
        <el-form label-width="120px">
          <el-form-item label="用户信息">
            <div v-if="selectedUser">
              <div><strong>用户ID：</strong>{{ selectedUser.id }}</div>
              <div style="margin-top: 8px;" v-if="selectedUser.name">
                <strong>姓名：</strong>{{ selectedUser.name }}
              </div>
              <div style="margin-top: 8px;" v-if="selectedUser.phone">
                <strong>手机号：</strong>{{ selectedUser.phone }}
              </div>
            </div>
          </el-form-item>

          <el-form-item label="选择优惠券" required>
            <el-select
              v-model="issueCouponForm.couponId"
              filterable
              placeholder="请选择优惠券"
              style="width: 100%"
              clearable
            >
              <el-option
                v-for="coupon in availableCoupons"
                :key="coupon.id"
                :label="getCouponLabel(coupon)"
                :value="coupon.id"
              >
                <div style="display: flex; justify-content: space-between;">
                  <span>{{ coupon.name }}</span>
                  <span style="color: #909399; font-size: 12px;">
                    {{ coupon.type === 'delivery_fee' ? '配送费券' : `¥${(coupon.discount_value || 0).toFixed(2)}` }}
                  </span>
                </div>
              </el-option>
            </el-select>
          </el-form-item>

          <el-form-item label="发放数量" required>
            <el-input-number
              v-model="issueCouponForm.quantity"
              :min="1"
              :max="100"
              :step="1"
              controls-position="right"
              style="width: 100%"
            />
            <div class="upload-tip">默认1张，最多可发放100张</div>
          </el-form-item>

          <el-form-item label="有效期设置">
            <el-radio-group v-model="issueCouponForm.expireType" @change="handleExpireTypeChange">
              <el-radio label="none">不限制</el-radio>
              <el-radio label="days">N天后过期</el-radio>
              <el-radio label="date">指定日期</el-radio>
            </el-radio-group>
            <div v-if="issueCouponForm.expireType === 'days'" style="margin-top: 10px;">
              <el-input-number
                v-model="issueCouponForm.expiresIn"
                :min="1"
                :max="365"
                :step="1"
                controls-position="right"
                placeholder="请输入天数"
                style="width: 100%"
              />
              <div class="upload-tip">从发放时开始计算，N天后过期</div>
            </div>
            <div v-if="issueCouponForm.expireType === 'date'" style="margin-top: 10px;">
              <el-date-picker
                v-model="issueCouponForm.expiresAt"
                type="datetime"
                placeholder="选择过期日期"
                format="YYYY-MM-DD HH:mm:ss"
                value-format="YYYY-MM-DD HH:mm:ss"
                style="width: 100%"
              />
            </div>
          </el-form-item>
        </el-form>

        <template #footer>
          <el-button @click="issueCouponDialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="issuingCoupon" @click="handleIssueCouponSubmit">确定发放</el-button>
        </template>
      </el-dialog>

      <!-- 查看用户优惠券弹窗 -->
      <el-dialog
        v-model="userCouponsDialogVisible"
        title="用户优惠券列表"
        width="900px"
        destroy-on-close
      >
        <div v-if="currentViewUser" style="margin-bottom: 20px; padding: 15px; background: #f5f7fa; border-radius: 4px;">
          <div><strong>用户ID：</strong>{{ currentViewUser.id }}</div>
          <div style="margin-top: 8px;" v-if="currentViewUser.name">
            <strong>姓名：</strong>{{ currentViewUser.name }}
          </div>
          <div style="margin-top: 8px;" v-if="currentViewUser.phone">
            <strong>手机号：</strong>{{ currentViewUser.phone }}
          </div>
        </div>

        <el-table
          v-loading="userCouponsLoading"
          :data="userCoupons"
          border
          stripe
          style="width: 100%"
        >
          <el-table-column prop="coupon.name" label="优惠券名称" min-width="150" />
          <el-table-column label="类型" width="100" align="center">
            <template #default="{ row }">
              <el-tag :type="row.coupon?.type === 'delivery_fee' ? 'success' : 'warning'">
                {{ row.coupon?.type === 'delivery_fee' ? '配送费券' : '金额券' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="优惠值" width="120" align="center">
            <template #default="{ row }">
              <span v-if="row.coupon?.type === 'delivery_fee'">免配送费</span>
              <span v-else>¥{{ (row.coupon?.discount_value || 0).toFixed(2) }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100" align="center">
            <template #default="{ row }">
              <el-tag 
                :type="row.status === 'used' ? 'success' : row.status === 'expired' ? 'danger' : 'info'"
              >
                {{ row.status === 'used' ? '已使用' : row.status === 'expired' ? '已过期' : '未使用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="有效期" min-width="200">
            <template #default="{ row }">
              <div v-if="row.expires_at">
                {{ formatDateTime(row.expires_at) }}
              </div>
              <div v-else style="color: #909399;">不限制</div>
            </template>
          </el-table-column>
          <el-table-column label="优惠券有效期" min-width="200">
            <template #default="{ row }">
              <div v-if="row.coupon">
                <div>{{ formatDate(row.coupon.valid_from) }}</div>
                <div style="color: #909399; font-size: 12px;">至 {{ formatDate(row.coupon.valid_to) }}</div>
              </div>
              <div v-else>-</div>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="发放时间" min-width="160">
            <template #default="{ row }">
              {{ formatDateTime(row.created_at) }}
            </template>
          </el-table-column>
        </el-table>

        <template #footer>
          <el-button @click="userCouponsDialogVisible = false">关闭</el-button>
        </template>
      </el-dialog>

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
  </div>
</template>

<script setup>
import { reactive, ref, onMounted, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Picture, Plus } from '@element-plus/icons-vue'
import { getMiniUsers, getMiniUserDetail, updateMiniUser, getAdminAddressDetail, updateAdminAddress, deleteAdminAddress, getSalesEmployees, uploadUserAvatar, getUserCoupons, geocodeAddress, reverseGeocode, saveInvoice } from '../api/miniUsers'
import { getMapSettings } from '../api/settings'
import { getCoupons, issueCouponToUser } from '../api/coupons'

const loading = ref(false)
const users = ref([])
const searchKeyword = ref('')
const pagination = reactive({
  pageNum: 1,
  pageSize: 10,
  total: 0
})

// 用户详情相关
const detailDialogVisible = ref(false)
const detailLoading = ref(false)
const userDetail = ref(null)

// 编辑相关
const editDialogVisible = ref(false)
const editSubmitting = ref(false)
const editFormRef = ref(null)
const editForm = reactive({
  name: '',
  phone: '',
  storeType: '',
  salesCode: '',
  salesEmployeeId: null,
  avatar: '',
  userType: 'unknown',
  profileCompleted: false,
  isSalesEmployee: false
})

// 上传头像相关
const uploadAvatarUrl = computed(() => {
  if (!userDetail.value) return ''
  // 使用完整的 API 路径（与 request.js 中的 baseURL 保持一致）
  // 根据环境选择 API 地址
  const baseURL = (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1')
    ? 'http://localhost:8082/api/mini'
    : '/api_mall/mini'
  return `${baseURL}/admin/mini-app/users/${userDetail.value.id}/avatar`
})

const uploadHeaders = computed(() => {
  const token = localStorage.getItem('token')
  return {
    'Authorization': `Bearer ${token}`
  }
})

// 地址头像上传URL（管理员接口）
const uploadAddressAvatarUrl = computed(() => {
  // 根据环境选择 API 地址
  const baseURL = (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1')
    ? 'http://localhost:8082/api/mini'
    : '/api_mall/mini'
  return `${baseURL}/admin/mini-app/addresses/avatar`
})
const salesEmployees = ref([])

// 发放优惠券相关
const issueCouponDialogVisible = ref(false)
const selectedUser = ref(null)
const availableCoupons = ref([])
const issuingCoupon = ref(false)
const issueCouponForm = reactive({
  couponId: null,
  quantity: 1,
  expireType: 'none', // none, days, date
  expiresIn: 30, // 天数
  expiresAt: null // 指定日期
})

// 查看用户优惠券相关
const userCouponsDialogVisible = ref(false)
const currentViewUser = ref(null)
const userCoupons = ref([])
const userCouponsLoading = ref(false)

const editFormRules = {
  phone: [
    { required: false, message: '请输入手机号', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (!value || value.trim() === '') {
          callback()
        } else if (!/^1[3-9]\d{9}$/.test(value)) {
          callback(new Error('请输入正确的手机号码'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

// 地址编辑相关
const addressEditDialogVisible = ref(false)
const addressEditSubmitting = ref(false)
const geocoding = ref(false)
const addressEditFormRef = ref(null)
const addressEditForm = reactive({
  name: '',
  contact: '',
  phone: '',
  address: '',
  avatar: '',
  storeType: '',
  latitude: null,
  longitude: null,
  isDefault: false
})
const editingAddressId = ref(null)

// 地址选择器相关
const showAddressPicker = ref(false)
const selectedAddressText = ref('')
const selectedLatitude = ref(null)
const selectedLongitude = ref(null)
let addressPickerMap = null
let addressPickerMarker = null
let addressPickerGeocoder = null

// 发票信息相关
const invoiceForm = ref({
  invoice_type: 'company',
  title: '',
  tax_number: '',
  company_address: '',
  company_phone: '',
  bank_name: '',
  bank_account: ''
})
const invoiceSaving = ref(false)

// Tab 切换
const activeTab = ref('basic')

const addressEditFormRules = {
  name: [
    { required: true, message: '请输入地址名称', trigger: 'blur' }
  ],
  contact: [
    { required: true, message: '请输入联系人', trigger: 'blur' }
  ],
  phone: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (!value || value.trim() === '') {
          callback(new Error('请输入手机号'))
        } else if (!/^1[3-9]\d{9}$/.test(value)) {
          callback(new Error('请输入正确的手机号码'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ],
  address: [
    { required: true, message: '请输入详细地址', trigger: 'blur' }
  ]
}

const loadUsers = async () => {
  loading.value = true
  try {
    const res = await getMiniUsers({
      pageNum: pagination.pageNum,
      pageSize: pagination.pageSize,
      keyword: searchKeyword.value
    })
    if (res.code === 200) {
      users.value = Array.isArray(res.data) ? res.data : []
      pagination.total = res.total || users.value.length
    }
  } catch (error) {
    console.error('获取用户失败:', error)
    ElMessage.error('获取用户列表失败，请稍后再试')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.pageNum = 1
  loadUsers()
}

const handlePageChange = (page) => {
  pagination.pageNum = page
  loadUsers()
}

const formatDate = (value) => {
  if (!value) return '-'
  return new Date(value).toLocaleString()
}

const formatUserType = (type) => {
  if (type === 'wholesale') return '批发用户'
  if (type === 'retail') return '零售用户'
  return '未选择'
}

const handleViewDetail = async (id) => {
  detailDialogVisible.value = true
  detailLoading.value = true
  userDetail.value = null
  activeTab.value = 'basic' // 重置到第一个 tab
  
  // 立即初始化发票表单，避免模板渲染时出错
  invoiceForm.value = {
    invoice_type: 'company',
    title: '',
    tax_number: '',
    company_address: '',
    company_phone: '',
    bank_name: '',
    bank_account: ''
  }
  
  try {
    const res = await getMiniUserDetail(id)
    if (res.code === 200) {
      userDetail.value = res.data
      // 初始化发票表单
      if (res.data && res.data.invoice) {
        invoiceForm.value = {
          invoice_type: res.data.invoice.invoice_type || 'company',
          title: res.data.invoice.title || '',
          tax_number: res.data.invoice.tax_number || '',
          company_address: res.data.invoice.company_address || '',
          company_phone: res.data.invoice.company_phone || '',
          bank_name: res.data.invoice.bank_name || '',
          bank_account: res.data.invoice.bank_account || ''
        }
      }
    } else {
      ElMessage.error(res.message || '获取用户详情失败')
      detailDialogVisible.value = false
    }
  } catch (error) {
    console.error('获取用户详情失败:', error)
    ElMessage.error('获取用户详情失败，请稍后再试')
    detailDialogVisible.value = false
  } finally {
    detailLoading.value = false
  }
}

const handleSaveInvoice = async () => {
  if (!userDetail.value || !invoiceForm.value) return
  
  // 验证必填字段
  if (!invoiceForm.value.title || invoiceForm.value.title.trim() === '') {
    ElMessage.warning('请输入发票抬头')
    return
  }
  
  if (invoiceForm.value.invoice_type === 'company' && (!invoiceForm.value.tax_number || invoiceForm.value.tax_number.trim() === '')) {
    ElMessage.warning('企业发票纳税人识别号不能为空')
    return
  }
  
  invoiceSaving.value = true
  try {
    const res = await saveInvoice(userDetail.value.id, {
      invoice_type: invoiceForm.value.invoice_type,
      title: invoiceForm.value.title.trim(),
      tax_number: invoiceForm.value.tax_number.trim(),
      company_address: invoiceForm.value.company_address.trim(),
      company_phone: invoiceForm.value.company_phone.trim(),
      bank_name: invoiceForm.value.bank_name.trim(),
      bank_account: invoiceForm.value.bank_account.trim(),
      is_default: true
    })
    
    if (res.code === 200) {
      ElMessage.success('保存成功')
      // 更新用户详情中的发票信息
      if (userDetail.value) {
        userDetail.value.invoice = res.data
      }
    } else {
      ElMessage.error(res.message || '保存失败')
    }
  } catch (error) {
    console.error('保存发票失败:', error)
    ElMessage.error('保存失败，请稍后再试')
  } finally {
    invoiceSaving.value = false
  }
}

const loadSalesEmployees = async () => {
  try {
    const res = await getSalesEmployees()
    if (res.code === 200) {
      salesEmployees.value = res.data || []
    }
  } catch (error) {
    console.error('获取销售员列表失败:', error)
  }
}

const handleEdit = async () => {
  if (!userDetail.value) return
  
  // 加载销售员列表
  await loadSalesEmployees()
  
  // 填充编辑表单
  editForm.name = userDetail.value.name || ''
  editForm.phone = userDetail.value.phone || ''
  editForm.storeType = userDetail.value.store_type || ''
  editForm.salesCode = userDetail.value.sales_code || ''
  editForm.salesEmployeeId = null
  
  // 根据sales_code找到对应的销售员ID
  if (editForm.salesCode && salesEmployees.value.length > 0) {
    const salesEmployee = salesEmployees.value.find(emp => emp.employee_code === editForm.salesCode)
    if (salesEmployee) {
      editForm.salesEmployeeId = salesEmployee.id
    }
  }
  
  editForm.avatar = userDetail.value.avatar || ''
  editForm.userType = userDetail.value.user_type || 'unknown'
  editForm.profileCompleted = userDetail.value.profile_completed || false
  editForm.isSalesEmployee = userDetail.value.is_sales_employee || false
  
  editDialogVisible.value = true
}

// 头像上传成功
const handleAvatarSuccess = (response) => {
  if (response.code === 200 && response.data && response.data.avatar) {
    editForm.avatar = response.data.avatar
    ElMessage.success('头像上传成功')
  } else {
    ElMessage.error(response.message || '头像上传失败')
  }
}

// 上传前验证
const beforeAvatarUpload = (file) => {
  const isImage = file.type.startsWith('image/')
  const isLt5M = file.size / 1024 / 1024 < 5

  if (!isImage) {
    ElMessage.error('只能上传图片文件!')
    return false
  }
  if (!isLt5M) {
    ElMessage.error('图片大小不能超过 5MB!')
    return false
  }
  return true
}

const handleSalesEmployeeChange = (employeeId) => {
  if (employeeId) {
    const employee = salesEmployees.value.find(emp => emp.id === employeeId)
    if (employee) {
      editForm.salesCode = employee.employee_code
    }
  } else {
    editForm.salesCode = ''
  }
}

// 处理销售员开关变化
const handleSalesEmployeeSwitchChange = (value) => {
  if (value) {
    // 如果设置为销售员，但没有选择销售员ID，提示用户
    if (!editForm.salesEmployeeId) {
      ElMessage.warning('请先选择绑定的销售员')
      // 延迟重置，让用户看到提示
      setTimeout(() => {
        editForm.isSalesEmployee = false
      }, 100)
    }
  } else {
    // 取消销售员身份时，清空销售员ID
    editForm.salesEmployeeId = null
    editForm.salesCode = ''
  }
}

const handleSaveEdit = async () => {
  if (!editFormRef.value) return
  
  await editFormRef.value.validate(async (valid) => {
    if (!valid) return
    
    if (!userDetail.value) return
    
    editSubmitting.value = true
    try {
      const updateData = {
        name: editForm.name,
        phone: editForm.phone,
        storeType: editForm.storeType,
        avatar: editForm.avatar,
        userType: editForm.userType,
        profileCompleted: editForm.profileCompleted,
        isSalesEmployee: editForm.isSalesEmployee
      }
      
      // 处理销售员绑定
      if (editForm.salesEmployeeId) {
        updateData.salesEmployeeId = editForm.salesEmployeeId
      } else {
        // 如果清空了选择，清除绑定
        updateData.salesCode = ''
      }
      
      // 如果设置为销售员，必须要有销售员ID
      if (editForm.isSalesEmployee && !editForm.salesEmployeeId) {
        ElMessage.warning('设置为销售员时，必须选择绑定的销售员')
        return
      }
      
      const res = await updateMiniUser(userDetail.value.id, updateData)
      if (res.code === 200) {
        ElMessage.success('更新成功')
        editDialogVisible.value = false
        // 刷新用户详情
        await handleViewDetail(userDetail.value.id)
        // 刷新用户列表
        await loadUsers()
      } else {
        ElMessage.error(res.message || '更新失败')
      }
    } catch (error) {
      console.error('更新用户失败:', error)
      ElMessage.error('更新用户失败，请稍后再试')
    } finally {
      editSubmitting.value = false
    }
  })
}

const handleEditAddress = async (address) => {
  editingAddressId.value = address.id
  addressEditForm.name = address.name || ''
  addressEditForm.contact = address.contact || ''
  addressEditForm.phone = address.phone || ''
  addressEditForm.address = address.address || ''
  addressEditForm.avatar = address.avatar || ''
  addressEditForm.storeType = address.store_type || ''
  addressEditForm.latitude = address.latitude || null
  addressEditForm.longitude = address.longitude || null
  addressEditForm.isDefault = address.is_default || false
  
  addressEditDialogVisible.value = true
}

const handleDeleteAddress = async (address) => {
  if (!userDetail.value || !address) return

  try {
    await ElMessageBox.confirm(
      `确定要删除该地址吗？\n\n地址名称：${address.name || '-'}\n联系人：${address.contact || '-'}\n电话：${address.phone || '-'}\n详细地址：${address.address || '-'}`,
      '删除地址',
      {
        type: 'warning',
        confirmButtonText: '删除',
        cancelButtonText: '取消'
      }
    )
  } catch {
    return
  }

  try {
    const res = await deleteAdminAddress(address.id)
    if (res.code === 200) {
      ElMessage.success('地址已删除')
      await handleViewDetail(userDetail.value.id)
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch (error) {
    console.error('删除地址失败:', error)
    ElMessage.error(error?.response?.data?.message || '删除失败，请稍后再试')
  }
}

const handleGeocodeAddress = async () => {
  if (!addressEditForm.address || !addressEditForm.address.trim()) {
    ElMessage.warning('请先输入详细地址')
    return
  }

  geocoding.value = true
  try {
    const res = await geocodeAddress(addressEditForm.address.trim())
    if (res.code === 200 && res.data && res.data.success) {
      addressEditForm.latitude = res.data.latitude
      addressEditForm.longitude = res.data.longitude
      ElMessage.success('地址解析成功')
    } else {
      ElMessage.error(res.message || res.data?.message || '地址解析失败，请检查地址是否正确')
    }
  } catch (error) {
    console.error('地址解析失败:', error)
    ElMessage.error('地址解析失败，请稍后再试')
  } finally {
    geocoding.value = false
  }
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
  const defaultCenter = addressEditForm.longitude && addressEditForm.latitude 
    ? [addressEditForm.longitude, addressEditForm.latitude]
    : [102.712251, 25.040609] // 默认昆明市

  addressPickerMap = new AMap.Map('addressPickerMap', {
    zoom: 15,
    center: defaultCenter
  })

  // 创建地理编码实例
  addressPickerGeocoder = new AMap.Geocoder({
    city: '全国'
  })

  // 如果已有地址，创建并显示标记
  if (addressEditForm.longitude && addressEditForm.latitude) {
    addressPickerMarker = new AMap.Marker({
      position: [addressEditForm.longitude, addressEditForm.latitude],
      draggable: false
    })
    addressPickerMap.add(addressPickerMarker)
    updateAddressFromCoordinates(addressEditForm.longitude, addressEditForm.latitude)
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
  addressEditForm.longitude = selectedLongitude.value
  addressEditForm.latitude = selectedLatitude.value

  // 调用接口进行逆地理编码，获取地址
  try {
    geocoding.value = true
    const res = await reverseGeocode(selectedLongitude.value, selectedLatitude.value)
    if (res.code === 200 && res.data && res.data.success && res.data.address) {
      addressEditForm.address = res.data.address
      ElMessage.success('地址解析成功')
    } else {
      // 如果接口解析失败，使用地图反解析的地址作为备用
      addressEditForm.address = selectedAddressText.value || ''
      ElMessage.warning(res.message || '接口解析失败，使用地图解析的地址')
    }
  } catch (error) {
    console.error('逆地理编码失败:', error)
    // 如果接口解析失败，使用地图反解析的地址作为备用
    addressEditForm.address = selectedAddressText.value || ''
    ElMessage.warning('接口解析失败，使用地图解析的地址')
  } finally {
    geocoding.value = false
  }

  showAddressPicker.value = false
}

// 监听地址选择器弹窗
watch(showAddressPicker, (newVal) => {
  if (newVal) {
    // 延迟初始化，确保DOM已渲染
    setTimeout(() => {
      initAddressPicker()
    }, 100)
  } else {
    // 销毁地图实例
    if (addressPickerMap) {
      addressPickerMap.destroy()
      addressPickerMap = null
      addressPickerMarker = null
      addressPickerGeocoder = null
    }
  }
})

// 地址头像上传成功
const handleAddressAvatarSuccess = (response) => {
  if (response.code === 200 && response.data) {
    // 兼容多种返回字段名
    const imageUrl = response.data.url || response.data.imageUrl || response.data.avatar
    if (imageUrl) {
      addressEditForm.avatar = imageUrl
      ElMessage.success('门头照片上传成功')
    } else {
      ElMessage.error('上传成功但未返回图片URL')
    }
  } else {
    ElMessage.error(response.message || '上传失败')
  }
}

// 地址头像上传前验证
const beforeAddressAvatarUpload = (file) => {
  const isImage = file.type.startsWith('image/')
  const isLt5M = file.size / 1024 / 1024 < 5

  if (!isImage) {
    ElMessage.error('只能上传图片文件!')
    return false
  }
  if (!isLt5M) {
    ElMessage.error('图片大小不能超过 5MB!')
    return false
  }
  return true
}

// 地址头像上传失败
const handleAddressAvatarError = (error) => {
  console.error('门头照片上传失败:', error)
  ElMessage.error('门头照片上传失败，请稍后再试')
}

const handleSaveAddressEdit = async () => {
  if (!addressEditFormRef.value) return
  
  await addressEditFormRef.value.validate(async (valid) => {
    if (!valid) return
    
    if (!editingAddressId.value) return
    
    addressEditSubmitting.value = true
    try {
      const updateData = {
        name: addressEditForm.name,
        contact: addressEditForm.contact,
        phone: addressEditForm.phone,
        address: addressEditForm.address,
        avatar: addressEditForm.avatar,
        storeType: addressEditForm.storeType,
        latitude: addressEditForm.latitude,
        longitude: addressEditForm.longitude,
        isDefault: addressEditForm.isDefault
      }
      
      const res = await updateAdminAddress(editingAddressId.value, updateData)
      if (res.code === 200) {
        ElMessage.success('地址更新成功')
        addressEditDialogVisible.value = false
        // 刷新用户详情
        if (userDetail.value) {
          await handleViewDetail(userDetail.value.id)
        }
      } else {
        ElMessage.error(res.message || '更新失败')
      }
    } catch (error) {
      console.error('更新地址失败:', error)
      ElMessage.error('更新地址失败，请稍后再试')
    } finally {
      addressEditSubmitting.value = false
    }
  })
}

// 有效期类型改变
const handleExpireTypeChange = () => {
  if (issueCouponForm.expireType === 'none') {
    issueCouponForm.expiresIn = 30
    issueCouponForm.expiresAt = null
  }
}

// 打开发放优惠券弹窗
const handleIssueCoupon = async (user) => {
  selectedUser.value = user
  issueCouponForm.couponId = null
  issueCouponForm.quantity = 1
  issueCouponForm.expireType = 'none'
  issueCouponForm.expiresIn = 30
  issueCouponForm.expiresAt = null
  
  // 先打开弹窗
  issueCouponDialogVisible.value = true
  
  // 加载可用优惠券列表
  try {
    const response = await getCoupons()
    
    // 如果响应直接是数组（某些情况下可能直接返回数组）
    if (Array.isArray(response)) {
      availableCoupons.value = response
        .map(coupon => ({
          ...coupon,
          type: coupon.type || '',
          discount_value: coupon.discount_value || 0,
          status: coupon.status !== undefined ? coupon.status : 1
        }))
        .filter(coupon => coupon.status === 1)
      return
    }
    
    // 标准响应格式：{ code, data, message }
    if (response && response.code === 200) {
      let coupons = response.data
      // 确保是数组
      if (!Array.isArray(coupons)) {
        coupons = []
      }
      // 只显示启用状态的优惠券
      availableCoupons.value = coupons
        .map(coupon => ({
          ...coupon,
          type: coupon.type || '',
          discount_value: coupon.discount_value || 0,
          status: coupon.status !== undefined ? coupon.status : 1
        }))
        .filter(coupon => coupon.status === 1)
    } else {
      availableCoupons.value = []
    }
  } catch (error) {
    console.error('加载优惠券列表失败:', error)
    ElMessage.error('加载优惠券列表失败')
    availableCoupons.value = []
  }
}

// 获取优惠券显示标签
const getCouponLabel = (coupon) => {
  let label = coupon.name
  if (coupon.type === 'delivery_fee') {
    label += ' (配送费券)'
  } else {
    label += ` (¥${(coupon.discount_value || 0).toFixed(2)})`
  }
  return label
}

// 提交发放优惠券
const handleIssueCouponSubmit = async () => {
  if (!issueCouponForm.couponId) {
    ElMessage.warning('请选择要发放的优惠券')
    return
  }
  
  if (!selectedUser.value) {
    ElMessage.error('用户信息错误')
    return
  }
  
  if (issueCouponForm.quantity < 1) {
    ElMessage.warning('发放数量必须大于0')
    return
  }
  
  issuingCoupon.value = true
  try {
    const issueData = {
      coupon_id: issueCouponForm.couponId,
      user_id: selectedUser.value.id,
      quantity: issueCouponForm.quantity
    }
    
    // 添加有效期参数
    if (issueCouponForm.expireType === 'days') {
      issueData.expires_in = issueCouponForm.expiresIn
    } else if (issueCouponForm.expireType === 'date' && issueCouponForm.expiresAt) {
      issueData.expires_at = issueCouponForm.expiresAt
    }
    
    await issueCouponToUser(issueData)
    ElMessage.success(`成功发放 ${issueCouponForm.quantity} 张优惠券`)
    issueCouponDialogVisible.value = false
    // 刷新用户列表
    loadUsers()
  } catch (error) {
    const errorMsg = error.response?.data?.message || error.message || '发放失败'
    ElMessage.error(errorMsg)
  } finally {
    issuingCoupon.value = false
  }
}

// 查看用户优惠券
const handleViewUserCoupons = async (user) => {
  currentViewUser.value = user
  userCoupons.value = []
  userCouponsLoading.value = true
  userCouponsDialogVisible.value = true
  
  try {
    const response = await getUserCoupons(user.id)
    if (response.code === 200 && Array.isArray(response.data)) {
      userCoupons.value = response.data
    } else {
      userCoupons.value = []
    }
  } catch (error) {
    console.error('获取用户优惠券失败:', error)
    ElMessage.error('获取用户优惠券失败')
    userCoupons.value = []
  } finally {
    userCouponsLoading.value = false
  }
}

// 格式化日期时间
const formatDateTime = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  if (isNaN(date.getTime())) return dateStr
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

onMounted(() => {
  loadUsers()
})
</script>

<style scoped>
.mini-users-page {
  padding: 20px 0;
}

.mini-users-card {
  border: none;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.04);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 12px;
}

.title .main {
  font-size: 20px;
  font-weight: 600;
  margin-right: 12px;
}

.title .sub {
  color: #909399;
  font-size: 14px;
}

.actions {
  display: flex;
  gap: 12px;
  min-width: 320px;
}

.mini-users-table {
  margin-top: 10px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

/* 用户详情对话框样式 */
:deep(.user-detail-dialog .el-dialog) {
  position: fixed;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  margin: 0;
}

:deep(.user-detail-dialog .el-dialog__body) {
  padding: 24px;
  max-height: 70vh;
  overflow-y: auto;
}

:deep(.user-detail-dialog .el-dialog__header) {
  padding: 20px 24px 16px;
  border-bottom: 1px solid #f0f0f0;
}

:deep(.user-detail-dialog .el-dialog__title) {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

:deep(.user-detail-dialog .el-dialog__footer) {
  padding: 16px 24px;
  border-top: 1px solid #f0f0f0;
}

.user-detail {
  min-height: 200px;
}

.detail-content {
  padding: 0;
}

/* 优化整体间距 */
:deep(.user-detail-dialog .el-dialog__body) {
  padding: 20px 24px;
}

.detail-section {
  margin-bottom: 24px;
  background: #fff;
  padding: 20px;
  border-radius: 8px;
  border: 1px solid #ebeef5;
}

.detail-section:last-child {
  margin-bottom: 0;
}

.avatar-section {
  margin-bottom: 28px;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 16px;
  padding-bottom: 10px;
  /* border-bottom: 2px solid #409eff; */
  position: relative;
}

.section-title::after {
  content: '';
  position: absolute;
  bottom: -2px;
  left: 0;
  width: 60px;
  height: 2px;
  background: #409eff;
}

.section-content {
  margin-top: 16px;
}

/* 头像样式 */
.avatar-image {
  width: 120px;
  height: 120px;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
  cursor: pointer;
  transition: all 0.3s;
}

.avatar-image:hover {
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  transform: scale(1.05);
}

.no-avatar {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 120px;
  height: 120px;
  border-radius: 8px;
  border: 1px dashed #dcdfe6;
  background: #f5f7fa;
  color: #909399;
  font-size: 14px;
  gap: 8px;
}

/* 基本信息包装器 */
.basic-info-wrapper {
  display: flex;
  gap: 16px;
  align-items: flex-start;
}

.avatar-container {
  flex-shrink: 0;
}

/* 头像样式（小尺寸） */
.avatar-image-small {
  width: 50px;
  height: 50px;
  border-radius: 6px;
  border: 1px solid #e4e7ed;
  cursor: pointer;
  transition: all 0.3s;
}

.avatar-image-small:hover {
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  transform: scale(1.05);
}

.no-avatar-small {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 50px;
  height: 50px;
  border-radius: 6px;
  border: 1px dashed #dcdfe6;
  background: #f5f7fa;
  color: #909399;
}

.info-list {
  flex: 1;
  min-width: 0;
}

/* 描述列表样式 */
:deep(.custom-descriptions .el-descriptions__table) {
  border-collapse: separate;
  border-spacing: 0;
}

:deep(.custom-descriptions .el-descriptions__label) {
  background: #f8f9fa;
  font-weight: 500;
  color: #606266;
  width: 120px;
  padding: 10px 14px;
  border-right: 1px solid #ebeef5;
  font-size: 13px;
}

:deep(.custom-descriptions .el-descriptions__content) {
  padding: 10px 14px;
  color: #303133;
  background: #fff;
  font-size: 13px;
}

:deep(.custom-descriptions .el-descriptions__cell) {
  border-bottom: 1px solid #ebeef5;
}

:deep(.custom-descriptions .el-descriptions__cell:last-child) {
  border-bottom: none;
}

:deep(.custom-descriptions .desc-label) {
  background: #f8f9fa !important;
}

.desc-value {
  color: #303133;
  font-size: 14px;
  word-break: break-all;
}

.unique-id {
  font-family: 'Courier New', monospace;
  font-size: 13px;
  color: #606266;
}

.phone-number {
  font-weight: 500;
  color: #303133;
}

.address-text {
  line-height: 1.6;
  color: #303133;
}

.coordinates {
  font-family: 'Courier New', monospace;
  color: #606266;
}

.time-text {
  color: #606266;
  font-size: 13px;
}

.user-type-tag,
.profile-tag {
  font-weight: 500;
  padding: 4px 12px;
  border-radius: 12px;
}

/* 头像上传样式 */
.avatar-uploader {
  display: inline-block;
}

.avatar-uploader .avatar {
  width: 100px;
  height: 100px;
  display: block;
  border-radius: 8px;
  border: 1px solid #dcdfe6;
}

.avatar-uploader-icon {
  font-size: 28px;
  color: #8c939d;
  width: 100px;
  height: 100px;
  line-height: 100px;
  text-align: center;
  border: 1px dashed #dcdfe6;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s;
}

.avatar-uploader-icon:hover {
  border-color: #409eff;
  color: #409eff;
}

.upload-tip {
  margin-top: 8px;
  color: #909399;
  font-size: 12px;
}

.user-code-text {
  font-weight: 600;
  color: #409eff;
  font-size: 15px;
}

/* 对话框底部按钮 */
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.dialog-footer .el-button {
  padding: 10px 20px;
  font-size: 14px;
  border-radius: 4px;
}

.dialog-footer .el-button--primary {
  background: #409eff;
  border-color: #409eff;
}

.dialog-footer .el-button--primary:hover {
  background: #66b1ff;
  border-color: #66b1ff;
}

/* 地址列表样式 */
.addresses-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.address-item {
  border: 1px solid #ebeef5;
  border-radius: 8px;
  padding: 16px;
  background: #fff;
  transition: all 0.3s;
}

.address-item:hover {
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  border-color: #c0c4cc;
}

.address-item.is-default {
  border-color: #67c23a;
  background: linear-gradient(135deg, #f0f9ff 0%, #f0fdf4 100%);
}

.address-header {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
  gap: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f0f0f0;
}

.address-count {
  font-size: 14px;
  color: #909399;
  font-weight: normal;
  margin-left: 8px;
}

:deep(.address-descriptions) {
  margin-top: 0;
}

:deep(.address-descriptions .el-descriptions__label) {
  width: 100px;
  font-size: 13px;
}

:deep(.address-descriptions .el-descriptions__content) {
  font-size: 13px;
}

.address-avatar-image {
  width: 80px;
  height: 80px;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
  cursor: pointer;
  transition: all 0.3s;
}

.address-avatar-image:hover {
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  transform: scale(1.05);
}

/* 编辑地址对话框样式优化 */
.address-edit-dialog :deep(.el-dialog__body) {
  padding: 24px;
}

.address-edit-form {
  max-height: 70vh;
  overflow-y: auto;
}

.address-edit-form :deep(.el-form-item) {
  margin-bottom: 20px;
}

.address-edit-form :deep(.el-form-item__label) {
  font-weight: 500;
  color: #606266;
}

/* 门头照片上传样式优化 */
.avatar-upload-wrapper {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 12px;
}

.avatar-uploader {
  border: 2px dashed #d9d9d9;
  border-radius: 8px;
  cursor: pointer;
  position: relative;
  overflow: hidden;
  transition: all 0.3s;
  background-color: #fafafa;
  width: 120px;
  height: 120px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.avatar-uploader:hover {
  border-color: #409eff;
  background-color: #f0f9ff;
}

.avatar-upload-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #8c939d;
  gap: 8px;
}

.avatar-upload-icon {
  font-size: 32px;
  color: #8c939d;
}

.avatar-upload-text {
  font-size: 12px;
  color: #8c939d;
}

.avatar-image {
  width: 120px;
  height: 120px;
  border-radius: 8px;
  object-fit: cover;
  cursor: pointer;
  transition: all 0.3s;
}

.avatar-image:hover {
  transform: scale(1.05);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.upload-tip {
  font-size: 12px;
  color: #909399;
  line-height: 1.5;
}

.sales-employee-info {
  display: flex;
  align-items: center;
}

.sales-employee-name {
  font-weight: 500;
  color: #303133;
}

/* 发票表单样式 */
.invoice-form {
  padding: 20px 0;
}

.invoice-form :deep(.el-form-item) {
  margin-bottom: 20px;
}

.invoice-form :deep(.el-form-item__label) {
  font-weight: 500;
  color: #606266;
}

/* 编辑地址对话框样式优化 */
.address-edit-dialog :deep(.el-dialog__body) {
  padding: 24px;
}

.address-edit-form {
  max-height: 70vh;
  overflow-y: auto;
  padding-right: 8px;
}

.address-edit-form :deep(.el-form-item) {
  margin-bottom: 20px;
}

.address-edit-form :deep(.el-form-item__label) {
  font-weight: 500;
  color: #606266;
}

/* 门头照片上传样式优化 */
.avatar-upload-wrapper {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 12px;
}

.avatar-uploader {
  border: 2px dashed #d9d9d9;
  border-radius: 8px;
  cursor: pointer;
  position: relative;
  overflow: hidden;
  transition: all 0.3s;
  background-color: #fafafa;
  width: 120px;
  height: 120px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.avatar-uploader:hover {
  border-color: #409eff;
  background-color: #f0f9ff;
}

.avatar-upload-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #8c939d;
  gap: 8px;
}

.avatar-upload-icon {
  font-size: 32px;
  color: #8c939d;
}

.avatar-upload-text {
  font-size: 12px;
  color: #8c939d;
}

.avatar-image {
  width: 120px;
  height: 120px;
  border-radius: 8px;
  object-fit: cover;
  cursor: pointer;
  transition: all 0.3s;
}

.avatar-image:hover {
  transform: scale(1.05);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.upload-tip {
  font-size: 12px;
  color: #909399;
  line-height: 1.5;
}

/* Tab 样式 */
.user-detail-tabs {
  margin-top: 0;
}

.user-detail-tabs :deep(.el-tabs__header) {
  margin-bottom: 20px;
}

.tab-content {
  padding: 20px 0;
  min-height: 200px;
}

/* 发票表单样式 */
.invoice-form {
  padding: 20px 0;
}

.invoice-form :deep(.el-form-item) {
  margin-bottom: 20px;
}

.invoice-form :deep(.el-form-item__label) {
  font-weight: 500;
  color: #606266;
}
</style>

