package main

import (
	"fmt"
	"log"

	"go_backend/internal/api"
	"go_backend/internal/config"
	"go_backend/internal/database"
	"go_backend/internal/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化配置
	config.InitConfig()

	// 初始化数据库
	if err := database.InitDB(); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer database.CloseDB()

	// 初始化MinIO客户端
	if err := utils.InitMinIO(); err != nil {
		log.Printf("MinIO初始化失败，但程序继续运行: %v", err)
	}

	// 创建路由引擎
	router := gin.Default()

	// 配置CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins: config.Config.CORS.AllowOrigins,
		AllowMethods: config.Config.CORS.AllowMethods,
		AllowHeaders: config.Config.CORS.AllowHeaders,
	}))

	// 静态文件服务
	router.Static("/static", "./static")

	// API路由
	apiGroup := router.Group("/api/mini")
	{
		// 首页相关接口
		apiGroup.GET("/carousels", api.GetCarousels)              // 获取首页轮播图
		apiGroup.GET("/categories", api.GetCategories)            // 获取商品分类列表
		apiGroup.GET("/products/special", api.GetSpecialProducts) // 获取特色商品列表
		apiGroup.GET("/products/hot", api.GetHotProducts)         // 获取热销商品列表
		apiGroup.POST("/auth/login", api.MiniAppLogin)            // 小程序登录
		apiGroup.PUT("/mini-app/users/type", api.UpdateMiniAppUserType)
		apiGroup.PUT("/mini-app/users/profile", api.UpdateMiniAppUserProfile)

		// 需要认证的小程序用户接口
		miniAppProtectedGroup := apiGroup.Group("/mini-app/users")
		miniAppProtectedGroup.Use(api.MiniAppAuthMiddleware())
		{
			miniAppProtectedGroup.GET("/info", api.GetMiniAppCurrentUser)            // 获取当前用户信息
			miniAppProtectedGroup.POST("/avatar", api.UploadMiniAppUserAvatar)       // 上传用户头像
			miniAppProtectedGroup.POST("/addresses/avatar", api.UploadAddressAvatar) // 上传地址头像（门头照片）
			miniAppProtectedGroup.PUT("/name", api.UpdateMiniAppUserName)            // 更新用户姓名
			miniAppProtectedGroup.PUT("/phone", api.UpdateMiniAppUserPhone)          // 更新用户电话

			// 地址相关接口
			miniAppProtectedGroup.GET("/addresses", api.GetMiniAppAddresses)                  // 获取用户的所有地址
			miniAppProtectedGroup.GET("/addresses/default", api.GetMiniAppDefaultAddress)     // 获取用户的默认地址
			miniAppProtectedGroup.DELETE("/addresses/:id", api.DeleteMiniAppAddress)          // 删除地址
			miniAppProtectedGroup.PUT("/addresses/:id/default", api.SetDefaultMiniAppAddress) // 设置默认地址
			miniAppProtectedGroup.POST("/addresses/geocode", api.GeocodeAddress)              // 地址解析（将地址文本转换为经纬度）
			miniAppProtectedGroup.POST("/addresses/reverse-geocode", api.ReverseGeocode)      // 逆地理编码（将经纬度转换为地址）

			// 发票抬头相关接口
			miniAppProtectedGroup.GET("/invoice", api.GetMiniAppInvoice)   // 获取用户的发票抬头
			miniAppProtectedGroup.POST("/invoice", api.SaveMiniAppInvoice) // 保存发票抬头

			// 采购单接口
			miniAppProtectedGroup.GET("/purchase-list", api.GetPurchaseListItems)
			miniAppProtectedGroup.POST("/purchase-list", api.AddPurchaseListItem)
			miniAppProtectedGroup.GET("/purchase-list/summary", api.GetPurchaseListSummary)
			miniAppProtectedGroup.PUT("/purchase-list/:id", api.UpdatePurchaseListItem)
			miniAppProtectedGroup.DELETE("/purchase-list/:id", api.DeletePurchaseListItem)
			miniAppProtectedGroup.DELETE("/purchase-list", api.ClearPurchaseList)

			// 常购商品接口
			miniAppProtectedGroup.GET("/frequent-products", api.GetFrequentProducts)

			// 订单接口
			miniAppProtectedGroup.POST("/orders", api.CreateOrderFromCart)        // 从当前采购单创建订单
			miniAppProtectedGroup.GET("/orders", api.GetUserOrders)               // 获取用户订单列表
			miniAppProtectedGroup.GET("/orders/:id", api.GetUserOrderDetail)      // 获取订单详情
			miniAppProtectedGroup.POST("/orders/:id/cancel", api.CancelUserOrder) // 取消订单

			// 配送员位置接口（小程序端查看配送员位置）
			miniAppProtectedGroup.GET("/delivery-employee-location/:code", api.GetEmployeeLocationByCode) // 根据员工码获取配送员位置

			// 优惠券接口
			miniAppProtectedGroup.GET("/coupons", api.GetUserCoupons)                // 获取用户的优惠券列表
			miniAppProtectedGroup.GET("/coupons/available", api.GetAvailableCoupons) // 获取可用优惠券

			// 新品需求接口
			miniAppProtectedGroup.POST("/product-requests", api.CreateProductRequest)  // 创建新品需求
			miniAppProtectedGroup.GET("/product-requests", api.GetUserProductRequests) // 获取用户的新品需求列表

			// 供应商合作申请接口
			miniAppProtectedGroup.GET("/supplier-applications", api.GetUserSupplierApplications) // 获取用户的申请列表

			// 收藏接口
			miniAppProtectedGroup.GET("/favorites", api.GetUserFavorites)                                // 获取用户收藏列表
			miniAppProtectedGroup.POST("/favorites", api.AddFavorite)                                    // 添加收藏
			miniAppProtectedGroup.DELETE("/favorites/:id", api.DeleteFavorite)                           // 删除收藏（通过收藏ID）
			miniAppProtectedGroup.DELETE("/favorites/product/:productId", api.DeleteFavoriteByProductID) // 删除收藏（通过商品ID）
			miniAppProtectedGroup.GET("/favorites/check", api.CheckFavorite)                             // 检查商品是否已收藏
		}

		// 供应商合作申请接口（不需要登录也可以提交）
		apiGroup.POST("/supplier-applications", api.CreateSupplierApplication) // 创建供应商合作申请

		// 价格反馈接口（需要登录）
		miniAppProtectedGroup.POST("/price-feedback", api.CreatePriceFeedback) // 创建价格反馈

		// 分类相关接口
		apiGroup.GET("/products/category", api.GetProductsByCategory) // 根据分类ID获取该分类下的商品列表

		// 商品相关接口
		apiGroup.GET("/products/search/suggestions", api.SearchProductSuggestions) // 搜索商品建议
		apiGroup.GET("/products/search", api.SearchProducts)                       // 搜索商品
		apiGroup.GET("/hot-search-keywords", api.GetHotSearchKeywords)             // 获取热门搜索关键词
		apiGroup.GET("/products/:id", api.GetProductDetail)                        // 根据商品ID获取商品详情

		// 管理员相关接口
		adminGroup := apiGroup.Group("/admin")
		{
			// 不需要认证的接口
			adminGroup.POST("/login", api.AdminLogin)   // 管理员登录
			adminGroup.POST("/logout", api.AdminLogout) // 管理员退出登录
			// WebSocket位置查看（管理后台）- 不需要认证中间件，在函数内部验证token
			adminGroup.GET("/employee-locations/ws", api.HandleAdminWebSocket) // WebSocket连接，用于实时接收位置更新

			// 需要认证的接口
			protectedGroup := adminGroup.Group("")
			protectedGroup.Use(api.AuthMiddleware())
			{
				protectedGroup.GET("/info", api.GetAdminInfo)       // 获取管理员信息
				protectedGroup.PUT("/password", api.ChangePassword) // 修改管理员密码

				// 系统设置接口
				protectedGroup.GET("/settings", api.GetSystemSettings)            // 获取所有系统设置
				protectedGroup.PUT("/settings", api.UpdateSystemSettings)         // 更新系统设置
				protectedGroup.GET("/settings/map", api.GetMapSettings)           // 获取地图设置
				protectedGroup.PUT("/settings/map", api.UpdateMapSettings)        // 更新地图设置
				protectedGroup.GET("/settings/websocket", api.GetWebSocketConfig) // 获取WebSocket配置

				// 分类管理接口
				protectedGroup.GET("/categories", api.GetAllCategoriesForAdmin)    // 获取所有商品分类（后台管理）
				protectedGroup.POST("/categories", api.CreateCategory)             // 创建新的商品分类
				protectedGroup.PUT("/categories/:id", api.UpdateCategory)          // 根据分类ID更新商品分类信息
				protectedGroup.DELETE("/categories/:id", api.DeleteCategory)       // 根据分类ID删除商品分类
				protectedGroup.POST("/categories/upload", api.UploadCategoryImage) // 上传分类图标

				// 轮播图管理接口
				protectedGroup.GET("/carousels", api.GetAllCarouselsForAdmin)     // 获取所有轮播图（管理后台用）
				protectedGroup.POST("/carousels", api.CreateCarousel)             // 创建轮播图
				protectedGroup.PUT("/carousels/:id", api.UpdateCarousel)          // 更新轮播图
				protectedGroup.DELETE("/carousels/:id", api.DeleteCarousel)       // 删除轮播图
				protectedGroup.POST("/carousels/upload", api.UploadCarouselImage) // 上传轮播图图片

				// 热销产品管理接口
				protectedGroup.GET("/hot-products", api.GetAllHotProductsForAdmin) // 获取所有热销产品（管理后台）
				protectedGroup.POST("/hot-products", api.CreateHotProduct)         // 创建热销产品关联
				protectedGroup.PUT("/hot-products/:id", api.UpdateHotProduct)      // 更新热销产品关联
				protectedGroup.DELETE("/hot-products/:id", api.DeleteHotProduct)   // 删除热销产品关联
				protectedGroup.PUT("/hot-products/sort", api.UpdateHotProductSort) // 批量更新热销产品排序

				// 热门搜索关键词管理接口
				protectedGroup.GET("/hot-search-keywords", api.GetAllHotSearchKeywordsForAdmin) // 获取所有热门搜索关键词
				protectedGroup.POST("/hot-search-keywords", api.CreateHotSearchKeyword)         // 创建热门搜索关键词
				protectedGroup.PUT("/hot-search-keywords/:id", api.UpdateHotSearchKeyword)      // 更新热门搜索关键词
				protectedGroup.DELETE("/hot-search-keywords/:id", api.DeleteHotSearchKeyword)   // 删除热门搜索关键词

				// 配送费设置
				protectedGroup.GET("/delivery-fee/settings", api.GetDeliveryFeeSettings)           // 获取配送费基础设置
				protectedGroup.PUT("/delivery-fee/settings", api.UpdateDeliveryFeeSettings)        // 更新配送费基础设置
				protectedGroup.GET("/delivery-fee/exclusions", api.ListDeliveryFeeExclusions)      // 获取配送费排除项
				protectedGroup.POST("/delivery-fee/exclusions", api.CreateDeliveryFeeExclusion)    // 新建配送费排除项
				protectedGroup.PUT("/delivery-fee/exclusions/:id", api.UpdateDeliveryFeeExclusion) // 更新配送费排除项
				protectedGroup.DELETE("/delivery-fee/exclusions/:id", api.DeleteDeliveryFeeExclusion)

				// 商品管理接口
				protectedGroup.GET("/products", api.GetAllProductsForAdmin)                 // 获取所有商品（管理后台）
				protectedGroup.POST("/products/upload", api.UploadProductImage)             // 上传商品图片（必须在 /:id 之前）
				protectedGroup.GET("/products/:id", api.GetProductDetail)                   // 获取商品详情（管理后台）
				protectedGroup.POST("/products", api.CreateProduct)                         // 创建商品
				protectedGroup.PUT("/products/:id", api.UpdateProduct)                      // 更新商品
				protectedGroup.PUT("/products/:id/special", api.UpdateProductSpecialStatus) // 更新商品精选状态
				protectedGroup.DELETE("/products/:id", api.DeleteProduct)                   // 删除商品

				// 供应商管理接口
				protectedGroup.GET("/suppliers", api.GetAllSuppliers)       // 获取所有供应商
				protectedGroup.GET("/suppliers/:id", api.GetSupplierByID)   // 获取供应商详情
				protectedGroup.POST("/suppliers", api.CreateSupplier)       // 创建供应商
				protectedGroup.PUT("/suppliers/:id", api.UpdateSupplier)    // 更新供应商
				protectedGroup.DELETE("/suppliers/:id", api.DeleteSupplier) // 删除供应商

				// 供应商付款统计接口
				protectedGroup.GET("/suppliers/payments/stats", api.GetSupplierPaymentsStats)      // 获取供应商付款统计列表
				protectedGroup.GET("/suppliers/:id/payments/detail", api.GetSupplierPaymentDetail) // 获取供应商详细付款清单
				protectedGroup.POST("/suppliers/payments", api.CreateSupplierPayment)              // 创建供应商付款记录
				protectedGroup.GET("/suppliers/payments", api.GetSupplierPayments)                 // 获取供应商付款记录列表
				protectedGroup.DELETE("/suppliers/payments/:id", api.CancelSupplierPayment)        // 撤销供应商付款

				// 小程序用户
				protectedGroup.GET("/mini-app/users", api.GetMiniAppUsers)                            // 查看小程序用户列表
				protectedGroup.GET("/mini-app/users/:id/coupons", api.GetAdminUserCoupons)            // 管理员获取用户优惠券列表（必须在 /:id 之前）
				protectedGroup.GET("/mini-app/users/:id", api.GetMiniAppUserDetail)                   // 查看小程序用户详情
				protectedGroup.POST("/mini-app/users/:id/invoice", api.SaveAdminInvoice)              // 保存发票抬头
				protectedGroup.PUT("/mini-app/users/:id", api.UpdateMiniAppUserByAdmin)               // 管理员更新小程序用户信息
				protectedGroup.POST("/mini-app/users/:id/avatar", api.UploadMiniAppUserAvatarByAdmin) // 管理员上传用户头像
				protectedGroup.GET("/mini-app/addresses/:id", api.GetAdminAddressByID)                // 管理员获取地址详情
				protectedGroup.PUT("/mini-app/addresses/:id", api.UpdateAdminAddress)                 // 管理员更新地址
				protectedGroup.DELETE("/mini-app/addresses/:id", api.DeleteAdminAddress)              // 管理员删除地址
				protectedGroup.POST("/mini-app/addresses/avatar", api.UploadAddressAvatarByAdmin)     // 管理员上传地址头像（门头照片）
				protectedGroup.POST("/mini-app/addresses/geocode", api.GeocodeAddress)                // 地址解析（将地址文本转换为经纬度）
				protectedGroup.POST("/mini-app/addresses/reverse-geocode", api.ReverseGeocode)        // 逆地理编码（将经纬度转换为地址）

				// 员工管理
				protectedGroup.GET("/employees", api.GetEmployees)            // 获取员工列表
				protectedGroup.GET("/employees/sales", api.GetSalesEmployees) // 获取销售员列表（用于下拉选择）
				protectedGroup.GET("/employees/:id", api.GetEmployee)         // 获取员工详情
				protectedGroup.POST("/employees", api.CreateEmployee)         // 创建员工
				protectedGroup.PUT("/employees/:id", api.UpdateEmployee)      // 更新员工
				protectedGroup.DELETE("/employees/:id", api.DeleteEmployee)   // 删除员工

				// 优惠券管理
				protectedGroup.GET("/coupons", api.GetAllCoupons)             // 获取所有优惠券
				protectedGroup.GET("/coupons/:id", api.GetCouponByID)         // 获取优惠券详情
				protectedGroup.POST("/coupons", api.CreateCoupon)             // 创建优惠券
				protectedGroup.PUT("/coupons/:id", api.UpdateCoupon)          // 更新优惠券
				protectedGroup.DELETE("/coupons/:id", api.DeleteCoupon)       // 删除优惠券
				protectedGroup.POST("/coupons/issue", api.IssueCouponToUser)  // 发放优惠券给用户
				protectedGroup.GET("/coupons/issues", api.GetCouponIssueLogs) // 优惠券发放记录列表
				protectedGroup.GET("/coupons/usages", api.GetCouponUsageLogs) // 优惠券使用记录列表

				// 订单管理
				protectedGroup.GET("/orders", api.GetAllOrdersForAdmin)                       // 获取所有订单（后台管理）
				protectedGroup.GET("/orders/:id", api.GetOrderByIDForAdmin)                   // 获取订单详情（后台管理）
				protectedGroup.PUT("/orders/:id/status", api.UpdateOrderStatus)               // 更新订单状态（后台管理）
				protectedGroup.GET("/orders/:id/delivery-fee", api.GetDeliveryFeeCalculation) // 获取配送费计算结果（管理员）

				// 配送记录管理
				protectedGroup.GET("/delivery-records", api.GetAllDeliveryRecordsForAdmin)                     // 获取所有配送记录（后台管理）
				protectedGroup.GET("/delivery-records/:id", api.GetDeliveryRecordByIDForAdmin)                 // 获取配送记录详情（后台管理）
				protectedGroup.GET("/delivery-records/order/:orderId", api.GetDeliveryRecordByOrderIDForAdmin) // 根据订单ID获取配送记录（后台管理）

				// 配送费结算管理
				protectedGroup.GET("/delivery-income/stats", api.GetDeliveryIncomeStatsForAdmin) // 获取配送员收入统计（管理员）
				protectedGroup.POST("/delivery-income/settle", api.BatchSettleDeliveryFees)      // 批量结算配送费

				// 销售分成管理（管理员）
				protectedGroup.GET("/sales-commission/stats", api.AdminGetSalesCommissionStats)                 // 获取销售员的分成统计（可查看所有销售员）
				protectedGroup.GET("/sales-commission/list", api.AdminGetSalesCommissions)                      // 获取销售员的分成记录列表
				protectedGroup.GET("/sales-commission/config", api.AdminGetSalesCommissionConfig)               // 获取销售员的分成配置
				protectedGroup.POST("/sales-commission/account", api.AdminAccountSalesCommissions)              // 批量计入销售分成
				protectedGroup.POST("/sales-commission/settle", api.AdminSettleSalesCommissions)                // 批量结算销售分成
				protectedGroup.POST("/sales-commission/cancel-account", api.AdminCancelAccountSalesCommissions) // 取消计入销售分成
				protectedGroup.POST("/sales-commission/reset-account", api.AdminResetAccountSalesCommissions)   // 重新计入销售分成（重置分成）
				protectedGroup.PUT("/sales-commission/config", api.AdminUpdateSalesCommissionConfig)            // 更新销售员的分成配置

				// 新品需求管理
				protectedGroup.GET("/product-requests", api.GetAllProductRequests)                 // 获取所有新品需求列表
				protectedGroup.PUT("/product-requests/:id/status", api.UpdateProductRequestStatus) // 更新新品需求状态

				// 供应商合作申请管理
				protectedGroup.GET("/supplier-applications", api.GetAllSupplierApplications)                 // 获取所有申请列表
				protectedGroup.PUT("/supplier-applications/:id/status", api.UpdateSupplierApplicationStatus) // 更新申请状态

				// 价格反馈管理接口
				protectedGroup.GET("/price-feedback", api.GetAllPriceFeedbacks)                 // 获取所有价格反馈列表
				protectedGroup.PUT("/price-feedback/:id/status", api.UpdatePriceFeedbackStatus) // 更新价格反馈状态

				// 仪表盘统计
				protectedGroup.GET("/dashboard/stats", api.GetDashboardStats) // 获取仪表盘统计数据

				// 员工位置管理
				protectedGroup.GET("/employee-locations", api.GetEmployeeLocations)    // 获取所有员工位置
				protectedGroup.GET("/employee-locations/:id", api.GetEmployeeLocation) // 获取指定员工位置

				// 收款审核管理
				protectedGroup.GET("/payment-verification", api.GetPaymentVerificationRequests)           // 获取收款审核列表
				protectedGroup.POST("/payment-verification/review", api.ReviewPaymentVerificationRequest) // 审核收款申请

				// 富文本内容管理
				protectedGroup.GET("/rich-contents", api.GetRichContentList)             // 获取富文本内容列表
				protectedGroup.GET("/rich-contents/:id", api.GetRichContent)             // 获取富文本内容详情
				protectedGroup.POST("/rich-contents", api.CreateRichContent)             // 创建富文本内容
				protectedGroup.PUT("/rich-contents/:id", api.UpdateRichContent)          // 更新富文本内容
				protectedGroup.PUT("/rich-contents/:id/publish", api.PublishRichContent) // 发布富文本内容
				protectedGroup.PUT("/rich-contents/:id/archive", api.ArchiveRichContent) // 归档富文本内容
				protectedGroup.DELETE("/rich-contents/:id", api.DeleteRichContent)       // 删除富文本内容
			}
		}

		// 富文本内容相关接口（小程序端）
		apiGroup.GET("/rich-contents", api.GetPublishedRichContentList)       // 获取已发布的富文本内容列表
		apiGroup.GET("/rich-contents/:id", api.GetPublishedRichContentDetail) // 获取已发布的富文本内容详情

		// 移动端接口（不需要token，通过参数验证）
		mobileGroup := apiGroup.Group("/mobile")
		{
			mobileGroup.GET("/pending-goods", api.GetMobilePendingGoods) // 移动端获取待备货货物列表
		}

		// 供应商相关接口
		supplierGroup := apiGroup.Group("/supplier")
		{
			// 不需要认证的接口
			supplierGroup.POST("/login", api.SupplierLogin) // 供应商登录

			// 需要认证的接口
			supplierProtectedGroup := supplierGroup.Group("")
			supplierProtectedGroup.Use(api.SupplierAuthMiddleware())
			{
				supplierProtectedGroup.GET("/dashboard", api.GetSupplierDashboard)        // 供应商数据总览
				supplierProtectedGroup.GET("/products", api.GetSupplierProducts)          // 供应商查看自己的商品列表
				supplierProtectedGroup.GET("/products/:id", api.GetSupplierProductDetail) // 供应商查看自己的商品详情
				supplierProtectedGroup.GET("/orders", api.GetSupplierOrders)              // 供应商查看包含自己商品的订单列表
				supplierProtectedGroup.GET("/orders/:id", api.GetSupplierOrderDetail)     // 供应商查看订单详情

				// 货物管理接口
				supplierProtectedGroup.GET("/goods/today/stats", api.GetTodayGoodsStats)     // 获取今日货物统计
				supplierProtectedGroup.GET("/goods/today/pending", api.GetTodayPendingGoods) // 获取今日待备货货物列表
				supplierProtectedGroup.GET("/goods/today/picked", api.GetTodayPickedGoods)   // 获取今日已取货货物列表

				// 历史记录接口
				supplierProtectedGroup.GET("/history", api.GetHistoryByDate)       // 获取历史记录列表（按天）
				supplierProtectedGroup.GET("/history/:date", api.GetHistoryDetail) // 获取某天的历史详情

				// 供应商对账功能
				supplierProtectedGroup.GET("/payments/paid", api.GetSupplierPaidItems)       // 获取已付款清单
				supplierProtectedGroup.GET("/payments/pending", api.GetSupplierPendingItems) // 获取待付款清单
				supplierProtectedGroup.GET("/payments/stats", api.GetSupplierPaymentStats)   // 获取对账统计
			}
		}

		// 员工相关接口
		employeeGroup := apiGroup.Group("/employee")
		{
			// 不需要认证的接口
			employeeGroup.POST("/login", api.EmployeeLogin)                // 员工登录
			employeeGroup.GET("/websocket-config", api.GetWebSocketConfig) // 获取WebSocket配置（不需要认证）
			// WebSocket位置上报（配送员端）- 不需要认证中间件，在函数内部验证token
			employeeGroup.GET("/location/ws", api.HandleEmployeeWebSocket) // WebSocket连接，用于实时上报位置

			// 需要认证的接口
			employeeProtectedGroup := employeeGroup.Group("")
			employeeProtectedGroup.Use(api.EmployeeAuthMiddleware())
			{
				employeeProtectedGroup.GET("/info", api.GetEmployeeInfo)           // 获取当前员工信息
				employeeProtectedGroup.GET("/dashboard", api.GetEmployeeDashboard) // 员工首页概览

				// 配送员相关接口
				employeeProtectedGroup.GET("/delivery/orders", api.GetDeliveryOrders)                                    // 获取待配送订单列表
				employeeProtectedGroup.GET("/delivery/orders/:id", api.GetDeliveryOrderDetail)                           // 获取订单详情
				employeeProtectedGroup.GET("/delivery/orders/:id/delivery-fee", api.GetDeliveryFeeCalculationForRider)   // 获取配送费计算结果（配送员）
				employeeProtectedGroup.PUT("/delivery/orders/:id/accept", api.AcceptDeliveryOrder)                       // 接单
				employeeProtectedGroup.PUT("/delivery/orders/:id/start", api.StartDeliveryOrder)                         // 开始配送
				employeeProtectedGroup.POST("/delivery/orders/:id/complete", api.CompleteDeliveryOrder)                  // 完成配送（支持上传图片）
				employeeProtectedGroup.POST("/delivery/orders/:id/report", api.ReportOrderIssue)                         // 问题上报
				employeeProtectedGroup.GET("/delivery/my-orders", api.GetDeliveryOrders)                                 // 获取我的配送订单（通过status参数筛选）
				employeeProtectedGroup.GET("/delivery/pickup/suppliers", api.GetPickupSuppliers)                         // 获取待取货供应商列表
				employeeProtectedGroup.GET("/delivery/pickup/suppliers/:supplierId/items", api.GetPickupItemsBySupplier) // 获取供应商的待取货商品
				employeeProtectedGroup.POST("/delivery/pickup/mark-picked", api.MarkItemsAsPicked)                       // 标记商品已取货
				employeeProtectedGroup.POST("/delivery/route/calculate", api.CalculateRoute)                             // 计算路线规划
				employeeProtectedGroup.GET("/delivery/route/orders", api.GetRouteOrders)                                 // 获取排序后的订单列表
				employeeProtectedGroup.GET("/delivery/income/stats", api.GetDeliveryIncomeStats)                         // 获取配送员收入统计
				employeeProtectedGroup.GET("/delivery/income/details", api.GetDeliveryIncomeDetails)                     // 获取配送员收入明细

				// 销售员相关接口
				employeeProtectedGroup.GET("/sales/customers", api.GetSalesCustomers)                                            // 获取我的客户列表
				employeeProtectedGroup.GET("/sales/customer-by-code", api.GetSalesCustomerByCode)                                // 通过编号查客户
				employeeProtectedGroup.GET("/sales/customers/:id", api.GetSalesCustomerDetail)                                   // 获取客户详情
				employeeProtectedGroup.GET("/sales/customers/:id/orders", api.GetSalesCustomerOrders)                            // 获取客户的订单列表
				employeeProtectedGroup.GET("/sales/customers/:id/frequent-products", api.GetSalesCustomerFrequentProducts)       // 获取客户的常购商品列表
				employeeProtectedGroup.GET("/sales/customers/:id/coupons", api.GetAdminUserCoupons)                              // 获取客户的优惠券列表（销售员查看）
				employeeProtectedGroup.GET("/sales/customers/:id/purchase-list", api.GetSalesCustomerPurchaseList)               // 获取客户的采购单
				employeeProtectedGroup.POST("/sales/customers/:id/purchase-list", api.AddSalesCustomerPurchaseItem)              // 新增客户采购单条目
				employeeProtectedGroup.PUT("/sales/customers/:id/purchase-list/:itemId", api.UpdateSalesCustomerPurchaseItem)    // 更新客户采购单条目
				employeeProtectedGroup.DELETE("/sales/customers/:id/purchase-list/:itemId", api.DeleteSalesCustomerPurchaseItem) // 删除客户采购单条目
				employeeProtectedGroup.GET("/sales/customers/:id/rider-delivery-fee-preview", api.PreviewRiderDeliveryFee)       // 预览配送员配送费
				employeeProtectedGroup.PUT("/sales/customers/:id/profile", api.UpdateSalesCustomerProfile)                       // 更新客户基础资料
				employeeProtectedGroup.POST("/sales/customers/:id/addresses", api.CreateSalesCustomerAddress)                    // 为客户新增地址
				employeeProtectedGroup.PUT("/sales/addresses/:id", api.UpdateSalesCustomerAddress)                               // 更新客户地址
				employeeProtectedGroup.POST("/upload/address-avatar", api.UploadAddressAvatarByEmployee)                         // 上传门头照
				employeeProtectedGroup.POST("/addresses/reverse-geocode", api.ReverseGeocode)                                    // 逆地理编码（将经纬度转换为地址，用于选点回填）
				employeeProtectedGroup.POST("/sales/orders", api.CreateOrderForCustomer)                                         // 为客户创建订单
				employeeProtectedGroup.GET("/sales/products", api.GetSalesProducts)                                              // 获取商品列表
				employeeProtectedGroup.GET("/sales/pending-orders", api.GetMyPendingOrders)                                      // 获取待配送订单列表
				employeeProtectedGroup.GET("/sales/coupons", api.GetAllCoupons)                                                  // 销售员查看优惠券列表
				employeeProtectedGroup.POST("/sales/coupons/issue", api.IssueCouponToUser)                                       // 销售员为客户发放优惠券
				employeeProtectedGroup.GET("/sales/orders", api.GetSalesOrders)                                                  // 销售员查看名下订单列表
				employeeProtectedGroup.GET("/sales/orders/:id", api.GetSalesOrderDetail)                                         // 销售员查看订单详情
				employeeProtectedGroup.POST("/sales/orders/:id/lock", api.LockOrderForEdit)                                      // 锁定订单用于修改
				employeeProtectedGroup.POST("/sales/orders/:id/unlock", api.UnlockOrderAfterEdit)                                // 解锁订单
				employeeProtectedGroup.POST("/sales/orders/:id/sync-to-purchase-list", api.SyncOrderItemsToPurchaseList)         // 将订单商品同步到采购单
				employeeProtectedGroup.PUT("/sales/orders/:id", api.UpdateOrderForCustomer)                                      // 修改订单
				employeeProtectedGroup.POST("/sales/orders/:id/cancel", api.CancelSalesOrder)                                    // 取消订单
				employeeProtectedGroup.GET("/delivery-employee-location/:code", api.GetEmployeeLocationByCode)                   // 获取配送员位置（员工端）

				// 销售分成相关接口
				employeeProtectedGroup.POST("/sales/commission/preview", api.PreviewSalesCommission)                                 // 预览销售分成（开单时）
				employeeProtectedGroup.GET("/sales/commission/list", api.GetSalesCommissions)                                        // 获取销售员的分成记录列表
				employeeProtectedGroup.GET("/sales/commission/stats", api.GetSalesCommissionMonthlyStats)                            // 获取销售员的分成月统计
				employeeProtectedGroup.GET("/sales/commission/overview", api.GetSalesCommissionOverview)                             // 获取销售员的分成总览统计
				employeeProtectedGroup.GET("/sales/commission/unpaid-orders", api.GetUnpaidOrdersWithCommissionPreview)              // 获取未收款订单及其分润预览
				employeeProtectedGroup.POST("/sales/payment-verification", api.SubmitPaymentVerificationRequest)                     // 提交收款申请
				employeeProtectedGroup.GET("/sales/payment-verification/order/:orderId", api.GetPaymentVerificationRequestByOrderID) // 获取订单的收款申请状态
				employeeProtectedGroup.GET("/sales/commission/config", api.GetSalesCommissionConfig)                                 // 获取销售员的分成配置
			}
		}
	}

	// 启动服务器
	port := config.Config.Server.Port
	fmt.Printf("服务器启动成功，访问地址: http://localhost:%d\n", port)
	router.Run(fmt.Sprintf(":%d", port))
}
