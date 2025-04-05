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

func (r *LoginRepo) AddUser(user model.User) error {
	return r.DB.Create(&user).Error
}
