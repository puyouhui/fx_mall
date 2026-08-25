package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

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
		Endpoint   string `json:"endpoint"`
		AccessKey  string `json:"access_key"`
		SecretKey  string `json:"secret_key"`
		Bucket     string `json:"bucket"`
		UseSSL     bool   `json:"use_ssl"`     // 是否使用 HTTPS 连接 MinIO 服务器
		BaseURL    string `json:"base_url"`    // MinIO 文件访问的基础 URL（用于生成文件访问链接）
		PublicRead bool   `json:"public_read"` // 是否允许匿名读取对象（仅建议本地开发使用）
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
	Config.MinIO.PublicRead = false
	// 小程序配置（用于用户登录）
	Config.MiniApp.AppID = "wx216ea69c5507523c"
	Config.MiniApp.AppSecret = "482be6b809c61ce122a3f3fbc535ec80"
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

	// 本地开发和部署环境可以通过环境变量覆盖默认配置。
	// 这样无需为了切换数据库或对象存储而修改源码。
	Config.Server.Port = envInt("APP_PORT", Config.Server.Port)
	Config.Database.Host = envString("DB_HOST", Config.Database.Host)
	Config.Database.Port = envInt("DB_PORT", Config.Database.Port)
	Config.Database.Username = envString("DB_USERNAME", Config.Database.Username)
	Config.Database.Password = envString("DB_PASSWORD", Config.Database.Password)
	Config.Database.DBName = envString("DB_NAME", Config.Database.DBName)
	Config.Database.Charset = envString("DB_CHARSET", Config.Database.Charset)
	Config.MinIO.Endpoint = envString("MINIO_ENDPOINT", Config.MinIO.Endpoint)
	Config.MinIO.AccessKey = envString("MINIO_ACCESS_KEY", Config.MinIO.AccessKey)
	Config.MinIO.SecretKey = envString("MINIO_SECRET_KEY", Config.MinIO.SecretKey)
	Config.MinIO.Bucket = envString("MINIO_BUCKET", Config.MinIO.Bucket)
	Config.MinIO.UseSSL = envBool("MINIO_USE_SSL", Config.MinIO.UseSSL)
	Config.MinIO.BaseURL = envString("MINIO_BASE_URL", Config.MinIO.BaseURL)
	Config.MinIO.PublicRead = envBool("MINIO_PUBLIC_READ", Config.MinIO.PublicRead)
	Config.MiniApp.AppID = envString("MINI_APP_ID", Config.MiniApp.AppID)
	Config.MiniApp.AppSecret = envString("MINI_APP_SECRET", Config.MiniApp.AppSecret)
	Config.Map.AmapKey = envString("AMAP_KEY", Config.Map.AmapKey)
	Config.Map.TencentKey = envString("TENCENT_MAP_KEY", Config.Map.TencentKey)
}

func envString(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
