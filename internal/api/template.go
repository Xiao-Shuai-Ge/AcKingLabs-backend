package api

import (
	"github.com/gin-gonic/gin"
	"tgwp/log/zlog"
	"tgwp/logic"
	"tgwp/response"
	"tgwp/types"
	"tgwp/utils/jwtUtils"
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
	//BindReq里面用泛型进行了处理绑定
	req, err := types.BindReq[types.TemplateReq](c)
	if err != nil {
		return
	}
	zlog.CtxInfof(ctx, "test request: %v", req)
	resp, err := logic.NewTemplateLogic().Way(ctx, req)
	response.Response(c, resp, err)
}

func SigninList(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	//BindReq里面用泛型进行了处理绑定
	req, err := types.BindReq[types.SigninListReq](c)
	req.ID = jwtUtils.GetUserId(c)
	if err != nil {
		return
	}
	zlog.CtxInfof(ctx, "signinlist request: %v", req)
	resp, err := logic.NewTemplateLogic().SigninList(ctx, req)
	response.Response(c, resp, err)
}

func Signin(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.SigninReq](c)
	req.ID = jwtUtils.GetUserId(c)
	if err != nil {
		return
	}
	zlog.CtxInfof(ctx, "sign request: %v", req)
	resp, err := logic.NewTemplateLogic().Signin(ctx, req)
	response.Response(c, resp, err)
}

func SigninTeacher(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.SigninTeacherReq](c)
	req.ID = jwtUtils.GetUserId(c)
	if err != nil {
		return
	}
	zlog.CtxInfof(ctx, "sign request: %v", req)
	resp, err := logic.NewTemplateLogic().SigninTeacher(ctx, req)
	response.Response(c, resp, err)
}

func GetAutoList(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.GetAutoListReq](c)
	req.UserID = jwtUtils.GetUserId(c)
	if err != nil {
		return
	}
	zlog.CtxInfof(ctx, "auto list request: %v", req)
	resp, err := logic.NewTemplateLogic().GetAutoList(ctx, req)
	response.Response(c, resp, err)
}

func AutoSetting(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.AutoSettingReq](c)
	req.UserID = jwtUtils.GetUserId(c)
	if err != nil {
		return
	}
	zlog.CtxInfof(ctx, "auto setting request: %v", req)
	resp, err := logic.NewTemplateLogic().AutoSetting(ctx, req)
	response.Response(c, resp, err)
}
