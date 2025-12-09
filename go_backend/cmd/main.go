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

			// 采购单接口
			miniAppProtectedGroup.GET("/purchase-list", api.GetPurchaseListItems)
			miniAppProtectedGroup.POST("/purchase-list", api.AddPurchaseListItem)
			miniAppProtectedGroup.GET("/purchase-list/summary", api.GetPurchaseListSummary)
			miniAppProtectedGroup.PUT("/purchase-list/:id", api.UpdatePurchaseListItem)
			miniAppProtectedGroup.DELETE("/purchase-list/:id", api.DeletePurchaseListItem)
			miniAppProtectedGroup.DELETE("/purchase-list", api.ClearPurchaseList)

			// 订单接口
			miniAppProtectedGroup.POST("/orders", api.CreateOrderFromCart)   // 从当前采购单创建订单
			miniAppProtectedGroup.GET("/orders", api.GetUserOrders)          // 获取用户订单列表
			miniAppProtectedGroup.GET("/orders/:id", api.GetUserOrderDetail) // 获取订单详情

			// 优惠券接口
			miniAppProtectedGroup.GET("/coupons", api.GetUserCoupons)                // 获取用户的优惠券列表
			miniAppProtectedGroup.GET("/coupons/available", api.GetAvailableCoupons) // 获取可用优惠券
		}

		// 分类相关接口
		apiGroup.GET("/products/category", api.GetProductsByCategory) // 根据分类ID获取该分类下的商品列表

		// 商品相关接口
		apiGroup.GET("/products/search/suggestions", api.SearchProductSuggestions) // 搜索商品建议
		apiGroup.GET("/products/search", api.SearchProducts)                       // 搜索商品
		apiGroup.GET("/products/:id", api.GetProductDetail)                        // 根据商品ID获取商品详情

		// 管理员相关接口
		adminGroup := apiGroup.Group("/admin")
		{
			// 不需要认证的接口
			adminGroup.POST("/login", api.AdminLogin)   // 管理员登录
			adminGroup.POST("/logout", api.AdminLogout) // 管理员退出登录

			// 需要认证的接口
			protectedGroup := adminGroup.Group("")
			protectedGroup.Use(api.AuthMiddleware())
			{
				protectedGroup.GET("/info", api.GetAdminInfo)       // 获取管理员信息
				protectedGroup.PUT("/password", api.ChangePassword) // 修改管理员密码

				// 系统设置接口
				protectedGroup.GET("/settings", api.GetSystemSettings)     // 获取所有系统设置
				protectedGroup.PUT("/settings", api.UpdateSystemSettings)  // 更新系统设置
				protectedGroup.GET("/settings/map", api.GetMapSettings)    // 获取地图设置
				protectedGroup.PUT("/settings/map", api.UpdateMapSettings) // 更新地图设置

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

				// 配送费设置
				protectedGroup.GET("/delivery-fee/settings", api.GetDeliveryFeeSettings)           // 获取配送费基础设置
				protectedGroup.PUT("/delivery-fee/settings", api.UpdateDeliveryFeeSettings)        // 更新配送费基础设置
				protectedGroup.GET("/delivery-fee/exclusions", api.ListDeliveryFeeExclusions)      // 获取配送费排除项
				protectedGroup.POST("/delivery-fee/exclusions", api.CreateDeliveryFeeExclusion)    // 新建配送费排除项
				protectedGroup.PUT("/delivery-fee/exclusions/:id", api.UpdateDeliveryFeeExclusion) // 更新配送费排除项
				protectedGroup.DELETE("/delivery-fee/exclusions/:id", api.DeleteDeliveryFeeExclusion)

				// 商品管理接口
				protectedGroup.GET("/products", api.GetAllProductsForAdmin)                 // 获取所有商品（管理后台）
				protectedGroup.POST("/products", api.CreateProduct)                         // 创建商品
				protectedGroup.PUT("/products/:id", api.UpdateProduct)                      // 更新商品
				protectedGroup.PUT("/products/:id/special", api.UpdateProductSpecialStatus) // 更新商品精选状态
				protectedGroup.DELETE("/products/:id", api.DeleteProduct)                   // 删除商品
				protectedGroup.POST("/products/upload", api.UploadProductImage)             // 上传商品图片

				// 供应商管理接口
				protectedGroup.GET("/suppliers", api.GetAllSuppliers)       // 获取所有供应商
				protectedGroup.GET("/suppliers/:id", api.GetSupplierByID)   // 获取供应商详情
				protectedGroup.POST("/suppliers", api.CreateSupplier)       // 创建供应商
				protectedGroup.PUT("/suppliers/:id", api.UpdateSupplier)    // 更新供应商
				protectedGroup.DELETE("/suppliers/:id", api.DeleteSupplier) // 删除供应商

				// 小程序用户
				protectedGroup.GET("/mini-app/users", api.GetMiniAppUsers)                            // 查看小程序用户列表
				protectedGroup.GET("/mini-app/users/:id/coupons", api.GetAdminUserCoupons)            // 管理员获取用户优惠券列表（必须在 /:id 之前）
				protectedGroup.GET("/mini-app/users/:id", api.GetMiniAppUserDetail)                   // 查看小程序用户详情
				protectedGroup.PUT("/mini-app/users/:id", api.UpdateMiniAppUserByAdmin)               // 管理员更新小程序用户信息
				protectedGroup.POST("/mini-app/users/:id/avatar", api.UploadMiniAppUserAvatarByAdmin) // 管理员上传用户头像
				protectedGroup.GET("/mini-app/addresses/:id", api.GetAdminAddressByID)                // 管理员获取地址详情
				protectedGroup.PUT("/mini-app/addresses/:id", api.UpdateAdminAddress)                 // 管理员更新地址
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
			}
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
				supplierProtectedGroup.GET("/products", api.GetSupplierProducts) // 供应商查看自己的商品
			}
		}

		// 员工相关接口
		employeeGroup := apiGroup.Group("/employee")
		{
			// 不需要认证的接口
			employeeGroup.POST("/login", api.EmployeeLogin) // 员工登录

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
				employeeProtectedGroup.POST("/delivery/route/plan", api.PlanDeliveryRoute)                               // 规划配送路线

				// 销售员相关接口
				employeeProtectedGroup.GET("/sales/customers", api.GetSalesCustomers)                                            // 获取我的客户列表
				employeeProtectedGroup.GET("/sales/customer-by-code", api.GetSalesCustomerByCode)                                // 通过编号查客户
				employeeProtectedGroup.GET("/sales/customers/:id", api.GetSalesCustomerDetail)                                   // 获取客户详情
				employeeProtectedGroup.GET("/sales/customers/:id/orders", api.GetSalesCustomerOrders)                            // 获取客户的订单列表
				employeeProtectedGroup.GET("/sales/customers/:id/coupons", api.GetAdminUserCoupons)                              // 获取客户的优惠券列表（销售员查看）
				employeeProtectedGroup.GET("/sales/customers/:id/purchase-list", api.GetSalesCustomerPurchaseList)               // 获取客户的采购单
				employeeProtectedGroup.POST("/sales/customers/:id/purchase-list", api.AddSalesCustomerPurchaseItem)              // 新增客户采购单条目
				employeeProtectedGroup.PUT("/sales/customers/:id/purchase-list/:itemId", api.UpdateSalesCustomerPurchaseItem)    // 更新客户采购单条目
				employeeProtectedGroup.DELETE("/sales/customers/:id/purchase-list/:itemId", api.DeleteSalesCustomerPurchaseItem) // 删除客户采购单条目
				employeeProtectedGroup.PUT("/sales/customers/:id/profile", api.UpdateSalesCustomerProfile)                       // 更新客户基础资料
				employeeProtectedGroup.POST("/sales/customers/:id/addresses", api.CreateSalesCustomerAddress)                    // 为客户新增地址
				employeeProtectedGroup.PUT("/sales/addresses/:id", api.UpdateSalesCustomerAddress)                               // 更新客户地址
				employeeProtectedGroup.POST("/upload/address-avatar", api.UploadAddressAvatarByEmployee)                         // 上传门头照
				employeeProtectedGroup.POST("/sales/orders", api.CreateOrderForCustomer)                                         // 为客户创建订单
				employeeProtectedGroup.GET("/sales/products", api.GetSalesProducts)                                              // 获取商品列表
				employeeProtectedGroup.GET("/sales/pending-orders", api.GetMyPendingOrders)                                      // 获取待配送订单列表
				employeeProtectedGroup.GET("/sales/coupons", api.GetAllCoupons)                                                  // 销售员查看优惠券列表
				employeeProtectedGroup.POST("/sales/coupons/issue", api.IssueCouponToUser)                                       // 销售员为客户发放优惠券
				employeeProtectedGroup.GET("/sales/orders", api.GetSalesOrders)                                                  // 销售员查看名下订单列表
				employeeProtectedGroup.GET("/sales/orders/:id", api.GetSalesOrderDetail)                                         // 销售员查看订单详情
			}
		}
	}

	// 启动服务器
	port := config.Config.Server.Port
	fmt.Printf("服务器启动成功，访问地址: http://localhost:%d\n", port)
	router.Run(fmt.Sprintf(":%d", port))
}
