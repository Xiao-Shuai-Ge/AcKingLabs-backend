package api

import (
	"github.com/gin-gonic/gin"
	"tgwp/log/zlog"
	"tgwp/logic"
	"tgwp/response"
	"tgwp/types"
)

// GetContestList 获取帖子列表
func GetContestList(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.GetContestListReq](c)
	if err != nil {
		return
	}
	zlog.CtxInfof(ctx, "获取更多帖子请求: %v", req)
	resp, err := logic.NewContestLogic().GetContestList(ctx, req)
	response.Response(c, resp, err)
}
