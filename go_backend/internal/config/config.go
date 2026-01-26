package config

import "time"

// Config 应用配置
var Config = struct {
	Server struct {
		Port         int           `json:"port"`
		ReadTimeout  time.Duration `json:"read_timeout"`
		WriteTimeout time.Duration `json:"write_timeout"`
	} `json:"server"`
	CORS struct {
		AllowOrigins []string `json:"allow_origins"`
		AllowMethods []string `json:"allow_methods"`
		AllowHeaders []string `json:"allow_headers"`
	} `json:"cors"`
	Database struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
		DBName   string `json:"dbname"`
		Charset  string `json:"charset"`
	} `json:"database"`
	MinIO struct {
		Endpoint  string `json:"endpoint"`
		AccessKey string `json:"access_key"`
		SecretKey string `json:"secret_key"`
		Bucket    string `json:"bucket"`
		UseSSL    bool   `json:"use_ssl"`  // 是否使用 HTTPS 连接 MinIO 服务器
		BaseURL   string `json:"base_url"` // MinIO 文件访问的基础 URL（用于生成文件访问链接）
	} `json:"minio"`
	MiniApp struct {
		AppID     string `json:"app_id"`
		AppSecret string `json:"app_secret"`
	} `json:"mini_app"`
	Map struct {
		AmapKey    string `json:"amap_key"`    // 高德地图API Key
		TencentKey string `json:"tencent_key"` // 腾讯地图API Key
	} `json:"map"`
	WebSocket struct {
		EmployeeLocationURL string `json:"employee_location_url"` // 配送员位置上报WebSocket URL
		AdminLocationURL    string `json:"admin_location_url"`    // 管理后台位置查看WebSocket URL
	} `json:"websocket"`
}{}

// InitConfig 初始化配置
func InitConfig() {
	// 设置默认配置
	Config.Server.Port = 8082
	Config.Server.ReadTimeout = 60 * time.Second  // 读取超时60秒（用于上传大文件）
	Config.Server.WriteTimeout = 60 * time.Second // 写入超时60秒（用于上传大文件）
	Config.CORS.AllowOrigins = []string{"*"}
	Config.CORS.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	Config.CORS.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	// 设置MySQL数据库配置
	Config.Database.Host = "localhost"
	Config.Database.Port = 3306
	Config.Database.Username = "root"
	Config.Database.Password = "12580abc"
	Config.Database.DBName = "fx_shop"
	Config.Database.Charset = "utf8mb4"
	// 设置MinIO配置
	Config.MinIO.Endpoint = "106.55.167.44:19000" // MinIO 服务器地址（仅主机名和端口，不包含协议和路径）
	Config.MinIO.AccessKey = "minio"
	Config.MinIO.SecretKey = "3ZMb3HWRyyhyFRwH"
	Config.MinIO.Bucket = "fengxing"
	Config.MinIO.UseSSL = false                            // 是否使用 HTTPS 连接 MinIO 服务器（根据实际服务器配置设置）
	Config.MinIO.BaseURL = "https://mall.sscchh.com/minio" // MinIO 文件访问的基础 URL（用于生成文件访问链接）
	// 小程序配置（用于用户登录）
	Config.MiniApp.AppID = "wxa2535727aedb00cc"
	Config.MiniApp.AppSecret = "4e39a349d4eff820c3d4fa8f6441f3f0"
	// 地图API配置（用于地址解析）
	// 高德地图API Key（可选，如果配置了则优先使用高德）
	Config.Map.AmapKey = "91ab3706ba83aaacdb80aa9bbe0b5da5"
	// 腾讯地图API Key（可选，如果配置了则使用腾讯）
	Config.Map.TencentKey = ""
	// WebSocket配置
	// 配送员位置上报WebSocket URL（相对路径，会自动拼接服务器地址）
	Config.WebSocket.EmployeeLocationURL = "/api/mini/employee/location/ws"
	// 管理后台位置查看WebSocket URL（相对路径，会自动拼接服务器地址）
	Config.WebSocket.AdminLocationURL = "/api/mini/admin/employee-locations/ws"
}
