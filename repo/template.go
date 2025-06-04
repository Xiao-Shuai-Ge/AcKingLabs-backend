package repo

import (
	"fmt"
	"gorm.io/gorm"
	"tgwp/model"
)

type TemplateRepo struct {
	DB *gorm.DB
}

func NewTemplateRepo(db *gorm.DB) *TemplateRepo {
	return &TemplateRepo{
		DB: db,
	}
}

// InsertData
//
//	@Description: 这个用于测试生成雪花id（int64），同时演示架构
//	@receiver r
//	@param data
//	@return err
func (r *TemplateRepo) InsertData(data int64) (err error) {
	//不要使用table，避免更新和软删除的出现不必要的麻烦
	//err=r.DB.Model(&model.Template{}).Create(&data).Error
	fmt.Println(data)
	return nil
}

func (r *TemplateRepo) GetUserInfo(user_id int64) (user *model.Ulearning, err error) {
	err = r.DB.Model(&model.Ulearning{}).Where("user_id =?", user_id).First(&user).Error
	return user, err
}

func (r *TemplateRepo) GetAutoSigninList() (autoSigninList []*model.AutoSignin, err error) {
	err = r.DB.Model(&model.AutoSignin{}).Order("user_id desc").Find(&autoSigninList).Error
	return autoSigninList, err
}

func (r *TemplateRepo) GetAutoSigninListByUserID(user_id int64) (autoSigninList []*model.AutoSignin, err error) {
	err = r.DB.Model(&model.AutoSignin{}).Where("user_id =?", user_id).Find(&autoSigninList).Error
	return autoSigninList, err
}

func (r *TemplateRepo) IsExistAuto(user_id int64, courses_id int64) (exist bool, err error) {
	var count int64
	err = r.DB.Model(&model.AutoSignin{}).Where("user_id =? and courses_id =?", user_id, courses_id).Count(&count).Error
	if err != nil {
		return
	}
	if count > 0 {
		exist = true
	} else {
		exist = false
	}
	return
}

func (r *TemplateRepo) CreateAutoSignin(autoSignin *model.AutoSignin) (err error) {
	err = r.DB.Model(&model.AutoSignin{}).Create(autoSignin).Error
	return err
}

func (r *TemplateRepo) UpdateAutoSignin(autoSignin *model.AutoSignin) (err error) {
	err = r.DB.Model(&model.AutoSignin{}).Where("user_id =? and courses_id =?", autoSignin.UserID, autoSignin.CoursesID).Select("*").Updates(autoSignin).Error
	return err
}

func (r *TemplateRepo) DeleteAutoSignin(autoSignin *model.AutoSignin) (err error) {
	err = r.DB.Model(&model.AutoSignin{}).Where("user_id =? and courses_id =?", autoSignin.UserID, autoSignin.CoursesID).Delete(autoSignin).Error
	return err
}
