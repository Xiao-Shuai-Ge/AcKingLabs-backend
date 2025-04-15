package repo

import (
	"gorm.io/gorm"
	"tgwp/model"
)

type MessageRepo struct {
	DB *gorm.DB
}

func NewMessageRepo(db *gorm.DB) *MessageRepo {
	return &MessageRepo{
		DB: db,
	}
}

func (r *MessageRepo) GetMessageCount(user_id int64) (system_count int64, like_count int64, comment_count int64, err error) {
	err = r.DB.Model(&model.Message{}).Where("user_id = ? and type = 'system' and is_read = 0 ", user_id).Count(&system_count).Error
	if err != nil {
		return
	}
	err = r.DB.Model(&model.Message{}).Where("user_id = ? and type = 'like' and is_read = 0 ", user_id).Count(&like_count).Error
	if err != nil {
		return
	}
	err = r.DB.Model(&model.Message{}).Where("user_id = ? and type = 'comment' and is_read = 0 ", user_id).Count(&comment_count).Error
	if err != nil {
		return
	}
	return
}

func (r *MessageRepo) GetMessageList(user_id int64, message_type string) (Messages []model.Message, err error) {
	err = r.DB.Model(&model.Message{}).Where("user_id = ? and type = ?", user_id, message_type).Order("id desc").Find(&Messages).Limit(100).Error
	return
}

func (r *MessageRepo) MarkReadMessage(user_id int64, message_id int64) (err error) {
	err = r.DB.Model(&model.Message{}).Where("user_id = ? and id = ?", user_id, message_id).Update("is_read", 1).Error
	return
}

func (r *MessageRepo) SendMessage(message model.Message) (err error) {
	err = r.DB.Model(&model.Message{}).Create(&message).Error
	return
}
