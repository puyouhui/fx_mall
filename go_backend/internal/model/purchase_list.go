package model

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go_backend/internal/database"
)

// PurchaseSpecSnapshot 保存加入采购单时的规格快照
type PurchaseSpecSnapshot struct {
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	Cost           float64 `json:"cost"`
	WholesalePrice float64 `json:"wholesale_price"`
	RetailPrice    float64 `json:"retail_price"`
	DeliveryCount  float64 `json:"delivery_count"` // 配送计件数（默认1.0，用于计算件数补贴）
}

// PurchaseListItem 采购单中的商品
type PurchaseListItem struct {
	ID           int                  `json:"id"`
	UserID       int                  `json:"user_id"`
	ProductID    int                  `json:"product_id"`
	ProductName  string               `json:"product_name"`
	ProductImage string               `json:"product_image"`
	SpecName     string               `json:"spec_name"`
	SpecSnapshot PurchaseSpecSnapshot `json:"spec_snapshot"`
	Quantity     int                  `json:"quantity"`
	IsSpecial    bool                 `json:"is_special"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
}

// AddOrUpdatePurchaseListItem 新增或更新采购单项（如果存在则累加数量）
func AddOrUpdatePurchaseListItem(item *PurchaseListItem) (*PurchaseListItem, error) {
	if item.Quantity <= 0 {
		return nil, fmt.Errorf("数量必须大于0")
	}

	specBytes, err := json.Marshal(item.SpecSnapshot)
	if err != nil {
		return nil, err
	}
	isSpecial := 0
	if item.IsSpecial {
		isSpecial = 1
	}

	query := `
		INSERT INTO purchase_list_items (user_id, product_id, product_name, product_image, spec_name, spec_snapshot, quantity, is_special, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE
			product_name = VALUES(product_name),
			product_image = VALUES(product_image),
			spec_snapshot = VALUES(spec_snapshot),
			is_special = VALUES(is_special),
			quantity = quantity + VALUES(quantity),
			updated_at = NOW()
	`

	_, err = database.DB.Exec(query,
		item.UserID,
		item.ProductID,
		item.ProductName,
		item.ProductImage,
		item.SpecName,
		string(specBytes),
		item.Quantity,
		isSpecial,
	)
	if err != nil {
		return nil, err
	}

	return GetPurchaseListItemByKey(item.UserID, item.ProductID, item.SpecName)
}

// GetPurchaseListItemByKey 根据唯一键获取采购单项
func GetPurchaseListItemByKey(userID, productID int, specName string) (*PurchaseListItem, error) {
	query := `
		SELECT id, user_id, product_id, product_name, product_image, spec_name, spec_snapshot, quantity, is_special, created_at, updated_at
		FROM purchase_list_items
		WHERE user_id = ? AND product_id = ? AND spec_name = ?
		LIMIT 1
	`

	row := database.DB.QueryRow(query, userID, productID, specName)
	return scanPurchaseListItem(row)
}

// GetPurchaseListItemsByUserID 获取用户的采购单
func GetPurchaseListItemsByUserID(userID int) ([]PurchaseListItem, error) {
	query := `
		SELECT id, user_id, product_id, product_name, product_image, spec_name, spec_snapshot, quantity, is_special, created_at, updated_at
		FROM purchase_list_items
		WHERE user_id = ?
		ORDER BY created_at ASC, id ASC
	`

	rows, err := database.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]PurchaseListItem, 0)
	for rows.Next() {
		item, err := scanPurchaseListItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}

	return items, nil
}

// BackupPurchaseList 备份用户的采购单（返回采购单的副本）
func BackupPurchaseList(userID int) ([]PurchaseListItem, error) {
	return GetPurchaseListItemsByUserID(userID)
}

// RestorePurchaseList 恢复用户的采购单（从备份恢复）
func RestorePurchaseList(userID int, backup []PurchaseListItem) error {
	// 先清空当前采购单
	_, err := database.DB.Exec("DELETE FROM purchase_list_items WHERE user_id = ?", userID)
	if err != nil {
		return fmt.Errorf("清空采购单失败: %w", err)
	}

	// 恢复备份的采购单
	for _, item := range backup {
		specBytes, err := json.Marshal(item.SpecSnapshot)
		if err != nil {
			log.Printf("[RestorePurchaseList] 序列化规格快照失败: 商品ID=%d, 错误=%v", item.ProductID, err)
			continue
		}
		isSpecial := 0
		if item.IsSpecial {
			isSpecial = 1
		}

		_, err = database.DB.Exec(`
			INSERT INTO purchase_list_items (user_id, product_id, product_name, product_image, spec_name, spec_snapshot, quantity, is_special, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
		`, item.UserID, item.ProductID, item.ProductName, item.ProductImage, item.SpecName, string(specBytes), item.Quantity, isSpecial)
		if err != nil {
			log.Printf("[RestorePurchaseList] 恢复采购单项失败: 商品ID=%d, 错误=%v", item.ProductID, err)
			continue
		}
	}

	return nil
}

// UpdatePurchaseListItemQuantity 更新采购单项数量
func UpdatePurchaseListItemQuantity(itemID, userID, quantity int) error {
	if quantity <= 0 {
		return fmt.Errorf("数量必须大于0")
	}

	result, err := database.DB.Exec(`
		UPDATE purchase_list_items
		SET quantity = ?, updated_at = NOW()
		WHERE id = ? AND user_id = ?
	`, quantity, itemID, userID)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("采购单项不存在")
	}
	return nil
}

// DeletePurchaseListItem 删除单个采购单项
func DeletePurchaseListItem(itemID, userID int) error {
	result, err := database.DB.Exec(`
		DELETE FROM purchase_list_items
		WHERE id = ? AND user_id = ?
	`, itemID, userID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("采购单项不存在")
	}
	return nil
}

// ClearPurchaseList 清空用户采购单
func ClearPurchaseList(userID int) error {
	_, err := database.DB.Exec(`
		DELETE FROM purchase_list_items WHERE user_id = ?
	`, userID)
	return err
}

func scanPurchaseListItem(scanner interface {
	Scan(dest ...interface{}) error
}) (*PurchaseListItem, error) {
	var (
		item            PurchaseListItem
		specSnapshotStr string
		isSpecialInt    int
	)

	err := scanner.Scan(
		&item.ID,
		&item.UserID,
		&item.ProductID,
		&item.ProductName,
		&item.ProductImage,
		&item.SpecName,
		&specSnapshotStr,
		&item.Quantity,
		&isSpecialInt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	item.IsSpecial = isSpecialInt == 1
	if specSnapshotStr != "" {
		_ = json.Unmarshal([]byte(specSnapshotStr), &item.SpecSnapshot)
		// 修复旧数据：如果规格快照没有 delivery_count 或为0，设置为默认值1.0
		if item.SpecSnapshot.DeliveryCount <= 0 {
			item.SpecSnapshot.DeliveryCount = 1.0
		}
	}

	return &item, nil
}
