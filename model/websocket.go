package model

type ChatMessage struct {
	ID int64 `json:"id" gorm:"column:id;primaryKey;type:bigint;type:bigint"`
	TimeModel
	Content string `json:"content" gorm:"column:content;type:text;type:json"`
	UserID  int64  `json:"user_id" gorm:"column:user_id;type:bigint;type:bigint"`
}
