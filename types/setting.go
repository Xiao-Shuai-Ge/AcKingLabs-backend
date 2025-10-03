package types

import "tgwp/model"

// GetSettingReq 获取用户设置请求
type GetSettingReq struct {
	// 空请求，从token中获取用户ID
}

// GetSettingResp 获取用户设置响应
type GetSettingResp struct {
	UserID   int64              `json:"user_id"`
	Settings model.SettingsJSON `json:"settings"`
}

// UpdateSettingReq 更新用户设置请求
type UpdateSettingReq struct {
	Settings model.SettingsJSON `json:"settings" binding:"required"`
}

// UpdateSettingResp 更新用户设置响应
type UpdateSettingResp struct {
	UserID   int64              `json:"user_id"`
	Settings model.SettingsJSON `json:"settings"`
	Message  string             `json:"message"`
}
