package api

import (
	"tgwp/log/zlog"
	"tgwp/logic"
	"tgwp/response"
	"tgwp/types"
	"tgwp/utils/jwtUtils"

	"github.com/gin-gonic/gin"
)

// GetSetting 获取用户设置
func GetSetting(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)

	// 从token中获取用户ID
	userID := jwtUtils.GetUserId(c)

	zlog.CtxInfof(ctx, "获取用户设置请求，用户ID: %s", userID)

	resp, err := logic.NewSettingLogic().GetSetting(ctx, userID)
	response.Response(c, resp, err)
}

// UpdateSetting 更新用户设置
func UpdateSetting(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)

	// 从token中获取用户ID
	userID := jwtUtils.GetUserId(c)

	// 绑定请求参数
	req, err := types.BindReq[types.UpdateSettingReq](c)
	if err != nil {
		zlog.CtxErrorf(ctx, "绑定请求参数失败: %v", err)
		return
	}

	zlog.CtxInfof(ctx, "更新用户设置请求，用户ID: %s, 请求: %+v", userID, req)

	resp, err := logic.NewSettingLogic().UpdateSetting(ctx, userID, req)
	response.Response(c, resp, err)
}

// ResetSetting 重置用户设置为默认值
func ResetSetting(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)

	// 从token中获取用户ID
	userID := jwtUtils.GetUserId(c)

	zlog.CtxInfof(ctx, "重置用户设置请求，用户ID: %s", userID)

	resp, err := logic.NewSettingLogic().ResetSetting(ctx, userID)
	response.Response(c, resp, err)
}
