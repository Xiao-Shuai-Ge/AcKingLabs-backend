package repo

import (
	"errors"
	"tgwp/model"

	"gorm.io/gorm"
)

type SettingRepo struct {
	db *gorm.DB
}

func NewSettingRepo(db *gorm.DB) *SettingRepo {
	return &SettingRepo{db: db}
}

// GetByUserID 根据用户ID获取设置
func (r *SettingRepo) GetByUserID(userID int64) (*model.UserSetting, error) {
	var setting model.UserSetting
	err := r.db.Where("user_id = ?", userID).First(&setting).Error
	return &setting, err
}

// GetOrCreate 获取或创建用户设置，如果不存在则创建默认设置
func (r *SettingRepo) GetOrCreate(userID int64) (*model.UserSetting, error) {
	var setting model.UserSetting
	err := r.db.Where("user_id = ?", userID).First(&setting).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 不存在，创建默认设置
			defaultSettings := model.GetDefaultSettings()
			setting = model.UserSetting{
				UserID:         userID,
				Settings:       defaultSettings,
				HelpPostNotify: defaultSettings.HelpPostNotify, // 同步冗余字段
			}
			if err := r.db.Create(&setting).Error; err != nil {
				return nil, err
			}
			return &setting, nil
		}
		return nil, err
	}

	return &setting, nil
}

// Create 创建用户设置
func (r *SettingRepo) Create(setting *model.UserSetting) error {
	return r.db.Create(setting).Error
}

// Update 更新用户设置
func (r *SettingRepo) Update(setting *model.UserSetting) error {
	return r.db.Model(setting).Where("user_id = ?", setting.UserID).Updates(map[string]interface{}{
		"settings":         setting.Settings,
		"help_post_notify": setting.HelpPostNotify, // 同步更新冗余字段
	}).Error
}

// UpdateOrCreate 更新或创建用户设置
func (r *SettingRepo) UpdateOrCreate(userID int64, settings model.SettingsJSON) (*model.UserSetting, error) {
	var setting model.UserSetting
	err := r.db.Where("user_id = ?", userID).First(&setting).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 不存在，创建新记录
			setting = model.UserSetting{
				UserID:         userID,
				Settings:       settings,
				HelpPostNotify: settings.HelpPostNotify, // 同步冗余字段
			}
			if err := r.db.Create(&setting).Error; err != nil {
				return nil, err
			}
			return &setting, nil
		}
		return nil, err
	}

	// 存在，更新记录
	setting.Settings = settings
	setting.HelpPostNotify = settings.HelpPostNotify // 同步冗余字段
	if err := r.Update(&setting); err != nil {
		return nil, err
	}

	return &setting, nil
}

// Delete 删除用户设置（软删除）
func (r *SettingRepo) Delete(userID int64) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.UserSetting{}).Error
}

// GetUserIDsWithHelpPostNotify 获取所有开启了求助帖通知的用户ID列表
func (r *SettingRepo) GetUserIDsWithHelpPostNotify() ([]int64, error) {
	var userIDs []int64
	err := r.db.Model(&model.UserSetting{}).
		Where("help_post_notify = ?", true).
		Pluck("user_id", &userIDs).Error
	return userIDs, err
}
