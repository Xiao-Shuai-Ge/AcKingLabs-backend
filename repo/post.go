package repo

import (
	"errors"
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

func (r *PostRepo) GetPostDetail(id int64) (model.Post, error) {
	var post model.Post
	err := r.DB.First(&post, id).Error
	return post, err
}

func (r *PostRepo) IsPostLikeExists(post_id int64, user_id int64) (is_like bool, err error) {
	var postLike model.PostLike
	err = r.DB.Model(&model.PostLike{}).Where("post_id =? AND user_id = ?", post_id, user_id).First(&postLike).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		} else {
			return false, err
		}
	} else {
		return true, nil
	}
}

func (r *PostRepo) CancelPostLike(post_id int64, user_id int64) error {
	var postLike model.PostLike
	err := r.DB.Model(&model.PostLike{}).Where("post_id =? AND user_id = ?", post_id, user_id).Delete(&postLike).Error
	if err != nil {
		return err
	}
	err = r.DB.Model(&model.Post{}).Where("id = ?", post_id).Update("likes", gorm.Expr("likes - ?", 1)).Error
	return err
}

func (r *PostRepo) AddPostLike(postLike model.PostLike) error {
	err := r.DB.Model(&model.Post{}).Where("id = ?", postLike.PostID).Update("likes", gorm.Expr("likes + ?", 1)).Error
	if err != nil {
		return err
	}
	return r.DB.Create(&postLike).Error
}

func (r *PostRepo) MarkAdminLike(post_id int64) error {
	err := r.DB.Model(&model.Post{}).Where("id = ?", post_id).Update("is_admin_like", true).Error
	return err
}
