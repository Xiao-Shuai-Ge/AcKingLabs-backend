package api

import (
	"tgwp/log/zlog"
	"tgwp/logic"
	"tgwp/response"
	"tgwp/types"
	"tgwp/utils/jwtUtils"

	"github.com/gin-gonic/gin"
)

// SubmitResume 投递简历
func SubmitResume(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.SubmitResumeReq](c)
	if err != nil {
		return
	}
	zlog.CtxInfof(ctx, "投递简历请求: %v", req)
	resp, err := logic.NewResumeLogic().SubmitResume(ctx, req)
	response.Response(c, resp, err)
}

// UpdateResume 修改简历
func UpdateResume(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.UpdateResumeReq](c)
	if err != nil {
		return
	}
	zlog.CtxInfof(ctx, "修改简历请求: %v", req)
	resp, err := logic.NewResumeLogic().UpdateResume(ctx, req)
	response.Response(c, resp, err)
}

// GetResumeDetail 查询简历详细信息
func GetResumeDetail(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.GetResumeDetailReq](c)
	if err != nil {
		return
	}
	// 获取用户角色
	userRole := jwtUtils.GetRole(c)
	zlog.CtxInfof(ctx, "查询简历详细信息请求: %v, 用户角色: %d", req, userRole)
	resp, err := logic.NewResumeLogic().GetResumeDetail(ctx, req, userRole)
	response.Response(c, resp, err)
}

// GetResumeList 获取简历列表（管理员功能）
func GetResumeList(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.GetResumeListReq](c)
	if err != nil {
		return
	}
	zlog.CtxInfof(ctx, "获取简历列表请求: %v", req)
	resp, err := logic.NewResumeLogic().GetResumeList(ctx, req)
	response.Response(c, resp, err)
}

// DeleteResume 删除简历（管理员功能）
func DeleteResume(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.DeleteResumeReq](c)
	if err != nil {
		return
	}
	zlog.CtxInfof(ctx, "删除简历请求: %v", req)
	resp, err := logic.NewResumeLogic().DeleteResume(ctx, req)
	response.Response(c, resp, err)
}

// AcceptResume 通过简历（管理员功能）
func AcceptResume(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.AcceptResumeReq](c)
	if err != nil {
		return
	}
	zlog.CtxInfof(ctx, "通过简历请求: %v", req)
	resp, err := logic.NewResumeLogic().AcceptResume(ctx, req)
	response.Response(c, resp, err)
}
