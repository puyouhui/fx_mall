package utils

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"strings"
	"time"

	"go_backend/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
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

	// 限制单个文件最大 15MB
	const maxSize = 15 * 1024 * 1024
	if header.Size > maxSize {
		return "", fmt.Errorf("文件大小不能超过 15MB")
	}

	// 生成唯一的对象名称
	objectName := fmt.Sprintf("%s_%d%s", fileName, time.Now().Unix(), getFileExtension(header.Filename))

	// 读取到内存以便压缩 / 处理
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(file); err != nil {
		return "", fmt.Errorf("读取文件失败: %v", err)
	}

	contentType := header.Header.Get("Content-Type")
	readerToUpload := bytes.NewReader(buf.Bytes())
	sizeToUpload := int64(buf.Len())

	// 如果是图片，进行压缩和优化
	if strings.HasPrefix(strings.ToLower(contentType), "image/") {
		compressedData, compressedSize, err := compressImage(buf.Bytes(), fileName)
		if err == nil && compressedSize > 0 {
			readerToUpload = bytes.NewReader(compressedData)
			sizeToUpload = compressedSize
			contentType = "image/jpeg"
			log.Printf("图片压缩成功: 原始大小 %d 字节, 压缩后 %d 字节, 压缩率 %.1f%%\n",
				buf.Len(), compressedSize, float64(compressedSize)/float64(buf.Len())*100)
		}
	}

	// 上传文件
	uploadInfo, err := minioClient.PutObject(
		context.Background(),
		cfg.Bucket,     // 存储桶名称
		objectName,     // 对象名称
		readerToUpload, // 文件内容
		sizeToUpload,   // 文件大小
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return "", fmt.Errorf("上传文件失败: %v", err)
	}

	// 生成可访问的URL
	// 注意：这里使用的是公共访问方式，如果需要私有访问，需要生成带有签名的URL
	fileURL := fmt.Sprintf("%s/%s/%s", cfg.BaseURL, cfg.Bucket, objectName)

	log.Printf("成功上传文件: %s, 大小: %d 字节\n", uploadInfo.Key, uploadInfo.Size)
	return fileURL, nil
}

// UploadFileByFieldName 根据字段名上传文件到MinIO（支持多文件上传）
func UploadFileByFieldName(fieldName string, fileName string, reader *http.Request) (string, error) {
	if minioClient == nil {
		// 如果客户端未初始化，先初始化
		if err := InitMinIO(); err != nil {
			return "", fmt.Errorf("初始化MinIO客户端失败: %v", err)
		}
	}

	file, header, err := reader.FormFile(fieldName)
	if err != nil {
		return "", fmt.Errorf("获取文件失败: %v", err)
	}
	defer file.Close()

	cfg := config.Config.MinIO

	// 限制单个文件最大 15MB
	const maxSize = 15 * 1024 * 1024
	if header.Size > maxSize {
		return "", fmt.Errorf("文件大小不能超过 15MB")
	}

	// 生成唯一的对象名称
	objectName := fmt.Sprintf("%s_%d%s", fileName, time.Now().Unix(), getFileExtension(header.Filename))

	// 读取到内存以便压缩 / 处理
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(file); err != nil {
		return "", fmt.Errorf("读取文件失败: %v", err)
	}

	contentType := header.Header.Get("Content-Type")
	readerToUpload := bytes.NewReader(buf.Bytes())
	sizeToUpload := int64(buf.Len())

	// 如果是图片，进行压缩和优化
	if strings.HasPrefix(strings.ToLower(contentType), "image/") {
		originalSize := int64(buf.Len())
		log.Printf("开始压缩图片: %s, 原始大小: %d 字节 (%.2f KB)\n", fileName, originalSize, float64(originalSize)/1024)
		compressedData, compressedSize, err := compressImage(buf.Bytes(), fileName)
		if err != nil {
			log.Printf("图片压缩失败: %v, 使用原始图片 (大小: %d 字节, %.2f KB)\n", err, originalSize, float64(originalSize)/1024)
		} else if compressedSize > 0 {
			// 强制使用压缩后的图片（即使压缩后大小没有减小，也使用压缩版本以确保质量一致）
			readerToUpload = bytes.NewReader(compressedData)
			sizeToUpload = compressedSize
			contentType = "image/jpeg"
			log.Printf("图片压缩成功: 原始大小 %d 字节 (%.2f KB), 压缩后 %d 字节 (%.2f KB), 压缩率 %.1f%%\n",
				originalSize, float64(originalSize)/1024, compressedSize, float64(compressedSize)/1024,
				float64(compressedSize)/float64(originalSize)*100)
		} else {
			log.Printf("图片压缩返回空数据, 使用原始图片\n")
		}
	}

	// 上传文件
	uploadInfo, err := minioClient.PutObject(
		context.Background(),
		cfg.Bucket,     // 存储桶名称
		objectName,     // 对象名称
		readerToUpload, // 文件内容
		sizeToUpload,   // 文件大小
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return "", fmt.Errorf("上传文件失败: %v", err)
	}

	// 生成可访问的URL
	fileURL := fmt.Sprintf("%s/%s/%s", cfg.BaseURL, cfg.Bucket, objectName)

	log.Printf("成功上传文件: %s, 大小: %d 字节 (%.2f KB)\n", uploadInfo.Key, uploadInfo.Size, float64(uploadInfo.Size)/1024)
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

// resizeImage 调整图片尺寸（使用简单的最近邻插值）
func resizeImage(img image.Image, newWidth, newHeight int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	bounds := img.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	// 使用简单的最近邻插值进行缩放
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := x * srcWidth / newWidth
			srcY := y * srcHeight / newHeight
			dst.Set(x, y, img.At(bounds.Min.X+srcX, bounds.Min.Y+srcY))
		}
	}

	return dst
}

// compressImage 压缩图片（使用Go标准库image包，轻量且高效）
// fileName: 文件名，用于判断是否需要高压缩率（配送完成图片需要更高压缩率）
func compressImage(imageData []byte, fileName string) ([]byte, int64, error) {
	// 解码图片
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, 0, err
	}

	// 获取原始尺寸
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 判断是否为配送完成图片（需要更高压缩率）
	isDeliveryImage := strings.Contains(fileName, "delivery_product") ||
		strings.Contains(fileName, "delivery_doorplate")

	// 设置压缩参数
	var maxWidth, maxHeight int
	var quality int

	if isDeliveryImage {
		// 配送完成图片：极高压缩率，更小尺寸（节省存储空间）
		maxWidth = 600  // 最大宽度600px（足够清晰，文件更小）
		maxHeight = 600 // 最大高度600px
		quality = 30    // 质量30（极高压缩率，文件更小，但仍可清晰识别货物和门牌）
	} else {
		// 其他图片：中等压缩率
		maxWidth = 1920  // 最大宽度1920px
		maxHeight = 1920 // 最大高度1920px
		quality = 75     // 质量75（中等压缩率）
	}

	// 如果图片尺寸超过限制，进行缩放
	var resizedImg image.Image = img
	if width > maxWidth || height > maxHeight {
		// 计算缩放比例
		scaleX := float64(maxWidth) / float64(width)
		scaleY := float64(maxHeight) / float64(height)
		scale := scaleX
		if scaleY < scaleX {
			scale = scaleY
		}

		newWidth := int(float64(width) * scale)
		newHeight := int(float64(height) * scale)

		// 创建新图片并缩放（使用简单的最近邻插值）
		resizedImg = resizeImage(img, newWidth, newHeight)
		log.Printf("图片尺寸调整: %dx%d -> %dx%d\n", width, height, newWidth, newHeight)
	} else if isDeliveryImage {
		// 配送完成图片即使尺寸不超过限制，也强制压缩以减小文件大小
		resizedImg = img
	}

	// 压缩为JPEG（所有图片都压缩，即使尺寸未超过限制）
	var compressed bytes.Buffer
	err = jpeg.Encode(&compressed, resizedImg, &jpeg.Options{Quality: quality})
	if err != nil {
		return nil, 0, fmt.Errorf("JPEG编码失败: %v", err)
	}

	compressedSize := int64(compressed.Len())
	log.Printf("图片压缩完成: 原始 %d 字节 (%.2f KB), 压缩后 %d 字节 (%.2f KB), 质量 %d, 压缩率 %.1f%%\n",
		len(imageData), float64(len(imageData))/1024, compressedSize, float64(compressedSize)/1024,
		quality, float64(compressedSize)/float64(len(imageData))*100)

	return compressed.Bytes(), compressedSize, nil
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
