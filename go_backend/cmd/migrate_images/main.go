package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"go_backend/internal/config"
	"go_backend/internal/database"
	"go_backend/internal/model"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	log.Println("=========================================")
	log.Println("MinIO 图片数据迁移工具")
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

	// 初始化MinIO客户端
	cfg := config.Config.MinIO
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		log.Fatalf("MinIO客户端初始化失败: %v", err)
	}

	// 检查存储桶是否存在
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		log.Fatalf("检查存储桶失败: %v", err)
	}
	if !exists {
		log.Fatalf("存储桶 %s 不存在", cfg.Bucket)
	}
	log.Println("MinIO连接成功，存储桶验证通过")

	// 从MinIO扫描所有图片并写入数据库
	log.Println("开始扫描MinIO桶中的图片...")

	objectCh := client.ListObjects(ctx, cfg.Bucket, minio.ListObjectsOptions{
		Recursive: true,
	})

	imageExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	count := 0
	skipCount := 0
	errorCount := 0

	for object := range objectCh {
		if object.Err != nil {
			log.Printf("列出对象时出错: %v", object.Err)
			errorCount++
			continue
		}

		// 检查是否为图片
		objectName := object.Key
		isImage := false
		for _, ext := range imageExtensions {
			if len(objectName) >= len(ext) &&
				strings.ToLower(objectName[len(objectName)-len(ext):]) == ext {
				isImage = true
				break
			}
		}

		if !isImage {
			continue
		}

		// 提取分类（从路径中提取，如 products/xxx.jpg -> products）
		category := "others"
		if strings.Contains(objectName, "/") {
			parts := strings.Split(objectName, "/")
			if len(parts) > 0 {
				category = parts[0]
				// 验证分类是否合法
				validCategories := map[string]bool{
					"products":     true,
					"carousels":    true,
					"categories":   true,
					"users":        true,
					"delivery":     true,
					"others":       true,
					"rich-content": true,
				}
				if !validCategories[category] {
					category = "others"
				}
			}
		}

		// 生成URL
		imageURL := fmt.Sprintf("%s/%s/%s", cfg.BaseURL, cfg.Bucket, objectName)

		// 提取文件名（从完整路径中提取）
		fileName := objectName
		if strings.Contains(objectName, "/") {
			parts := strings.Split(objectName, "/")
			fileName = parts[len(parts)-1]
		}

		// 创建索引记录
		imgIndex := &model.ImageIndex{
			ObjectName: objectName,
			ObjectURL:  imageURL,
			Category:   category,
			FileName:   fileName,
			FileSize:   object.Size,
			FileType:   "image/jpeg", // 默认，实际可以从扩展名判断
			UploadedAt: object.LastModified,
		}

		// 根据扩展名设置文件类型
		lastDotIndex := strings.LastIndex(objectName, ".")
		if lastDotIndex > 0 && lastDotIndex < len(objectName)-1 {
			ext := strings.ToLower(objectName[lastDotIndex:])
			switch ext {
		case ".jpg", ".jpeg":
			imgIndex.FileType = "image/jpeg"
		case ".png":
			imgIndex.FileType = "image/png"
		case ".gif":
			imgIndex.FileType = "image/gif"
		case ".webp":
			imgIndex.FileType = "image/webp"
			}
		}

		if err := model.CreateImageIndex(database.DB, imgIndex); err != nil {
			// 如果已存在（UNIQUE约束），跳过
			if strings.Contains(err.Error(), "Duplicate entry") {
				skipCount++
				if skipCount%100 == 0 {
					log.Printf("已跳过 %d 条重复记录...", skipCount)
				}
				continue
			}
			log.Printf("创建图片索引失败: %v, 对象: %s", err, objectName)
			errorCount++
			continue
		}

		count++
		if count%100 == 0 {
			log.Printf("已迁移 %d 张图片（跳过 %d 条重复记录，错误 %d 条）...", count, skipCount, errorCount)
		}
	}

	log.Println("=========================================")
	log.Printf("迁移完成！")
	log.Printf("成功迁移: %d 张图片", count)
	log.Printf("跳过重复: %d 条记录", skipCount)
	log.Printf("错误记录: %d 条", errorCount)
	log.Println("=========================================")
}

