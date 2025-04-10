package api

import (
	"github.com/gin-gonic/gin"
	"tgwp/log/zlog"
	"tgwp/logic"
	"tgwp/response"
	"tgwp/types"
	"tgwp/utils/jwtUtils"
)

// CreatePost 获取用户基础信息
func CreatePost(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.CreatePostReq](c)
	if err != nil {
		return
	}
	req.UserID = jwtUtils.GetUserId(c)
	zlog.CtxInfof(ctx, "创建帖子请求: %v", req)
	resp, err := logic.NewPostLogic().CreatePost(ctx, req)
	response.Response(c, resp, err)
}

// GetPostDetail 获取帖子详情
func GetPostDetail(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.GetPostDetailReq](c)
	if err != nil {
		return
	}
	req.OperatorID = jwtUtils.GetUserId(c)
	req.OperatorRole = jwtUtils.GetRole(c)
	zlog.CtxInfof(ctx, "获取帖子详情请求: %v", req)
	resp, err := logic.NewPostLogic().GetPostDetail(ctx, req)
	response.Response(c, resp, err)
}
