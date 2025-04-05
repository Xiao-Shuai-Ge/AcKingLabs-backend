package model

type User struct {
	ID int64 `json:"id" gorm:"primaryKey;type:bigint;type:bigint;column:id"`
	TimeModel
	Username string `json:"username" gorm:"type:varchar(255);column:'用户名'"`
	Password string `json:"password" gorm:"type:varchar(255);column:'密码'"`
	Email    string `json:"email" gorm:"type:varchar(255);column:'email'"`
}
