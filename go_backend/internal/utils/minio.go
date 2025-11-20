package utils

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go_backend/internal/config"
)

var minioClient *minio.Client

// InitMinIO 初始化MinIO客户端
func InitMinIO() error {
	cfg := config.Config.MinIO

	// 初始化MinIO客户端
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: false, // 使用HTTP而不是HTTPS
	})
	if err != nil {
		log.Printf("初始化MinIO客户端失败: %v", err)
		return err
	}

	// 检查存储桶是否存在，如果不存在则创建
	exists, err := client.BucketExists(context.Background(), cfg.Bucket)
	if err != nil {
		log.Printf("检查存储桶是否存在失败: %v", err)
		return err
	}

	if !exists {
		// 创建存储桶
		err = client.MakeBucket(context.Background(), cfg.Bucket, minio.MakeBucketOptions{Region: "us-east-1"})
		if err != nil {
			log.Printf("创建存储桶失败: %v", err)
			return err
		}
		log.Printf("存储桶 %s 创建成功\n", cfg.Bucket)
	}

	minioClient = client
	log.Println("MinIO客户端初始化成功")
	return nil
}

// UploadFile 上传文件到MinIO
func UploadFile(fileName string, reader *http.Request) (string, error) {
	if minioClient == nil {
		// 如果客户端未初始化，先初始化
		if err := InitMinIO(); err != nil {
			return "", fmt.Errorf("初始化MinIO客户端失败: %v", err)
		}
	}

	file, header, err := reader.FormFile("file")
	if err != nil {
		return "", fmt.Errorf("获取文件失败: %v", err)
	}
	defer file.Close()

	cfg := config.Config.MinIO

	// 生成唯一的对象名称
	objectName := fmt.Sprintf("%s_%d%s", fileName, time.Now().Unix(), getFileExtension(header.Filename))

	// 上传文件
	uploadInfo, err := minioClient.PutObject(
		context.Background(),
		cfg.Bucket,  // 存储桶名称
		objectName,  // 对象名称
		file,        // 文件内容
		header.Size, // 文件大小
		minio.PutObjectOptions{ContentType: header.Header.Get("Content-Type")},
	)
	if err != nil {
		return "", fmt.Errorf("上传文件失败: %v", err)
	}

	// 生成可访问的URL
	// 注意：这里使用的是公共访问方式，如果需要私有访问，需要生成带有签名的URL
	fileURL := fmt.Sprintf("http://%s/%s/%s", cfg.Endpoint, cfg.Bucket, objectName)

	log.Printf("成功上传文件: %s, 大小: %d 字节\n", uploadInfo.Key, uploadInfo.Size)
	return fileURL, nil
}

// DeleteFile 从MinIO删除文件
func DeleteFile(objectName string) error {
	if minioClient == nil {
		// 如果客户端未初始化，先初始化
		if err := InitMinIO(); err != nil {
			return fmt.Errorf("初始化MinIO客户端失败: %v", err)
		}
	}

	cfg := config.Config.MinIO

	// 从URL中提取对象名称
	// 假设URL格式为: http://endpoint/bucket/objectName
	objectName = extractObjectNameFromURL(objectName, cfg.Endpoint, cfg.Bucket)

	// 删除文件
	err := minioClient.RemoveObject(context.Background(), cfg.Bucket, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("删除文件失败: %v", err)
	}

	log.Printf("成功删除文件: %s\n", objectName)
	return nil
}

// 获取文件扩展名
func getFileExtension(filename string) string {
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			return filename[i:]
		}
	}
	return ""
}

// 从URL中提取对象名称
func extractObjectNameFromURL(url, endpoint, bucket string) string {
	// 构建前缀
	prefix := fmt.Sprintf("http://%s/%s/", endpoint, bucket)

	// 移除前缀获取对象名称
	if len(url) > len(prefix) && url[:len(prefix)] == prefix {
		return url[len(prefix):]
	}

	// 如果URL格式不符合预期，直接返回原始URL
	return url
}