package initalize

import (
	"context"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"tgwp/global"
	"tgwp/log/zlog"
	"tgwp/repo"
	"time"
)

func Cron() {
	zone, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		logrus.Warn("加载时区错误:%v", err)
	}
	crontab := cron.New(cron.WithSeconds(), cron.WithLocation(zone))
	// 每天2点同步文章数据
	_, err = crontab.AddFunc("@every 10m", ComputePostWeight)
	if err != nil {
		logrus.Warn("添加帖子热度计算任务失败:%v", err)
	}
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
