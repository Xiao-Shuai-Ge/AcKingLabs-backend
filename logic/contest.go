package logic

import (
	"context"
	"tgwp/global"
	"tgwp/log/zlog"
	"tgwp/model"
	"tgwp/repo"
	"tgwp/response"
	"tgwp/types"
	"tgwp/utils"
	"time"
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

	contests, resp.PageTotal, err = repo.NewContestRepo(global.DB).GetContestList(req.Type, req.Page, req.Count)

	// 数据库查询失败
	if err != nil {
		zlog.CtxErrorf(ctx, "查询比赛失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	//zlog.CtxDebugf(ctx, "查询比赛成功: %v", contests)
	for _, contest := range contests {
		// 组装返回数据
		resp.Contests = append(resp.Contests, types.ContestInfo{
			ID:        contest.ID,
			Title:     contest.Title,
			StartTime: contest.StartTime,
			EndTime:   contest.EndTime,
			Duration:  contest.Duration,
			Platform:  contest.Platform,
			Url:       contest.Url,
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
