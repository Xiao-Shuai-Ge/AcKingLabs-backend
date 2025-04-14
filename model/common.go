package model

import (
	"gorm.io/gorm"
	"tgwp/global"
	"tgwp/utils/snowflake"
	"time"
)

// CommonModel 每张表都有的四个东西，最好不要用 gorm.model（虽然他们一模一样）
type CommonModel struct {
	ID        int64 `gorm:"primaryKey;column:id;type:bigint"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type TimeModel struct {
	CreatedTime int64 `gorm:"column:created_time;type:bigint"`
	UpdatedTime int64 `gorm:"column:updated_time;type:bigint"`
}

func (b *CommonModel) BeforeCreate(db *gorm.DB) error {
	// 生成雪花ID
	if b.ID == 0 {
		b.ID = snowflake.GetIntId(global.Node)
	}

	return nil
}

func (b *TimeModel) BeforeCreate(db *gorm.DB) error {
	// 生成雪花ID
	b.CreatedTime = time.Now().UnixMilli()
	b.UpdatedTime = time.Now().UnixMilli()

	return nil
}

type Ulearning struct {
	ID int64 `json:"id" gorm:"column:id;primaryKey;type:bigint;type:bigint"`

	UserID   int64  `json:"user_id" gorm:"column:user_id;type:bigint;comment:用户ID;uniqueIndex:idx_post_user_unique"`
	UserName string `json:"user_name" gorm:"column:user_name;type:varchar(255);comment:用户名;"`
	Password string `json:"password" gorm:"column:password;type:varchar(255);comment:密码;"`
}
