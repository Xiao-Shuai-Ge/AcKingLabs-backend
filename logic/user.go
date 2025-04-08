package logic

import (
	"context"
	"errors"
	"gorm.io/gorm"
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
	resp.Role = user.Role

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
	resp.RealName = user.RealName
	resp.Avatar = user.Avatar
	resp.Xp = user.Xp
	resp.Grade = user.Grade
	resp.StudentNo = user.StudentNo
	resp.CodeforcesID = user.CodeforcesID
	resp.CodeforcesRating = user.CodeforcesRating
	resp.Role = user.Role

	return resp, nil
}

func (l *UserLogic) SetUserProfile(ctx context.Context, req types.SetUserProfileReq) (resp types.SetUserProfileResp, err error) {
	defer utils.RecordTime(time.Now())()
	// ID 转化为 int64
	userID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.ID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	operatorID, err := strconv.ParseInt(req.OperatorID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.OperatorID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 检查权限
	if operatorID != userID {
		// 查询操作者是不是管理员
		if req.OperatorRole < 3 {
			zlog.CtxErrorf(ctx, "非法操作: %v", req.OperatorID)
			return resp, response.ErrResp(err, response.PERMISSION_DENIED)
		}
	}
	// 检验数据
	// 1.用户名去除所有空格，且不能为空，且长度不超过 30
	req.Username = strings.ReplaceAll(req.Username, " ", "")
	if req.Username == "" {
		zlog.CtxErrorf(ctx, "用户名不能为空")
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	} else if len(req.Username) > 30 {
		zlog.CtxErrorf(ctx, "用户名长度不能超过 30")
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 2.真实姓名去除所有空格，不能为空，且长度在 [2,20] 之间
	req.RealName = strings.ReplaceAll(req.RealName, " ", "")
	if len(req.RealName) < 2 || len(req.RealName) > 20 {
		zlog.CtxErrorf(ctx, "真实姓名长度必须在 [2,20] 之间")
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	if len(req.Avatar) > 100 {
		zlog.CtxErrorf(ctx, "头像 URL 长度不能超过 100")
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 3. 年级在 [0,99]
	if req.Grade < 0 || req.Grade > 99 {
		zlog.CtxErrorf(ctx, "年级必须在 [0,99] 之间")
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 4.其他一律不超过30即可
	if utils.Max(len(req.CodeforcesID), len(req.StudentNo)) > 30 {
		zlog.CtxErrorf(ctx, "其他字段长度不能超过 30")
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}

	// 拿出先前的用户信息
	user, err := repo.NewUserRepo(global.DB).GetUserProfileByID(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		zlog.CtxErrorf(ctx, "用户并不存在!: %v", err)
		return resp, response.ErrResp(err, response.USER_NOT_EXIST)
	} else if err != nil {
		zlog.CtxErrorf(ctx, "获取用户信息失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	// 更新用户信息
	user.Username = req.Username
	user.Avatar = req.Avatar
	user.Grade = req.Grade
	user.StudentNo = req.StudentNo
	user.RealName = req.RealName
	user.CodeforcesID = req.CodeforcesID
	err = repo.NewUserRepo(global.DB).UpdateUserProfile(user)
	if err != nil {
		zlog.CtxErrorf(ctx, "更新用户信息失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	return resp, nil
}

// SetUserRole 设置用户权限
func (l *UserLogic) SetUserRole(ctx context.Context, req types.SetUserRoleReq) (resp types.SetUserRoleResp, err error) {
	// ID 转化为 int64
	userID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.ID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 检查权限(必须是超级管理员)
	if req.OperatorRole < 4 {
		zlog.CtxErrorf(ctx, "非法操作")
		return resp, response.ErrResp(err, response.PERMISSION_DENIED)
	}
	// 修改权限
	err = repo.NewUserRepo(global.DB).SetUserRole(userID, req.Role)
	if err != nil {
		zlog.CtxErrorf(ctx, "修改权限失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	zlog.CtxInfof(ctx, "修改权限成功")
	// 修改用户经验值
	// 获取用户当前经验值
	user, err := repo.NewUserRepo(global.DB).GetUserProfileByID(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		zlog.CtxErrorf(ctx, "用户并不存在!: %v", err)
		return resp, response.ErrResp(err, response.USER_NOT_EXIST)
	} else if err != nil {
		zlog.CtxErrorf(ctx, "获取用户信息失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	if req.Role == 2 {
		// 至少为 200
		user.Xp = utils.Max(user.Xp, 200)
	} else if req.Role == 3 {
		// 至少为 400
		user.Xp = utils.Max(user.Xp, 400)
	}
	// 更新用户信息
	err = repo.NewUserRepo(global.DB).UpdateUserProfile(user)
	if err != nil {
		zlog.CtxErrorf(ctx, "更新用户信息失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}

	return resp, nil
}
