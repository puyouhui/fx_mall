package main

import (
	"fmt"
	"log"

	"go_backend/internal/database"
	"go_backend/internal/config"
)

func main() {
	// 初始化配置
	config.InitConfig()

	// 初始化数据库连接
	if err := database.InitDB(); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 添加original_price字段
	addColumnSQL := `
	ALTER TABLE products
	ADD COLUMN original_price DECIMAL(10,2) NOT NULL COMMENT '原价' AFTER description;
	`

	_, err := database.DB.Exec(addColumnSQL)
	if err != nil {
		log.Printf("添加original_price字段失败: %v", err)
		// 检查字段是否已存在
		var columnExists int
		checkColumnSQL := `
		SELECT COUNT(*) 
		FROM information_schema.columns 
		WHERE table_schema = ? AND table_name = 'products' AND column_name = 'original_price';
		`
		err = database.DB.QueryRow(checkColumnSQL, config.Config.Database.DBName).Scan(&columnExists)
		if err != nil {
			log.Fatalf("检查字段是否存在失败: %v", err)
		}

		if columnExists > 0 {
			fmt.Println("original_price字段已存在，跳过添加")
		} else {
			log.Fatalf("添加original_price字段失败，且字段不存在")
		}
	} else {
		fmt.Println("成功添加original_price字段到products表")

		// 更新现有数据，将price值复制到original_price
		updateExistingDataSQL := `
		UPDATE products
		SET original_price = price;
		`

		_, err = database.DB.Exec(updateExistingDataSQL)
		if err != nil {
			log.Printf("更新现有数据失败: %v", err)
		} else {
			fmt.Println("成功更新现有数据的original_price值")
		}
	}
}