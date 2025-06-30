package repo

import (
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"tgwp/log/zlog"
	"tgwp/model"
	"tgwp/utils/cacheUtils"
	"time"
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
	// redis缓存查询
	value, err := cacheUtils.Get(fmt.Sprintf("cache:message_count:%d", user_id))
	if value != "" {
		// 缓存命中
		zlog.Debugf("缓存命中")
		var data map[string]interface{}
		err = json.Unmarshal([]byte(value), &data)
		if err != nil {
			return system_count, like_count, comment_count, err
		}
		system_count = int64(int(data["system_count"].(float64)))
		like_count = int64(int(data["like_count"].(float64)))
		comment_count = int64(int(data["comment_count"].(float64)))
		return system_count, like_count, comment_count, nil
	}

	// system_count, err = r.DB.Model(&model.Message{}).Where("user_id = ? and type = 'system' ", user_id).Count(&system_count).Error
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

	// 缓存写入
	data := map[string]interface{}{
		"system_count":  system_count,
		"like_count":    like_count,
		"comment_count": comment_count,
	}
	newValue, err := json.Marshal(data)
	if err != nil {
		return
	}
	err = cacheUtils.Set(fmt.Sprintf("cache:message_count:%d", user_id), string(newValue), 5*time.Minute)
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
