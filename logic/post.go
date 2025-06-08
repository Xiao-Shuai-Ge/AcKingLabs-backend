package logic

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"tgwp/global"
	"tgwp/log/zlog"
	"tgwp/model"
	"tgwp/repo"
	"tgwp/response"
	"tgwp/types"
	"tgwp/utils"
	"tgwp/utils/elasticSearchUtils"
	"time"
	"unicode/utf8"
)

const (
	REDIS_LIKE_MESSAGE = "like_message:%d:%d"
)

type PostLogic struct {
}

func NewPostLogic() *PostLogic {
	return &PostLogic{}
}

// CreatePost 创建帖子
func (l *PostLogic) CreatePost(ctx context.Context, req types.CreatePostReq) (resp types.CreatePostResp, err error) {
	defer utils.RecordTime(time.Now())()
	// ID 转化为 int64
	userID, err := strconv.ParseInt(req.UserID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.UserID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	if req.Type == "diary" {
		// 如果是周记打卡，先检查时间是否正确
		req.Source = GetWeekCode()
		if len(req.Source) == 0 {
			zlog.CtxErrorf(ctx, "周记打卡时间错误: %v", err)
			return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
		}
		zlog.CtxInfof(ctx, "解析出打卡周数: %s", req.Source)
		// 判断周记打卡是否已经存在
		var exist bool
		exist, err = repo.NewPostRepo(global.DB).ExistDiary(userID, req.Source)
		if err != nil {
			zlog.CtxErrorf(ctx, "查询周记打卡是否存在失败: %v", err)
			return resp, response.ErrResp(err, response.DATABASE_ERROR)
		}
		if exist {
			zlog.CtxErrorf(ctx, "周记打卡已经存在: %v", err)
			return resp, response.ErrResp(err, response.DIARY_ALREADY_EXIST)
		}
	}
	// 判断数据范围
	// 1. 标题不能超过 30 个字符
	if utf8.RuneCountInString(req.Title) > 30 {
		zlog.CtxErrorf(ctx, "标题不能超过 30 个字符: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 2. 内容不能超过 20000 个字符
	zlog.CtxInfof(ctx, "内容长度: %d", utf8.RuneCountInString(req.Content))
	if utf8.RuneCountInString(req.Content) > 20000 {
		zlog.CtxErrorf(ctx, "内容不能超过 20000 个字: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 3. 除了周记打卡可以私密，其他类型都不可以私密
	if req.Type != "diary" && req.IsPrivate {
		zlog.CtxErrorf(ctx, "非周记打卡不能私密: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 4. 不允许出现不存在的类型
	if !global.TYPE_SET[req.Type] {
		zlog.CtxErrorf(ctx, "不存在的类型: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 创建帖子
	id := global.SnowflakeNode.Generate().Int64()
	post := model.Post{
		ID:        id,
		UserID:    userID,
		Title:     req.Title,
		Content:   req.Content,
		Type:      req.Type,
		Source:    req.Source,
		IsPrivate: req.IsPrivate,
		Weight:    time.Now().UnixMilli(),
	}
	err = repo.NewPostRepo(global.DB).CreatePost(post)
	if err != nil {
		zlog.CtxErrorf(ctx, "创建帖子失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	resp.ID = id
	// 给作者加 XP
	addXp := 4
	if req.IsPrivate == false {
		addXp = 8
	}
	err = repo.NewUserRepo(global.DB).AddUserXp(userID, addXp)
	if err != nil {
		zlog.CtxErrorf(ctx, "给作者加 XP 失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}

	return
}

// EditPost 编辑帖子
func (l *PostLogic) EditPost(ctx context.Context, req types.EditPostReq) (resp types.EditPostResp, err error) {
	defer utils.RecordTime(time.Now())()
	// ID 转化为 int64
	postID, err := strconv.ParseInt(req.PostID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.PostID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	operatorID, err := strconv.ParseInt(req.OperatorID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.OperatorID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 判断数据范围
	// 1. 标题不能超过 30 个字符
	if utf8.RuneCountInString(req.Title) > 30 {
		zlog.CtxErrorf(ctx, "标题不能超过 30 个字符: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 2. 内容不能超过 5000 个字符
	zlog.CtxInfof(ctx, "内容长度: %d", utf8.RuneCountInString(req.Content))
	if utf8.RuneCountInString(req.Content) > 20000 {
		zlog.CtxErrorf(ctx, "内容不能超过 20000 个字: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 3. 除了周记打卡可以私密，其他类型都不可以私密
	if req.Type != "diary" && req.IsPrivate {
		zlog.CtxErrorf(ctx, "非周记打卡不能私密: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 4. 不允许出现不存在的类型
	if !global.TYPE_SET[req.Type] {
		zlog.CtxErrorf(ctx, "不存在的类型: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 拿取原帖子
	post, err := repo.NewPostRepo(global.DB).GetPostDetail(postID)
	if err != nil {
		zlog.CtxErrorf(ctx, "查询原帖子失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	// 判断是否有权限编辑
	if post.UserID != operatorID && !(req.OperatorRole >= global.ROLE_ADMIN) {
		zlog.CtxErrorf(ctx, "非作者或管理员无权编辑帖子: %v", err)
		return resp, response.ErrResp(err, response.PERMISSION_DENIED)
	}
	// 更新帖子
	post.Title = req.Title
	post.Content = req.Content
	post.Type = req.Type
	post.Source = req.Source
	post.IsPrivate = req.IsPrivate
	err = repo.NewPostRepo(global.DB).UpdatePost(post)
	if err != nil {
		zlog.CtxErrorf(ctx, "创建帖子失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	return
}

// DeletePost 编辑帖子
func (l *PostLogic) DeletePost(ctx context.Context, req types.DeletePostReq) (resp types.DeletePostResp, err error) {
	defer utils.RecordTime(time.Now())()
	// ID 转化为 int64
	postID, err := strconv.ParseInt(req.PostID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.PostID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	operatorID, err := strconv.ParseInt(req.OperatorID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.OperatorID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 拿取原帖子
	post, err := repo.NewPostRepo(global.DB).GetPostDetail(postID)
	if err != nil {
		zlog.CtxErrorf(ctx, "查询原帖子失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	// 判断是否有权限删除
	if post.UserID != operatorID && !(req.OperatorRole >= global.ROLE_ADMIN) {
		zlog.CtxErrorf(ctx, "非作者或管理员无权编辑帖子: %v", err)
		return resp, response.ErrResp(err, response.PERMISSION_DENIED)
	}
	// 如果是周记打卡，不允许用户自己删除
	if post.Type == "diary" && req.OperatorRole < global.ROLE_ADMIN {
		zlog.CtxErrorf(ctx, "周记打卡不允许用户自己删除: %v", err)
		return resp, response.ErrResp(err, response.DIARY_CANT_DELETE)
	}
	// 删除帖子
	err = repo.NewPostRepo(global.DB).DeletePost(post)
	if err != nil {
		zlog.CtxErrorf(ctx, "删除帖子失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	return
}

func (l *PostLogic) GetPostDetail(ctx context.Context, req types.GetPostDetailReq) (resp types.GetPostDetailResp, err error) {
	defer utils.RecordTime(time.Now())()
	// ID 转化为 int64
	postID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.ID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 查询帖子详情
	post, err := repo.NewPostRepo(global.DB).GetPostDetail(postID)
	if err != nil {
		zlog.CtxErrorf(ctx, "查询帖子详情失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	// 判断是否有权限查看
	if post.IsPrivate {
		// 如果是私密，只有自己和管理员可以查看
		if post.UserID != postID && !(req.OperatorRole >= global.ROLE_ADMIN) {
			zlog.CtxErrorf(ctx, "非作者或管理员无权查看私密帖子: %v", err)
			return resp, response.ErrResp(err, response.PERMISSION_DENIED)
		}
	}
	// 转换为响应结构
	resp.ID = post.ID
	resp.UserID = post.UserID
	resp.Title = post.Title
	resp.Content = post.Content
	resp.Type = post.Type
	resp.Source = post.Source
	resp.Likes = post.Likes
	resp.Comments = post.Comments
	resp.CreatedAt = post.CreatedTime
	resp.UpdatedAt = post.UpdatedTime

	resp.IsAdminLike = post.IsAdminLike
	resp.IsPrivate = post.IsPrivate
	resp.IsFeatured = post.IsFeatured
	return
}

func (l *PostLogic) LikePost(ctx context.Context, req types.LikePostReq) (resp types.LikePostResp, err error) {
	defer utils.RecordTime(time.Now())()
	// ID 转化为 int64
	postID, err := strconv.ParseInt(req.PostID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.PostID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	operatorID, err := strconv.ParseInt(req.OperatorID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.OperatorID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 判断是否已经点赞
	isLike, err := repo.NewPostRepo(global.DB).IsPostLikeExists(postID, operatorID)
	if err != nil {
		zlog.CtxErrorf(ctx, "查询点赞状态失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	if isLike {
		// 已经点赞，取消点赞
		err = repo.NewPostRepo(global.DB).CancelPostLike(postID, operatorID)
		if err != nil {
			zlog.CtxErrorf(ctx, "取消点赞失败: %v", err)
			return resp, response.ErrResp(err, response.DATABASE_ERROR)
		}
		resp.IsLike = false
		return
	} else {
		// 先判断是否为管理员点赞
		if req.OperatorRole >= global.ROLE_ADMIN {
			// 判断是否已有管理员点赞，如果第一次有管理员点赞，应该为用户增加经验
			var post model.Post
			post, err = repo.NewPostRepo(global.DB).GetPostDetail(postID)
			if err != nil {
				zlog.CtxErrorf(ctx, "查询帖子详情失败: %v", err)
				return resp, response.ErrResp(err, response.DATABASE_ERROR)
			}
			if !post.IsAdminLike {
				// 第一次有管理员点赞，增加经验 +5，并标记为管理员点赞
				err = repo.NewPostRepo(global.DB).MarkAdminLikePost(postID)
				if err != nil {
					zlog.CtxErrorf(ctx, "标记管理员点赞失败: %v", err)
					return resp, response.ErrResp(err, response.DATABASE_ERROR)
				}
				// 增加经验
				err = repo.NewUserRepo(global.DB).AddUserXp(post.UserID, 5)
				if err != nil {
					zlog.CtxErrorf(ctx, "增加经验失败: %v", err)
					return resp, response.ErrResp(err, response.DATABASE_ERROR)
				}
			}
		}
		// 点赞
		id := global.SnowflakeNode.Generate().Int64()
		postLike := model.PostLike{
			PostID: postID,
			UserID: operatorID,
			ID:     id,
		}
		err = repo.NewPostRepo(global.DB).AddPostLike(postLike)
		if err != nil {
			zlog.CtxErrorf(ctx, "点赞失败: %v", err)
			return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
		}
		resp.IsLike = true
	}
	// 发送点赞通知
	// 先用redis判断两小时内是否有过点赞通知，如果有，则不再发送
	key := fmt.Sprintf(REDIS_LIKE_MESSAGE, postID, operatorID)
	if global.Rdb.Exists(ctx, key).Val() == 1 {
		zlog.CtxInfof(ctx, "两小时内有过点赞通知，不再发送")
		return
	}
	// redis 记录点赞通知
	err = global.Rdb.Set(ctx, key, "1", time.Hour*2).Err()
	if err != nil {
		zlog.CtxErrorf(ctx, "%v", err)
		return resp, response.ErrResp(err, response.REDIS_ERROR)
	}
	// 获取帖子详情
	var post model.Post
	post, err = repo.NewPostRepo(global.DB).GetPostDetail(postID)
	if err != nil {
		zlog.CtxErrorf(ctx, "查询帖子详情失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	// 发送通知
	messageID := global.SnowflakeNode.Generate().Int64()
	var url string
	if post.Type == "diary" {
		url = fmt.Sprintf("/diary/%d", post.ID)
	} else {
		url = fmt.Sprintf("/learn/%d", post.ID)
	}
	message := model.Message{
		ID:       messageID,
		UserID:   post.UserID,
		SenderID: operatorID,
		Type:     "like",
		Content:  fmt.Sprintf("赞了你的帖子 《%s》", post.Title),
		Url:      url,
		IsRead:   false,
	}
	err = repo.NewMessageRepo(global.DB).SendMessage(message)
	if err != nil {
		zlog.CtxErrorf(ctx, "发送点赞通知失败: %v", err)
	}
	return
}

func (l *PostLogic) GetLikePost(ctx context.Context, req types.GetLikePostReq) (resp types.GetLikePostResp, err error) {
	defer utils.RecordTime(time.Now())()
	// ID 转化为 int64
	postID, err := strconv.ParseInt(req.PostID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.PostID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	operatorID, err := strconv.ParseInt(req.OperatorID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.OperatorID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 查询是否有点赞记录
	IsLike, err := repo.NewPostRepo(global.DB).IsPostLikeExists(postID, operatorID)
	if err != nil {
		zlog.CtxErrorf(ctx, "查询点赞状态失败: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	resp.IsLike = IsLike
	return
}

func (l *PostLogic) CreateComment(ctx context.Context, req types.CreateCommentReq) (resp types.CreateCommentResp, err error) {
	defer utils.RecordTime(time.Now())()
	// ID 转化为 int64
	postID, err := strconv.ParseInt(req.PostID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.PostID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	fatherID, err := strconv.ParseInt(req.FatherID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.FatherID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	userID, err := strconv.ParseInt(req.UserID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.UserID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 判断内容长度
	if utf8.RuneCountInString(req.Content) > 1000 {
		zlog.CtxErrorf(ctx, "评论内容不能超过 1000 个字符: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 获取帖子详情，如果帖子不存在，则返回错误
	var post model.Post
	post, err = repo.NewPostRepo(global.DB).GetPostDetail(postID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		zlog.CtxErrorf(ctx, "帖子不存在: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	} else if err != nil {
		zlog.CtxErrorf(ctx, "查询帖子详情失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	// 创建评论
	id := global.SnowflakeNode.Generate().Int64()
	comment := model.Comment{
		ID:       id,
		PostID:   postID,
		FatherID: fatherID,
		UserID:   userID,
		Content:  req.Content,
		Likes:    0,
	}
	err = repo.NewPostRepo(global.DB).CreateComment(comment)
	if err != nil {
		zlog.CtxErrorf(ctx, "创建评论失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	resp.ID = id
	// 发送评论通知
	// 简化评论内容 (去掉换行符)
	contentShort := comment.Content
	contentShort = strings.ReplaceAll(contentShort, "\n", " ")
	contentShort = utils.TruncateString(contentShort, 20)
	// 发送通知
	messageID := global.SnowflakeNode.Generate().Int64()
	var url string
	if post.Type == "diary" {
		url = fmt.Sprintf("/diary/%d", post.ID)
	} else {
		url = fmt.Sprintf("/learn/%d", post.ID)
	}
	// 判断是几级评论，给出对应的提示
	var content string
	var receiverID int64
	if fatherID == 0 {
		content = fmt.Sprintf("在你的帖子 《%s》 评论了: [%s]", post.Title, contentShort)
		receiverID = post.UserID
	} else {
		var fatherComment model.Comment
		fatherComment, err = repo.NewPostRepo(global.DB).GetCommentDetail(fatherID)
		if err != nil {
			zlog.CtxErrorf(ctx, "查询父评论详情失败: %v", err)
			return resp, response.ErrResp(err, response.DATABASE_ERROR)
		}
		fatherContentShort := fatherComment.Content
		fatherContentShort = strings.ReplaceAll(fatherContentShort, "\n", " ")
		fatherContentShort = utils.TruncateString(fatherContentShort, 20)
		content = fmt.Sprintf("在你的评论 [%s] 回复了: [%s]", fatherContentShort, contentShort)
		receiverID = fatherComment.UserID
	}
	// 发送通知
	message := model.Message{
		ID:       messageID,
		UserID:   receiverID,
		SenderID: userID,
		Type:     "comment",
		Content:  content,
		Url:      url,
		IsRead:   false,
	}
	err = repo.NewMessageRepo(global.DB).SendMessage(message)
	if err != nil {
		zlog.CtxErrorf(ctx, "发送评论通知失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	return
}

func (l *PostLogic) GetMoreComments(ctx context.Context, req types.GetMoreCommentsReq) (resp types.GetMoreCommentsResp, err error) {
	defer utils.RecordTime(time.Now())()
	// ID 转化为 int64
	ID, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.ID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	beforeID, err := strconv.ParseInt(req.BeforeID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.BeforeID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 从数据库中查询评论
	var comments []model.Comment
	if req.IsChild {
		comments, err = repo.NewPostRepo(global.DB).GetMoreChildComments(ID, beforeID, req.Count)
	} else {
		comments, err = repo.NewPostRepo(global.DB).GetMoreComments(ID, beforeID, req.Count)
	}

	if err != nil {
		zlog.CtxErrorf(ctx, "查询评论失败: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	//zlog.CtxDebugf(ctx, "查询评论成功: %v", comments)
	for _, comment := range comments {
		resp.Comments = append(resp.Comments, types.Comment{
			ID:          comment.ID,
			UserID:      comment.UserID,
			Content:     comment.Content,
			Likes:       comment.Likes,
			CreatedAt:   comment.CreatedTime,
			IsAdminLike: comment.IsAdminLike,
		})
	}
	resp.Length = len(resp.Comments)
	return
}

func (l *PostLogic) LikeComment(ctx context.Context, req types.LikeCommentReq) (resp types.LikeCommentResp, err error) {
	defer utils.RecordTime(time.Now())()
	// ID 转化为 int64
	commentID, err := strconv.ParseInt(req.CommentID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.CommentID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	operatorID, err := strconv.ParseInt(req.OperatorID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.OperatorID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 判断是否已经点赞
	isLike, err := repo.NewPostRepo(global.DB).IsCommentLikeExists(commentID, operatorID)
	if err != nil {
		zlog.CtxErrorf(ctx, "查询点赞状态失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	if isLike {
		zlog.CtxDebugf(ctx, "取消点赞")
		err = repo.NewPostRepo(global.DB).CancelCommentLike(commentID, operatorID)
		if err != nil {
			zlog.CtxErrorf(ctx, "取消点赞失败: %v", err)
			return resp, response.ErrResp(err, response.DATABASE_ERROR)
		}
		resp.IsLike = false
		return
	} else {
		// 先判断是否为管理员点赞
		if req.OperatorRole >= global.ROLE_ADMIN {
			// 判断是否已有管理员点赞，如果第一次有管理员点赞，应该为用户增加经验
			var comment model.Comment
			comment, err = repo.NewPostRepo(global.DB).GetCommentDetail(commentID)
			if err != nil {
				zlog.CtxErrorf(ctx, "查询帖子详情失败: %v", err)
				return resp, response.ErrResp(err, response.DATABASE_ERROR)
			}
			if !comment.IsAdminLike {
				// 第一次有管理员点赞，增加经验 +2，并标记为管理员点赞
				err = repo.NewPostRepo(global.DB).MarkAdminLikeComment(commentID)
				if err != nil {
					zlog.CtxErrorf(ctx, "标记管理员点赞失败: %v", err)
					return resp, response.ErrResp(err, response.DATABASE_ERROR)
				}
				// 增加经验 (合理性有待商榷，暂时取消)
				//err = repo.NewUserRepo(global.DB).AddUserXp(comment.UserID, 2)
				//if err != nil {
				//	zlog.CtxErrorf(ctx, "增加经验失败: %v", err)
				//	return resp, response.ErrResp(err, response.DATABASE_ERROR)
				//}
			}
		}
		// 点赞
		id := global.SnowflakeNode.Generate().Int64()
		commentLike := model.CommentLike{
			CommentID: commentID,
			UserID:    operatorID,
			ID:        id,
		}
		err = repo.NewPostRepo(global.DB).AddCommentLike(commentLike)
		if err != nil {
			zlog.CtxErrorf(ctx, "点赞失败: %v", err)
			return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
		}
		resp.IsLike = true
	}
	// 发送点赞通知
	// 先用redis判断两小时内是否有过点赞通知，如果有，则不再发送
	key := fmt.Sprintf(REDIS_LIKE_MESSAGE, commentID, operatorID)
	if global.Rdb.Exists(ctx, key).Val() == 1 {
		zlog.CtxInfof(ctx, "两小时内有过点赞通知，不再发送")
		return
	}
	// redis 记录点赞通知
	err = global.Rdb.Set(ctx, key, "1", time.Hour*2).Err()
	if err != nil {
		zlog.CtxErrorf(ctx, "%v", err)
		return resp, response.ErrResp(err, response.REDIS_ERROR)
	}
	// 获取评论详情
	var comment model.Comment
	comment, err = repo.NewPostRepo(global.DB).GetCommentDetail(commentID)
	if err != nil {
		zlog.CtxErrorf(ctx, "查询评论详情失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	// 获取帖子详情
	var post model.Post
	post, err = repo.NewPostRepo(global.DB).GetPostDetail(comment.PostID)
	if err != nil {
		zlog.CtxErrorf(ctx, "查询帖子详情失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	// 简化评论内容 (去掉换行符)
	contentShort := comment.Content
	contentShort = strings.ReplaceAll(contentShort, "\n", " ")
	contentShort = utils.TruncateString(contentShort, 20)
	// 发送通知
	messageID := global.SnowflakeNode.Generate().Int64()
	var url string
	if post.Type == "diary" {
		url = fmt.Sprintf("/diary/%d", post.ID)
	} else {
		url = fmt.Sprintf("/learn/%d", post.ID)
	}
	message := model.Message{
		ID:       messageID,
		UserID:   comment.UserID,
		SenderID: operatorID,
		Type:     "like",
		Content:  fmt.Sprintf("赞了你的评论 [ %s ]", contentShort),
		Url:      url,
		IsRead:   false,
	}
	err = repo.NewMessageRepo(global.DB).SendMessage(message)
	if err != nil {
		zlog.CtxErrorf(ctx, "发送点赞通知失败: %v", err)
		// 发送失败，但不影响实际点赞
		err = nil
	}
	return
}

func (l *PostLogic) GetLikeComment(ctx context.Context, req types.GetLikeCommentReq) (resp types.GetLikeCommentResp, err error) {
	defer utils.RecordTime(time.Now())()
	// ID 转化为 int64
	commentID, err := strconv.ParseInt(req.CommentID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.CommentID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	operatorID, err := strconv.ParseInt(req.OperatorID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.OperatorID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 查询是否有点赞记录
	IsLike, err := repo.NewPostRepo(global.DB).IsCommentLikeExists(commentID, operatorID)
	if err != nil {
		zlog.CtxErrorf(ctx, "查询点赞状态失败: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	resp.IsLike = IsLike
	return
}

func (l *PostLogic) GetMorePosts(ctx context.Context, req types.GetMorePostsReq) (resp types.GetMorePostsResp, err error) {
	defer utils.RecordTime(time.Now())()
	// ID 转化为 int64
	beforeID, err := strconv.ParseInt(req.BeforeID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.BeforeID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 分各种情况查询帖子
	var posts []model.Post
	if req.Type == "diary" {
		// 周记类型
		if req.By == "popular" || req.By == "weight" {
			// 按热度排序
			posts, err = repo.NewPostRepo(global.DB).GetMoreDiaryByWeight(req.Source, beforeID, req.Count)
		} else if req.By == "new" {
			// 按最新排序
			posts, err = repo.NewPostRepo(global.DB).GetMoreDiaryByID(req.Source, beforeID, req.Count)
		} else if req.By == "user" {
			// 查看个人
			var userID int64
			userID, err = strconv.ParseInt(req.UserID, 10, 64)
			if err != nil {
				zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.UserID, err)
				return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
			}
			posts, err = repo.NewPostRepo(global.DB).GetMoreDiaryByUser(userID, beforeID, req.Count)
		} else {
			zlog.CtxErrorf(ctx, "类型错误: %v", req.Type)
			return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
		}
	} else if req.Type == "post" {
		// 各种帖子类型
	} else {
		// 不存在的类型
		zlog.CtxErrorf(ctx, "类型错误: %v", req.Type)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 数据库查询失败
	if err != nil {
		zlog.CtxErrorf(ctx, "查询帖子失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	//zlog.CtxDebugf(ctx, "查询帖子成功: %v", posts)
	for _, post := range posts {
		// 截短内容
		contentShort := post.Content
		// 去掉换行符
		contentShort = strings.ReplaceAll(contentShort, "\n", " ")
		if len(contentShort) > 300 {
			contentShort = contentShort[:300]
		}
		if post.IsPrivate {
			contentShort = "......"
		}
		// 组装返回数据
		resp.Posts = append(resp.Posts, types.PostInfo{
			ID:           post.ID,
			UserID:       post.UserID,
			Title:        post.Title,
			ContentShort: contentShort,
			Type:         post.Type,
			Source:       post.Source,
			Likes:        post.Likes,
			Comments:     post.Comments,
			CreatedAt:    post.CreatedTime,
			UpdatedAt:    post.UpdatedTime,

			IsAdminLike: post.IsAdminLike,
			IsPrivate:   post.IsPrivate,
			IsFeatured:  post.IsFeatured,

			Weight: post.Weight,
		})
	}
	resp.Length = len(resp.Posts)
	return
}

func (l *PostLogic) GetPagePosts(ctx context.Context, req types.GetPagePostsReq) (resp types.GetPagePostsResp, err error) {
	defer utils.RecordTime(time.Now())()
	// 分各种情况查询帖子
	var posts []model.Post

	if req.By == "popular" || req.By == "weight" || req.By == "hot" {
		// 按热度排序
		posts, resp.PageTotal, err = repo.NewPostRepo(global.DB).GetPagePostByWeight(req.Type, req.Page, req.Count)
	} else if req.By == "new" || req.By == "time" {
		// 按最新排序
		posts, resp.PageTotal, err = repo.NewPostRepo(global.DB).GetPagePostByID(req.Type, req.Page, req.Count)
	} else if req.By == "featured" {
		// 精选
		posts, resp.PageTotal, err = repo.NewPostRepo(global.DB).GetPagePostByFeatured(req.Type, req.Page, req.Count)
	} else if req.By == "source" {
		// 按来源排序
		posts, resp.PageTotal, err = repo.NewPostRepo(global.DB).GetPagePostBySource(req.Type, req.Source, req.Page, req.Count)
	} else {
		zlog.CtxErrorf(ctx, "类型错误: %v", req.Type)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}

	// 数据库查询失败
	if err != nil {
		zlog.CtxErrorf(ctx, "查询帖子失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	//zlog.CtxDebugf(ctx, "查询帖子成功: %v", posts)
	for _, post := range posts {
		// 截短内容
		contentShort := post.Content
		// 去掉换行符
		contentShort = strings.ReplaceAll(contentShort, "\n", " ")
		if len(contentShort) > 300 {
			contentShort = contentShort[:300]
		}
		if post.IsPrivate {
			contentShort = "......"
		}
		// 组装返回数据
		resp.Posts = append(resp.Posts, types.PostInfo{
			ID:           post.ID,
			UserID:       post.UserID,
			Title:        post.Title,
			ContentShort: contentShort,
			Type:         post.Type,
			Source:       post.Source,
			Likes:        post.Likes,
			Comments:     post.Comments,
			CreatedAt:    post.CreatedTime,
			UpdatedAt:    post.UpdatedTime,

			IsAdminLike: post.IsAdminLike,
			IsPrivate:   post.IsPrivate,
			IsFeatured:  post.IsFeatured,

			Weight: post.Weight,
		})
	}
	resp.Length = len(resp.Posts)
	if resp.PageTotal%int64(req.Count) == 0 {
		resp.PageTotal = resp.PageTotal / int64(req.Count)
	} else {
		resp.PageTotal = resp.PageTotal/int64(req.Count) + 1
	}
	return
}

func (l *PostLogic) SetPostFeature(ctx context.Context, req types.SetPostFeatureReq) (resp types.SetPostFeatureResp, err error) {
	defer utils.RecordTime(time.Now())()
	// ID 转化为 int64
	postID, err := strconv.ParseInt(req.PostID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.PostID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 数据库操作
	err = repo.NewPostRepo(global.DB).SetPostFeature(postID)
	if err != nil {
		zlog.CtxErrorf(ctx, "设置帖子为精华失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	// 获得作者id
	post, err := repo.NewPostRepo(global.DB).GetPostDetail(postID)
	if err != nil {
		zlog.CtxErrorf(ctx, "查询帖子详情失败: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 为作者增加经验
	err = repo.NewUserRepo(global.DB).AddUserXp(post.UserID, 20)
	if err != nil {
		zlog.CtxErrorf(ctx, "增加经验失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	return
}

func GetWeekCode() string {
	timeNow := time.Now()
	//timeNow = time.UnixMilli(1743914958000)
	timestamp := timeNow.UnixMilli()

	// 打卡时间为每周的周日中午到周二的中午，为了先确定当前周数，先把时间减去 2 天
	timestamp -= 2 * 24 * 60 * 60 * 1000
	// 周一到周日分别为 1 到 7
	weekday := int64((time.UnixMilli(timestamp).Weekday()+6)%7 + 1)

	// 时间返回到周一中午 12:00:00
	timestamp -= (weekday - 1) * 24 * 60 * 60 * 1000
	timestamp -= (timestamp + 8*3600*1000) % (24 * 60 * 60 * 1000) // 取整到天
	timestamp += 12 * 60 * 60 * 1000                               // 加上中午(UTC+8)
	// 如果当前时间不在合法打卡
	if utils.Abs(timeNow.UnixMilli()-(timestamp+7*24*60*60*1000)) > 24*60*60*1000 {
		zlog.Debugf("%v = %v", utils.Abs(timeNow.UnixMilli()-(timestamp+7*24*60*60*1000)), 24*60*60*1000)
		zlog.Warnf("当前时间不在合法打卡时间范围内 %v ~ %v", time.UnixMilli(timestamp+7*24*60*60*1000), timeNow)
		return ""
	}
	// 计算当前周数，确定年份和月份
	week := 0
	year := time.UnixMilli(timestamp).Year()
	month := time.UnixMilli(timestamp).Month()
	for month == time.UnixMilli(timestamp).Month() {
		timestamp -= 7 * 24 * 60 * 60 * 1000
		week++
	}
	// 格式化周数
	weekCode := fmt.Sprintf("%d-%d-%d", year, month, week)
	return weekCode
}

func (l *PostLogic) GetDiaryList(ctx context.Context, req types.GetDiaryListReq) (resp types.GetDiaryListResp, err error) {
	defer utils.RecordTime(time.Now())()
	// id 转化为 int64
	userID, err := strconv.ParseInt(req.UserID, 10, 64)
	if err != nil {
		zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.UserID, err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 查询周记列表
	var Posts []model.Post
	Posts, err = repo.NewPostRepo(global.DB).GetDiaryList(userID)
	if err != nil {
		zlog.CtxErrorf(ctx, "查询周记列表失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	// 放入resp
	for _, post := range Posts {
		resp.Posts = append(resp.Posts, types.DiaryInfo{
			PostID: post.ID,
			Source: post.Source,
		})
	}
	resp.Length = len(resp.Posts)
	return resp, nil
}

func (l *PostLogic) SearchPosts(ctx context.Context, req types.SearchPostsReq) (resp types.SearchPostsResp, err error) {
	defer utils.RecordTime(time.Now())()
	// 查询 ElasticSearch
	// 查询测试
	query := `{
	 "query": {
		"query_string": {
		  "query": "%s",
		  "fields": ["*"],
		  "analyze_wildcard": true
		}
	 },
     "from": %d,
	 "size": %d
	}`
	query = fmt.Sprintf(query, req.Keyword, (req.Page-1)*req.Count, req.Count)
	m, err := elasticSearchUtils.Search(global.ESClient, "post", query)
	if err != nil {
		zlog.CtxErrorf(ctx, "查询失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	// 解析结果
	resp.Length = len(m["hits"].(map[string]interface{})["hits"].([]interface{}))
	zlog.Debugf("查询数量为: %d", resp.Length)
	resp.PageTotal = int64(m["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64))
	zlog.Debugf("总数量为: %v", resp.PageTotal)
	if resp.PageTotal%int64(req.Count) == 0 {
		resp.PageTotal = resp.PageTotal / int64(req.Count)
	} else {
		resp.PageTotal = resp.PageTotal/int64(req.Count) + 1
	}
	// 拿取ID，然后从数据库中查询详细信息
	for _, hit := range m["hits"].(map[string]interface{})["hits"].([]interface{}) {
		postIDStr := hit.(map[string]interface{})["_id"].(string)
		zlog.Debugf("postID: %s", postIDStr)
		// 转换为 int64
		postID, err := strconv.ParseInt(postIDStr, 10, 64)
		if err != nil {
			zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", postID, err)
			return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
		}
		// 查询数据库
		post, err := repo.NewPostRepo(global.DB).GetPostDetail(postID)
		if err != nil {
			zlog.CtxErrorf(ctx, "查询帖子详情失败: %v", err)
			return resp, response.ErrResp(err, response.DATABASE_ERROR)
		}
		// 组装返回数据
		contentShort := post.Content
		// 去掉换行符
		contentShort = strings.ReplaceAll(contentShort, "\n", " ")
		if len(contentShort) > 300 {
			contentShort = contentShort[:300]
		}
		if post.IsPrivate {
			contentShort = "......"
		}
		resp.Posts = append(resp.Posts, types.PostInfo{
			ID:           post.ID,
			UserID:       post.UserID,
			Title:        post.Title,
			ContentShort: contentShort,
			Type:         post.Type,
			Source:       post.Source,
			Likes:        post.Likes,
			Comments:     post.Comments,
			CreatedAt:    post.CreatedTime,
			UpdatedAt:    post.UpdatedTime,

			IsAdminLike: post.IsAdminLike,
			IsPrivate:   post.IsPrivate,
			IsFeatured:  post.IsFeatured,

			Weight: post.Weight,
		})
	}

	//zlog.Debugf("查询结果: %v", m)

	return
}
