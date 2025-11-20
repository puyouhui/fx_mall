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
	} `json:"minio"`
}{}

// InitConfig 初始化配置
func InitConfig() {
	// 设置默认配置
	Config.Server.Port = 8082
	Config.Server.ReadTimeout = 5 * time.Second
	Config.Server.WriteTimeout = 5 * time.Second
	Config.CORS.AllowOrigins = []string{"*"}
	Config.CORS.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	Config.CORS.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	// 设置MySQL数据库配置
	Config.Database.Host = "113.44.164.151"
	Config.Database.Port = 3306
	Config.Database.Username = "root"
	Config.Database.Password = "hn02le.34lkdLKD"
	Config.Database.DBName = "product_shop"
	Config.Database.Charset = "utf8mb4"
	// 设置MinIO配置
	Config.MinIO.Endpoint = "113.44.164.151:9000"
	Config.MinIO.AccessKey = "admin"
	Config.MinIO.SecretKey = "hn02le.34lkdLKD"
	Config.MinIO.Bucket = "selected"
}