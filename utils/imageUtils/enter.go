package imageUtils

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png" // 注册PNG解码器
	"path/filepath"
	"strings"
)

// CompressImage 压缩图片到指定大小以下
// 支持PNG和JPG格式
// maxSizeBytes: 最大文件大小（字节）
func CompressImage(imageData []byte, filename string, maxSizeBytes int64) ([]byte, error) {
	// 检查文件大小，如果已经小于限制，直接返回
	if int64(len(imageData)) <= maxSizeBytes {
		return imageData, nil
	}

	// 获取文件扩展名
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return nil, fmt.Errorf("不支持的图片格式: %s", ext)
	}

	// 解码图片
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("解码图片失败: %v", err)
	}

	// 简化压缩流程：PNG转JPG，只尝试3种质量参数
	return compressWithSimpleQuality(img, format, maxSizeBytes), nil
}

// CompressImageWithFormat 压缩图片并返回压缩后的数据和格式信息
// 返回: (压缩后的数据, 是否转换为JPG, 错误)
func CompressImageWithFormat(imageData []byte, filename string, maxSizeBytes int64) ([]byte, bool, error) {
	// 检查文件大小，如果已经小于限制，直接返回
	if int64(len(imageData)) <= maxSizeBytes {
		return imageData, false, nil
	}

	// 获取文件扩展名
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return nil, false, fmt.Errorf("不支持的图片格式: %s", ext)
	}

	// 解码图片
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, false, fmt.Errorf("解码图片失败: %v", err)
	}

	// 简化压缩流程：PNG转JPG，只尝试3种质量参数
	result := compressWithSimpleQuality(img, format, maxSizeBytes)

	// 检查是否从PNG转换为JPG
	convertedToJPG := (format == "png")

	return result, convertedToJPG, nil
}

// compressWithSimpleQuality 简化的压缩函数
// PNG转JPG，只尝试3种质量参数：85, 70, 50
func compressWithSimpleQuality(img image.Image, format string, maxSizeBytes int64) []byte {
	var buf bytes.Buffer

	// 定义3种质量参数，从高到低
	qualities := []int{85, 70, 50}

	// 如果是PNG，直接转换为JPG；如果是JPG，保持JPG格式
	for _, quality := range qualities {
		buf.Reset()
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
		if err != nil {
			continue
		}

		// 如果大小合适，返回结果
		if int64(buf.Len()) <= maxSizeBytes {
			return buf.Bytes()
		}
	}

	// 如果所有质量参数都不满足，返回最低质量的结果
	return buf.Bytes()
}

// ValidateImageFormat 验证图片格式
func ValidateImageFormat(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return fmt.Errorf("只支持PNG和JPG格式的图片")
	}
	return nil
}

// GetImageInfo 获取图片信息
func GetImageInfo(imageData []byte) (width, height int, format string, err error) {
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return 0, 0, "", err
	}

	bounds := img.Bounds()
	return bounds.Dx(), bounds.Dy(), format, nil
}
