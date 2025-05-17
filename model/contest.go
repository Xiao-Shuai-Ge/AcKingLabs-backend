package model

type Contest struct {
	ID int64 `json:"id" gorm:"column:id;primaryKey;type:bigint;type:bigint"`
	TimeModel
	Url         string `json:"url" gorm:"column:url;type:varchar(512);index:idx_url"`
	Platform    string `json:"platform" gorm:"column:platform;type:varchar(255);"`
	Title       string `json:"title" gorm:"column:title;type:varchar(255)"`
	StartTime   int64  `json:"start_time" gorm:"column:start_time;type:bigint"`
	EndTime     int64  `json:"end_time" gorm:"column:end_time;type:bigint"`
	Duration    int64  `json:"duration" gorm:"column:duration;type:bigint"`
	IsRecommend bool   `json:"is_recommend" gorm:"column:is_recommend;type:bool"`
	IsCustomize bool   `json:"is_customize" gorm:"column:is_customize;type:bool"`
}
