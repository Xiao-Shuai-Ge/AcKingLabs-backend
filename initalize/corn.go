package initalize

import (
	"context"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"tgwp/global"
	"tgwp/log/zlog"
	"tgwp/model"
	"tgwp/repo"
	"tgwp/utils/contest"
	"time"
)

func Cron() {
	zone, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		logrus.Warn("加载时区错误:%v", err)
	}
	crontab := cron.New(cron.WithSeconds(), cron.WithLocation(zone))
	// 每10分钟计算帖子热度
	_, err = crontab.AddFunc("@every 10m", ComputePostWeight)
	if err != nil {
		logrus.Warn("添加帖子热度计算任务失败:%v", err)
	}
	// 每30分钟更新比赛列表
	_, err = crontab.AddFunc("@every 1m", UpdateContests)

	zlog.Infof("启动定时任务成功")
	crontab.Start()
}

func ComputePostWeight() {
	ctx := context.Background()
	zlog.CtxInfof(ctx, "开始计算帖子热度")
	err := repo.NewPostRepo(global.DB).ComputePostWeight()
	if err != nil {
		zlog.CtxErrorf(ctx, "计算帖子热度失败:%v", err)
	}
	zlog.CtxInfof(ctx, "计算帖子热度完成")
}

func UpdateContests() {
	zlog.Infof("开始更新比赛列表")
	contests := contest.GetCodeForcesContest()
	AddContest(contests)
	contests = contest.GetAtCoderContest()
	AddContest(contests)
	contests = contest.GetNowcoderContest()
	AddContest(contests)
	zlog.Infof("更新比赛列表完成")
}

func AddContest(contests []model.Contest) {
	for _, contest := range contests {
		// 先判断比赛是否已经在数据库中
		isExists, err := repo.NewContestRepo(global.DB).IsContestExists(contest.Url)
		if err != nil {
			zlog.Errorf("添加比赛失败: %v", err)
			return
		}
		if isExists {
			// 更新比赛时间
			err = repo.NewContestRepo(global.DB).UpdateContest(contest)
			if err != nil {
				zlog.Errorf("更新比赛失败: %v", err)
				return
			}
			continue
		} else {
			// 否则插入数据库
			contest.ID = global.SnowflakeNode.Generate().Int64()
			err = repo.NewContestRepo(global.DB).CreateContest(contest)
			zlog.Infof("添加比赛成功: %v", contest)
			if err != nil {
				zlog.Errorf("添加比赛失败: %v", err)
				return
			}
		}
	}
}
