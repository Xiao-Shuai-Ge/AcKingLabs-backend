package repo

import (
	"errors"
	"gorm.io/gorm"
	"tgwp/log/zlog"
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

func (r *PostRepo) ComputePostWeight() (err error) {
	// 权重 = 创建时间+点赞数量(一小时)+评论数量(一小时)+管理员是否点赞(三天)+精选(两周)
	err = r.DB.Model(&model.Post{}).Where("type != ?", "no_weight").Update("weight", gorm.Expr("created_time + (likes * 3600000) + (comments * 3600000) + (is_admin_like * 259200000) + (is_featured * 1209600000)")).Error
	return err
}

func (r *PostRepo) CreatePost(post model.Post) error {
	return r.DB.Create(&post).Error
}

func (r *PostRepo) UpdatePost(post model.Post) error {
	return r.DB.Save(&post).Error
}

func (r *PostRepo) DeletePost(post model.Post) error {
	return r.DB.Delete(&post).Error
}

func (r *PostRepo) GetPostDetail(id int64) (model.Post, error) {
	var post model.Post
	err := r.DB.First(&post, id).Error
	return post, err
}

func (r *PostRepo) IsPostLikeExists(post_id int64, user_id int64) (is_like bool, err error) {
	err = r.DB.Model(&model.PostLike{}).Where("post_id =? AND user_id = ?", post_id, user_id).First(&model.PostLike{}).Error
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

func (r *PostRepo) CancelPostLike(post_id int64, user_id int64) (err error) {
	result := r.DB.Model(&model.PostLike{}).Where("post_id =? AND user_id = ?", post_id, user_id).Delete(&model.PostLike{})
	if result.Error != nil {
		return result.Error
	} else if result.RowsAffected == 0 {
		zlog.Errorf("删除失败: %v", result.Error)
		return nil
	}

	err = r.DB.Model(&model.Post{}).Where("id = ?", post_id).Update("likes", gorm.Expr("likes - ?", 1)).Error
	return err
}

func (r *PostRepo) AddPostLike(postLike model.PostLike) (err error) {
	err = r.DB.Create(&postLike).Error
	if err != nil {
		return err
	}

	err = r.DB.Model(&model.Post{}).Where("id = ?", postLike.PostID).Update("likes", gorm.Expr("likes + ?", 1)).Error
	if err != nil {
		return err
	}
	return
}

func (r *PostRepo) MarkAdminLikePost(post_id int64) error {
	err := r.DB.Model(&model.Post{}).Where("id = ?", post_id).Update("is_admin_like", true).Error
	return err
}

func (r *PostRepo) MarkAdminLikeComment(comment_id int64) error {
	err := r.DB.Model(&model.Comment{}).Where("id = ?", comment_id).Update("is_admin_like", true).Error
	return err
}

func (r *PostRepo) CreateComment(comment model.Comment) error {
	// 创建评论
	err := r.DB.Create(&comment).Error
	if err != nil {
		return err
	}
	// 更新帖子评论数
	err = r.DB.Model(&model.Post{}).Where("id = ?", comment.PostID).Update("comments", gorm.Expr("comments + ?", 1)).Error
	return err
}

func (r *PostRepo) GetMoreComments(post_id int64, before int64, count int) ([]model.Comment, error) {
	var comments []model.Comment
	err := r.DB.Model(&model.Comment{}).Where("post_id = ? AND id < ? AND father_id = 0", post_id, before).Order("id DESC").Limit(count).Find(&comments).Error
	return comments, err
}

func (r *PostRepo) GetMoreChildComments(father_id int64, before int64, count int) ([]model.Comment, error) {
	var comments []model.Comment
	err := r.DB.Model(&model.Comment{}).Where("father_id = ? AND id < ? ", father_id, before).Order("id DESC").Limit(count).Find(&comments).Error
	return comments, err
}

func (r *PostRepo) GetCommentDetail(id int64) (model.Comment, error) {
	var comment model.Comment
	err := r.DB.First(&comment, id).Error
	return comment, err
}

func (r *PostRepo) IsCommentLikeExists(comment_id int64, user_id int64) (is_like bool, err error) {
	err = r.DB.Model(&model.CommentLike{}).Where("comment_id = ? AND user_id = ?", comment_id, user_id).First(&model.CommentLike{}).Error
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

func (r *PostRepo) CancelCommentLike(comment_id int64, user_id int64) (err error) {
	result := r.DB.Model(&model.CommentLike{}).Where("comment_id =? AND user_id = ?", comment_id, user_id).Delete(&model.CommentLike{})
	if result.Error != nil {
		return result.Error
	} else if result.RowsAffected == 0 {
		zlog.Errorf("删除失败: %v", result.Error)
		return nil
	}

	err = r.DB.Model(&model.Comment{}).Where("id = ?", comment_id).Update("likes", gorm.Expr("likes - ?", 1)).Error
	if err != nil {
		return err
	}

	return err
}

func (r *PostRepo) AddCommentLike(commentLike model.CommentLike) (err error) {
	err = r.DB.Create(&commentLike).Error
	if err != nil {
		return err
	}

	err = r.DB.Model(&model.Comment{}).Where("id = ?", commentLike.CommentID).Update("likes", gorm.Expr("likes + ?", 1)).Error
	if err != nil {
		return err
	}

	return err
}

func (r *PostRepo) GetMoreDiaryByWeight(source string, before int64, count int) ([]model.Post, error) {
	var posts []model.Post
	err := r.DB.Model(&model.Post{}).Where("type = 'diary' AND source = ? AND weight < ? ", source, before).Order("weight DESC,id DESC").Limit(count).Find(&posts).Error
	return posts, err
}

func (r *PostRepo) GetMoreDiaryByID(source string, before int64, count int) ([]model.Post, error) {
	var posts []model.Post
	err := r.DB.Model(&model.Post{}).Where("type = 'diary' AND source = ? AND id < ? ", source, before).Order("id DESC").Limit(count).Find(&posts).Error
	return posts, err
}

func (r *PostRepo) GetMoreDiaryByUser(user_id int64, before int64, count int) ([]model.Post, error) {
	var posts []model.Post
	err := r.DB.Model(&model.Post{}).Where("type = 'diary' AND user_id = ? AND id < ? ", user_id, before).Order("id DESC").Limit(count).Find(&posts).Error
	return posts, err
}

func (r *PostRepo) GetPagePostByWeight(post_type string, page int, count int) (posts []model.Post, total int64, err error) {
	offset := (page - 1) * count
	if post_type == "post" {
		err = r.DB.Model(&model.Post{}).Where("type != 'diary'").Order("weight DESC,id DESC").Offset(offset).Limit(count).Find(&posts).Error
		r.DB.Model(&model.Post{}).Where("type != 'diary'").Count(&total)
	} else {
		err = r.DB.Model(&model.Post{}).Where("type = ?", post_type).Order("weight DESC,id DESC").Offset(offset).Limit(count).Find(&posts).Error
		r.DB.Model(&model.Post{}).Where("type = ?", post_type).Count(&total)
	}
	return
}

func (r *PostRepo) GetPagePostByID(post_type string, page int, count int) (posts []model.Post, total int64, err error) {
	offset := (page - 1) * count
	if post_type == "post" {
		err = r.DB.Model(&model.Post{}).Where("type != 'diary'").Order("id DESC").Offset(offset).Limit(count).Find(&posts).Error
		r.DB.Model(&model.Post{}).Where("type != 'diary'").Count(&total)
	} else {
		err = r.DB.Model(&model.Post{}).Where("type = ?", post_type).Order("id DESC").Offset(offset).Limit(count).Find(&posts).Error
		r.DB.Model(&model.Post{}).Where("type = ?", post_type).Count(&total)
	}

	return
}

func (r *PostRepo) GetPagePostByFeatured(post_type string, page int, count int) (posts []model.Post, total int64, err error) {
	offset := (page - 1) * count
	if post_type == "post" {
		err = r.DB.Model(&model.Post{}).Where("type != 'diary' AND is_featured = 1").Order("id DESC").Offset(offset).Limit(count).Find(&posts).Error
		r.DB.Model(&model.Post{}).Where("type != 'diary' AND is_featured = 1").Count(&total)
	} else {
		err = r.DB.Model(&model.Post{}).Where("type = ? AND is_featured = 1", post_type).Order("id DESC").Offset(offset).Limit(count).Find(&posts).Error
		r.DB.Model(&model.Post{}).Where("type = ? AND is_featured = 1", post_type).Count(&total)
	}
	return
}

func (r *PostRepo) GetPagePostBySource(post_type string, source string, page int, count int) (posts []model.Post, total int64, err error) {
	offset := (page - 1) * count
	if post_type == "post" {
		err = r.DB.Model(&model.Post{}).Where("type != 'diary' AND source = ?", source).Order("weight DESC,id DESC").Offset(offset).Limit(count).Find(&posts).Error
		r.DB.Model(&model.Post{}).Where("type != 'diary' AND source = ?", source).Count(&total)
	} else {
		err = r.DB.Model(&model.Post{}).Where("type = ? AND source = ?", post_type, source).Order("weight DESC,id DESC").Offset(offset).Limit(count).Find(&posts).Error
		r.DB.Model(&model.Post{}).Where("type = ? AND source = ?", post_type, source).Count(&total)
	}
	return
}

func (r *PostRepo) SetPostFeature(post_id int64) (err error) {
	result := r.DB.Model(&model.Post{}).Where("id = ? AND is_featured = 0", post_id).Update("is_featured", true)
	err = result.Error
	if result.RowsAffected == 0 {
		zlog.Errorf("已经是精选状态: %v", err)
		err = errors.New("已经是精选状态")
	}
	return err
}

func (r *PostRepo) ExistDiary(user_id int64, source string) (exist bool, err error) {
	var count int64
	err = r.DB.Model(&model.Post{}).Where("type = 'diary' AND user_id = ? AND source = ?", user_id, source).Count(&count).Error
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

func (r *PostRepo) GetDiaryList(user_id int64) (posts []model.Post, err error) {
	err = r.DB.Model(&model.Post{}).Where("type = 'diary' AND user_id = ?", user_id).Order("id DESC").Find(&posts).Error
	return posts, err
}

func (r *PostRepo) GetPostAfterUpdateTime(update_time int64) (posts []model.Post, err error) {
	err = r.DB.Model(&model.Post{}).Where("updated_time > ?", update_time).Find(&posts).Error
	return posts, err
}
