package repo

import (
	"gorm.io/gorm"
	"tgwp/model"
)

type PostRepo struct {
	DB *gorm.DB
}

func NewPostRepo(db *gorm.DB) *PostRepo {
	return &PostRepo{
		DB: db,
	}
}

func (r *PostRepo) CreatePost(post model.Post) error {
	return r.DB.Create(&post).Error
}
