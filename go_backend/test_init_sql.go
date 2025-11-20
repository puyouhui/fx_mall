package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 远程数据库配置
	username := "root"
	password := "hn02le.34lkdLKD"
	host := "113.44.164.151"
	port := 3306
	dbName := "product_shop_test"
	charset := "utf8mb4"

	// 构建DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=%s&parseTime=True&loc=Local",
		username,
		password,
		host,
		port,
		charset,
	)

	// 连接数据库（不指定具体数据库）
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	// 测试连接
	if err = db.Ping(); err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}

	fmt.Println("成功连接到远程数据库服务器")

	// 删除测试数据库（如果存在）
	dropDB := fmt.Sprintf("DROP DATABASE IF EXISTS %s;", dbName)
	_, err = db.Exec(dropDB)
	if err != nil {
		log.Fatalf("删除测试数据库失败: %v", err)
	}
	fmt.Printf("已删除测试数据库 %s\n", dbName)

	// 创建测试数据库
	createDB := fmt.Sprintf("CREATE DATABASE %s DEFAULT CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci;", dbName)
	_, err = db.Exec(createDB)
	if err != nil {
		log.Fatalf("创建测试数据库失败: %v", err)
	}
	fmt.Printf("已创建测试数据库 %s\n", dbName)

	// 选择测试数据库
	useDB := fmt.Sprintf("USE %s;", dbName)
	_, err = db.Exec(useDB)
	if err != nil {
		log.Fatalf("选择测试数据库失败: %v", err)
	}

	// 读取并执行init.sql文件
	sqlFile, err := os.Open("init.sql")
	if err != nil {
		log.Fatalf("打开init.sql文件失败: %v", err)
	}
	defer sqlFile.Close()

	fmt.Println("开始执行init.sql文件...")

	// 逐行读取SQL文件并执行
	scanner := bufio.NewScanner(sqlFile)
	var sqlStmt strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		// 跳过注释行和空行
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "--") {
			continue
		}

		// 添加当前行到SQL语句构建器
		sqlStmt.WriteString(line)
		sqlStmt.WriteString(" ")

		// 如果当前行以分号结尾，则执行SQL语句
		if strings.HasSuffix(strings.TrimSpace(line), ";") {
			stmt := sqlStmt.String()
			sqlStmt.Reset()

			// 跳过USE语句，因为我们已经选择了数据库
			if strings.HasPrefix(strings.ToUpper(stmt), "USE ") {
				continue
			}

			// 跳过CREATE DATABASE语句，因为我们已经创建了数据库
			if strings.HasPrefix(strings.ToUpper(stmt), "CREATE DATABASE") {
				continue
			}

			// 执行SQL语句
			_, err = db.Exec(stmt)
			if err != nil {
				log.Fatalf("执行SQL语句失败: %v\nSQL: %s", err, stmt)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("读取init.sql文件失败: %v", err)
	}

	fmt.Println("init.sql文件执行成功")

	// 验证表结构和数据
	validateTables(db)

	fmt.Println("所有表结构和数据验证通过！init.sql文件已完善。")
}

// 验证表结构和数据
func validateTables(db *sql.DB) {
	tables := []string{"admins", "categories", "products", "carousels"}

	for _, table := range tables {
		// 检查表是否存在
		var exists bool
		err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?)", table).Scan(&exists)
		if err != nil {
			log.Fatalf("检查表 %s 是否存在失败: %v", table, err)
		}
		if !exists {
			log.Fatalf("表 %s 不存在", table)
		}

		// 检查表中是否有数据
		var count int
		err = db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
		if err != nil {
			log.Fatalf("查询表 %s 数据数量失败: %v", table, err)
		}
		fmt.Printf("表 %s 存在，包含 %d 条数据\n", table, count)

		// 特殊检查：categories表是否有icon字段
		if table == "categories" {
			var hasIcon bool
			err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'categories' AND column_name = 'icon')").Scan(&hasIcon)
			if err != nil {
				log.Fatalf("检查categories表icon字段失败: %v", err)
			}
			if !hasIcon {
				log.Fatalf("categories表缺少icon字段")
			}
			fmt.Println("categories表包含icon字段")
		}
	}
}