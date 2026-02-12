package repo

import (
	"tgwp/model"

	"gorm.io/gorm"
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

// SetUserRole  设置用户角色
func (r *UserRepo) SetUserRole(id int64, role int) error {
	err := r.DB.Model(&model.User{}).Where("id = ?", id).Update("role", role).Error
	return err
}

// SetCodeforcesRating  设置用户codeforces rating
func (r *UserRepo) SetCodeforcesRating(id int64, rating int) error {
	err := r.DB.Model(&model.User{}).Where("id = ?", id).Update("codeforces_rating", rating).Error
	return err
}

func (r *UserRepo) AddUserXp(id int64, xp int) error {
	err := r.DB.Model(&model.User{}).Where("id = ?", id).Update("xp", gorm.Expr("xp + ?", xp)).Error
	return err
}

func (r *UserRepo) GetRankings(page int, count int) ([]model.User, int64, error) {
	var users []model.User
	offset := (page - 1) * count
	err := r.DB.Model(&model.User{}).Order("xp DESC").Offset(offset).Limit(count).Find(&users).Error
	var total int64
	r.DB.Model(&model.User{}).Count(&total)
	return users, total, err
}

// DeleteUser 删除用户
func (r *UserRepo) DeleteUser(id int64) error {
	err := r.DB.Where("id = ?", id).Delete(&model.User{}).Error
	return err
}

// GetUserList 获取用户列表（按ID排序分页）
func (r *UserRepo) GetUserList(page int, count int, keyword string) ([]model.User, int64, error) {
	var users []model.User
	offset := (page - 1) * count

	db := r.DB.Model(&model.User{})
	if keyword != "" {
		likePattern := "%" + keyword + "%"
		db = db.Where("username LIKE ? OR email LIKE ? OR real_name LIKE ? OR student_no LIKE ?", likePattern, likePattern, likePattern, likePattern)
	}

	var total int64
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Order("id ASC").Offset(offset).Limit(count).Find(&users).Error
	return users, total, err
}

// SearchUsers 搜索用户（公开）
func (r *UserRepo) SearchUsers(page int, count int, keyword string) ([]model.User, int64, error) {
	var users []model.User
	offset := (page - 1) * count

	db := r.DB.Model(&model.User{})
	if keyword != "" {
		likePattern := "%" + keyword + "%"
		db = db.Where("username LIKE ?", likePattern)
	}

	var total int64
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = db.Order("xp DESC").Offset(offset).Limit(count).Find(&users).Error
	return users, total, err
}
