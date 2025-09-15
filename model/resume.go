package model

type Resume struct {
	ID int64 `json:"id" gorm:"column:id;primaryKey;type:bigint;type:bigint"`
	TimeModel

	Avatar string `json:"avatar" gorm:"column:avatar;type:varchar(512);comment:头像"`

	RealName  string `json:"real_name" gorm:"column:real_name;type:varchar(255);"`
	Grade     int    `json:"grade" gorm:"column:grade;type:int;comment:年级"`
	StudentNo string `json:"student_no" gorm:"column:student_no;type:varchar(255);comment:学号"`

	Email string `json:"email" gorm:"column:email;type:varchar(255);uniqueIndex"`

	// 包含 information 个人介绍、skills 专业能力、reason 为什么要加入实验室、understanding 对竞赛的理解、future_plan 未来计划
	Extra string `json:"extra" gorm:"column:extra;type:json;comment:额外信息"`

	// 通过后自动生成邀请码，只能用于投递的邮箱注册账号
	Code       string `json:"code" gorm:"column:code;type:varchar(255);comment:邀请码"`
	IsAccepted bool   `json:"is_accepted" gorm:"column:is_accepted;type:bool;comment:是否已通过"`
}
