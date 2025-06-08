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

// EditPost 编辑帖子
func EditPost(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.EditPostReq](c)
	if err != nil {
		return
	}
	req.OperatorID = jwtUtils.GetUserId(c)
	req.OperatorRole = jwtUtils.GetRole(c)
	zlog.CtxInfof(ctx, "编辑帖子请求: %v", req)
	resp, err := logic.NewPostLogic().EditPost(ctx, req)
	response.Response(c, resp, err)
}

// DeletePost 删除帖子
func DeletePost(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.DeletePostReq](c)
	if err != nil {
		return
	}
	req.OperatorID = jwtUtils.GetUserId(c)
	req.OperatorRole = jwtUtils.GetRole(c)
	zlog.CtxInfof(ctx, "删除帖子请求: %v", req)
	resp, err := logic.NewPostLogic().DeletePost(ctx, req)
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

func GetPostDetailVisitor(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.GetPostDetailReq](c)
	if err != nil {
		return
	}
	// 和GetPostDetailReq一致，但id设置为0，表示访客
	req.OperatorID = "0"
	req.OperatorRole = 0
	zlog.CtxInfof(ctx, "获取帖子详情请求(游客): %v", req)
	resp, err := logic.NewPostLogic().GetPostDetail(ctx, req)
	response.Response(c, resp, err)
}

// GetLikePost 获取帖子是否点赞
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

// CreateComment 创建评论
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

// GetMoreComments 获取更多评论
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

// GetLikeComment 获取评论是否点赞
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

// LikeComment 点赞评论
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

// GetMorePosts 获取更多帖子
func GetMorePosts(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.GetMorePostsReq](c)
	if err != nil {
		return
	}
	zlog.CtxInfof(ctx, "获取更多帖子请求: %v", req)
	resp, err := logic.NewPostLogic().GetMorePosts(ctx, req)
	response.Response(c, resp, err)
}

// GetPagePosts 按页获取帖子
func GetPagePosts(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.GetPagePostsReq](c)
	if err != nil {
		return
	}
	zlog.CtxInfof(ctx, "获取更多帖子请求: %v", req)
	resp, err := logic.NewPostLogic().GetPagePosts(ctx, req)
	response.Response(c, resp, err)
}

func SetPostFeature(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.SetPostFeatureReq](c)
	if err != nil {
		return
	}
	zlog.CtxInfof(ctx, "设置帖子精华请求: %v", req)
	resp, err := logic.NewPostLogic().SetPostFeature(ctx, req)
	response.Response(c, resp, err)
}

func GetDiaryList(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.GetDiaryListReq](c)
	if err != nil {
		return
	}
	zlog.CtxInfof(ctx, "获取个人周记列表请求: %v", req)
	resp, err := logic.NewPostLogic().GetDiaryList(ctx, req)
	response.Response(c, resp, err)
}

func SearchPosts(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.SearchPostsReq](c)
	if err != nil {
		return
	}
	zlog.CtxInfof(ctx, "搜索帖子请求: %v", req)
	resp, err := logic.NewPostLogic().SearchPosts(ctx, req)
	response.Response(c, resp, err)
}
