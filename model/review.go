package model

type Review struct {
	ID int64 `json:"id" gorm:"column:id;primaryKey;type:bigint"`
	TimeModel

	PostID     int64  `json:"post_id" gorm:"column:post_id;type:bigint;comment:帖子ID"`
	ReviewerID int64  `json:"reviewer_id" gorm:"column:reviewer_id;type:bigint;comment:审核人ID"`
	Status     int    `json:"status" gorm:"column:status;type:int;comment:审核状态 0:未审核 1:通过 2:拒绝"`
	Reason     string `json:"reason" gorm:"column:reason;type:varchar(255);comment:拒绝原因"`

	// 历史快照
	PostTitle   string `json:"post_title" gorm:"column:post_title;type:varchar(255);comment:帖子标题快照"`
	PostContent string `json:"post_content" gorm:"column:post_content;type:longtext;comment:帖子内容快照"`
	UserID      int64  `json:"user_id" gorm:"column:user_id;type:bigint;comment:作者ID快照"`
}

func (Review) TableName() string {
	return "reviews"
}
