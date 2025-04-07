package logic

import (
	"context"
	"errors"
	"gorm.io/gorm"
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
	// 获取用户信息
	var user model.User
	user, err = repo.NewUserRepo(global.DB).GetUserProfileByID(req.ID)
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
	// 获取用户信息
	var user model.User
	user, err = repo.NewUserRepo(global.DB).GetUserProfileByID(req.ID)
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
