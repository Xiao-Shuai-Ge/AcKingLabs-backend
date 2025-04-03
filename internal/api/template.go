package api

import (
	"github.com/gin-gonic/gin"
	"tgwp/log/zlog"
	"tgwp/logic"
	"tgwp/response"
	"tgwp/types"
)

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
