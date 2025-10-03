package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// UserSetting 用户设置表
type UserSetting struct {
	ID int64 `json:"id" gorm:"column:id;primaryKey;type:bigint"`
	TimeModel
	UserID   int64        `json:"user_id" gorm:"column:user_id;type:bigint;comment:用户ID;uniqueIndex:idx_user_id"`
	Settings SettingsJSON `json:"settings" gorm:"column:settings;type:json;comment:用户设置JSON"`
}

// SettingsJSON 用户设置的JSON结构体
type SettingsJSON struct {
	// 通知设置
	SystemMessageEmailNotify bool `json:"system_message_email_notify"` // 系统消息邮箱通知
	LikeNotify               bool `json:"like_notify"`                 // 点赞通知
	ReplyNotify              bool `json:"reply_notify"`                // 回复通知
	HelpPostNotify           bool `json:"help_post_notify"`            // 发布求助帖通知
}

// GetDefaultSettings 返回默认设置
func GetDefaultSettings() SettingsJSON {
	return SettingsJSON{
		SystemMessageEmailNotify: false, // 系统消息邮箱通知默认关闭
		LikeNotify:               false, // 点赞通知默认关闭
		ReplyNotify:              true,  // 回复通知默认开启
		HelpPostNotify:           false, // 发布求助帖通知默认关闭
	}
}

// Scan 实现 sql.Scanner 接口，用于从数据库读取
func (s *SettingsJSON) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("类型转换失败")
	}
	return json.Unmarshal(bytes, s)
}

// Value 实现 driver.Valuer 接口，用于写入数据库
func (s SettingsJSON) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Validate 验证设置字段是否合法
func (s *SettingsJSON) Validate() error {
	// 通知设置都是布尔值，无需特殊验证
	return nil
}

// MergeWithDefault 将当前设置与默认设置合并，填充缺失的字段
func (s *SettingsJSON) MergeWithDefault() SettingsJSON {
	// 布尔类型字段无需特殊合并处理
	// 所有字段都会有默认的零值（false）或从请求中传入的值
	return *s
}
