package repo

import (
	"gorm.io/gorm"
	"tgwp/log/zlog"
	"tgwp/model"
	"time"
)

type ContestRepo struct {
	DB *gorm.DB
}

func NewContestRepo(db *gorm.DB) *ContestRepo {
	return &ContestRepo{
		DB: db,
	}
}

func (r *ContestRepo) GetContestByID(ID int64) (contest model.Contest, err error) {
	err = r.DB.Where("id = ?", ID).First(&contest).Error
	return
}

func (r *ContestRepo) IsContestExistsByUrl(Url string) (is_exists bool, err error) {
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

func (r *ContestRepo) IsContestExistsByID(ID int64) (is_exists bool, err error) {
	var count int64
	err = r.DB.Model(&model.Contest{}).Where("id =?", ID).Count(&count).Error
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
		err = r.DB.Model(&model.Contest{}).Order("start_time DESC").Offset(offset).Limit(count).Find(&contests).Error
		r.DB.Model(&model.Contest{}).Count(&total)
	} else {
		err = r.DB.Model(&model.Contest{}).Where("platform = ?", contest_type).Order("start_time DESC").Offset(offset).Limit(count).Find(&contests).Error
		r.DB.Model(&model.Contest{}).Where("platform = ?", contest_type).Count(&total)
	}
	return
}

func (r *ContestRepo) GetContestListByRecommend(page int, count int) (contests []model.Contest, total int64, err error) {
	offset := (page - 1) * count
	err = r.DB.Model(&model.Contest{}).Where("is_recommend = ?", true).Order("start_time DESC").Offset(offset).Limit(count).Find(&contests).Error
	r.DB.Model(&model.Contest{}).Where("is_recommend = ?", true).Count(&total)
	return
}

func (r *ContestRepo) IsBooking(contestID int64, userID int64) (is_exists bool, err error) {
	var count int64
	err = r.DB.Model(&model.Booking{}).Where("contest_id =? AND user_id = ?", contestID, userID).Count(&count).Error
	if err != nil {
		return true, err
	}
	if count > 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func (r *ContestRepo) CreateBooking(booking model.Booking) error {
	return r.DB.Create(&booking).Error
}

func (r *ContestRepo) RemoveBooking(contestID int64, userID int64) error {
	return r.DB.Where("contest_id = ? AND user_id = ?", contestID, userID).Delete(&model.Booking{}).Error
}

func (r *ContestRepo) GetUnStartedContestList() (contests []model.Contest, err error) {
	err = r.DB.Model(&model.Contest{}).Where("start_time > ?", time.Now().UnixMilli()).Find(&contests).Error
	return
}

func (r *ContestRepo) GetBookingListByContestID(contestID int64) (bookings []model.Booking, err error) {
	err = r.DB.Where("contest_id = ?", contestID).Find(&bookings).Error
	return
}

func (r *ContestRepo) RemoveBookingByContestID(contestID int64) error {
	return r.DB.Where("contest_id = ?", contestID).Delete(&model.Booking{}).Error
}

func (r *ContestRepo) SetContestRecommend(contestID int64, isRecommend bool) error {
	return r.DB.Model(&model.Contest{}).Where("id = ?", contestID).Update("is_recommend", isRecommend).Error
}
