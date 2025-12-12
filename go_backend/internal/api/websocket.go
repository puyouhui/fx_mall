package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"go_backend/internal/model"
	"go_backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// 允许所有来源，生产环境应该限制
		return true
	},
}

// EmployeeLocation 员工位置信息
type EmployeeLocation struct {
	EmployeeID   int       `json:"employee_id"`
	EmployeeCode string    `json:"employee_code"`
	Name         string    `json:"name"`
	Phone        string    `json:"phone"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	Accuracy     float64   `json:"accuracy"` // 精度（米）
	UpdatedAt    time.Time `json:"updated_at"`
}

// LocationManager 位置管理器
type LocationManager struct {
	locations map[int]*EmployeeLocation // key: employee_id
	clients   map[*websocket.Conn]bool  // 管理后台WebSocket客户端
	mu        sync.RWMutex
}

var locationManager = &LocationManager{
	locations: make(map[int]*EmployeeLocation),
	clients:   make(map[*websocket.Conn]bool),
}

// UpdateLocation 更新员工位置（配送员端调用）
func (lm *LocationManager) UpdateLocation(employeeID int, employeeCode, name, phone string, latitude, longitude, accuracy float64) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.locations[employeeID] = &EmployeeLocation{
		EmployeeID:   employeeID,
		EmployeeCode: employeeCode,
		Name:         name,
		Phone:        phone,
		Latitude:     latitude,
		Longitude:    longitude,
		Accuracy:     accuracy,
		UpdatedAt:    time.Now(),
	}

	// 异步保存到数据库（不阻塞实时更新）
	go func() {
		if err := model.SaveEmployeeLocation(employeeID, employeeCode, latitude, longitude, accuracy); err != nil {
			log.Printf("保存配送员位置到数据库失败 (员工ID: %d): %v", employeeID, err)
		}
	}()

	// 广播给所有管理后台客户端
	lm.broadcastLocation(lm.locations[employeeID])
}

// broadcastLocation 广播位置信息给所有管理后台客户端
func (lm *LocationManager) broadcastLocation(location *EmployeeLocation) {
	message, err := json.Marshal(map[string]interface{}{
		"type":     "location_update",
		"location": location,
	})
	if err != nil {
		log.Printf("序列化位置信息失败: %v", err)
		return
	}

	for client := range lm.clients {
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("发送位置信息失败: %v", err)
			delete(lm.clients, client)
			client.Close()
		}
	}
}

// GetAllLocations 获取所有员工位置（包括离线员工的最后位置）
func (lm *LocationManager) GetAllLocations() []*EmployeeLocation {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	locations := make([]*EmployeeLocation, 0, len(lm.locations))
	for _, loc := range lm.locations {
		// 返回所有位置（包括超过5分钟的离线位置）
		locations = append(locations, loc)
	}
	return locations
}

// GetLocationByEmployeeID 根据员工ID获取位置
func (lm *LocationManager) GetLocationByEmployeeID(employeeID int) *EmployeeLocation {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	loc, exists := lm.locations[employeeID]
	if !exists {
		return nil
	}

	// 如果位置超过5分钟，返回nil
	if time.Since(loc.UpdatedAt) > 5*time.Minute {
		return nil
	}

	return loc
}

// GetLocationByEmployeeCode 根据员工码获取位置
func (lm *LocationManager) GetLocationByEmployeeCode(employeeCode string) *EmployeeLocation {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	for _, loc := range lm.locations {
		if loc.EmployeeCode == employeeCode {
			// 如果位置超过5分钟，返回nil
			if time.Since(loc.UpdatedAt) > 5*time.Minute {
				return nil
			}
			return loc
		}
	}
	return nil
}

// AddClient 添加管理后台WebSocket客户端
func (lm *LocationManager) AddClient(conn *websocket.Conn) {
	lm.mu.Lock()
	lm.clients[conn] = true
	lm.mu.Unlock()

	// 发送当前所有位置
	locations := lm.GetAllLocations()
	message, err := json.Marshal(map[string]interface{}{
		"type":      "initial_locations",
		"locations": locations,
	})
	if err == nil {
		conn.WriteMessage(websocket.TextMessage, message)
	}
}

// RemoveClient 移除管理后台WebSocket客户端
func (lm *LocationManager) RemoveClient(conn *websocket.Conn) {
	lm.mu.Lock()
	delete(lm.clients, conn)
	lm.mu.Unlock()
}

// HandleEmployeeWebSocket 处理配送员端的WebSocket连接
func HandleEmployeeWebSocket(c *gin.Context) {
	// 从URL参数或请求头中获取token（优先使用URL参数，因为WebSocket可能不支持自定义请求头）
	token := c.Query("token")
	if token == "" {
		token = c.GetHeader("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
	}

	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 使用ParseEmployeeToken解析员工token
	claims, err := utils.ParseEmployeeToken(token)
	if err != nil {
		log.Printf("解析员工token失败: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的token"})
		return
	}

	// Token验证成功，从数据库获取员工信息
	employee, err := model.GetEmployeeByID(claims.EmployeeID)
	if err != nil || employee == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "员工不存在"})
		return
	}

	// 验证员工状态
	if !employee.Status {
		c.JSON(http.StatusForbidden, gin.H{"error": "员工账号已被禁用"})
		return
	}

	empID := employee.ID

	// 升级为WebSocket连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("配送员 %s (ID: %d) 已连接WebSocket", employee.Name, empID)

	// 设置读取超时和心跳处理
	const pongWait = 60 * time.Second
	const pingPeriod = 30 * time.Second

	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// 使用互斥锁保护WebSocket写入操作（WebSocket连接不是并发安全的）
	var writeMu sync.Mutex

	// 安全的写入函数
	writeMessage := func(messageType int, data []byte) error {
		writeMu.Lock()
		defer writeMu.Unlock()
		conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		return conn.WriteMessage(messageType, data)
	}

	writeJSON := func(v interface{}) error {
		writeMu.Lock()
		defer writeMu.Unlock()
		conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		return conn.WriteJSON(v)
	}

	// 启动心跳发送goroutine
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	done := make(chan struct{})
	var once sync.Once
	closeDone := func() {
		once.Do(func() {
			close(done)
		})
	}

	go func() {
		defer closeDone()
		for {
			select {
			case <-ticker.C:
				if err := writeMessage(websocket.PingMessage, nil); err != nil {
					log.Printf("发送心跳失败 (配送员 %d): %v", empID, err)
					return
				}
			case <-done:
				return
			}
		}
	}()

	// 处理消息
	for {
		// 检查是否应该退出
		select {
		case <-done:
			log.Printf("配送员 %s (ID: %d) 心跳goroutine退出", employee.Name, empID)
			return
		default:
		}

		// 设置读取超时
		conn.SetReadDeadline(time.Now().Add(pongWait))

		// 读取消息类型
		messageType, reader, err := conn.NextReader()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket读取错误 (配送员 %d): %v", empID, err)
			}
			break
		}

		// 处理ping/pong消息
		if messageType == websocket.PingMessage {
			if err := writeMessage(websocket.PongMessage, nil); err != nil {
				log.Printf("发送pong失败 (配送员 %d): %v", empID, err)
				break
			}
			continue
		}

		// 处理文本消息（JSON）
		if messageType == websocket.TextMessage {
			var msg map[string]interface{}
			decoder := json.NewDecoder(reader)
			if err := decoder.Decode(&msg); err != nil {
				log.Printf("解析WebSocket消息失败 (配送员 %d): %v", empID, err)
				continue
			}

			// 处理位置上报
			if msgType, ok := msg["type"].(string); ok && msgType == "location" {
				latitude, _ := msg["latitude"].(float64)
				longitude, _ := msg["longitude"].(float64)
				accuracy, _ := msg["accuracy"].(float64)

				log.Printf("收到配送员 %s (ID: %d) 位置上报: %.6f, %.6f, 精度: %.2f米",
					employee.Name, empID, latitude, longitude, accuracy)

				// 更新位置
				locationManager.UpdateLocation(
					empID,
					employee.EmployeeCode,
					employee.Name,
					employee.Phone,
					latitude,
					longitude,
					accuracy,
				)

				// 回复确认
				if err := writeJSON(map[string]interface{}{
					"type":    "location_received",
					"success": true,
				}); err != nil {
					log.Printf("发送位置确认失败 (配送员 %d): %v", empID, err)
					break
				}
			} else if msgType == "ping" {
				// 处理前端发送的ping消息
				if err := writeJSON(map[string]interface{}{
					"type": "pong",
				}); err != nil {
					log.Printf("发送pong失败 (配送员 %d): %v", empID, err)
					break
				}
			} else {
				log.Printf("收到未知类型的消息 (配送员 %d): %v", empID, msg)
			}
		}
	}

	// 停止心跳（使用安全的关闭函数，避免重复关闭）
	closeDone()
	log.Printf("配送员 %s (ID: %d) 已断开WebSocket连接", employee.Name, empID)
}

// HandleAdminWebSocket 处理管理后台的WebSocket连接
func HandleAdminWebSocket(c *gin.Context) {
	// 从URL参数或请求头中获取token
	token := c.Query("token")
	if token == "" {
		token = c.GetHeader("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
	}

	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 验证token
	claims, err := utils.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的token"})
		return
	}

	// 验证通过，将管理员信息存入上下文
	c.Set("adminID", claims.UserID)
	c.Set("username", claims.Username)

	// 升级为WebSocket连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("管理后台WebSocket客户端已连接 (管理员ID: %d)", claims.UserID)

	// 添加到客户端列表
	locationManager.AddClient(conn)
	defer locationManager.RemoveClient(conn)

	// 设置读取超时和Pong处理器
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// 启动读取goroutine（用于接收Pong响应和处理关闭）
	done := make(chan struct{})
	var adminOnce sync.Once
	closeAdminDone := func() {
		adminOnce.Do(func() {
			close(done)
		})
	}

	go func() {
		defer closeAdminDone()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket读取错误: %v", err)
				}
				return
			}
		}
	}()

	// 保持连接，定期发送心跳
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			log.Printf("管理后台WebSocket客户端已断开 (管理员ID: %d)", claims.UserID)
			return
		case <-ticker.C:
			// 发送心跳
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("发送心跳失败: %v", err)
				return
			}
		}
	}
}

// GetEmployeeLocations 获取所有员工位置（HTTP API，供管理后台使用）
func GetEmployeeLocations(c *gin.Context) {
	locations := locationManager.GetAllLocations()
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    locations,
	})
}

// GetEmployeeLocation 获取指定员工位置
func GetEmployeeLocation(c *gin.Context) {
	employeeID := c.Param("id")
	var empID int
	if _, err := fmt.Sscanf(employeeID, "%d", &empID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的员工ID"})
		return
	}

	location := locationManager.GetLocationByEmployeeID(empID)
	if location == nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "该员工暂无位置信息",
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    location,
	})
}

// GetEmployeeLocationByCode 根据员工码获取配送员位置（小程序端调用）
// 优先返回实时位置，如果没有则返回数据库中的最新位置
func GetEmployeeLocationByCode(c *gin.Context) {
	employeeCode := c.Param("code")
	if employeeCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "员工码不能为空",
		})
		return
	}

	// 先尝试从内存中获取实时位置
	location := locationManager.GetLocationByEmployeeCode(employeeCode)

	// 如果内存中没有，从数据库获取最新位置
	if location == nil {
		historyLocation, err := model.GetLatestEmployeeLocationByCode(employeeCode)
		if err != nil {
			log.Printf("从数据库获取配送员位置失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取配送员位置失败",
			})
			return
		}

		if historyLocation == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "未找到该配送员的位置信息",
			})
			return
		}

		// 转换为API响应格式
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "获取成功",
			"data": gin.H{
				"employee_code": historyLocation.EmployeeCode,
				"latitude":      historyLocation.Latitude,
				"longitude":     historyLocation.Longitude,
				"accuracy":      historyLocation.Accuracy,
				"updated_at":    historyLocation.CreatedAt,
				"is_realtime":   false, // 标记为历史位置
			},
		})
		return
	}

	// 返回实时位置
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"employee_code": location.EmployeeCode,
			"latitude":      location.Latitude,
			"longitude":     location.Longitude,
			"accuracy":      location.Accuracy,
			"updated_at":    location.UpdatedAt,
			"is_realtime":   true, // 标记为实时位置
		},
	})
}
