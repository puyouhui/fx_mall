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

		// 创建mini_app_users表
		createMiniAppUsersTableSQL := `
		CREATE TABLE IF NOT EXISTS mini_app_users (
		    id INT PRIMARY KEY AUTO_INCREMENT,
		    unique_id VARCHAR(64) NOT NULL COMMENT '小程序用户唯一标识（OpenID）',
		    avatar VARCHAR(255) DEFAULT NULL COMMENT '用户头像',
		    name VARCHAR(100) DEFAULT NULL COMMENT '名称',
		    contact VARCHAR(50) DEFAULT NULL COMMENT '联系人',
		    phone VARCHAR(20) DEFAULT NULL COMMENT '手机号码',
		    address VARCHAR(255) DEFAULT NULL COMMENT '中文地址',
		    latitude DECIMAL(10,6) DEFAULT NULL COMMENT '纬度',
		    longitude DECIMAL(10,6) DEFAULT NULL COMMENT '经度',
		    sales_code VARCHAR(50) DEFAULT NULL COMMENT '绑定的销售员代码',
		    store_type VARCHAR(50) DEFAULT NULL COMMENT '店铺类型',
		    user_type VARCHAR(20) NOT NULL DEFAULT 'unknown' COMMENT '用户类型：retail/wholesale/unknown',
		    profile_completed TINYINT(1) NOT NULL DEFAULT 0 COMMENT '资料是否完善：0-未完善，1-已完善',
		    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		    UNIQUE KEY uk_unique_id (unique_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='小程序用户表';
		`

		if _, err = DB.Exec(createMiniAppUsersTableSQL); err != nil {
			log.Printf("创建mini_app_users表失败: %v", err)
			return
		}

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

		// 将未完善资料的用户类型重置为未选择
		if _, err := DB.Exec(`
			UPDATE mini_app_users
			SET user_type = 'unknown'
			WHERE (user_type IS NULL OR user_type = '' OR user_type = 'retail')
			  AND (profile_completed IS NULL OR profile_completed = 0)
		`); err != nil {
			log.Printf("更新未完善资料的用户类型失败: %v", err)
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
