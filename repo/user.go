package repo

import (
	"gorm.io/gorm"
	"tgwp/model"
)

type UserRepo struct {
	DB *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{
		DB: db,
	}
}

// GetUserProfileByID  获取用户信息
func (r *UserRepo) GetUserProfileByID(id int64) (model.User, error) {
	var user model.User
	err := r.DB.Where("id = ?", id).First(&user).Error
	return user, err
}

// UpdateUserProfile  更新用户信息
func (r *UserRepo) UpdateUserProfile(user model.User) error {
	err := r.DB.Save(&user).Error
	return err
}

func (r *UserRepo) SetUserRole(id int64, role int) error {
	err := r.DB.Model(&model.User{}).Where("id = ?", id).Update("role", role).Error
	return err
}

func (r *UserRepo) SetCodeforcesRating(id int64, rating int) error {
	err := r.DB.Model(&model.User{}).Where("id = ?", id).Update("codeforces_rating", rating).Error
	return err
}
