package logic

import (
	"context"
	"fmt"
	"strconv"
	"tgwp/global"
	"tgwp/log/zlog"
	"tgwp/model"
	"tgwp/repo"
	"tgwp/response"
	"tgwp/types"
	"tgwp/utils"
	"time"
)

type PostLogic struct {
}

func NewPostLogic() *PostLogic {
	return &PostLogic{}
}

// CreatePost 创建帖子
func (l *PostLogic) CreatePost(ctx context.Context, req types.CreatePostReq) (resp types.CreatePostResp, err error) {
	defer utils.RecordTime(time.Now())()
	// ID 转化为 int64
	userID, err := strconv.ParseInt(req.UserID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.UserID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 如果是周记打卡，先检查时间是否正确
	if req.Type == "diary" {
		req.Source = GetWeekCode()
		if len(req.Source) == 0 {
			zlog.CtxErrorf(ctx, "周记打卡时间错误: %v", err)
			return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
		}
		zlog.CtxInfof(ctx, "解析出打卡周数: %s", req.Source)
	}
	// 判断数据范围
	// 1. 标题不能超过 50 个字符
	if len(req.Title) > 50 {
		zlog.CtxErrorf(ctx, "标题不能超过 50 个字符: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 2. 内容不能超过 5000 个字符
	if len(req.Content) > 5000 {
		zlog.CtxErrorf(ctx, "内容不能超过 5000 个字符: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 3. 除了周记打卡可以私密，其他类型都不可以私密
	if req.Type != "diary" && req.IsPrivate {
		zlog.CtxErrorf(ctx, "非周记打卡不能私密: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 4. 不允许出现不存在的类型
	if !global.TYPE_SET[req.Type] {
		zlog.CtxErrorf(ctx, "不存在的类型: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 创建帖子
	id := global.SnowflakeNode.Generate().Int64()
	post := model.Post{
		ID:      id,
		UserID:  userID,
		Title:   req.Title,
		Content: req.Content,
		Type:    req.Type,
		Source:  req.Source,
	}
	err = repo.NewPostRepo(global.DB).CreatePost(post)
	if err != nil {
		zlog.CtxErrorf(ctx, "创建帖子失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	resp.ID = id
	return
}

func GetWeekCode() string {
	timeNow := time.Now()
	timeNow = time.UnixMilli(1744081921000)
	timestamp := timeNow.UnixMilli()

	// 打卡时间为每周的周日中午到周二的中午，为了先确定当前周数，先把时间减去 2 天
	timestamp -= 2 * 24 * 60 * 60 * 1000
	// 周一到周日分别为 1 到 7
	weekday := int64((time.UnixMilli(timestamp).Weekday()+6)%7 + 1)

	// 时间返回到周一中午 12:00:00
	timestamp -= (weekday - 1) * 24 * 60 * 60 * 1000
	timestamp -= (timestamp + 8*3600*1000) % (24 * 60 * 60 * 1000) // 取整到天
	timestamp += 12 * 60 * 60 * 1000                               // 加上中午(UTC+8)
	// 如果当前时间不在合法打卡
	if utils.Abs(timeNow.UnixMilli()-(timestamp+7*24*60*60*1000)) > 24*60*60*1000 {
		zlog.Debugf("%v = %v", utils.Abs(timeNow.UnixMilli()-(timestamp+7*24*60*60*1000)), 24*60*60*1000)
		zlog.Warnf("当前时间不在合法打卡时间范围内 %v ~ %v", time.UnixMilli(timestamp+7*24*60*60*1000), timeNow)
		return ""
	}
	// 计算当前周数，确定年份和月份
	week := 0
	year := time.UnixMilli(timestamp).Year()
	month := time.UnixMilli(timestamp).Month()
	for month == time.UnixMilli(timestamp).Month() {
		timestamp -= 7 * 24 * 60 * 60 * 1000
		week++
	}
	// 格式化周数
	weekCode := fmt.Sprintf("%d-%d-%d", year, month, week)
	return weekCode
}
