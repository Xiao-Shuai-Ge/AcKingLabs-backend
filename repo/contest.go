package repo

import (
	"gorm.io/gorm"
	"tgwp/log/zlog"
	"tgwp/model"
)

type ContestRepo struct {
	DB *gorm.DB
}

func NewContestRepo(db *gorm.DB) *ContestRepo {
	return &ContestRepo{
		DB: db,
	}
}

func (r *ContestRepo) IsContestExists(Url string) (is_exists bool, err error) {
	var count int64
	err = r.DB.Model(&model.Contest{}).Where("url =?", Url).Count(&count).Error
	if err != nil {
		return true, err
	}
	if count > 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func (r *ContestRepo) CreateContest(contest model.Contest) error {
	return r.DB.Create(&contest).Error
}

func (r *ContestRepo) UpdateContest(contest model.Contest) error {
	// 按 Url 字段进行更新，不更新 ID 字段
	return r.DB.Model(&model.Contest{}).Where("url = ?", contest.Url).Omit("ID").Updates(&contest).Error
}

func (r *ContestRepo) GetContestList(contest_type string, page int, count int) (contests []model.Contest, total int64, err error) {
	offset := (page - 1) * count
	zlog.Infof("offset: %d, count: %d", offset, count)
	if contest_type == "all" {
		err = r.DB.Model(&model.Contest{}).Order("end_time DESC").Offset(offset).Limit(count).Find(&contests).Error
		r.DB.Model(&model.Contest{}).Count(&total)
	} else {
		err = r.DB.Model(&model.Contest{}).Where("platform = ?", contest_type).Order("end_time DESC").Offset(offset).Limit(count).Find(&contests).Error
		r.DB.Model(&model.Contest{}).Where("platform = ?", contest_type).Count(&total)
	}
	return
}
