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

		// 分类相关接口
		apiGroup.GET("/products/category", api.GetProductsByCategory) // 根据分类ID获取该分类下的商品列表

		// 商品相关接口
		apiGroup.GET("/products/search/suggestions", api.SearchProductSuggestions) // 搜索商品建议
		apiGroup.GET("/products/search", api.SearchProducts)                       // 搜索商品
		apiGroup.GET("/products/:id", api.GetProductDetail)                         // 根据商品ID获取商品详情

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
				protectedGroup.GET("/info", api.GetAdminInfo) // 获取管理员信息

				// 分类管理接口
				protectedGroup.GET("/categories", api.GetAllCategoriesForAdmin) // 获取所有商品分类（后台管理）
				protectedGroup.POST("/categories", api.CreateCategory)       // 创建新的商品分类
				protectedGroup.PUT("/categories/:id", api.UpdateCategory)    // 根据分类ID更新商品分类信息
				protectedGroup.DELETE("/categories/:id", api.DeleteCategory) // 根据分类ID删除商品分类
				protectedGroup.POST("/categories/upload", api.UploadCategoryImage) // 上传分类图标

				// 轮播图管理接口
				protectedGroup.GET("/carousels", api.GetAllCarouselsForAdmin) // 获取所有轮播图（管理后台用）
				protectedGroup.POST("/carousels", api.CreateCarousel)         // 创建轮播图
				protectedGroup.PUT("/carousels/:id", api.UpdateCarousel)      // 更新轮播图
				protectedGroup.DELETE("/carousels/:id", api.DeleteCarousel)   // 删除轮播图
				protectedGroup.POST("/carousels/upload", api.UploadCarouselImage) // 上传轮播图图片

				// 商品管理接口
				protectedGroup.GET("/products", api.GetAllProductsForAdmin) // 获取所有商品（管理后台）
				protectedGroup.POST("/products", api.CreateProduct)       // 创建商品
				protectedGroup.PUT("/products/:id", api.UpdateProduct)    // 更新商品
				protectedGroup.DELETE("/products/:id", api.DeleteProduct) // 删除商品
				protectedGroup.POST("/products/upload", api.UploadProductImage) // 上传商品图片
			}
		}
	}
	// 添加测试路由（不需要身份验证）
	testGroup := apiGroup.Group("/test")
	{
		testGroup.GET("/categories", api.GetAllCategoriesForAdmin) // 测试获取所有商品分类
	}

	// 小程序相关接口
	miniGroup := apiGroup.Group("/mini")
	{
		// 首页相关接口
		miniGroup.GET("/carousels", api.GetCarousels)              // 获取小程序轮播图
		miniGroup.GET("/categories", api.GetCategories)            // 获取小程序分类列表
		miniGroup.GET("/products/special", api.GetSpecialProducts) // 获取小程序特价商品列表
		miniGroup.GET("/products/:id", api.GetProductDetail)       // 获取小程序商品详情
		miniGroup.GET("/categories/:id/products", api.GetProductsByCategory) // 获取小程序分类商品
	}

	// 启动服务器
	port := config.Config.Server.Port
	fmt.Printf("服务器启动成功，访问地址: http://localhost:%d\n", port)
	router.Run(fmt.Sprintf(":%d", port))
}
