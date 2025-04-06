package repo

import (
	"gorm.io/gorm"
	"tgwp/model"
)

type LoginRepo struct {
	DB *gorm.DB
}

func NewLoginRepo(db *gorm.DB) *LoginRepo {
	return &LoginRepo{
		DB: db,
	}
}

// AddUser 新增用户
func (r *LoginRepo) AddUser(user model.User) error {
	return r.DB.Create(&user).Error
}

// GetUserByEmail 根据邮箱获取用户信息
func (r *LoginRepo) GetUserByEmail(email string) (model.User, error) {
	var user model.User
	err := r.DB.Where("email = ?", email).First(&user).Error
	return user, err
}
