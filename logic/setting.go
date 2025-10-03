package logic

import (
	"context"
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

type SettingLogic struct {
	settingRepo *repo.SettingRepo
}

func NewSettingLogic() *SettingLogic {
	return &SettingLogic{
		settingRepo: repo.NewSettingRepo(global.DB),
	}
}

// GetSetting 获取用户设置，如果不存在则自动创建默认设置
func (l *SettingLogic) GetSetting(ctx context.Context, userIDStr string) (resp types.GetSettingResp, err error) {
	defer utils.CtxRecordTime(ctx, time.Now())()

	// 转换用户ID
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "用户ID转换失败: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}

	// 获取或创建用户设置
	setting, err := l.settingRepo.GetOrCreate(userID)
	if err != nil {
		zlog.CtxErrorf(ctx, "获取用户设置失败，用户ID: %d, 错误: %v", userID, err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}

	zlog.CtxInfof(ctx, "成功获取用户设置，用户ID: %d", userID)

	resp = types.GetSettingResp{
		UserID:   setting.UserID,
		Settings: setting.Settings,
	}

	return resp, nil
}

// UpdateSetting 更新用户设置，如果不存在则自动创建
func (l *SettingLogic) UpdateSetting(ctx context.Context, userIDStr string, req types.UpdateSettingReq) (resp types.UpdateSettingResp, err error) {
	defer utils.CtxRecordTime(ctx, time.Now())()

	// 转换用户ID
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "用户ID转换失败: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}

	// 验证设置字段
	if err := req.Settings.Validate(); err != nil {
		zlog.CtxErrorf(ctx, "设置字段验证失败: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}

	// 合并默认设置，防止缺失字段
	mergedSettings := req.Settings.MergeWithDefault()

	// 更新或创建用户设置
	setting, err := l.settingRepo.UpdateOrCreate(userID, mergedSettings)
	if err != nil {
		zlog.CtxErrorf(ctx, "更新用户设置失败，用户ID: %d, 错误: %v", userID, err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}

	zlog.CtxInfof(ctx, "成功更新用户设置，用户ID: %d", userID)

	resp = types.UpdateSettingResp{
		UserID:   setting.UserID,
		Settings: setting.Settings,
		Message:  "设置更新成功",
	}

	return resp, nil
}

// ResetSetting 重置用户设置为默认值
func (l *SettingLogic) ResetSetting(ctx context.Context, userIDStr string) (resp types.UpdateSettingResp, err error) {
	defer utils.CtxRecordTime(ctx, time.Now())()

	// 转换用户ID
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "用户ID转换失败: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}

	// 重置为默认设置
	defaultSettings := model.GetDefaultSettings()
	setting, err := l.settingRepo.UpdateOrCreate(userID, defaultSettings)
	if err != nil {
		zlog.CtxErrorf(ctx, "重置用户设置失败，用户ID: %d, 错误: %v", userID, err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}

	zlog.CtxInfof(ctx, "成功重置用户设置，用户ID: %d", userID)

	resp = types.UpdateSettingResp{
		UserID:   setting.UserID,
		Settings: setting.Settings,
		Message:  "设置已重置为默认值",
	}

	return resp, nil
}
