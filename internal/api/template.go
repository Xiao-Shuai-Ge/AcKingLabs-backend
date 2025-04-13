package api

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"path/filepath"
	"tgwp/global"
	"tgwp/log/zlog"
	"tgwp/response"
	"tgwp/types"
)

type OSSConfig struct {
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	BucketName      string
	Domain          string
}

// api层不要写复杂的东西，移步到logic层
func Template(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	// 获得上传文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		zlog.CtxErrorf(ctx, "获取上传文件失败: %v", err)
		return
	}
	defer file.Close()

	// 生成唯一文件名
	ext := filepath.Ext(header.Filename)
	newFilename := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// 上传到OSS
	err = global.OssBucket.PutObject(newFilename, file, oss.ACL(oss.ACLPublicRead))
	if err != nil {
		zlog.CtxErrorf(ctx, "获取上传文件失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件上传失败"})
		return
	}

	// 返回访问URL
	url := fmt.Sprintf("https://%s.%s/%s", global.Config.Oss.BucketName, global.Config.Oss.Endpoint, newFilename)

	zlog.CtxInfof(ctx, "上传成功，访问URL: %s", url)

	//zlog.CtxInfof(ctx, "test request: %v", req)
	//resp, err := logic.NewTemplateLogic().Way(ctx, req)
	resp := types.TemplateResp{}
	response.Response(c, resp, err)
}
