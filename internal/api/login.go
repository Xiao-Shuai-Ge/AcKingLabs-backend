package api

import (
	"github.com/gin-gonic/gin"
	"tgwp/log/zlog"
	"tgwp/logic"
	"tgwp/response"
	"tgwp/types"
	"tgwp/utils/jwtUtils"
)

// SendCode 发送验证码
func SendCode(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.SendCodeReq](c)
	if err != nil {
		return
	}
	zlog.CtxInfof(ctx, "注册请求: %v", req)
	resp, err := logic.NewLoginLogic().SendCode(ctx, req)
	response.Response(c, resp, err)
}

// Register 注册
func Register(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.RegisterReq](c)
	if err != nil {
		return
	}
	zlog.CtxInfof(ctx, "注册请求: %v", req)
	resp, err := logic.NewLoginLogic().Register(ctx, req)
	response.Response(c, resp, err)
}

// TokenTest 测试token
func TokenTest(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.TokenTestReq](c)
	if err != nil {
		return
	}
	userId := jwtUtils.GetUserId(c)
	zlog.CtxInfof(ctx, "解析token成功，userId: %v", userId)

	zlog.CtxInfof(ctx, "注册请求: %v", req)
	resp, err := logic.NewLoginLogic().TokenTest(ctx, req)
	response.Response(c, resp, err)
}
