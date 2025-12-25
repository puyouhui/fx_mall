package database

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"go_backend/internal/config"
	"go_backend/internal/utils"

	_ "github.com/go-sql-driver/mysql"
)

var (
	// DB 全局数据库连接池
	DB *sql.DB
	// 单例模式的锁
	dbOnce sync.Once
)

// InitDB 初始化数据库连接
func InitDB() error {
	var err error
	dbOnce.Do(func() {
		cfg := config.Config.Database

		// 先使用不指定数据库的DSN连接，确保数据库存在
		serverDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=%s&parseTime=True&loc=Local",
			cfg.Username,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.Charset,
		)

		serverDB, serverErr := sql.Open("mysql", serverDSN)
		if serverErr != nil {
			err = fmt.Errorf("打开数据库服务器连接失败: %w", serverErr)
			log.Println(err)
			return
		}
		defer serverDB.Close()

		if serverErr = serverDB.Ping(); serverErr != nil {
			err = fmt.Errorf("数据库服务器连接测试失败: %w", serverErr)
			log.Println(err)
			return
		}

		createDBSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARSET %s COLLATE %s_unicode_ci;", cfg.DBName, cfg.Charset, cfg.Charset)
		if _, serverErr = serverDB.Exec(createDBSQL); serverErr != nil {
			err = fmt.Errorf("创建数据库失败: %w", serverErr)
			log.Println(err)
			return
		}

		// 构建包含数据库名称的DSN
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			cfg.Username,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.DBName,
			cfg.Charset,
		)

		// 打开数据库连接池
		DB, err = sql.Open("mysql", dsn)
		if err != nil {
			log.Printf("打开数据库连接失败: %v", err)
			return
		}

		// 设置连接池参数
		DB.SetMaxOpenConns(100)                 // 最大打开连接数
		DB.SetMaxIdleConns(20)                  // 最大空闲连接数
		DB.SetConnMaxLifetime(30 * time.Minute) // 连接最大生存时间，避免长时间连接导致的问题

		// 测试连接
		if err = DB.Ping(); err != nil {
			log.Printf("数据库连接测试失败: %v", err)
			return
		}

		// 创建admins表
		createAdminsTableSQL := `
		CREATE TABLE IF NOT EXISTS admins (
		    id INT PRIMARY KEY AUTO_INCREMENT,
		    username VARCHAR(50) NOT NULL COMMENT '用户名',
		    password VARCHAR(255) NOT NULL COMMENT '密码',
		    created_at DATETIME NOT NULL,
		    updated_at DATETIME NOT NULL,
		    UNIQUE KEY uk_username (username)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='管理员表';
		`

		_, err = DB.Exec(createAdminsTableSQL)
		if err != nil {
			log.Printf("创建admins表失败: %v", err)
			return
		}

		// 检查管理员表是否有数据，如果没有则插入默认管理员
		var adminCount int
		err = DB.QueryRow("SELECT COUNT(*) FROM admins").Scan(&adminCount)
		if err != nil {
			log.Printf("查询管理员数据数量失败: %v", err)
			return
		}

		if adminCount == 0 {
			// 使用bcrypt加密默认密码
			hashedPassword, hashErr := utils.HashPassword("admin123")
			if hashErr != nil {
				log.Printf("加密默认管理员密码失败: %v", hashErr)
				err = hashErr
				return
			}

			// 插入默认管理员 (密码为admin123，已使用bcrypt加密存储)
			insertAdminSQL := `
			INSERT INTO admins (username, password, created_at, updated_at)
			VALUES (?, ?, NOW(), NOW());
			`

			_, err = DB.Exec(insertAdminSQL, "admin", hashedPassword)
			if err != nil {
				log.Printf("插入默认管理员失败: %v", err)
				return
			}

			log.Println("管理员表初始化成功，已创建默认管理员（密码已加密）")
		} else {
			log.Println("管理员表已存在，跳过初始化")
		}

		// 创建categories表
		createCategoriesTableSQL := `
		CREATE TABLE IF NOT EXISTS categories (
		    id INT PRIMARY KEY AUTO_INCREMENT,
		    name VARCHAR(50) NOT NULL COMMENT '分类名称',
		    parent_id INT DEFAULT 0 COMMENT '父分类ID，0表示一级分类',
		    sort INT DEFAULT 0 COMMENT '排序',
		    status TINYINT DEFAULT 1 COMMENT '状态：1-启用，0-禁用',
		    created_at DATETIME NOT NULL,
		    updated_at DATETIME NOT NULL,
		    icon VARCHAR(255) NULL COMMENT '分类图标URL',
		    UNIQUE KEY uk_name_parent_id (name, parent_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品分类表';
		`

		_, err = DB.Exec(createCategoriesTableSQL)
		if err != nil {
			log.Printf("创建categories表失败: %v", err)
			return
		}

		// 检查分类表是否有数据，如果没有则插入测试数据
		var categoryCount int
		err = DB.QueryRow("SELECT COUNT(*) FROM categories").Scan(&categoryCount)
		if err != nil {
			log.Printf("查询分类表数据数量失败: %v", err)
			return
		}

		if categoryCount == 0 {
			// 插入测试数据
			insertDataSQL := `
			INSERT INTO categories (name, parent_id, sort, status, created_at, updated_at, icon)
			VALUES 
			    ('电子产品', 0, 1, 1, NOW(), NOW(), ''),
			    ('家居用品', 0, 2, 1, NOW(), NOW(), ''),
			    ('服装鞋帽', 0, 3, 1, NOW(), NOW(), ''),
			    ('手机', 1, 1, 1, NOW(), NOW(), ''),
			    ('电脑', 1, 2, 1, NOW(), NOW(), ''),
			    ('平板电脑', 1, 3, 1, NOW(), NOW(), ''),
			    ('厨房用品', 2, 1, 1, NOW(), NOW(), ''),
			    ('床上用品', 2, 2, 1, NOW(), NOW(), ''),
			    ('清洁用品', 2, 3, 1, NOW(), NOW(), ''),
			    ('男装', 3, 1, 1, NOW(), NOW(), ''),
			    ('女装', 3, 2, 1, NOW(), NOW(), ''),
			    ('童装', 3, 3, 1, NOW(), NOW(), '');
			`

			_, err = DB.Exec(insertDataSQL)
			if err != nil {
				log.Printf("插入分类测试数据失败: %v", err)
				return
			}

			log.Println("分类表初始化成功，已插入完整的多级分类测试数据")
		} else {
			log.Println("分类表已有数据，跳过初始化")
		}

		log.Println("数据库连接成功")

		// 创建suppliers表
		createSuppliersTableSQL := `
		CREATE TABLE IF NOT EXISTS suppliers (
		    id INT PRIMARY KEY AUTO_INCREMENT,
		    name VARCHAR(100) NOT NULL COMMENT '供应商名称',
		    contact VARCHAR(50) COMMENT '联系人',
		    phone VARCHAR(20) COMMENT '联系电话',
		    email VARCHAR(100) COMMENT '邮箱',
		    address VARCHAR(255) COMMENT '地址',
		    latitude DECIMAL(10, 6) COMMENT '纬度',
		    longitude DECIMAL(10, 6) COMMENT '经度',
		    username VARCHAR(50) NOT NULL COMMENT '登录账号',
		    password VARCHAR(255) NOT NULL COMMENT '密码（加密存储）',
		    status TINYINT DEFAULT 1 COMMENT '状态：1-启用，0-禁用',
		    created_at DATETIME NOT NULL,
		    updated_at DATETIME NOT NULL,
		    UNIQUE KEY uk_username (username)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='供应商表';
		`

		_, err = DB.Exec(createSuppliersTableSQL)
		if err != nil {
			log.Printf("创建suppliers表失败: %v", err)
			return
		}

		// 检查并添加经纬度字段（如果不存在，用于兼容旧表）
		var latitudeExists int
		err = DB.QueryRow("SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = ? AND table_name = 'suppliers' AND column_name = 'latitude'", cfg.DBName).Scan(&latitudeExists)
		if err == nil && latitudeExists == 0 {
			_, err = DB.Exec("ALTER TABLE suppliers ADD COLUMN latitude DECIMAL(10, 6) COMMENT '纬度' AFTER address, ADD COLUMN longitude DECIMAL(10, 6) COMMENT '经度' AFTER latitude")
			if err != nil {
				log.Printf("添加经纬度字段失败: %v", err)
			} else {
				log.Println("已添加经纬度字段到suppliers表")
			}
		}

		// 检查供应商表是否有数据，如果没有则插入默认"自营"供应商
		var supplierCount int
		err = DB.QueryRow("SELECT COUNT(*) FROM suppliers").Scan(&supplierCount)
		if err != nil {
			log.Printf("查询供应商数据数量失败: %v", err)
			return
		}

		if supplierCount == 0 {
			// 使用bcrypt加密默认密码
			hashedPassword, hashErr := utils.HashPassword("self_operated_123")
			if hashErr != nil {
				log.Printf("加密默认供应商密码失败: %v", hashErr)
				err = hashErr
				return
			}

			// 插入默认"自营"供应商（昆明市中心坐标：纬度 25.040609，经度 102.712251）
			insertSupplierSQL := `
			INSERT INTO suppliers (name, contact, phone, email, address, latitude, longitude, username, password, status, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1, NOW(), NOW());
			`

			_, err = DB.Exec(insertSupplierSQL, "自营", "系统管理员", "", "", "云南省昆明市", 25.040609, 102.712251, "self_operated", hashedPassword)
			if err != nil {
				log.Printf("插入默认供应商失败: %v", err)
				return
			}

			log.Println("供应商表初始化成功，已创建默认'自营'供应商")
		} else {
			log.Println("供应商表已存在，跳过初始化")
		}

		// 创建products表
		createProductsTableSQL := `
		CREATE TABLE IF NOT EXISTS products (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL COMMENT '商品名称',
    description TEXT COMMENT '商品描述',
    price DECIMAL(10,2) COMMENT '商品价格（废弃，使用规格价格）',
    original_price DECIMAL(10,2) COMMENT '商品原价（废弃，使用规格价格）',
    category_id INT NOT NULL COMMENT '分类ID',
    supplier_id INT DEFAULT NULL COMMENT '供应商ID',
    is_special TINYINT DEFAULT 0 COMMENT '是否特价：1-是，0-否',
    images TEXT COMMENT '商品图片（JSON格式）',
    specs TEXT COMMENT '商品规格（JSON格式）',
    status TINYINT DEFAULT 1 COMMENT '状态：1-启用，0-禁用',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    KEY idx_category_id (category_id),
    KEY idx_supplier_id (supplier_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品表';
		`

		_, err = DB.Exec(createProductsTableSQL)
		if err != nil {
			log.Printf("创建products表失败: %v", err)
			return
		}

		// 创建配送费基础设置表
		createDeliveryFeeSettingsTableSQL := `
		CREATE TABLE IF NOT EXISTS delivery_fee_settings (
		    id INT PRIMARY KEY AUTO_INCREMENT,
		    base_fee DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '基础配送费',
		    free_shipping_threshold DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '免配送费金额阈值',
		    description VARCHAR(255) DEFAULT '' COMMENT '备注',
		    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='配送费用设置';
		`

		if _, err = DB.Exec(createDeliveryFeeSettingsTableSQL); err != nil {
			log.Printf("创建delivery_fee_settings表失败: %v", err)
			return
		}

		// 创建配送费排除项表
		createDeliveryFeeExclusionsTableSQL := `
		CREATE TABLE IF NOT EXISTS delivery_fee_exclusions (
		    id INT PRIMARY KEY AUTO_INCREMENT,
		    item_type ENUM('category','product') NOT NULL COMMENT '排除类型：分类或商品',
		    target_id INT NOT NULL COMMENT '目标ID（分类或商品）',
		    min_quantity_for_free INT DEFAULT NULL COMMENT '单品免配送费所需数量，仅针对商品',
		    remark VARCHAR(255) DEFAULT '' COMMENT '备注说明',
		    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		    UNIQUE KEY uk_item_scope (item_type, target_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='配送费用排除项';
		`

		if _, err = DB.Exec(createDeliveryFeeExclusionsTableSQL); err != nil {
			log.Printf("创建delivery_fee_exclusions表失败: %v", err)
			return
		}

		// 创建mini_app_users表（必须在purchase_list_items之前创建，因为purchase_list_items有外键依赖）
		createMiniAppUsersTableSQL := `
		CREATE TABLE IF NOT EXISTS mini_app_users (
		    id INT PRIMARY KEY AUTO_INCREMENT,
		    unique_id VARCHAR(64) NOT NULL COMMENT '小程序用户唯一标识（OpenID）',
		    user_code VARCHAR(10) DEFAULT NULL COMMENT '用户编号（4-5位数）',
		    name VARCHAR(50) DEFAULT NULL COMMENT '用户姓名',
		    avatar VARCHAR(255) DEFAULT NULL COMMENT '用户头像',
		    phone VARCHAR(20) DEFAULT NULL COMMENT '手机号码',
		    sales_code VARCHAR(50) DEFAULT NULL COMMENT '绑定的销售员代码',
		    store_type VARCHAR(50) DEFAULT NULL COMMENT '店铺类型',
		    user_type VARCHAR(20) NOT NULL DEFAULT 'unknown' COMMENT '用户类型：retail/wholesale/unknown',
		    profile_completed TINYINT(1) NOT NULL DEFAULT 0 COMMENT '资料是否完善：0-未完善，1-已完善',
		    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		    UNIQUE KEY uk_unique_id (unique_id),
		    UNIQUE KEY uk_user_code (user_code)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='小程序用户表';
		`

		if _, err = DB.Exec(createMiniAppUsersTableSQL); err != nil {
			log.Printf("创建mini_app_users表失败: %v", err)
			return
		}

		// 创建采购单表（类似购物车）
		createPurchaseListTableSQL := `
		CREATE TABLE IF NOT EXISTS purchase_list_items (
		    id INT PRIMARY KEY AUTO_INCREMENT,
		    user_id INT NOT NULL COMMENT '小程序用户ID',
		    product_id INT NOT NULL COMMENT '商品ID',
		    product_name VARCHAR(255) NOT NULL COMMENT '商品名称快照',
		    product_image VARCHAR(255) DEFAULT NULL COMMENT '商品图片快照',
		    spec_name VARCHAR(100) NOT NULL COMMENT '规格名称',
		    spec_snapshot TEXT NOT NULL COMMENT '规格快照（JSON）',
		    quantity INT NOT NULL DEFAULT 1 COMMENT '采购数量',
		    is_special TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否精选商品',
		    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		    UNIQUE KEY uk_user_product_spec (user_id, product_id, spec_name),
		    KEY idx_user_id (user_id),
		    CONSTRAINT fk_purchase_list_user FOREIGN KEY (user_id) REFERENCES mini_app_users(id) ON DELETE CASCADE,
		    CONSTRAINT fk_purchase_list_product FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='小程序采购单';
		`

		if _, err = DB.Exec(createPurchaseListTableSQL); err != nil {
			log.Printf("创建purchase_list_items表失败: %v", err)
			return
		}

		// 检查并更新products表结构，确保price和original_price字段是可空的
		// 首先检查表是否存在
		var tableExists bool
		err = DB.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = 'products'", cfg.DBName).Scan(&tableExists)
		if err != nil {
			log.Printf("检查表是否存在失败: %v", err)
			return
		}

		if tableExists {
			// 检查并添加supplier_id字段（如果不存在）
			var supplierIdExists int
			err = DB.QueryRow("SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = ? AND table_name = 'products' AND column_name = 'supplier_id'", cfg.DBName).Scan(&supplierIdExists)
			if err == nil && supplierIdExists == 0 {
				_, err = DB.Exec("ALTER TABLE products ADD COLUMN supplier_id INT DEFAULT NULL COMMENT '供应商ID' AFTER category_id, ADD KEY idx_supplier_id (supplier_id)")
				if err != nil {
					log.Printf("添加supplier_id字段失败: %v", err)
				} else {
					log.Println("已添加supplier_id字段到products表")
				}
			}
			// 检查price字段是否可空
			var priceNullable string
			err = DB.QueryRow("SELECT IS_NULLABLE FROM information_schema.columns WHERE table_schema = ? AND table_name = 'products' AND column_name = 'price'", cfg.DBName).Scan(&priceNullable)
			if err != nil {
				log.Printf("检查price字段可空性失败: %v", err)
				return
			}

			// 如果price字段不是可空的，则修改
			if priceNullable == "NO" {
				_, err = DB.Exec("ALTER TABLE products MODIFY COLUMN price DECIMAL(10,2) NULL COMMENT '商品价格（废弃，使用规格价格）'")
				if err != nil {
					log.Printf("修改price字段为可空失败: %v", err)
					return
				}
				log.Println("已将price字段修改为可空")
			}

			// 检查original_price字段是否可空
			var originalPriceNullable string
			err = DB.QueryRow("SELECT IS_NULLABLE FROM information_schema.columns WHERE table_schema = ? AND table_name = 'products' AND column_name = 'original_price'", cfg.DBName).Scan(&originalPriceNullable)
			if err != nil {
				log.Printf("检查original_price字段可空性失败: %v", err)
				return
			}

			// 如果original_price字段不是可空的，则修改
			if originalPriceNullable == "NO" {
				_, err = DB.Exec("ALTER TABLE products MODIFY COLUMN original_price DECIMAL(10,2) NULL COMMENT '商品原价（废弃，使用规格价格）'")
				if err != nil {
					log.Printf("修改original_price字段为可空失败: %v", err)
					return
				}
				log.Println("已将original_price字段修改为可空")
			}
		}

		// 创建carousels表
		createCarouselsTableSQL := `
		CREATE TABLE IF NOT EXISTS carousels (
		    id INT PRIMARY KEY AUTO_INCREMENT,
		    image VARCHAR(255) NOT NULL COMMENT '轮播图图片地址',
		    title VARCHAR(255) DEFAULT '' COMMENT '轮播图标题',
		    link VARCHAR(255) DEFAULT '' COMMENT '链接地址',
		    sort INT DEFAULT 0 COMMENT '排序',
		    status TINYINT DEFAULT 1 COMMENT '状态：1-启用，0-禁用',
		    created_at DATETIME NOT NULL,
		    updated_at DATETIME NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='轮播图表';
		`

		_, err = DB.Exec(createCarouselsTableSQL)
		if err != nil {
			log.Printf("创建carousels表失败: %v", err)
			return
		}

		// 创建hot_products表（热销产品关联表）
		createHotProductsTableSQL := `
		CREATE TABLE IF NOT EXISTS hot_products (
		    id INT PRIMARY KEY AUTO_INCREMENT,
		    product_id INT NOT NULL COMMENT '商品ID',
		    sort INT DEFAULT 0 COMMENT '排序',
		    status TINYINT DEFAULT 1 COMMENT '状态：1-启用，0-禁用',
		    created_at DATETIME NOT NULL,
		    updated_at DATETIME NOT NULL,
		    UNIQUE KEY uk_product_id (product_id),
		    KEY idx_sort (sort),
		    KEY idx_status (status)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='热销产品关联表';
		`

		_, err = DB.Exec(createHotProductsTableSQL)
		if err != nil {
			log.Printf("创建hot_products表失败: %v", err)
			return
		}

		// 创建热门搜索关键词表
		createHotSearchKeywordsTableSQL := `
		CREATE TABLE IF NOT EXISTS hot_search_keywords (
		    id INT AUTO_INCREMENT PRIMARY KEY,
		    keyword VARCHAR(100) NOT NULL COMMENT '关键词',
		    sort INT DEFAULT 0 COMMENT '排序（越小越靠前）',
		    status TINYINT DEFAULT 1 COMMENT '状态：1启用，0禁用',
		    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='热门搜索关键词';
		`

		_, err = DB.Exec(createHotSearchKeywordsTableSQL)
		if err != nil {
			log.Printf("创建hot_search_keywords表失败: %v", err)
			return
		}

		// 检查热门搜索关键词表是否有数据，如果没有则插入默认数据
		var hotKeywordCount int
		err = DB.QueryRow("SELECT COUNT(*) FROM hot_search_keywords").Scan(&hotKeywordCount)
		if err != nil {
			log.Printf("查询热门搜索关键词数量失败: %v", err)
			return
		}

		if hotKeywordCount == 0 {
			insertHotKeywordsSQL := `
			INSERT INTO hot_search_keywords (keyword, sort, status) VALUES
			    ('抽纸', 1, 1),
			    ('下拉抽纸', 2, 1),
			    ('纸碗', 3, 1),
			    ('筷子', 4, 1),
			    ('勺子', 5, 1),
			    ('打包袋', 6, 1),
			    ('定制打包袋', 7, 1),
			    ('保鲜膜', 8, 1);
			`
			_, err = DB.Exec(insertHotKeywordsSQL)
			if err != nil {
				log.Printf("插入热门搜索关键词失败: %v", err)
				return
			}
			log.Println("热门搜索关键词表初始化成功")
		} else {
			log.Println("热门搜索关键词表已有数据，跳过初始化")
		}

		// 创建mini_app_addresses表（用户地址表）
		createMiniAppAddressesTableSQL := `
		CREATE TABLE IF NOT EXISTS mini_app_addresses (
		    id INT PRIMARY KEY AUTO_INCREMENT,
		    user_id INT NOT NULL COMMENT '用户ID（关联mini_app_users.id）',
		    name VARCHAR(100) NOT NULL COMMENT '地址名称（如：明丽烧烤）',
		    contact VARCHAR(50) NOT NULL COMMENT '联系人',
		    phone VARCHAR(20) NOT NULL COMMENT '手机号码',
		    address VARCHAR(255) NOT NULL COMMENT '详细地址',
		    avatar VARCHAR(255) DEFAULT NULL COMMENT '地址照片（门头照片）',
		    latitude DECIMAL(10,6) DEFAULT NULL COMMENT '纬度',
		    longitude DECIMAL(10,6) DEFAULT NULL COMMENT '经度',
		    store_type VARCHAR(50) DEFAULT NULL COMMENT '店铺类型',
		    sales_code VARCHAR(50) DEFAULT NULL COMMENT '业务员代码',
		    is_default TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否默认地址：0-否，1-是',
		    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		    INDEX idx_user_id (user_id),
		    INDEX idx_is_default (is_default),
		    FOREIGN KEY (user_id) REFERENCES mini_app_users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='小程序用户地址表';
		`

		if _, err = DB.Exec(createMiniAppAddressesTableSQL); err != nil {
			log.Printf("创建mini_app_addresses表失败: %v", err)
			return
		}

		// 确保地址表有avatar字段（兼容旧表）
		ensureAddressColumn := func(columnName, columnDefinition string) {
			var columnCount int
			checkQuery := `
				SELECT COUNT(*)
				FROM information_schema.columns
				WHERE table_schema = ? AND table_name = 'mini_app_addresses' AND column_name = ?
			`
			if err := DB.QueryRow(checkQuery, cfg.DBName, columnName).Scan(&columnCount); err != nil {
				log.Printf("检查mini_app_addresses.%s字段失败: %v", columnName, err)
				return
			}
			if columnCount == 0 {
				alterSQL := fmt.Sprintf("ALTER TABLE mini_app_addresses ADD COLUMN %s", columnDefinition)
				if _, err := DB.Exec(alterSQL); err != nil {
					log.Printf("添加mini_app_addresses.%s字段失败: %v", columnName, err)
				} else {
					log.Printf("已为mini_app_addresses添加%s字段", columnName)
				}
			}
		}
		ensureAddressColumn("avatar", "avatar VARCHAR(255) DEFAULT NULL COMMENT '地址照片（门头照片）'")

		ensureColumn := func(columnName, columnDefinition string) {
			var columnCount int
			checkQuery := `
				SELECT COUNT(*)
				FROM information_schema.columns
				WHERE table_schema = ? AND table_name = 'mini_app_users' AND column_name = ?
			`
			if err := DB.QueryRow(checkQuery, cfg.DBName, columnName).Scan(&columnCount); err != nil {
				log.Printf("检查mini_app_users.%s字段失败: %v", columnName, err)
				return
			}
			if columnCount == 0 {
				alterSQL := fmt.Sprintf("ALTER TABLE mini_app_users ADD COLUMN %s", columnDefinition)
				if _, err := DB.Exec(alterSQL); err != nil {
					log.Printf("添加mini_app_users.%s字段失败: %v", columnName, err)
				} else {
					log.Printf("已为mini_app_users添加%s字段", columnName)
				}
			}
		}

		ensureColumn("user_type", "user_type VARCHAR(20) NOT NULL DEFAULT 'unknown' COMMENT '用户类型：retail/wholesale/unknown'")
		ensureColumn("profile_completed", "profile_completed TINYINT(1) NOT NULL DEFAULT 0 COMMENT '资料是否完善：0-未完善，1-已完善'")
		ensureColumn("user_code", "user_code VARCHAR(10) DEFAULT NULL COMMENT '用户编号（4-5位数）'")
		ensureColumn("name", "name VARCHAR(50) DEFAULT NULL COMMENT '用户姓名' AFTER user_code")
		ensureColumn("is_sales_employee", "is_sales_employee TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否是销售员：0-否，1-是'")
		ensureColumn("sales_employee_id", "sales_employee_id INT DEFAULT NULL COMMENT '绑定的销售员ID（员工表ID）'")

		// 删除不需要的字段（如果存在）
		dropColumn := func(columnName string) {
			var columnCount int
			checkQuery := `
				SELECT COUNT(*)
				FROM information_schema.columns
				WHERE table_schema = ? AND table_name = 'mini_app_users' AND column_name = ?
			`
			if err := DB.QueryRow(checkQuery, cfg.DBName, columnName).Scan(&columnCount); err != nil {
				log.Printf("检查mini_app_users.%s字段失败: %v", columnName, err)
				return
			}
			if columnCount > 0 {
				alterSQL := fmt.Sprintf("ALTER TABLE mini_app_users DROP COLUMN %s", columnName)
				if _, err := DB.Exec(alterSQL); err != nil {
					log.Printf("删除mini_app_users.%s字段失败: %v", columnName, err)
				} else {
					log.Printf("已删除mini_app_users.%s字段", columnName)
				}
			}
		}
		// 删除已迁移到地址表的字段
		dropColumn("address")
		dropColumn("latitude")
		dropColumn("longitude")
		dropColumn("contact")
		// 注意：name字段现在用于存储用户姓名，不再删除

		// 将未完善资料的用户类型重置为未选择
		if _, err := DB.Exec(`
			UPDATE mini_app_users
			SET user_type = 'unknown'
			WHERE (user_type IS NULL OR user_type = '' OR user_type = 'retail')
			  AND (profile_completed IS NULL OR profile_completed = 0)
		`); err != nil {
			log.Printf("更新未完善资料的用户类型失败: %v", err)
		}

		// 创建employees表（员工表）
		createEmployeesTableSQL := `
		CREATE TABLE IF NOT EXISTS employees (
		    id INT PRIMARY KEY AUTO_INCREMENT,
		    employee_code VARCHAR(10) NOT NULL COMMENT '员工码（5位数）',
		    phone VARCHAR(20) NOT NULL COMMENT '手机号（登录账号）',
		    password VARCHAR(255) NOT NULL COMMENT '密码（加密后）',
		    name VARCHAR(50) DEFAULT NULL COMMENT '员工姓名',
		    is_delivery TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否是配送员：0-否，1-是',
		    is_sales TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否是销售员：0-否，1-是',
		    status TINYINT(1) NOT NULL DEFAULT 1 COMMENT '状态：1-启用，0-禁用',
		    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		    UNIQUE KEY uk_employee_code (employee_code),
		    UNIQUE KEY uk_phone (phone),
		    INDEX idx_status (status)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='员工表';
		`

		if _, err = DB.Exec(createEmployeesTableSQL); err != nil {
			log.Printf("创建employees表失败: %v", err)
			return
		}

		// 创建优惠券表
		createCouponsTableSQL := `
		CREATE TABLE IF NOT EXISTS coupons (
		    id INT PRIMARY KEY AUTO_INCREMENT,
		    name VARCHAR(100) NOT NULL COMMENT '优惠券名称',
		    type ENUM('delivery_fee','amount') NOT NULL COMMENT '类型：delivery_fee-配送费券，amount-金额券',
		    discount_value DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '优惠值：配送费券为0（全免），金额券为具体金额',
		    min_amount DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '最低使用金额，0表示无门槛',
		    category_ids TEXT DEFAULT NULL COMMENT '适用分类ID（JSON数组），空表示全品类',
		    total_count INT NOT NULL DEFAULT 0 COMMENT '发放总数，0表示不限制',
		    used_count INT NOT NULL DEFAULT 0 COMMENT '已使用数量',
		    status TINYINT DEFAULT 1 COMMENT '状态：1-启用，0-禁用',
		    valid_from DATETIME NOT NULL COMMENT '有效期开始时间',
		    valid_to DATETIME NOT NULL COMMENT '有效期结束时间',
		    description VARCHAR(500) DEFAULT '' COMMENT '优惠券说明',
		    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		    KEY idx_status (status),
		    KEY idx_valid_time (valid_from, valid_to)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='优惠券表';
		`

		if _, err = DB.Exec(createCouponsTableSQL); err != nil {
			log.Printf("创建coupons表失败: %v", err)
			return
		}

		// 创建用户优惠券关联表
		createUserCouponsTableSQL := `
		CREATE TABLE IF NOT EXISTS user_coupons (
		    id INT PRIMARY KEY AUTO_INCREMENT,
		    user_id INT NOT NULL COMMENT '用户ID',
		    coupon_id INT NOT NULL COMMENT '优惠券ID',
		    status ENUM('unused','used','expired') DEFAULT 'unused' COMMENT '状态：unused-未使用，used-已使用，expired-已过期',
		    used_at DATETIME DEFAULT NULL COMMENT '使用时间',
		    order_id INT DEFAULT NULL COMMENT '订单ID（使用时的订单）',
		    expires_at DATETIME DEFAULT NULL COMMENT '有效期（发放时设置，过期后无法使用）',
		    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		    KEY idx_user_id (user_id),
		    KEY idx_coupon_id (coupon_id),
		    KEY idx_status (status),
		    KEY idx_expires_at (expires_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户优惠券关联表';
		`

		if _, err = DB.Exec(createUserCouponsTableSQL); err != nil {
			log.Printf("创建user_coupons表失败: %v", err)
			return
		}

		// 创建优惠券发放记录表
		createCouponIssueLogsTableSQL := `
		CREATE TABLE IF NOT EXISTS coupon_issue_logs (
		    id INT PRIMARY KEY AUTO_INCREMENT,
		    user_id INT NOT NULL COMMENT '客户ID',
		    coupon_id INT NOT NULL COMMENT '优惠券ID',
		    coupon_name VARCHAR(255) NOT NULL COMMENT '优惠券名称快照',
		    quantity INT NOT NULL DEFAULT 1 COMMENT '发放数量',
		    reason VARCHAR(255) NOT NULL COMMENT '发放原因',
		    operator_type VARCHAR(20) NOT NULL COMMENT '操作人类型：admin/employee',
		    operator_id INT NOT NULL COMMENT '操作人ID',
		    operator_name VARCHAR(100) NOT NULL COMMENT '操作人名称',
		    expires_at DATETIME DEFAULT NULL COMMENT '到期时间',
		    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '发放时间',
		    KEY idx_user_id (user_id),
		    KEY idx_coupon_id (coupon_id),
		    KEY idx_created_at (created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='优惠券发放记录表';
		`

		if _, err = DB.Exec(createCouponIssueLogsTableSQL); err != nil {
			log.Printf("创建coupon_issue_logs表失败: %v", err)
			return
		}

		// 创建订单主表
		createOrdersTableSQL := `
		CREATE TABLE IF NOT EXISTS orders (
		    id INT PRIMARY KEY AUTO_INCREMENT,
		    order_number VARCHAR(32) UNIQUE COMMENT '订单编号',
		    user_id INT NOT NULL COMMENT '用户ID',
		    address_id INT NOT NULL COMMENT '地址ID',
		    status VARCHAR(20) NOT NULL DEFAULT 'pending_delivery' COMMENT '订单状态',
		    goods_amount DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '商品总金额',
		    delivery_fee DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '配送费',
		    points_discount DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '积分抵扣金额',
		    coupon_discount DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '优惠券抵扣金额',
		    is_urgent TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否加急订单',
		    urgent_fee DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '加急费用',
		    total_amount DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '实际应付金额',
		    remark VARCHAR(500) DEFAULT '' COMMENT '备注',
		    out_of_stock_strategy VARCHAR(20) NOT NULL DEFAULT 'contact_me' COMMENT '缺货处理策略',
		    trust_receipt TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否信任签收',
		    hide_price TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否隐藏价格',
		    require_phone_contact TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否要求配送时电话联系',
		    expected_delivery_at DATETIME DEFAULT NULL COMMENT '预计送达时间',
		    weather_info JSON DEFAULT NULL COMMENT '天气信息（JSON格式，存储温度、天气状况等）',
		    is_isolated TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否孤立订单（8公里内无其他订单）',
		    delivery_fee_calculation JSON DEFAULT NULL COMMENT '配送费计算结果（JSON格式，存储基础配送费、补贴、利润分成等）',
		    order_profit DECIMAL(10,2) DEFAULT NULL COMMENT '订单总利润（商品金额-商品成本）',
		    net_profit DECIMAL(10,2) DEFAULT NULL COMMENT '净利润（总利润-配送费成本）',
		    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		    KEY idx_user_id (user_id),
		    KEY idx_status (status),
		    KEY idx_order_number (order_number),
		    KEY idx_is_urgent (is_urgent),
		    KEY idx_is_isolated (is_isolated),
		    KEY idx_status_urgent_isolated (status, is_urgent, is_isolated) COMMENT '复合索引：用于订单池筛选'
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单主表';
		`

		if _, err = DB.Exec(createOrdersTableSQL); err != nil {
			log.Printf("创建orders表失败: %v", err)
			return
		}

		// 检查并添加缺失的字段（用于表已存在的情况）
		// 检查 is_urgent 字段
		var isUrgentExists int
		checkIsUrgentQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'orders' AND COLUMN_NAME = 'is_urgent'`
		if err := DB.QueryRow(checkIsUrgentQuery).Scan(&isUrgentExists); err == nil && isUrgentExists == 0 {
			if _, err = DB.Exec(`ALTER TABLE orders ADD COLUMN is_urgent TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否加急订单'`); err != nil {
				log.Printf("添加is_urgent字段失败: %v", err)
			} else {
				log.Println("已添加is_urgent字段到orders表")
			}
		}

		// 检查 urgent_fee 字段
		var urgentFeeExists int
		checkUrgentFeeQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'orders' AND COLUMN_NAME = 'urgent_fee'`
		if err := DB.QueryRow(checkUrgentFeeQuery).Scan(&urgentFeeExists); err == nil && urgentFeeExists == 0 {
			if _, err = DB.Exec(`ALTER TABLE orders ADD COLUMN urgent_fee DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '加急费用'`); err != nil {
				log.Printf("添加urgent_fee字段失败: %v", err)
			} else {
				log.Println("已添加urgent_fee字段到orders表")
			}
		}

		// 检查 weather_info 字段
		var weatherInfoExists int
		checkWeatherInfoQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'orders' AND COLUMN_NAME = 'weather_info'`
		if err := DB.QueryRow(checkWeatherInfoQuery).Scan(&weatherInfoExists); err == nil && weatherInfoExists == 0 {
			if _, err = DB.Exec(`ALTER TABLE orders ADD COLUMN weather_info JSON DEFAULT NULL COMMENT '天气信息（JSON格式，存储温度、天气状况等）'`); err != nil {
				log.Printf("添加weather_info字段失败: %v", err)
			} else {
				log.Println("已添加weather_info字段到orders表")
			}
		}

		// 检查 is_isolated 字段
		var isIsolatedExists int
		checkIsIsolatedQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'orders' AND COLUMN_NAME = 'is_isolated'`
		if err := DB.QueryRow(checkIsIsolatedQuery).Scan(&isIsolatedExists); err == nil && isIsolatedExists == 0 {
			if _, err = DB.Exec(`ALTER TABLE orders ADD COLUMN is_isolated TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否孤立订单（8公里内无其他订单）'`); err != nil {
				log.Printf("添加is_isolated字段失败: %v", err)
			} else {
				log.Println("已添加is_isolated字段到orders表")
			}
		}

		// 检查 delivery_fee_calculation 字段
		var deliveryFeeCalcExists int
		checkDeliveryFeeCalcQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'orders' AND COLUMN_NAME = 'delivery_fee_calculation'`
		if err := DB.QueryRow(checkDeliveryFeeCalcQuery).Scan(&deliveryFeeCalcExists); err == nil && deliveryFeeCalcExists == 0 {
			if _, err = DB.Exec(`ALTER TABLE orders ADD COLUMN delivery_fee_calculation JSON DEFAULT NULL COMMENT '配送费计算结果（JSON格式，存储基础配送费、补贴、利润分成等）'`); err != nil {
				log.Printf("添加delivery_fee_calculation字段失败: %v", err)
			} else {
				log.Println("已添加delivery_fee_calculation字段到orders表")
			}
		}

		// 检查 order_profit 字段
		var orderProfitExists int
		checkOrderProfitQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'orders' AND COLUMN_NAME = 'order_profit'`
		if err := DB.QueryRow(checkOrderProfitQuery).Scan(&orderProfitExists); err == nil && orderProfitExists == 0 {
			if _, err = DB.Exec(`ALTER TABLE orders ADD COLUMN order_profit DECIMAL(10,2) DEFAULT NULL COMMENT '订单总利润（商品金额-商品成本）'`); err != nil {
				log.Printf("添加order_profit字段失败: %v", err)
			} else {
				log.Println("已添加order_profit字段到orders表")
			}
		}

		// 检查 net_profit 字段
		var netProfitExists int
		checkNetProfitQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'orders' AND COLUMN_NAME = 'net_profit'`
		if err := DB.QueryRow(checkNetProfitQuery).Scan(&netProfitExists); err == nil && netProfitExists == 0 {
			if _, err = DB.Exec(`ALTER TABLE orders ADD COLUMN net_profit DECIMAL(10,2) DEFAULT NULL COMMENT '净利润（总利润-配送费成本）'`); err != nil {
				log.Printf("添加net_profit字段失败: %v", err)
			} else {
				log.Println("已添加net_profit字段到orders表")
			}
		}

		// 检查 delivery_employee_code 字段
		var deliveryEmployeeCodeExists int
		checkDeliveryEmployeeCodeQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'orders' AND COLUMN_NAME = 'delivery_employee_code'`
		if err := DB.QueryRow(checkDeliveryEmployeeCodeQuery).Scan(&deliveryEmployeeCodeExists); err == nil && deliveryEmployeeCodeExists == 0 {
			if _, err = DB.Exec(`ALTER TABLE orders ADD COLUMN delivery_employee_code VARCHAR(10) DEFAULT NULL COMMENT '配送员员工码（接单时记录）' AFTER status, ADD KEY idx_delivery_employee_code (delivery_employee_code)`); err != nil {
				log.Printf("添加delivery_employee_code字段失败: %v", err)
			} else {
				log.Println("已添加delivery_employee_code字段到orders表")
			}
		}

		// 检查 delivery_fee_settled 字段（配送费是否已结算）
		var deliveryFeeSettledExists int
		checkDeliveryFeeSettledQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'orders' AND COLUMN_NAME = 'delivery_fee_settled'`
		if err := DB.QueryRow(checkDeliveryFeeSettledQuery).Scan(&deliveryFeeSettledExists); err == nil && deliveryFeeSettledExists == 0 {
			if _, err = DB.Exec(`ALTER TABLE orders ADD COLUMN delivery_fee_settled TINYINT(1) NOT NULL DEFAULT 0 COMMENT '配送费是否已结算' AFTER delivery_fee_calculation, ADD KEY idx_delivery_fee_settled (delivery_fee_settled)`); err != nil {
				log.Printf("添加delivery_fee_settled字段失败: %v", err)
			} else {
				log.Println("已添加delivery_fee_settled字段到orders表")
			}
		}

		// 检查 settlement_date 字段（结算日期）
		var settlementDateExists int
		checkSettlementDateQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'orders' AND COLUMN_NAME = 'settlement_date'`
		if err := DB.QueryRow(checkSettlementDateQuery).Scan(&settlementDateExists); err == nil && settlementDateExists == 0 {
			if _, err = DB.Exec(`ALTER TABLE orders ADD COLUMN settlement_date DATETIME DEFAULT NULL COMMENT '结算日期' AFTER delivery_fee_settled, ADD KEY idx_settlement_date (settlement_date)`); err != nil {
				log.Printf("添加settlement_date字段失败: %v", err)
			} else {
				log.Println("已添加settlement_date字段到orders表")
			}
		}

		// 检查 is_locked 字段（订单锁定，防止修改时被接单）
		var isLockedExists int
		checkIsLockedQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'orders' AND COLUMN_NAME = 'is_locked'`
		if err := DB.QueryRow(checkIsLockedQuery).Scan(&isLockedExists); err == nil && isLockedExists == 0 {
			// 先检查settlement_date是否存在，如果不存在则添加到表末尾
			var settlementDateExists int
			checkSettlementDateQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
				WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'orders' AND COLUMN_NAME = 'settlement_date'`
			var alterSQL string
			if err := DB.QueryRow(checkSettlementDateQuery).Scan(&settlementDateExists); err == nil && settlementDateExists > 0 {
				alterSQL = `ALTER TABLE orders ADD COLUMN is_locked TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否被锁定（修改中）' AFTER settlement_date`
			} else {
				alterSQL = `ALTER TABLE orders ADD COLUMN is_locked TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否被锁定（修改中）'`
			}
			if _, err = DB.Exec(alterSQL); err != nil {
				log.Printf("添加is_locked字段失败: %v", err)
			} else {
				log.Println("已添加is_locked字段到orders表")
			}
		}

		// 检查 locked_by 字段（锁定者员工码）
		var lockedByExists int
		checkLockedByQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'orders' AND COLUMN_NAME = 'locked_by'`
		if err := DB.QueryRow(checkLockedByQuery).Scan(&lockedByExists); err == nil && lockedByExists == 0 {
			var alterSQL string
			var isLockedExistsCheck int
			checkIsLockedExistsQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
				WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'orders' AND COLUMN_NAME = 'is_locked'`
			if err := DB.QueryRow(checkIsLockedExistsQuery).Scan(&isLockedExistsCheck); err == nil && isLockedExistsCheck > 0 {
				alterSQL = `ALTER TABLE orders ADD COLUMN locked_by VARCHAR(10) DEFAULT NULL COMMENT '锁定者员工码' AFTER is_locked`
			} else {
				alterSQL = `ALTER TABLE orders ADD COLUMN locked_by VARCHAR(10) DEFAULT NULL COMMENT '锁定者员工码'`
			}
			if _, err = DB.Exec(alterSQL); err != nil {
				log.Printf("添加locked_by字段失败: %v", err)
			} else {
				log.Println("已添加locked_by字段到orders表")
			}
		}

		// 检查 locked_at 字段（锁定时间）
		var lockedAtExists int
		checkLockedAtQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'orders' AND COLUMN_NAME = 'locked_at'`
		if err := DB.QueryRow(checkLockedAtQuery).Scan(&lockedAtExists); err == nil && lockedAtExists == 0 {
			var alterSQL string
			var lockedByExistsCheck int
			checkLockedByExistsQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
				WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'orders' AND COLUMN_NAME = 'locked_by'`
			if err := DB.QueryRow(checkLockedByExistsQuery).Scan(&lockedByExistsCheck); err == nil && lockedByExistsCheck > 0 {
				alterSQL = `ALTER TABLE orders ADD COLUMN locked_at DATETIME DEFAULT NULL COMMENT '锁定时间' AFTER locked_by`
			} else {
				alterSQL = `ALTER TABLE orders ADD COLUMN locked_at DATETIME DEFAULT NULL COMMENT '锁定时间'`
			}
			if _, err = DB.Exec(alterSQL); err != nil {
				log.Printf("添加locked_at字段失败: %v", err)
			} else {
				log.Println("已添加locked_at字段到orders表")
			}
		}

		// 检查 order_items 表的 is_picked 字段
		var isPickedExists int
		checkIsPickedQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'order_items' AND COLUMN_NAME = 'is_picked'`
		if err := DB.QueryRow(checkIsPickedQuery).Scan(&isPickedExists); err == nil && isPickedExists == 0 {
			if _, err = DB.Exec(`ALTER TABLE order_items ADD COLUMN is_picked TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已取货' AFTER image, ADD KEY idx_is_picked (is_picked)`); err != nil {
				log.Printf("添加is_picked字段失败: %v", err)
			} else {
				log.Println("已添加is_picked字段到order_items表")
			}
		}

		// 检查并添加索引
		// 检查 idx_is_urgent 索引
		var idxIsUrgentExists int
		checkIdxIsUrgentQuery := `SELECT COUNT(*) FROM information_schema.STATISTICS 
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'orders' AND INDEX_NAME = 'idx_is_urgent'`
		if err := DB.QueryRow(checkIdxIsUrgentQuery).Scan(&idxIsUrgentExists); err == nil && idxIsUrgentExists == 0 {
			if _, err = DB.Exec(`ALTER TABLE orders ADD INDEX idx_is_urgent (is_urgent)`); err != nil {
				log.Printf("添加idx_is_urgent索引失败: %v", err)
			}
		}

		// 检查 idx_is_isolated 索引
		var idxIsIsolatedExists int
		checkIdxIsIsolatedQuery := `SELECT COUNT(*) FROM information_schema.STATISTICS 
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'orders' AND INDEX_NAME = 'idx_is_isolated'`
		if err := DB.QueryRow(checkIdxIsIsolatedQuery).Scan(&idxIsIsolatedExists); err == nil && idxIsIsolatedExists == 0 {
			if _, err = DB.Exec(`ALTER TABLE orders ADD INDEX idx_is_isolated (is_isolated)`); err != nil {
				log.Printf("添加idx_is_isolated索引失败: %v", err)
			}
		}

		// 检查 idx_status_urgent_isolated 复合索引
		var idxStatusUrgentIsolatedExists int
		checkIdxStatusUrgentIsolatedQuery := `SELECT COUNT(*) FROM information_schema.STATISTICS 
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'orders' AND INDEX_NAME = 'idx_status_urgent_isolated'`
		if err := DB.QueryRow(checkIdxStatusUrgentIsolatedQuery).Scan(&idxStatusUrgentIsolatedExists); err == nil && idxStatusUrgentIsolatedExists == 0 {
			if _, err = DB.Exec(`ALTER TABLE orders ADD INDEX idx_status_urgent_isolated (status, is_urgent, is_isolated) COMMENT '复合索引：用于订单池筛选'`); err != nil {
				log.Printf("添加idx_status_urgent_isolated索引失败: %v", err)
			}
		}

		// 统一历史状态：将旧状态 pending 归一为 pending_delivery
		if _, err = DB.Exec(`UPDATE orders SET status = 'pending_delivery' WHERE status = 'pending'`); err != nil {
			log.Printf("归一化订单状态(pending -> pending_delivery)失败: %v", err)
		}

		// 创建配送记录表（记录配送完成时的照片等信息）
		createDeliveryRecordsTableSQL := `
		CREATE TABLE IF NOT EXISTS delivery_records (
		    id INT PRIMARY KEY AUTO_INCREMENT,
		    order_id INT NOT NULL COMMENT '订单ID',
		    delivery_employee_code VARCHAR(10) NOT NULL COMMENT '配送员员工码',
		    product_image_url VARCHAR(500) DEFAULT NULL COMMENT '货物照片URL',
		    doorplate_image_url VARCHAR(500) DEFAULT NULL COMMENT '门牌照片URL',
		    completed_at DATETIME NOT NULL COMMENT '完成时间',
		    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		    KEY idx_order_id (order_id),
		    KEY idx_delivery_employee_code (delivery_employee_code),
		    KEY idx_completed_at (completed_at),
		    UNIQUE KEY uk_order_id (order_id) COMMENT '一个订单只能有一条配送记录'
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='配送记录表（配送完成时的照片等信息）';
		`

		if _, err = DB.Exec(createDeliveryRecordsTableSQL); err != nil {
			log.Printf("创建delivery_records表失败: %v", err)
		} else {
			log.Println("配送记录表初始化成功")
		}

		// 创建配送流程日志表（记录配送的整个流程：创建、接单、取货、配送、完成）
		createDeliveryLogsTableSQL := `
		CREATE TABLE IF NOT EXISTS delivery_logs (
		    id INT PRIMARY KEY AUTO_INCREMENT,
		    order_id INT NOT NULL COMMENT '订单ID',
		    action VARCHAR(50) NOT NULL COMMENT '操作类型：created-订单创建, accepted-接单, pickup_started-开始取货, pickup_completed-取货完成, delivering_started-开始配送, delivering_completed-配送完成',
		    delivery_employee_code VARCHAR(10) DEFAULT NULL COMMENT '配送员员工码（接单、取货、配送时记录）',
		    action_time DATETIME NOT NULL COMMENT '操作时间',
		    remark VARCHAR(500) DEFAULT NULL COMMENT '备注信息',
		    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		    KEY idx_order_id (order_id),
		    KEY idx_delivery_employee_code (delivery_employee_code),
		    KEY idx_action (action),
		    KEY idx_action_time (action_time),
		    KEY idx_order_action (order_id, action) COMMENT '复合索引：用于查询订单的某个操作'
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='配送流程日志表（记录配送的整个流程）';
		`

		if _, err = DB.Exec(createDeliveryLogsTableSQL); err != nil {
			log.Printf("创建delivery_logs表失败: %v", err)
		} else {
			log.Println("配送流程日志表初始化成功")
		}

		// 创建系统设置表
		createSystemSettingsTableSQL := `
		CREATE TABLE IF NOT EXISTS system_settings (
		    id INT PRIMARY KEY AUTO_INCREMENT,
		    setting_key VARCHAR(100) NOT NULL UNIQUE COMMENT '设置键名',
		    setting_value TEXT DEFAULT NULL COMMENT '设置值',
		    description VARCHAR(255) DEFAULT '' COMMENT '设置说明',
		    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		    INDEX idx_setting_key (setting_key)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统设置表';
		`

		if _, err = DB.Exec(createSystemSettingsTableSQL); err != nil {
			log.Printf("创建system_settings表失败: %v", err)
			return
		}

		// 创建配送路线排序表（记录配送员已接单订单的排序）
		createDeliveryRouteOrdersTableSQL := `
		CREATE TABLE IF NOT EXISTS delivery_route_orders (
		    id INT PRIMARY KEY AUTO_INCREMENT,
		    delivery_employee_code VARCHAR(10) NOT NULL COMMENT '配送员员工码',
		    batch_id VARCHAR(50) NOT NULL COMMENT '批次ID（用于区分不同的趟）',
		    order_id INT NOT NULL COMMENT '订单ID',
		    route_sequence INT NOT NULL COMMENT '路线排序序号（从1开始）',
		    calculated_distance DECIMAL(10,2) DEFAULT NULL COMMENT '计算的距离（公里）',
		    calculated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '计算时间',
		    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		    UNIQUE KEY uk_employee_order (delivery_employee_code, order_id) COMMENT '同一配送员的同一订单只能有一条记录',
		    KEY idx_delivery_employee_code (delivery_employee_code),
		    KEY idx_batch_id (delivery_employee_code, batch_id) COMMENT '用于按批次查询',
		    KEY idx_order_id (order_id),
		    KEY idx_route_sequence (delivery_employee_code, batch_id, route_sequence) COMMENT '用于按批次和排序查询'
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='配送路线排序表';
		`

		if _, err = DB.Exec(createDeliveryRouteOrdersTableSQL); err != nil {
			log.Printf("创建delivery_route_orders表失败: %v", err)
		} else {
			log.Println("配送路线排序表初始化成功")
		}

		// 为已存在的表添加 batch_id 字段（兼容旧数据）
		// 先检查字段是否存在
		var batchColumnExists int
		checkBatchColumnSQL := `
			SELECT COUNT(*) 
			FROM INFORMATION_SCHEMA.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() 
			AND TABLE_NAME = 'delivery_route_orders' 
			AND COLUMN_NAME = 'batch_id'
		`
		if err := DB.QueryRow(checkBatchColumnSQL).Scan(&batchColumnExists); err == nil && batchColumnExists == 0 {
			// 字段不存在，添加字段和索引
			alterDeliveryRouteOrdersTableSQL := `
			ALTER TABLE delivery_route_orders 
			ADD COLUMN batch_id VARCHAR(50) NOT NULL DEFAULT '' COMMENT '批次ID（用于区分不同的趟）' AFTER delivery_employee_code,
			ADD INDEX idx_batch_id (delivery_employee_code, batch_id) COMMENT '用于按批次查询',
			ADD INDEX idx_route_sequence_batch (delivery_employee_code, batch_id, route_sequence) COMMENT '用于按批次和排序查询'
			`
			if _, alterErr := DB.Exec(alterDeliveryRouteOrdersTableSQL); alterErr != nil {
				log.Printf("添加 batch_id 字段失败: %v", alterErr)
			} else {
				log.Println("为 delivery_route_orders 表添加 batch_id 字段成功")
			}
		} else {
			log.Println("delivery_route_orders 表的 batch_id 字段已存在，跳过添加")
		}

		// 初始化默认系统设置
		initSystemSettings := []struct {
			key         string
			value       string
			description string
		}{
			{"map_amap_key", "", "高德地图API Key"},
			{"map_tencent_key", "", "腾讯地图API Key"},
			{"order_urgent_fee", "0", "加急订单费用（元）"},
			// 配送费计算配置
			{"delivery_base_fee", "4.0", "基础配送费（元）"},
			{"delivery_isolated_distance", "8.0", "孤立订单判断距离（公里）"},
			{"delivery_isolated_subsidy", "3.0", "孤立订单补贴（元）"},
			{"delivery_item_threshold_low", "5", "件数补贴低阈值（件）"},
			{"delivery_item_rate_low", "0.5", "件数补贴低档费率（元/件）"},
			{"delivery_item_threshold_high", "10", "件数补贴高阈值（件）"},
			{"delivery_item_rate_high", "0.6", "件数补贴高档费率（元/件）"},
			{"delivery_item_max_count", "50", "件数补贴最大计件数"},
			{"delivery_urgent_subsidy", "10.0", "加急订单补贴（元）"},
			{"delivery_weather_subsidy", "1.0", "极端天气补贴（元）"},
			{"delivery_extreme_temp", "37.0", "极端高温阈值（摄氏度）"},
			{"delivery_profit_threshold", "25.0", "利润分成阈值（元）"},
			{"delivery_profit_share_rate", "0.08", "利润分成比例（8%）"},
			{"delivery_max_profit_share", "50.0", "利润分成上限（元）"},
		}

		for _, setting := range initSystemSettings {
			var count int
			checkQuery := `SELECT COUNT(*) FROM system_settings WHERE setting_key = ?`
			if err := DB.QueryRow(checkQuery, setting.key).Scan(&count); err != nil {
				log.Printf("检查系统设置 %s 失败: %v", setting.key, err)
				continue
			}
			if count == 0 {
				insertQuery := `INSERT INTO system_settings (setting_key, setting_value, description) VALUES (?, ?, ?)`
				if _, err := DB.Exec(insertQuery, setting.key, setting.value, setting.description); err != nil {
					log.Printf("初始化系统设置 %s 失败: %v", setting.key, err)
				}
			}
		}

		// 注意：如果 orders 表已存在但没有 order_number 字段，需要手动执行迁移脚本
		// 迁移脚本位置：go_backend/migration_add_order_number.sql

		// 创建订单明细表
		createOrderItemsTableSQL := `
		CREATE TABLE IF NOT EXISTS order_items (
		    id INT PRIMARY KEY AUTO_INCREMENT,
		    order_id INT NOT NULL COMMENT '订单ID',
		    product_id INT NOT NULL COMMENT '商品ID',
		    product_name VARCHAR(200) NOT NULL COMMENT '商品名称',
		    spec_name VARCHAR(100) DEFAULT '' COMMENT '规格名称',
		    quantity INT NOT NULL COMMENT '数量',
		    unit_price DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '成交单价',
		    subtotal DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '小计',
		    image VARCHAR(255) DEFAULT '' COMMENT '商品图片',
		    KEY idx_order_id (order_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单明细表';
		`

		if _, err = DB.Exec(createOrderItemsTableSQL); err != nil {
			log.Printf("创建order_items表失败: %v", err)
			return
		}

		// 检查并添加 expires_at 字段（如果表已存在但字段不存在）
		// MySQL 不支持 IF NOT EXISTS 在 ADD COLUMN，需要先检查
		var columnExists int
		checkColumnSQL := `
		SELECT COUNT(*) FROM information_schema.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'user_coupons' 
		AND COLUMN_NAME = 'expires_at'
		`
		err = DB.QueryRow(checkColumnSQL).Scan(&columnExists)
		if err == nil && columnExists == 0 {
			// 字段不存在，添加字段
			alterSQL := `ALTER TABLE user_coupons ADD COLUMN expires_at DATETIME DEFAULT NULL COMMENT '有效期（发放时设置，过期后无法使用）' AFTER order_id`
			if _, err = DB.Exec(alterSQL); err != nil {
				log.Printf("添加expires_at字段失败: %v", err)
			} else {
				log.Println("成功添加expires_at字段")
			}
			// 添加索引
			alterIndexSQL := `ALTER TABLE user_coupons ADD INDEX idx_expires_at (expires_at)`
			if _, err = DB.Exec(alterIndexSQL); err != nil {
				// 索引可能已存在，忽略错误
				log.Printf("添加expires_at索引失败（可能已存在）: %v", err)
			}
		}

		// 检查并删除唯一键约束 uk_user_coupon（因为现在支持一个用户拥有多张相同的优惠券）
		var constraintExists int
		checkConstraintSQL := `
		SELECT COUNT(*) FROM information_schema.TABLE_CONSTRAINTS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'user_coupons' 
		AND CONSTRAINT_NAME = 'uk_user_coupon'
		AND CONSTRAINT_TYPE = 'UNIQUE'
		`
		err = DB.QueryRow(checkConstraintSQL).Scan(&constraintExists)
		if err == nil && constraintExists > 0 {
			// 唯一键约束存在，删除它
			dropConstraintSQL := `ALTER TABLE user_coupons DROP INDEX uk_user_coupon`
			if _, err = DB.Exec(dropConstraintSQL); err != nil {
				log.Printf("删除唯一键约束uk_user_coupon失败: %v", err)
			} else {
				log.Println("成功删除唯一键约束uk_user_coupon")
			}
		}

		// 创建配送员位置历史表
		createEmployeeLocationHistoryTableSQL := `
		CREATE TABLE IF NOT EXISTS employee_location_history (
		    id INT PRIMARY KEY AUTO_INCREMENT,
		    employee_id INT NOT NULL COMMENT '员工ID',
		    employee_code VARCHAR(20) NOT NULL COMMENT '员工码',
		    latitude DECIMAL(10, 8) NOT NULL COMMENT '纬度',
		    longitude DECIMAL(11, 8) NOT NULL COMMENT '经度',
		    accuracy DECIMAL(10, 2) DEFAULT NULL COMMENT '精度（米）',
		    created_at DATETIME NOT NULL COMMENT '创建时间',
		    INDEX idx_employee_id (employee_id),
		    INDEX idx_employee_code (employee_code),
		    INDEX idx_created_at (created_at),
		    KEY idx_employee_created (employee_id, created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='配送员位置历史表';
		`

		if _, err = DB.Exec(createEmployeeLocationHistoryTableSQL); err != nil {
			log.Printf("创建employee_location_history表失败: %v", err)
		} else {
			log.Println("配送员位置历史表初始化成功")
		}

		// 创建销售分成配置表
		createSalesCommissionConfigTableSQL := `
		CREATE TABLE IF NOT EXISTS sales_commission_config (
		    id INT PRIMARY KEY AUTO_INCREMENT,
		    employee_code VARCHAR(10) NOT NULL UNIQUE COMMENT '销售员员工码',
		    base_commission_rate DECIMAL(5,4) NOT NULL DEFAULT 0.4500 COMMENT '基础提成比例（默认45%）',
		    new_customer_bonus_rate DECIMAL(5,4) NOT NULL DEFAULT 0.2000 COMMENT '新客开发激励比例（默认20%）',
		    tier1_threshold DECIMAL(10,2) NOT NULL DEFAULT 50000.00 COMMENT '阶梯1阈值（默认50000元）',
		    tier1_rate DECIMAL(5,4) NOT NULL DEFAULT 0.0500 COMMENT '阶梯1提成比例（默认5%）',
		    tier2_threshold DECIMAL(10,2) NOT NULL DEFAULT 100000.00 COMMENT '阶梯2阈值（默认100000元）',
		    tier2_rate DECIMAL(5,4) NOT NULL DEFAULT 0.1000 COMMENT '阶梯2提成比例（默认10%）',
		    tier3_threshold DECIMAL(10,2) NOT NULL DEFAULT 200000.00 COMMENT '阶梯3阈值（默认200000元）',
		    tier3_rate DECIMAL(5,4) NOT NULL DEFAULT 0.2000 COMMENT '阶梯3提成比例（默认20%）',
		    min_profit_threshold DECIMAL(10,2) NOT NULL DEFAULT 5.00 COMMENT '最小利润阈值（默认5元）',
		    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		    KEY idx_employee_code (employee_code)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='销售分成配置表';
		`

		if _, err = DB.Exec(createSalesCommissionConfigTableSQL); err != nil {
			log.Printf("创建sales_commission_config表失败: %v", err)
		} else {
			log.Println("销售分成配置表初始化成功")
		}

		// 创建销售分成记录表
		createSalesCommissionsTableSQL := `
		CREATE TABLE IF NOT EXISTS sales_commissions (
		    id INT PRIMARY KEY AUTO_INCREMENT,
		    order_id INT NOT NULL COMMENT '订单ID',
		    employee_code VARCHAR(10) NOT NULL COMMENT '销售员员工码',
		    user_id INT NOT NULL COMMENT '客户用户ID',
		    order_number VARCHAR(32) NOT NULL COMMENT '订单编号',
		    order_date DATE NOT NULL COMMENT '订单日期',
		    settlement_date DATE DEFAULT NULL COMMENT '结算日期（订单结算时记录）',
		    is_valid_order TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否有效订单（已结算且利润>阈值）',
		    is_new_customer_order TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否新客户首单',
		    order_amount DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '订单金额（平台总收入）',
		    goods_cost DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '商品总成本',
		    delivery_cost DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '配送成本',
		    order_profit DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '订单利润（订单金额-商品成本-配送成本）',
		    base_commission DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '基础提成',
		    new_customer_bonus DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '新客开发激励',
		    tier_commission DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '阶梯提成',
		    total_commission DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '总分成',
		    tier_level INT NOT NULL DEFAULT 0 COMMENT '达到的阶梯等级（0-未达到，1-阶梯1，2-阶梯2，3-阶梯3）',
		    calculation_month VARCHAR(7) NOT NULL COMMENT '计算月份（YYYY-MM格式）',
		    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		    UNIQUE KEY uk_order_employee (order_id, employee_code) COMMENT '同一订单同一销售员只能有一条记录',
		    KEY idx_employee_code (employee_code),
		    KEY idx_order_id (order_id),
		    KEY idx_settlement_date (settlement_date),
		    KEY idx_calculation_month (employee_code, calculation_month),
		    KEY idx_is_valid_order (is_valid_order),
		    KEY idx_order_date (order_date)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='销售分成记录表';
		`

		if _, err = DB.Exec(createSalesCommissionsTableSQL); err != nil {
			log.Printf("创建sales_commissions表失败: %v", err)
		} else {
			log.Println("销售分成记录表初始化成功")
		}

		// 添加计入和结算相关字段（如果不存在）
		// 检查 is_accounted 字段
		var isAccountedExists int
		checkIsAccountedQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sales_commissions' AND COLUMN_NAME = 'is_accounted'`
		if err := DB.QueryRow(checkIsAccountedQuery).Scan(&isAccountedExists); err == nil && isAccountedExists == 0 {
			if _, err = DB.Exec(`ALTER TABLE sales_commissions ADD COLUMN is_accounted TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已计入（平台承认了销售员这个分润收入）' AFTER calculation_month, ADD KEY idx_is_accounted (is_accounted)`); err != nil {
				log.Printf("添加is_accounted字段失败: %v", err)
			} else {
				log.Println("已添加is_accounted字段到sales_commissions表")
			}
		}

		// 检查 accounted_at 字段
		var accountedAtExists int
		checkAccountedAtQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sales_commissions' AND COLUMN_NAME = 'accounted_at'`
		if err := DB.QueryRow(checkAccountedAtQuery).Scan(&accountedAtExists); err == nil && accountedAtExists == 0 {
			if _, err = DB.Exec(`ALTER TABLE sales_commissions ADD COLUMN accounted_at DATETIME DEFAULT NULL COMMENT '计入时间' AFTER is_accounted, ADD KEY idx_accounted_at (accounted_at)`); err != nil {
				log.Printf("添加accounted_at字段失败: %v", err)
			} else {
				log.Println("已添加accounted_at字段到sales_commissions表")
			}
		}

		// 检查 is_settled 字段
		var isSettledExists int
		checkIsSettledQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sales_commissions' AND COLUMN_NAME = 'is_settled'`
		if err := DB.QueryRow(checkIsSettledQuery).Scan(&isSettledExists); err == nil && isSettledExists == 0 {
			if _, err = DB.Exec(`ALTER TABLE sales_commissions ADD COLUMN is_settled TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否已结算（平台已经将该费用结算给销售员）' AFTER accounted_at, ADD KEY idx_is_settled (is_settled)`); err != nil {
				log.Printf("添加is_settled字段失败: %v", err)
			} else {
				log.Println("已添加is_settled字段到sales_commissions表")
			}
		}

		// 检查 settled_at 字段
		var settledAtExists int
		checkSettledAtQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sales_commissions' AND COLUMN_NAME = 'settled_at'`
		if err := DB.QueryRow(checkSettledAtQuery).Scan(&settledAtExists); err == nil && settledAtExists == 0 {
			if _, err = DB.Exec(`ALTER TABLE sales_commissions ADD COLUMN settled_at DATETIME DEFAULT NULL COMMENT '结算时间' AFTER is_settled, ADD KEY idx_settled_at (settled_at)`); err != nil {
				log.Printf("添加settled_at字段失败: %v", err)
			} else {
				log.Println("已添加settled_at字段到sales_commissions表")
			}
		}

		// 检查 is_accounted_cancelled 字段
		var isAccountedCancelledExists int
		checkIsAccountedCancelledQuery := `SELECT COUNT(*) FROM information_schema.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sales_commissions' AND COLUMN_NAME = 'is_accounted_cancelled'`
		if err := DB.QueryRow(checkIsAccountedCancelledQuery).Scan(&isAccountedCancelledExists); err == nil && isAccountedCancelledExists == 0 {
			if _, err = DB.Exec(`ALTER TABLE sales_commissions ADD COLUMN is_accounted_cancelled TINYINT(1) NOT NULL DEFAULT 0 COMMENT '计入是否已取消' AFTER settled_at, ADD KEY idx_is_accounted_cancelled (is_accounted_cancelled)`); err != nil {
				log.Printf("添加is_accounted_cancelled字段失败: %v", err)
			} else {
				log.Println("已添加is_accounted_cancelled字段到sales_commissions表")
			}
		}

		// 创建销售分成月统计表
		createSalesCommissionMonthlyStatsTableSQL := `
		CREATE TABLE IF NOT EXISTS sales_commission_monthly_stats (
		    id INT PRIMARY KEY AUTO_INCREMENT,
		    employee_code VARCHAR(10) NOT NULL COMMENT '销售员员工码',
		    stat_month VARCHAR(7) NOT NULL COMMENT '统计月份（YYYY-MM格式）',
		    total_sales_amount DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '总销售额（有效订单金额）',
		    total_valid_orders INT NOT NULL DEFAULT 0 COMMENT '有效订单数',
		    total_new_customers INT NOT NULL DEFAULT 0 COMMENT '新客户数',
		    total_profit DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '总利润',
		    total_base_commission DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '总基础提成',
		    total_new_customer_bonus DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '总新客激励',
		    total_tier_commission DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '总阶梯提成',
		    total_commission DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '总分成',
		    tier_level INT NOT NULL DEFAULT 0 COMMENT '达到的阶梯等级（0-未达到，1-阶梯1，2-阶梯2，3-阶梯3）',
		    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		    UNIQUE KEY uk_employee_month (employee_code, stat_month) COMMENT '同一销售员同一月份只能有一条记录',
		    KEY idx_employee_code (employee_code),
		    KEY idx_stat_month (stat_month)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='销售分成月统计表';
		`

		if _, err = DB.Exec(createSalesCommissionMonthlyStatsTableSQL); err != nil {
			log.Printf("创建sales_commission_monthly_stats表失败: %v", err)
		} else {
			log.Println("销售分成月统计表初始化成功")
		}

		// 创建收款审核申请表
		createPaymentVerificationRequestsTableSQL := `
		CREATE TABLE IF NOT EXISTS payment_verification_requests (
		    id INT AUTO_INCREMENT PRIMARY KEY,
		    order_id INT NOT NULL COMMENT '订单ID',
		    order_number VARCHAR(50) NOT NULL COMMENT '订单号',
		    sales_employee_code VARCHAR(50) NOT NULL COMMENT '销售员代码',
		    sales_employee_name VARCHAR(100) COMMENT '销售员姓名',
		    customer_id INT NOT NULL COMMENT '客户ID',
		    customer_name VARCHAR(100) COMMENT '客户姓名',
		    order_amount DECIMAL(10, 2) NOT NULL COMMENT '订单金额',
		    request_reason TEXT COMMENT '申请原因/备注',
		    status ENUM('pending', 'approved', 'rejected') NOT NULL DEFAULT 'pending' COMMENT '审核状态：pending-待审核，approved-已通过，rejected-已拒绝',
		    admin_id INT COMMENT '审核管理员ID',
		    admin_name VARCHAR(100) COMMENT '审核管理员姓名',
		    reviewed_at DATETIME COMMENT '审核时间',
		    review_remark TEXT COMMENT '审核备注',
		    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
		    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
		    INDEX idx_order_id (order_id),
		    INDEX idx_sales_employee_code (sales_employee_code),
		    INDEX idx_status (status),
		    INDEX idx_created_at (created_at),
		    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='收款审核申请表';
		`

		if _, err = DB.Exec(createPaymentVerificationRequestsTableSQL); err != nil {
			log.Printf("创建payment_verification_requests表失败: %v", err)
		} else {
			log.Println("收款审核申请表初始化成功")
		}

		// 创建富文本内容表
		createRichContentsTableSQL := `
		CREATE TABLE IF NOT EXISTS rich_contents (
		    id INT AUTO_INCREMENT PRIMARY KEY,
		    title VARCHAR(200) NOT NULL COMMENT '富文本标题',
		    content LONGTEXT NOT NULL COMMENT '富文本HTML内容',
		    content_type VARCHAR(50) NOT NULL DEFAULT 'notice' COMMENT '内容类型：notice(通知), activity(活动), other(其他)',
		    status VARCHAR(20) NOT NULL DEFAULT 'draft' COMMENT '状态：draft(草稿), published(已发布), archived(已归档)',
		    published_at DATETIME NULL COMMENT '发布时间',
		    view_count INT DEFAULT 0 COMMENT '浏览次数',
		    created_by VARCHAR(100) COMMENT '创建人',
		    updated_by VARCHAR(100) COMMENT '更新人',
		    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		    INDEX idx_content_type (content_type),
		    INDEX idx_status (status),
		    INDEX idx_published_at (published_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='富文本内容表';
		`

		if _, err = DB.Exec(createRichContentsTableSQL); err != nil {
			log.Printf("创建rich_contents表失败: %v", err)
		} else {
			log.Println("富文本内容表初始化成功")
		}

		// 创建发票抬头表
		createInvoicesTableSQL := `
		CREATE TABLE IF NOT EXISTS mini_app_invoices (
		    id INT PRIMARY KEY AUTO_INCREMENT,
		    user_id INT NOT NULL COMMENT '用户ID',
		    invoice_type VARCHAR(20) NOT NULL DEFAULT 'personal' COMMENT '发票类型：personal(个人), company(企业)',
		    title VARCHAR(200) NOT NULL COMMENT '发票抬头',
		    tax_number VARCHAR(50) DEFAULT '' COMMENT '纳税人识别号（企业必填）',
		    company_address VARCHAR(255) DEFAULT '' COMMENT '公司地址（企业可选）',
		    company_phone VARCHAR(50) DEFAULT '' COMMENT '公司电话（企业可选）',
		    bank_name VARCHAR(100) DEFAULT '' COMMENT '开户银行（企业可选）',
		    bank_account VARCHAR(100) DEFAULT '' COMMENT '银行账号（企业可选）',
		    is_default TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否默认',
		    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		    INDEX idx_user_id (user_id),
		    INDEX idx_is_default (is_default),
		    FOREIGN KEY (user_id) REFERENCES mini_app_users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='发票抬头表';
		`

		if _, err = DB.Exec(createInvoicesTableSQL); err != nil {
			log.Printf("创建mini_app_invoices表失败: %v", err)
		} else {
			log.Println("发票抬头表初始化成功")
		}

		// 创建新品需求表
		createProductRequestsTableSQL := `
		CREATE TABLE IF NOT EXISTS product_requests (
		    id INT PRIMARY KEY AUTO_INCREMENT,
		    user_id INT NOT NULL COMMENT '用户ID',
		    product_name VARCHAR(255) NOT NULL COMMENT '需求产品名称',
		    brand VARCHAR(100) DEFAULT '' COMMENT '品牌',
		    monthly_quantity INT DEFAULT 0 COMMENT '月消耗数量',
		    description TEXT COMMENT '需求说明',
		    status ENUM('pending', 'processing', 'completed', 'rejected') NOT NULL DEFAULT 'pending' COMMENT '状态：pending-待处理，processing-处理中，completed-已完成，rejected-已拒绝',
		    admin_remark TEXT COMMENT '管理员备注',
		    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		    INDEX idx_user_id (user_id),
		    INDEX idx_status (status),
		    INDEX idx_created_at (created_at),
		    FOREIGN KEY (user_id) REFERENCES mini_app_users(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='新品需求表';
		`

		if _, err = DB.Exec(createProductRequestsTableSQL); err != nil {
			log.Printf("创建product_requests表失败: %v", err)
		} else {
			log.Println("新品需求表初始化成功")
		}

		// 创建供应商合作申请表
		createSupplierApplicationsTableSQL := `
		CREATE TABLE IF NOT EXISTS supplier_applications (
		    id INT PRIMARY KEY AUTO_INCREMENT,
		    user_id INT COMMENT '用户ID（如果已登录）',
		    company_name VARCHAR(255) NOT NULL COMMENT '公司名称',
		    contact_name VARCHAR(100) NOT NULL COMMENT '联系人',
		    contact_phone VARCHAR(20) NOT NULL COMMENT '联系电话',
		    email VARCHAR(100) DEFAULT '' COMMENT '邮箱',
		    address VARCHAR(500) DEFAULT '' COMMENT '公司地址',
		    main_category VARCHAR(100) NOT NULL COMMENT '主营类目',
		    company_intro TEXT COMMENT '公司简介',
		    cooperation_intent TEXT COMMENT '合作意向说明',
		    status ENUM('pending', 'approved', 'rejected') NOT NULL DEFAULT 'pending' COMMENT '状态：pending-待审核，approved-已通过，rejected-已拒绝',
		    admin_remark TEXT COMMENT '管理员备注',
		    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		    INDEX idx_user_id (user_id),
		    INDEX idx_status (status),
		    INDEX idx_created_at (created_at),
		    FOREIGN KEY (user_id) REFERENCES mini_app_users(id) ON DELETE SET NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='供应商合作申请表';
		`

		if _, err = DB.Exec(createSupplierApplicationsTableSQL); err != nil {
			log.Printf("创建supplier_applications表失败: %v", err)
		} else {
			log.Println("供应商合作申请表初始化成功")
		}

		log.Println("所有表创建成功")
	})

	return err
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
