package api

import (
	"github.com/gin-gonic/gin"
	"tgwp/log/zlog"
	"tgwp/logic"
	"tgwp/response"
	"tgwp/types"
	"tgwp/utils/jwtUtils"
)

func GetMessageCount(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.GetMessageCountReq](c)
	if err != nil {
		return
	}
	req.UserID = jwtUtils.GetUserId(c)
	zlog.CtxInfof(ctx, "获取用户消息数量请求: %v", req)
	resp, err := logic.NewMessageLogic().GetMessageCount(ctx, req)
	response.Response(c, resp, err)
}

func GetMessageList(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.GetMessageListReq](c)
	if err != nil {
		return
	}
	req.UserID = jwtUtils.GetUserId(c)
	zlog.CtxInfof(ctx, "获取用户消息列表请求: %v", req)
	resp, err := logic.NewMessageLogic().GetMessageList(ctx, req)
	response.Response(c, resp, err)
}

func MarkReadMessage(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.MarkReadMessageReq](c)
	if err != nil {
		return
	}
	req.UserID = jwtUtils.GetUserId(c)
	zlog.CtxInfof(ctx, "标记已读消息请求: %v", req)
	err = logic.NewMessageLogic().MarkReadMessage(ctx, req)
	response.Response(c, nil, err)
}
