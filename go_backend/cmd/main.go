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

		// 分类相关接口
		apiGroup.GET("/products/category", api.GetProductsByCategory) // 根据分类ID获取该分类下的商品列表

		// 商品相关接口
		apiGroup.GET("/products/search/suggestions", api.SearchProductSuggestions) // 搜索商品建议
		apiGroup.GET("/products/search", api.SearchProducts)                       // 搜索商品
		apiGroup.GET("/products/:id", api.GetProductDetail)                        // 根据商品ID获取商品详情

		// 购物车相关接口
		apiGroup.POST("/cart", api.AddToCart)            // 添加商品到购物车
		apiGroup.GET("/cart", api.GetCartItems)          // 获取购物车中的商品列表
		apiGroup.DELETE("/cart/:id", api.DeleteCartItem) // 根据购物车项ID删除购物车中的商品
		apiGroup.DELETE("/cart/clear", api.ClearCart)    // 清空购物车

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
				protectedGroup.GET("/mini-app/users", api.GetMiniAppUsers) // 查看小程序用户
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
	}

	// 启动服务器
	port := config.Config.Server.Port
	fmt.Printf("服务器启动成功，访问地址: http://localhost:%d\n", port)
	router.Run(fmt.Sprintf(":%d", port))
}
