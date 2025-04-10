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

func GetLikePost(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.GetLikePostReq](c)
	if err != nil {
		return
	}
	req.OperatorID = jwtUtils.GetUserId(c)
	zlog.CtxInfof(ctx, "获取点赞帖子请求: %v", req)
	resp, err := logic.NewPostLogic().GetLikePost(ctx, req)
	response.Response(c, resp, err)
}

// LikePost 点赞帖子
func LikePost(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.LikePostReq](c)
	if err != nil {
		return
	}
	req.OperatorID = jwtUtils.GetUserId(c)
	req.OperatorRole = jwtUtils.GetRole(c)
	zlog.CtxInfof(ctx, "点赞帖子请求: %v", req)
	resp, err := logic.NewPostLogic().LikePost(ctx, req)
	response.Response(c, resp, err)
}

func CreateComment(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.CreateCommentReq](c)
	if err != nil {
		return
	}
	req.UserID = jwtUtils.GetUserId(c)
	zlog.CtxInfof(ctx, "创建评论请求: %v", req)
	resp, err := logic.NewPostLogic().CreateComment(ctx, req)
	response.Response(c, resp, err)
}

func GetMoreComments(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.GetMoreCommentsReq](c)
	if err != nil {
		return
	}
	zlog.CtxInfof(ctx, "获取更多评论请求: %v", req)
	resp, err := logic.NewPostLogic().GetMoreComments(ctx, req)
	response.Response(c, resp, err)
}

func GetLikeComment(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.GetLikeCommentReq](c)
	if err != nil {
		return
	}
	req.OperatorID = jwtUtils.GetUserId(c)
	zlog.CtxInfof(ctx, "获取点赞评论请求: %v", req)
	resp, err := logic.NewPostLogic().GetLikeComment(ctx, req)
	response.Response(c, resp, err)
}

func LikeComment(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.LikeCommentReq](c)
	if err != nil {
		return
	}
	req.OperatorID = jwtUtils.GetUserId(c)
	req.OperatorRole = jwtUtils.GetRole(c)
	zlog.CtxInfof(ctx, "点赞评论请求: %v", req)
	resp, err := logic.NewPostLogic().LikeComment(ctx, req)
	response.Response(c, resp, err)
}
