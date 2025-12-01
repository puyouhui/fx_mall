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

			// 插入默认"自营"供应商
			insertSupplierSQL := `
			INSERT INTO suppliers (name, contact, phone, email, address, username, password, status, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, 1, NOW(), NOW());
			`

			_, err = DB.Exec(insertSupplierSQL, "自营", "系统管理员", "", "", "", "self_operated", hashedPassword)
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
		    total_amount DECIMAL(10,2) NOT NULL DEFAULT 0 COMMENT '实际应付金额',
		    remark VARCHAR(500) DEFAULT '' COMMENT '备注',
		    out_of_stock_strategy VARCHAR(20) NOT NULL DEFAULT 'contact_me' COMMENT '缺货处理策略',
		    trust_receipt TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否信任签收',
		    hide_price TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否隐藏价格',
		    require_phone_contact TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否要求配送时电话联系',
		    expected_delivery_at DATETIME DEFAULT NULL COMMENT '预计送达时间',
		    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		    KEY idx_user_id (user_id),
		    KEY idx_status (status),
		    KEY idx_order_number (order_number)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单主表';
		`

		if _, err = DB.Exec(createOrdersTableSQL); err != nil {
			log.Printf("创建orders表失败: %v", err)
			return
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
