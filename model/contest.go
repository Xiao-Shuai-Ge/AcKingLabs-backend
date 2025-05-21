package model

type Contest struct {
	ID int64 `json:"id" gorm:"column:id;primaryKey;type:bigint;type:bigint"`
	TimeModel
	Url         string `json:"url" gorm:"column:url;type:varchar(512);uniqueIndex"`
	Platform    string `json:"platform" gorm:"column:platform;type:varchar(255);"`
	Title       string `json:"title" gorm:"column:title;type:varchar(255)"`
	StartTime   int64  `json:"start_time" gorm:"column:start_time;type:bigint"`
	EndTime     int64  `json:"end_time" gorm:"column:end_time;type:bigint"`
	Duration    int64  `json:"duration" gorm:"column:duration;type:bigint"`
	IsRecommend bool   `json:"is_recommend" gorm:"column:is_recommend;type:bool"`
}

type Booking struct {
	ID int64 `json:"id" gorm:"column:id;primaryKey;type:bigint;type:bigint"`
	TimeModel
	ContestID int64  `json:"contest_id" gorm:"column:contest_id;type:bigint"`
	UserID    int64  `json:"user_id" gorm:"column:user_id;type:bigint"`
	Email     string `json:"email" gorm:"column:email;type:varchar(255);"`
}
