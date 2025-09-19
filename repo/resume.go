package repo

import (
	"tgwp/model"

	"gorm.io/gorm"
)

type ResumeRepo struct {
	DB *gorm.DB
}

func NewResumeRepo(db *gorm.DB) *ResumeRepo {
	return &ResumeRepo{
		DB: db,
	}
}

// CreateResume 创建简历
func (r *ResumeRepo) CreateResume(resume model.Resume) error {
	err := r.DB.Create(&resume).Error
	return err
}

// GetResumeByID 根据ID获取简历
func (r *ResumeRepo) GetResumeByID(id int64) (model.Resume, error) {
	var resume model.Resume
	err := r.DB.Where("id = ?", id).First(&resume).Error
	return resume, err
}

// GetResumeByEmail 根据邮箱获取简历
func (r *ResumeRepo) GetResumeByEmail(email string) (model.Resume, error) {
	var resume model.Resume
	err := r.DB.Where("email = ?", email).First(&resume).Error
	return resume, err
}

// UpdateResume 更新简历
func (r *ResumeRepo) UpdateResume(resume model.Resume) error {
	err := r.DB.Save(&resume).Error
	return err
}

// DeleteResume 删除简历
func (r *ResumeRepo) DeleteResume(id int64) error {
	err := r.DB.Where("id = ?", id).Delete(&model.Resume{}).Error
	return err
}

// GetResumeList 获取简历列表（按ID排序分页）
func (r *ResumeRepo) GetResumeList(page int, count int) ([]model.Resume, int64, error) {
	var resumes []model.Resume
	offset := (page - 1) * count
	err := r.DB.Model(&model.Resume{}).Order("id ASC").Offset(offset).Limit(count).Find(&resumes).Error
	var total int64
	r.DB.Model(&model.Resume{}).Count(&total)
	return resumes, total, err
}

// AcceptResume 通过简历（设置邀请码和通过状态）
func (r *ResumeRepo) AcceptResume(id int64, code string) error {
	err := r.DB.Model(&model.Resume{}).Where("id = ?", id).Updates(map[string]interface{}{
		"code":   code,
		"status": 1, // 1表示已通过
	}).Error
	return err
}

// RejectResume 不通过简历（设置不通过状态）
func (r *ResumeRepo) RejectResume(id int64) error {
	err := r.DB.Model(&model.Resume{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status": -1, // -1表示未通过
	}).Error
	return err
}
