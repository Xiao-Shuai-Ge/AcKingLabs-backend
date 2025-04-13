package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gin-gonic/gin"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"tgwp/global"
	"tgwp/log/zlog"
	"tgwp/response"
	"tgwp/types"
)

// UploadFile 上传文件,因为比较难以分类，所以只在API中实现
func UploadFile(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	resp := types.UploadFileResp{}
	// 限制上传文件大小（示例为10MB）
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 10*1024*1024)

	// 获得上传文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		zlog.CtxErrorf(ctx, "获取上传文件失败: %v", err)
		err = response.ErrResp(err, response.INTERNAL_ERROR)
		response.Response(c, resp, err)
		return
	}
	defer file.Close()

	// 读取文件内容到内存
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		zlog.CtxErrorf(ctx, "读取文件内容失败: %v", err)
		err = response.ErrResp(err, response.INTERNAL_ERROR)
		response.Response(c, resp, err)
		return
	}

	// 计算SHA-256哈希值
	hasher := sha256.New()
	hasher.Write(fileBytes)
	hash := hex.EncodeToString(hasher.Sum(nil))

	// 获取文件扩展名并构造新文件名
	ext := filepath.Ext(header.Filename)
	newFilename := hash + ext

	// 检查OSS中是否已存在该文件
	exist, err := global.OssBucket.IsObjectExist(newFilename)
	if err != nil {
		zlog.CtxErrorf(ctx, "检查文件存在失败: %v", err)
		err = response.ErrResp(err, response.INTERNAL_ERROR)
		response.Response(c, resp, err)
		return
	}

	if !exist {
		// 设置正确的Content-Type
		contentType := mime.TypeByExtension(ext)
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		// 上传到OSS
		reader := bytes.NewReader(fileBytes)
		err = global.OssBucket.PutObject(newFilename, reader,
			oss.ACL(oss.ACLPublicRead),
			oss.ContentType(contentType),
		)
		if err != nil {
			zlog.CtxErrorf(ctx, "上传文件到OSS失败: %v", err)
			err = response.ErrResp(err, response.INTERNAL_ERROR)
			response.Response(c, resp, err)
			return
		}
	}

	// 构造访问URL
	url := fmt.Sprintf("https://%s.%s/%s", global.Config.Oss.BucketName, global.Config.Oss.Endpoint, newFilename)
	resp.Url = url

	zlog.CtxInfof(ctx, "上传成功，访问URL: %s", url)
	response.Response(c, resp, nil)
}
