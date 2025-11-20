package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 远程数据库配置
	username := "root"
	password := "hn02le.34lkdLKD"
	host := "113.44.164.151"
	port := 3306
	dbName := "product_shop"
	charset := "utf8mb4"

	// 构建DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		username,
		password,
		host,
		port,
		dbName,
		charset,
	)

	// 连接远程数据库
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	// 测试连接
	if err = db.Ping(); err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}

	fmt.Println("成功连接到远程数据库")

	// 检查categories表是否已经有icon字段
	var count int
	checkColumnSQL := `
	SELECT COUNT(*)
	FROM information_schema.columns
	WHERE table_schema = ? AND table_name = 'categories' AND column_name = 'icon';
	`
	err = db.QueryRow(checkColumnSQL, dbName).Scan(&count)
	if err != nil {
		log.Printf("检查字段失败: %v", err)
	} else if count > 0 {
		fmt.Println("icon字段已经存在于categories表中")
		return
	}

	// 向categories表添加icon字段
	addColumnSQL := `
	ALTER TABLE categories
	ADD COLUMN icon VARCHAR(255) NULL COMMENT '分类图标URL' AFTER updated_at;
	`
	_, err = db.Exec(addColumnSQL)
	if err != nil {
		log.Fatalf("添加icon字段失败: %v", err)
	}

	fmt.Println("成功向categories表添加icon字段")
}