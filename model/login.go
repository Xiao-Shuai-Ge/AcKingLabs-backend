package model

type User struct {
	ID int64 `json:"id" gorm:"column:id;primaryKey;type:bigint;type:bigint"`
	TimeModel
	Username string `json:"username" gorm:"column:username;type:varchar(255);comment:用户名"`
	Password string `json:"password" gorm:"column:password;type:varchar(255);comment:密码"`
	Email    string `json:"email" gorm:"column:email;type:varchar(255);comment:邮箱"`
}
