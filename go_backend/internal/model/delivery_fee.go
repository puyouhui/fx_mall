package model

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"go_backend/internal/database"
)

// DeliveryFeeSetting 配送费基础设置
type DeliveryFeeSetting struct {
	ID                    int       `json:"id"`
	BaseFee               float64   `json:"base_fee"`
	FreeShippingThreshold float64   `json:"free_shipping_threshold"`
	Description           string    `json:"description"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// DeliveryFeeExclusion 配送费排除项
type DeliveryFeeExclusion struct {
	ID                 int       `json:"id"`
	ItemType           string    `json:"item_type"`
	TargetID           int       `json:"target_id"`
	TargetName         string    `json:"target_name"`
	ParentCategoryName string    `json:"parent_category_name,omitempty"`
	MinQuantityForFree *int      `json:"min_quantity_for_free"`
	Remark             string    `json:"remark"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// DeliveryFeeBlockingItem 未满足免配送条件的提示项
type DeliveryFeeBlockingItem struct {
	ItemType         string `json:"item_type"`
	TargetID         int    `json:"target_id"`
	TargetName       string `json:"target_name"`
	RequiredQuantity *int   `json:"required_quantity,omitempty"`
	CurrentQuantity  int    `json:"current_quantity"`
	ProductID        *int   `json:"product_id,omitempty"`
	ProductName      string `json:"product_name,omitempty"`
	CategoryID       *int   `json:"category_id,omitempty"`
}

// DeliveryFeeSummary 配送费计算结果
type DeliveryFeeSummary struct {
	BaseFee               float64                   `json:"base_fee"`
	FreeShippingThreshold float64                   `json:"free_shipping_threshold"`
	EligibleAmount        float64                   `json:"eligible_amount"`
	IneligibleAmount      float64                   `json:"ineligible_amount"`
	TotalAmount           float64                   `json:"total_amount"`
	DeliveryFee           float64                   `json:"delivery_fee"`
	IsFreeShipping        bool                      `json:"is_free_shipping"`
	ShortOfAmount         float64                   `json:"short_of_amount"`
	Tips                  []DeliveryFeeBlockingItem `json:"tips"`
	EligibleQuantity      int                       `json:"eligible_quantity"`
	IneligibleQuantity    int                       `json:"ineligible_quantity"`
	TotalQuantity         int                       `json:"total_quantity"`
	BlockedItemIDs        []int                     `json:"blocked_item_ids"`
}

// GetDeliveryFeeSetting 获取当前配送费设置
func GetDeliveryFeeSetting() (*DeliveryFeeSetting, error) {
	query := "SELECT id, base_fee, free_shipping_threshold, description, created_at, updated_at FROM delivery_fee_settings ORDER BY id ASC LIMIT 1"
	row := database.DB.QueryRow(query)

	var setting DeliveryFeeSetting
	err := row.Scan(&setting.ID, &setting.BaseFee, &setting.FreeShippingThreshold, &setting.Description, &setting.CreatedAt, &setting.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &setting, nil
}

// UpsertDeliveryFeeSetting 更新或创建配送费设置
func UpsertDeliveryFeeSetting(setting *DeliveryFeeSetting) error {
	if setting == nil {
		return fmt.Errorf("配送费设置不能为空")
	}

	if setting.ID > 0 {
		query := "UPDATE delivery_fee_settings SET base_fee = ?, free_shipping_threshold = ?, description = ?, updated_at = NOW() WHERE id = ?"
		_, err := database.DB.Exec(query, setting.BaseFee, setting.FreeShippingThreshold, setting.Description, setting.ID)
		return err
	}

	query := "INSERT INTO delivery_fee_settings (base_fee, free_shipping_threshold, description, created_at, updated_at) VALUES (?, ?, ?, NOW(), NOW())"
	result, err := database.DB.Exec(query, setting.BaseFee, setting.FreeShippingThreshold, setting.Description)
	if err != nil {
		return err
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	setting.ID = int(lastID)
	return nil
}

// DeliveryFeeExclusionList 获取排除项列表
func DeliveryFeeExclusionList() ([]DeliveryFeeExclusion, error) {
	query := "SELECT id, item_type, target_id, min_quantity_for_free, remark, created_at, updated_at FROM delivery_fee_exclusions ORDER BY updated_at DESC"
	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exclusions []DeliveryFeeExclusion
	categoryIDs := make(map[int]struct{})
	productIDs := make(map[int]struct{})

	for rows.Next() {
		var exclusion DeliveryFeeExclusion
		var minQuantity sql.NullInt64
		if err := rows.Scan(&exclusion.ID, &exclusion.ItemType, &exclusion.TargetID, &minQuantity, &exclusion.Remark, &exclusion.CreatedAt, &exclusion.UpdatedAt); err != nil {
			return nil, err
		}

		if minQuantity.Valid {
			exclusion.MinQuantityForFree = intPointer(int(minQuantity.Int64))
		}

		if exclusion.ItemType == "category" {
			categoryIDs[exclusion.TargetID] = struct{}{}
		} else if exclusion.ItemType == "product" {
			productIDs[exclusion.TargetID] = struct{}{}
		}

		exclusions = append(exclusions, exclusion)
	}

	// 构建名称映射
	categoryNames, parentNames, err := fetchCategoryNames(categoryIDs)
	if err != nil {
		return nil, err
	}
	productNames, err := fetchProductNames(productIDs)
	if err != nil {
		return nil, err
	}

	for i := range exclusions {
		ex := &exclusions[i]
		if ex.ItemType == "category" {
			ex.TargetName = categoryNames[ex.TargetID]
			ex.ParentCategoryName = parentNames[ex.TargetID]
		} else if ex.ItemType == "product" {
			ex.TargetName = productNames[ex.TargetID]
		}
	}

	return exclusions, nil
}

// GetDeliveryFeeExclusionByScope 根据类型和目标ID获取排除项
func GetDeliveryFeeExclusionByScope(itemType string, targetID int) (*DeliveryFeeExclusion, error) {
	query := "SELECT id, item_type, target_id, min_quantity_for_free, remark, created_at, updated_at FROM delivery_fee_exclusions WHERE item_type = ? AND target_id = ? LIMIT 1"
	row := database.DB.QueryRow(query, itemType, targetID)

	var exclusion DeliveryFeeExclusion
	var minQuantity sql.NullInt64
	if err := row.Scan(&exclusion.ID, &exclusion.ItemType, &exclusion.TargetID, &minQuantity, &exclusion.Remark, &exclusion.CreatedAt, &exclusion.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if minQuantity.Valid {
		exclusion.MinQuantityForFree = intPointer(int(minQuantity.Int64))
	}

	return &exclusion, nil
}

// GetDeliveryFeeExclusionByID 根据ID获取排除项
func GetDeliveryFeeExclusionByID(id int) (*DeliveryFeeExclusion, error) {
	query := "SELECT id, item_type, target_id, min_quantity_for_free, remark, created_at, updated_at FROM delivery_fee_exclusions WHERE id = ? LIMIT 1"
	row := database.DB.QueryRow(query, id)

	var exclusion DeliveryFeeExclusion
	var minQuantity sql.NullInt64
	if err := row.Scan(&exclusion.ID, &exclusion.ItemType, &exclusion.TargetID, &minQuantity, &exclusion.Remark, &exclusion.CreatedAt, &exclusion.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if minQuantity.Valid {
		exclusion.MinQuantityForFree = intPointer(int(minQuantity.Int64))
	}

	return &exclusion, nil
}

// CreateDeliveryFeeExclusion 创建排除项
func CreateDeliveryFeeExclusion(exclusion *DeliveryFeeExclusion) error {
	if exclusion == nil {
		return fmt.Errorf("排除项不能为空")
	}

	query := "INSERT INTO delivery_fee_exclusions (item_type, target_id, min_quantity_for_free, remark, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW())"
	var minQuantity interface{}
	if exclusion.MinQuantityForFree != nil {
		minQuantity = *exclusion.MinQuantityForFree
	}

	result, err := database.DB.Exec(query, exclusion.ItemType, exclusion.TargetID, minQuantity, exclusion.Remark)
	if err != nil {
		return err
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	exclusion.ID = int(lastID)
	return nil
}

// UpdateDeliveryFeeExclusion 更新排除项
func UpdateDeliveryFeeExclusion(exclusion *DeliveryFeeExclusion) error {
	if exclusion == nil {
		return fmt.Errorf("排除项不能为空")
	}

	query := "UPDATE delivery_fee_exclusions SET min_quantity_for_free = ?, remark = ?, updated_at = NOW() WHERE id = ?"
	var minQuantity interface{}
	if exclusion.MinQuantityForFree != nil {
		minQuantity = *exclusion.MinQuantityForFree
	}

	_, err := database.DB.Exec(query, minQuantity, exclusion.Remark, exclusion.ID)
	return err
}

// DeleteDeliveryFeeExclusion 删除排除项
func DeleteDeliveryFeeExclusion(id int) error {
	query := "DELETE FROM delivery_fee_exclusions WHERE id = ?"
	_, err := database.DB.Exec(query, id)
	return err
}

// CalculateDeliveryFee 根据采购单计算配送费用
func CalculateDeliveryFee(items []PurchaseListItem) (*DeliveryFeeSummary, error) {
	summary := &DeliveryFeeSummary{
		Tips:           []DeliveryFeeBlockingItem{},
		BlockedItemIDs: []int{},
	}
	blockedItemMap := make(map[int]struct{})

	setting, err := GetDeliveryFeeSetting()
	if err != nil {
		return nil, err
	}
	if setting == nil {
		setting = &DeliveryFeeSetting{BaseFee: 0, FreeShippingThreshold: 0}
	}
	summary.BaseFee = setting.BaseFee
	summary.FreeShippingThreshold = setting.FreeShippingThreshold

	if len(items) == 0 {
		summary.DeliveryFee = summary.BaseFee
		if summary.BaseFee == 0 && summary.FreeShippingThreshold == 0 {
			summary.IsFreeShipping = true
		}
		return summary, nil
	}

	exclusions, err := DeliveryFeeExclusionList()
	if err != nil {
		return nil, err
	}

	productRules := make(map[int]*DeliveryFeeExclusion)
	categoryRules := make(map[int]*DeliveryFeeExclusion)
	for _, ex := range exclusions {
		exCopy := ex
		if ex.ItemType == "product" {
			productRules[ex.TargetID] = &exCopy
		} else {
			categoryRules[ex.TargetID] = &exCopy
		}
	}

	productIDs := uniqueProductIDs(items)
	categoryInfo, err := fetchProductCategoryInfo(productIDs)
	if err != nil {
		return nil, err
	}

	categoryQuantities := aggregateCategoryQuantities(items, categoryInfo)
	tipTracker := make(map[string]bool)

	for _, item := range items {
		info := categoryInfo[item.ProductID]
		summary.TotalQuantity += item.Quantity
		amount := calculateItemAmount(item)
		summary.TotalAmount += amount

		rule := pickDeliveryRule(item.ProductID, info, productRules, categoryRules)
		eligible, currentQty, requiredQty := evaluateRule(rule, item.Quantity, categoryQuantities)
		if eligible {
			summary.EligibleAmount += amount
			summary.EligibleQuantity += item.Quantity
		} else {
			summary.IneligibleAmount += amount
			summary.IneligibleQuantity += item.Quantity
			if _, exists := blockedItemMap[item.ID]; !exists {
				blockedItemMap[item.ID] = struct{}{}
				summary.BlockedItemIDs = append(summary.BlockedItemIDs, item.ID)
			}
			if rule != nil {
				key := fmt.Sprintf("%s-%d", rule.ItemType, rule.TargetID)
				if !tipTracker[key] {
					tipTracker[key] = true
					summary.Tips = append(summary.Tips, buildBlockingItem(rule, currentQty, requiredQty, item))
				}
			}
		}
	}

	summary.DeliveryFee = summary.BaseFee
	if summary.FreeShippingThreshold > 0 && summary.EligibleAmount >= summary.FreeShippingThreshold {
		summary.DeliveryFee = 0
		summary.IsFreeShipping = true
	} else if summary.FreeShippingThreshold == 0 && summary.BaseFee == 0 {
		summary.DeliveryFee = 0
		summary.IsFreeShipping = true
	} else {
		summary.IsFreeShipping = summary.DeliveryFee == 0
		if summary.FreeShippingThreshold > 0 {
			diff := summary.FreeShippingThreshold - summary.EligibleAmount
			if diff > 0 {
				summary.ShortOfAmount = diff
			}
		}
	}

	return summary, nil
}

// CalculateDeliveryFeeByUser 根据用户ID计算配送费
func CalculateDeliveryFeeByUser(userID int) (*DeliveryFeeSummary, error) {
	items, err := GetPurchaseListItemsByUserID(userID)
	if err != nil {
		return nil, err
	}
	return CalculateDeliveryFee(items)
}

func fetchCategoryNames(categoryIDs map[int]struct{}) (map[int]string, map[int]string, error) {
	names := make(map[int]string)
	parentNames := make(map[int]string)
	if len(categoryIDs) == 0 {
		return names, parentNames, nil
	}

	ids := make([]int, 0, len(categoryIDs))
	for id := range categoryIDs {
		ids = append(ids, id)
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf("SELECT c.id, c.name, c.parent_id, p.name FROM categories c LEFT JOIN categories p ON c.parent_id = p.id WHERE c.id IN (%s)", stringList(placeholders))
	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id, parentID int
		var name string
		var parentName sql.NullString
		if err := rows.Scan(&id, &name, &parentID, &parentName); err != nil {
			return nil, nil, err
		}
		names[id] = name
		if parentName.Valid {
			parentNames[id] = parentName.String
		}
	}

	return names, parentNames, nil
}

func fetchProductNames(productIDs map[int]struct{}) (map[int]string, error) {
	names := make(map[int]string)
	if len(productIDs) == 0 {
		return names, nil
	}

	ids := make([]int, 0, len(productIDs))
	for id := range productIDs {
		ids = append(ids, id)
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf("SELECT id, name FROM products WHERE id IN (%s)", stringList(placeholders))
	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		names[id] = name
	}

	return names, nil
}

func stringList(items []string) string {
	return strings.Join(items, ",")
}

func intPointer(v int) *int {
	value := v
	return &value
}

// ProductCategoryInfo 商品分类信息
type ProductCategoryInfo struct {
	CategoryID int
	ParentID   int
}

type productCategoryInfo struct {
	CategoryID int
	ParentID   int
}

func uniqueProductIDs(items []PurchaseListItem) []int {
	ids := make([]int, 0)
	seen := make(map[int]struct{})
	for _, item := range items {
		if _, ok := seen[item.ProductID]; !ok {
			seen[item.ProductID] = struct{}{}
			ids = append(ids, item.ProductID)
		}
	}
	return ids
}

// FetchProductCategoryInfo 获取商品分类信息（公开函数）
func FetchProductCategoryInfo(productIDs []int) (map[int]ProductCategoryInfo, error) {
	infoMap, err := fetchProductCategoryInfo(productIDs)
	if err != nil {
		return nil, err
	}
	result := make(map[int]ProductCategoryInfo)
	for k, v := range infoMap {
		result[k] = ProductCategoryInfo{
			CategoryID: v.CategoryID,
			ParentID:   v.ParentID,
		}
	}
	return result, nil
}

func fetchProductCategoryInfo(productIDs []int) (map[int]productCategoryInfo, error) {
	result := make(map[int]productCategoryInfo)
	if len(productIDs) == 0 {
		return result, nil
	}

	placeholders := make([]string, len(productIDs))
	args := make([]interface{}, len(productIDs))
	for i, id := range productIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT p.id, p.category_id, COALESCE(c.parent_id, 0) AS parent_id
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.id IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var productID, categoryID, parentID int
		if err := rows.Scan(&productID, &categoryID, &parentID); err != nil {
			return nil, err
		}
		result[productID] = productCategoryInfo{
			CategoryID: categoryID,
			ParentID:   parentID,
		}
	}

	return result, nil
}

func aggregateCategoryQuantities(items []PurchaseListItem, infoMap map[int]productCategoryInfo) map[int]int {
	result := make(map[int]int)
	for _, item := range items {
		info := infoMap[item.ProductID]
		if info.CategoryID > 0 {
			result[info.CategoryID] += item.Quantity
		}
		if info.ParentID > 0 {
			result[info.ParentID] += item.Quantity
		}
	}
	return result
}

func calculateItemAmount(item PurchaseListItem) float64 {
	price := item.SpecSnapshot.WholesalePrice
	if price <= 0 {
		price = item.SpecSnapshot.RetailPrice
	}
	if price <= 0 {
		price = item.SpecSnapshot.Cost
	}
	if price < 0 {
		price = 0
	}
	return price * float64(item.Quantity)
}

func pickDeliveryRule(productID int, info productCategoryInfo, productRules, categoryRules map[int]*DeliveryFeeExclusion) *DeliveryFeeExclusion {
	if rule, ok := productRules[productID]; ok {
		return rule
	}
	if rule, ok := categoryRules[info.CategoryID]; ok {
		return rule
	}
	if info.ParentID > 0 {
		if rule, ok := categoryRules[info.ParentID]; ok {
			return rule
		}
	}
	return nil
}

func evaluateRule(rule *DeliveryFeeExclusion, itemQuantity int, categoryQuantities map[int]int) (bool, int, *int) {
	if rule == nil {
		return true, itemQuantity, nil
	}

	var current int
	if rule.ItemType == "product" {
		current = itemQuantity
	} else {
		current = categoryQuantities[rule.TargetID]
	}

	if rule.MinQuantityForFree == nil {
		return false, current, nil
	}

	required := *rule.MinQuantityForFree
	if current >= required {
		return true, current, nil
	}

	return false, current, intPointer(required)
}

func buildBlockingItem(rule *DeliveryFeeExclusion, currentQty int, requiredQty *int, item PurchaseListItem) DeliveryFeeBlockingItem {
	tip := DeliveryFeeBlockingItem{
		ItemType:        rule.ItemType,
		TargetID:        rule.TargetID,
		TargetName:      rule.TargetName,
		CurrentQuantity: currentQty,
	}
	if requiredQty != nil {
		tip.RequiredQuantity = requiredQty
	}

	if rule.ItemType == "product" {
		tip.ProductID = intPointer(item.ProductID)
		tip.ProductName = item.ProductName
	} else {
		tip.CategoryID = intPointer(rule.TargetID)
	}

	return tip
}
