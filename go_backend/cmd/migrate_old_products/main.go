package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"go_backend/internal/config"
	"go_backend/internal/database"
	"go_backend/internal/model"
)

// oldProductRow 用于扫描 products 表的最小字段集合
type oldProductRow struct {
	ID            int
	SpecsJSON     sql.NullString
	UomCategoryID sql.NullInt64
}

func main() {
	log.Println("=========================================")
	log.Println("老商品规格迁移工具：绑定默认「件」单位类别和规格单位")
	log.Println("=========================================")

	// 初始化配置
	config.InitConfig()
	log.Println("配置初始化完成")

	// 初始化数据库
	if err := database.InitDB(); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer database.CloseDB()
	log.Println("数据库连接成功")

	// 1. 获取默认「件」类别 ID 和默认「件」基准单位 ID
	defaultCatID, err := model.GetDefaultUomCategoryID()
	if err != nil {
		log.Fatalf("获取默认「件」单位类别失败: %v\n", err)
	}
	defaultUnitID, err := model.GetDefaultUomUnitID()
	if err != nil {
		log.Fatalf("获取默认「件」单位失败: %v\n", err)
	}

	log.Printf("默认「件」类别ID: %d, 默认「件」单位ID: %d\n", defaultCatID, defaultUnitID)

	// 2. 查询所有商品（只取需要的字段）
	rows, err := database.DB.Query(`SELECT id, specs, uom_category_id FROM products`)
	if err != nil {
		log.Fatalf("查询商品失败: %v\n", err)
	}
	defer rows.Close()

	var (
		totalCount   int
		updatedCount int
	)

	for rows.Next() {
		var row oldProductRow
		if err := rows.Scan(&row.ID, &row.SpecsJSON, &row.UomCategoryID); err != nil {
			log.Printf("扫描商品(id=%d)失败: %v\n", row.ID, err)
			continue
		}
		totalCount++

		// 解析规格
		var specs []model.Spec
		if row.SpecsJSON.Valid && row.SpecsJSON.String != "" {
			if err := json.Unmarshal([]byte(row.SpecsJSON.String), &specs); err != nil {
				log.Printf("解析商品(id=%d)规格JSON失败，跳过该商品: %v\n", row.ID, err)
				continue
			}
		}

		// 如果没有规格，就不强制创建规格，但仍然可以给商品补 uom_category_id
		//（你的系统里创建/更新已经强制至少一个规格，新老数据兼容这里宽松处理）

		// 标记是否有修改
		changed := false

		// 2.1 如果商品没有绑定单位类别，补上默认「件」类别
		if !row.UomCategoryID.Valid || row.UomCategoryID.Int64 == 0 {
			row.UomCategoryID.Int64 = int64(defaultCatID)
			row.UomCategoryID.Valid = true
			changed = true
		}

		// 2.2 遍历规格，补充 delivery_count 和 uom_unit_id
		for i := range specs {
			// 配送计件数 <= 0 时设为 1.0
			if specs[i].DeliveryCount <= 0 {
				specs[i].DeliveryCount = 1.0
				changed = true
			}
			// 如果没有绑定单位，绑定到默认「件」单位
			if specs[i].UomUnitID == nil || *specs[i].UomUnitID == 0 {
				uid := defaultUnitID
				specs[i].UomUnitID = &uid
				changed = true
			}
		}

		if !changed {
			continue
		}

		// 重新序列化规格
		newSpecsJSON, err := json.Marshal(specs)
		if err != nil {
			log.Printf("序列化商品(id=%d)规格失败: %v\n", row.ID, err)
			continue
		}

		// 执行更新
		_, err = database.DB.Exec(
			`UPDATE products SET specs = ?, uom_category_id = ? WHERE id = ?`,
			string(newSpecsJSON),
			row.UomCategoryID.Int64,
			row.ID,
		)
		if err != nil {
			log.Printf("更新商品(id=%d)失败: %v\n", row.ID, err)
			continue
		}

		updatedCount++
		if updatedCount%100 == 0 {
			log.Printf("已更新 %d 个商品...\n", updatedCount)
		}
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("遍历商品结果集出错: %v\n", err)
	}

	log.Printf("迁移完成：总商品数=%d，已更新=%d\n", totalCount, updatedCount)
	fmt.Println("迁移结束。请检查部分商品数据确认无误。")
}

