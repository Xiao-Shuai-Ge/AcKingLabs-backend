package initalize

import (
	"context"
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"tgwp/global"
	"tgwp/log/zlog"
	"tgwp/model"
	"tgwp/repo"
	"tgwp/utils/contest"
	"tgwp/utils/email"
	"tgwp/utils/ulearning"
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
		zlog.Errorf("添加定时任务失败:%v", err)
	}
	// 每30分钟更新比赛列表
	_, err = crontab.AddFunc("@every 30m", UpdateContests)
	if err != nil {
		zlog.Errorf("添加定时任务失败:%v", err)
	}
	// 每1分钟检查订阅比赛通知
	_, err = crontab.AddFunc("@every 1m", BookingContestNotice)
	if err != nil {
		zlog.Errorf("添加定时任务失败:%v", err)
	}

	// 每1分钟检查自动签到
	_, err = crontab.AddFunc("@every 1m", AutoSignin)
	if err != nil {
		zlog.Errorf("添加定时任务失败:%v", err)
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
		isExists, err := repo.NewContestRepo(global.DB).IsContestExistsByUrl(contest.Url)
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

func BookingContestNotice() {
	zlog.Infof("开始检查订阅比赛通知")
	contests, err := repo.NewContestRepo(global.DB).GetUnStartedContestList()
	if err != nil {
		zlog.Errorf("获取未开始比赛列表失败: %v", err)
		return
	}
	//zlog.Infof("获取未开始比赛列表成功: %v", contests)
	for _, contest := range contests {
		// 如果比赛开始时间不止20分钟，跳过
		if time.Now().Add(20*time.Minute).UnixMilli() < contest.StartTime {
			continue
		}
		// 获取订阅者列表
		var bookings []model.Booking
		bookings, err = repo.NewContestRepo(global.DB).GetBookingListByContestID(contest.ID)
		if err != nil {
			zlog.Errorf("获取订阅者列表失败: %v", err)
			continue
		}
		// 整合订阅者邮箱
		var sendMails []string
		for _, booking := range bookings {
			sendMails = append(sendMails, booking.Email)
		}
		if len(sendMails) == 0 {
			continue
		}
		// 发送邮件通知
		err = email.BookingContest(sendMails, contest.Url, contest.Title, "20")
		if err != nil {
			zlog.Errorf("发送邮件通知失败: %v", err)
			continue
		}
		// 删除订阅记录
		err = repo.NewContestRepo(global.DB).RemoveBookingByContestID(contest.ID)
		if err != nil {
			zlog.Errorf("删除订阅记录失败: %v", err)
		}
	}
	zlog.Infof("检查订阅比赛通知完成")
}

func AutoSignin() {
	zlog.Infof("开始检查自动签到列表")
	// 获取所有参与自动签到的数据行
	autoSignins, err := repo.NewTemplateRepo(global.DB).GetAutoSigninList()
	if err != nil {
		zlog.Errorf("获取自动签到列表失败: %v", err)
		return
	}
	var now_userid int64 // 记录当前的用户ID
	user := ulearning.NewUser()
	for _, autoSignin := range autoSignins {
		// 判断是否需要登录
		if autoSignin.UserID != now_userid {
			now_userid = autoSignin.UserID
			// 从数据库里获取账号密码
			var userInfo *model.Ulearning
			userInfo, err = repo.NewTemplateRepo(global.DB).GetUserInfo(autoSignin.UserID)
			if err != nil {
				zlog.Errorf("获取用户信息失败: %v", err)
				continue
			}
			err = user.Login(userInfo.UserName, userInfo.Password)
			if err != nil {
				zlog.Errorf("登录失败: %v", err)
				continue
			}
		}
		// 查看签到活动
		var Activities ulearning.GetCourseActivitiesResp
		Activities, err = user.GetCourseActivities(int(autoSignin.CoursesID))
		if err != nil {
			zlog.Errorf("获取课程活动失败: %v", err)
			continue
		}
		for _, activity := range Activities.OtherActivityDTOList {
			if activity.RelationType == 1 && activity.PersonStatus == 0 && activity.Status != 3 {
				zlog.Infof("发现需要处理的签到: 【%s】 %s", autoSignin.CourseName, activity.Title)
				// 获取签到详情
				teacher := ulearning.NewUser()
				err = teacher.TeacherLogin()
				if err != nil {
					zlog.Errorf("登录失败: %v", err)
					continue
				}
				var detail ulearning.GetActivityDetailResp
				detail, err = teacher.GetActivityDetail(activity.RelationID)
				if err != nil {
					zlog.Errorf("获取课程活动详细信息失败: %v", err)
					continue
				}
				// 检查签到人数占比
				p := float64(detail.AbsenceNum) / float64(detail.AbsenceNum+detail.NotAbsenceNum)
				zlog.Infof("签到人数占比: %.1f%% (%d/%d)", p*100, detail.AbsenceNum, detail.AbsenceNum+detail.NotAbsenceNum)
				if int(p*100) < autoSignin.Percentage {
					zlog.Warnf("签到人数占比不足 %d%% : (%d/%d)", autoSignin.Percentage, detail.AbsenceNum, detail.AbsenceNum+detail.NotAbsenceNum)
					continue
				}
				// 签到
				err = user.SigninByStudent(activity.RelationID, int(autoSignin.ClassID))
				if err != nil {
					zlog.Errorf("签到失败: %v", err)
					continue
				}
				zlog.Infof("签到成功: 【%s】 %s", autoSignin.CourseName, activity.Title)
				// 发送邮箱通知
				err = email.AutoSignin(autoSignin.Email, fmt.Sprintf("【%s】 %s", autoSignin.CourseName, activity.Title))
				if err != nil {
					zlog.Errorf("发送邮件通知失败: %v", err)
					continue
				}
			}
		}
	}
}
