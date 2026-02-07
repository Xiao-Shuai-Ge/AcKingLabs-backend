package api

import (
	"tgwp/log/zlog"
	"tgwp/logic"
	"tgwp/response"
	"tgwp/types"
	"tgwp/utils/jwtUtils"

	"github.com/gin-gonic/gin"
)

// GetReviewList 获取审核列表
func GetReviewList(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.GetReviewListReq](c)
	if err != nil {
		zlog.CtxErrorf(ctx, "获取审核列表请求绑定失败, err: %v", err)
		return
	}
	req.OperatorID = jwtUtils.GetUserId(c)
	req.OperatorRole = jwtUtils.GetRole(c)
	zlog.CtxInfof(ctx, "获取审核列表请求: %v", req)
	resp, err := logic.NewReviewLogic().GetReviewList(ctx, req)
	response.Response(c, resp, err)
}

// AuditReview 审核帖子
func AuditReview(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.AuditReviewReq](c)
	if err != nil {
		zlog.CtxErrorf(ctx, "审核帖子请求绑定失败, err: %v", err)
		return
	}
	req.OperatorID = jwtUtils.GetUserId(c)
	req.OperatorRole = jwtUtils.GetRole(c)
	zlog.CtxInfof(ctx, "审核帖子请求: %v", req)
	resp, err := logic.NewReviewLogic().AuditReview(ctx, req)
	response.Response(c, resp, err)
}
