package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"go_backend/internal/config"
)

func main() {
	// 初始化配置
	config.InitConfig()

	// 构建DSN
	cfg := config.Config.Database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.Charset,
	)

	// 连接数据库
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	// 测试连接
	if err = db.Ping(); err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}

	// 添加title字段到carousels表
	addColumnSQL := `ALTER TABLE carousels ADD COLUMN title VARCHAR(255) DEFAULT '' COMMENT '轮播图标题' AFTER image;`
	_, err = db.Exec(addColumnSQL)
	if err != nil {
		log.Printf("添加字段失败（可能已存在）: %v", err)
	} else {
		fmt.Println("成功添加title字段到carousels表")
	}

	// 更新database.go中的创建表语句以包含title字段
	fmt.Println("请手动更新internal/model/carousel.go文件中的Carousel结构体，添加title字段")
}