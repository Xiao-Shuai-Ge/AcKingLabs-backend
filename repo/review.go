package repo

import (
	"tgwp/model"

	"gorm.io/gorm"
)

type ReviewRepo struct {
	DB *gorm.DB
}

func NewReviewRepo(db *gorm.DB) *ReviewRepo {
	return &ReviewRepo{
		DB: db,
	}
}

// CreateReview 创建审核记录
func (r *ReviewRepo) CreateReview(review model.Review) error {
	return r.DB.Create(&review).Error
}

func (r *ReviewRepo) GetReviewList(offset, limit int) ([]model.Review, int64, error) {
	var reviews []model.Review
	var count int64
	db := r.DB.Model(&model.Review{}) //.Where("status = ?", 0)
	err := db.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}
	err = db.Order("created_time desc").Offset(offset).Limit(limit).Find(&reviews).Error
	return reviews, count, err
}

func (r *ReviewRepo) UpdateReview(review model.Review) error {
	return r.DB.Save(&review).Error
}

func (r *ReviewRepo) GetReviewByID(id int64) (model.Review, error) {
	var review model.Review
	err := r.DB.First(&review, id).Error
	return review, err
}
