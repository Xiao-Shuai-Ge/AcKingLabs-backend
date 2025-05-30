package logic

import (
	"context"
	"strconv"
	"strings"
	"tgwp/global"
	"tgwp/log/zlog"
	"tgwp/model"
	"tgwp/repo"
	"tgwp/response"
	"tgwp/types"
	"tgwp/utils"
	"time"
	"unicode/utf8"
)

type ContestLogic struct {
}

func NewContestLogic() *ContestLogic {
	return &ContestLogic{}
}

func (l *ContestLogic) GetContestList(ctx context.Context, req types.GetContestListReq) (resp types.GetContestListResp, err error) {
	defer utils.RecordTime(time.Now())()
	// 分各种情况查询比赛
	var contests []model.Contest

	if req.Type == "recommend" {
		contests, resp.PageTotal, err = repo.NewContestRepo(global.DB).GetContestListByRecommend(req.Page, req.Count)
	} else {
		contests, resp.PageTotal, err = repo.NewContestRepo(global.DB).GetContestList(req.Type, req.Page, req.Count)
	}

	// 数据库查询失败
	if err != nil {
		zlog.CtxErrorf(ctx, "查询比赛失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	//zlog.CtxDebugf(ctx, "查询比赛成功: %v", contests)
	for _, contest := range contests {
		// 组装返回数据
		resp.Contests = append(resp.Contests, types.ContestInfo{
			ID:          contest.ID,
			Title:       contest.Title,
			StartTime:   contest.StartTime,
			EndTime:     contest.EndTime,
			Duration:    contest.Duration,
			Platform:    contest.Platform,
			Url:         contest.Url,
			IsRecommend: contest.IsRecommend,
		})
	}
	resp.Length = len(resp.Contests)
	if resp.PageTotal%int64(req.Count) == 0 {
		resp.PageTotal = resp.PageTotal / int64(req.Count)
	} else {
		resp.PageTotal = resp.PageTotal/int64(req.Count) + 1
	}
	return
}

func (l *ContestLogic) CreateContest(ctx context.Context, req types.CreateContestReq) (resp types.CreateContestResp, err error) {
	defer utils.RecordTime(time.Now())()
	// 验证数据
	// 1. 标题不能超过 30 个字符
	if utf8.RuneCountInString(req.Title) > 50 {
		zlog.CtxErrorf(ctx, "标题不能超过 50 个字符: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 2. Url 不能超过 255 个字符
	if utf8.RuneCountInString(req.Url) > 255 {
		zlog.CtxErrorf(ctx, "Url 不能超过 255 个字符: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 3. 标题 和 Url 去掉前后空格后均不能为空
	if strings.Trim(req.Title, " ") == "" || strings.Trim(req.Url, " ") == "" {
		zlog.CtxErrorf(ctx, "标题和 Url 不能为空: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 4. 开始时间不能大于结束时间
	if req.StartTime > req.EndTime {
		zlog.CtxErrorf(ctx, "开始时间不能大于结束时间: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 组装数据
	id := global.SnowflakeNode.Generate().Int64()
	contest := model.Contest{
		ID:          id,
		Title:       req.Title,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Duration:    (req.EndTime - req.StartTime) / 1000,
		Url:         req.Url,
		IsRecommend: false,
		Platform:    "AcKing",
	}
	// 插入数据库
	err = repo.NewContestRepo(global.DB).CreateContest(contest)
	if err != nil {
		zlog.CtxErrorf(ctx, "创建比赛失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	zlog.CtxDebugf(ctx, "创建比赛成功: %v", contest)
	resp.ID = contest.ID
	return
}

func (l *ContestLogic) GetContestDetail(ctx context.Context, req types.GetContestDetailReq) (resp types.GetContestDetailResp, err error) {
	defer utils.RecordTime(time.Now())()
	// ID 转化为 int64
	contestID, err := strconv.ParseInt(req.ContestID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.ContestID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 查询比赛
	var contest model.Contest
	contest, err = repo.NewContestRepo(global.DB).GetContestByID(contestID)
	if err != nil {
		zlog.CtxErrorf(ctx, "查询比赛失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	// 组装返回数据
	resp.Contest = types.ContestInfo{
		ID:          contest.ID,
		Title:       contest.Title,
		StartTime:   contest.StartTime,
		EndTime:     contest.EndTime,
		Duration:    contest.Duration,
		Platform:    contest.Platform,
		Url:         contest.Url,
		IsRecommend: contest.IsRecommend,
	}
	return
}

func (l *ContestLogic) BookingContest(ctx context.Context, req types.BookingContestReq) (resp types.BookingContestResp, err error) {
	defer utils.RecordTime(time.Now())()
	// ID 转化为 int64
	contestID, err := strconv.ParseInt(req.ContestID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.ContestID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	userID, err := strconv.ParseInt(req.UserID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.UserID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 验证比赛是否存在
	var isExists bool
	isExists, err = repo.NewContestRepo(global.DB).IsContestExistsByID(contestID)
	if err != nil {
		zlog.CtxErrorf(ctx, "查询比赛失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	} else if !isExists {
		zlog.CtxErrorf(ctx, "比赛不存在: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 判断是否已经预约
	isExists, err = repo.NewContestRepo(global.DB).IsBooking(contestID, userID)
	if isExists {
		// 取消预约
		err = repo.NewContestRepo(global.DB).RemoveBooking(contestID, userID)
		resp.IsBooking = false
	} else {
		//  获取个人邮箱
		var user model.User
		user, err = repo.NewUserRepo(global.DB).GetUserProfileByID(userID)
		if err != nil {
			zlog.CtxErrorf(ctx, "查询用户信息失败: %v", err)
			return resp, response.ErrResp(err, response.DATABASE_ERROR)
		}
		// 预约比赛
		id := global.SnowflakeNode.Generate().Int64()
		booking := model.Booking{
			ID:        id,
			ContestID: contestID,
			UserID:    userID,
			Email:     user.Email,
		}
		err = repo.NewContestRepo(global.DB).CreateBooking(booking)
		resp.IsBooking = true
	}

	if err != nil {
		zlog.CtxErrorf(ctx, "操作预约比赛失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}

	return
}

func (l *ContestLogic) IsBookingContest(ctx context.Context, req types.IsBookingContestReq) (resp types.IsBookingContestResp, err error) {
	defer utils.RecordTime(time.Now())()
	// ID 转化为 int64
	contestID, err := strconv.ParseInt(req.ContestID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.UserID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	userID, err := strconv.ParseInt(req.UserID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.UserID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 查询是否已经预约
	var isBooking bool
	isBooking, err = repo.NewContestRepo(global.DB).IsBooking(contestID, userID)
	if err != nil {
		zlog.CtxErrorf(ctx, "查询是否已经预约失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	resp.IsBooking = isBooking
	return
}

func (l *ContestLogic) RecommendContest(ctx context.Context, req types.RecommendContestReq) (resp types.RecommendContestResp, err error) {
	defer utils.RecordTime(time.Now())()
	// ID 转化为 int64
	contestID, err := strconv.ParseInt(req.ContestID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.ContestID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 数据库操作
	err = repo.NewContestRepo(global.DB).SetContestRecommend(contestID, req.IsRecommend)
	if err != nil {
		zlog.CtxErrorf(ctx, "设置推荐状态失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	// 组装返回数据
	resp.IsRecommend = req.IsRecommend
	return
}
