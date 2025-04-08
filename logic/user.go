package logic

import (
	"context"
	"errors"
	"gorm.io/gorm"
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

type UserLogic struct {
}

func NewUserLogic() *UserLogic {
	return &UserLogic{}
}

// GetUserInfo 获取用户信息
func (l *UserLogic) GetUserInfo(ctx context.Context, req types.GetUserInfoReq) (resp types.GetUserInfoResp, err error) {
	defer utils.RecordTime(time.Now())()
	zlog.CtxInfof(ctx, "获取用户信息 %s", req.ID)
	// ID 转化为 int64
	userID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%s 转换 int64 错误: %v", req.ID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 获取用户信息
	var user model.User
	user, err = repo.NewUserRepo(global.DB).GetUserProfileByID(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		zlog.CtxErrorf(ctx, "用户并不存在!: %v", err)
		return resp, response.ErrResp(err, response.USER_NOT_EXIST)
	}
	// 填入参数
	resp.ID = user.ID
	resp.Username = user.Username
	resp.Avatar = user.Avatar
	resp.Xp = user.Xp

	return resp, nil
}

// GetUserProfile 获取用户信息
func (l *UserLogic) GetUserProfile(ctx context.Context, req types.GetUserProfileReq) (resp types.GetUserProfileResp, err error) {
	defer utils.RecordTime(time.Now())()
	// ID 转化为 int64
	userID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.ID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 获取用户信息
	var user model.User
	user, err = repo.NewUserRepo(global.DB).GetUserProfileByID(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		zlog.CtxErrorf(ctx, "用户并不存在!: %v", err)
		return resp, response.ErrResp(err, response.USER_NOT_EXIST)
	}
	// 填入参数
	resp.ID = user.ID
	resp.Username = user.Username
	resp.Avatar = user.Avatar
	resp.Xp = user.Xp
	resp.Grade = user.Grade
	resp.StudentNo = user.StudentNo
	resp.CodeforcesID = user.CodeforcesID
	resp.CodeforcesRating = user.CodeforcesRating

	return resp, nil
}
