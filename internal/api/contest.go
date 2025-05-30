package api

import (
	"github.com/gin-gonic/gin"
	"tgwp/log/zlog"
	"tgwp/logic"
	"tgwp/response"
	"tgwp/types"
	"tgwp/utils/jwtUtils"
)

// GetContestList 获取比赛列表请求
func GetContestList(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.GetContestListReq](c)
	if err != nil {
		return
	}
	zlog.CtxInfof(ctx, "获取比赛列表请求: %v", req)
	resp, err := logic.NewContestLogic().GetContestList(ctx, req)
	response.Response(c, resp, err)
}

// CreateContest 创建比赛请求
func CreateContest(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.CreateContestReq](c)
	if err != nil {
		return
	}
	zlog.CtxInfof(ctx, "创建比赛请求: %v", req)
	resp, err := logic.NewContestLogic().CreateContest(ctx, req)
	response.Response(c, resp, err)
}

func GetContestDetail(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.GetContestDetailReq](c)
	if err != nil {
		return
	}
	zlog.CtxInfof(ctx, "获取比赛详情请求: %v", req)
	resp, err := logic.NewContestLogic().GetContestDetail(ctx, req)
	response.Response(c, resp, err)
}

// BookingContest 预约比赛请求
func BookingContest(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.BookingContestReq](c)
	if err != nil {
		return
	}
	req.UserID = jwtUtils.GetUserId(c)
	zlog.CtxInfof(ctx, "预约比赛请求: %v", req)
	resp, err := logic.NewContestLogic().BookingContest(ctx, req)
	response.Response(c, resp, err)
}

// IsBookingContest 查询是否预约比赛请求
func IsBookingContest(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.IsBookingContestReq](c)
	if err != nil {
		return
	}
	req.UserID = jwtUtils.GetUserId(c)
	zlog.CtxInfof(ctx, "查询是否预约比赛请求: %v", req)
	resp, err := logic.NewContestLogic().IsBookingContest(ctx, req)
	response.Response(c, resp, err)
}

func RecommendContest(c *gin.Context) {
	ctx := zlog.GetCtxFromGin(c)
	req, err := types.BindReq[types.RecommendContestReq](c)
	if err != nil {
		return
	}
	zlog.CtxInfof(ctx, "推荐比赛请求: %v", req)
	resp, err := logic.NewContestLogic().RecommendContest(ctx, req)
	response.Response(c, resp, err)
}
