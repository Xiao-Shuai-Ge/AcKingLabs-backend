package logic

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"tgwp/global"
	"tgwp/internal/utils/messageService"
	"tgwp/internal/utils/notifyService"
	"tgwp/internal/utils/task"
	"tgwp/log/zlog"
	"tgwp/model"
	"tgwp/repo"
	"tgwp/response"
	"tgwp/types"
	"tgwp/utils"
	"tgwp/utils/elasticSearchUtils"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"
)

const (
	REDIS_LIKE_MESSAGE = "like_message:%d:%d"
)

type PostLogic struct {
}

// getPostMaxLength 根据用户角色获取帖子最大长度限制
func getPostMaxLength(role int) int {
	if role >= global.ROLE_ADMIN {
		return global.POST_MAX_LENGTH_ADMIN
	}
	if role >= global.ROLE_PLAYER {
		return global.POST_MAX_LENGTH_PLAYER
	}
	return global.POST_MAX_LENGTH_USER
}

// getCommentMaxLength 根据用户角色获取评论最大长度限制
func getCommentMaxLength(role int) int {
	if role >= global.ROLE_ADMIN {
		return global.COMMENT_MAX_LENGTH_ADMIN
	}
	if role >= global.ROLE_PLAYER {
		return global.COMMENT_MAX_LENGTH_PLAYER
	}
	return global.COMMENT_MAX_LENGTH_USER
}

func NewPostLogic() *PostLogic {
	return &PostLogic{}
}

// CreatePost 创建帖子
func (l *PostLogic) CreatePost(ctx context.Context, req types.CreatePostReq) (resp types.CreatePostResp, err error) {
	defer utils.CtxRecordTime(ctx, time.Now())()
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
	// 1. 标题不能超过 50 个字符
	if utf8.RuneCountInString(req.Title) > 50 {
		zlog.CtxErrorf(ctx, "标题不能超过 50 个字符: %v", err)
		return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
	}
	// 2. 根据用户角色判断内容长度限制
	maxLength := getPostMaxLength(req.UserRole)
	contentLength := utf8.RuneCountInString(req.Content)
	zlog.CtxInfof(ctx, "内容长度: %d, 用户角色: %d, 最大长度限制: %d", contentLength, req.UserRole, maxLength)
	if contentLength > maxLength {
		zlog.CtxErrorf(ctx, "内容不能超过 %d 个字: %v", maxLength, err)
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

	// 如果是求助帖，发送通知给开启了求助帖通知的用户
	if req.Type == "help" {
		url := fmt.Sprintf("/learn/%d", id)
		go notifyService.SendHelpPostNotify(ctx, req.Title, url, userID)
	}
	// 给作者加 XP
	addXp := 4
	if req.IsPrivate == false {
		addXp = 8
	}

	// 添加 AI 审核任务
	task.GlobalDispatcher.AddJob(task.Job{
		Type: task.JOB_TYPE_AI_AUDIT,
		Payload: task.AIAuditPayload{
			PostID:     id,
			UserID:     userID,
			Content:    req.Title + "\n" + req.Content,
			SenderRole: req.UserRole,
		},
	})

	// 如果是帖子而不是周记，不参与经验值计算
	if req.Type != "diary" {
		return
	}

	err = repo.NewUserRepo(global.DB).AddUserXp(userID, addXp)
	if err != nil {
		zlog.CtxErrorf(ctx, "给作者加 XP 失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}

	// 发送经验增加通知
	var url string
	if req.Type == "diary" {
		url = fmt.Sprintf("/diary/%d", id)
	} else {
		url = fmt.Sprintf("/learn/%d", id)
	}
	reason := "发布帖子"
	if req.Type == "diary" {
		reason = "发布周记"
	}
	notifyService.SendSystemMessageNotify(ctx, userID, fmt.Sprintf("获得 %d 经验值：%s《%s》", addXp, reason, req.Title), url)

	return
}

// EditPost 编辑帖子
func (l *PostLogic) EditPost(ctx context.Context, req types.EditPostReq) (resp types.EditPostResp, err error) {
	defer utils.CtxRecordTime(ctx, time.Now())()
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
	// 2. 根据用户角色判断内容长度限制
	maxLength := getPostMaxLength(req.OperatorRole)
	contentLength := utf8.RuneCountInString(req.Content)
	zlog.CtxInfof(ctx, "内容长度: %d, 用户角色: %d, 最大长度限制: %d", contentLength, req.OperatorRole, maxLength)
	if contentLength > maxLength {
		zlog.CtxErrorf(ctx, "内容不能超过 %d 个字: %v", maxLength, err)
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
	// 如果不是管理员，不允许更改类型和是否私密
	if req.OperatorRole < global.ROLE_ADMIN {
		if req.Type != post.Type || req.IsPrivate != post.IsPrivate {
			zlog.CtxErrorf(ctx, "非管理员不能更改类型和是否私密: %v", err)
			return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
		}
	}
	// 周记类型不允许改来源和类型
	if post.Type == "diary" {
		if req.Type != "diary" || req.Source != post.Source {
			zlog.CtxErrorf(ctx, "周记类型不允许改来源和类型: %v", err)
			return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
		}
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
	// 如果是官方贴，需要重新计算热度
	task.GlobalDispatcher.AddJob(task.Job{Type: task.JOB_TYPE_COMPUTE_POST_WEIGHT, Payload: task.ComputePostWeightPayload{PostID: postID}})

	// 添加 AI 审核任务
	task.GlobalDispatcher.AddJob(task.Job{
		Type: task.JOB_TYPE_AI_AUDIT,
		Payload: task.AIAuditPayload{
			PostID:     postID,
			UserID:     post.UserID,
			Content:    req.Title + "\n" + req.Content,
			SenderRole: req.OperatorRole,
		},
	})

	return
}

// DeletePost 编辑帖子
func (l *PostLogic) DeletePost(ctx context.Context, req types.DeletePostReq) (resp types.DeletePostResp, err error) {
	defer utils.CtxRecordTime(ctx, time.Now())()
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
	// 删除帖子相关点赞记录
	err = repo.NewPostRepo(global.DB).DeletePostLikeByPostID(postID)
	// 获取所有评论
	comments, err := repo.NewPostRepo(global.DB).GetAllCommentsByPostID(postID)
	if err != nil {
		zlog.CtxErrorf(ctx, "查询评论失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	for _, comment := range comments {
		// 删除点赞记录
		err = repo.NewPostRepo(global.DB).DeleteCommentLikeByCommentID(comment.ID)
		// 获取所有子评论
		childComments, err := repo.NewPostRepo(global.DB).GetAllChildCommentsByCommentID(comment.ID)
		if err != nil {
			zlog.CtxErrorf(ctx, "查询子评论失败: %v", err)
			return resp, response.ErrResp(err, response.DATABASE_ERROR)
		}
		for _, childComment := range childComments {
			// 删除子评论点赞记录
			err = repo.NewPostRepo(global.DB).DeleteCommentLikeByCommentID(childComment.ID)
			if err != nil {
				zlog.CtxErrorf(ctx, "删除子评论点赞记录失败: %v", err)
				return resp, response.ErrResp(err, response.DATABASE_ERROR)
			}
			// 删除子评论
			err = repo.NewPostRepo(global.DB).DeleteCommentByCommentID(childComment.ID, comment.PostID)
		}
		// 删除评论
		err = repo.NewPostRepo(global.DB).DeleteCommentByCommentID(comment.ID, comment.PostID)
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
	defer utils.CtxRecordTime(ctx, time.Now())()
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
	defer utils.CtxRecordTime(ctx, time.Now())()
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
		task.GlobalDispatcher.AddJob(task.Job{Type: task.JOB_TYPE_COMPUTE_POST_WEIGHT, Payload: task.ComputePostWeightPayload{PostID: postID}})
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
				// 发送经验增加通知
				var url string
				if post.Type == "diary" {
					url = fmt.Sprintf("/diary/%d", post.ID)
				} else {
					url = fmt.Sprintf("/learn/%d", post.ID)
				}
				notifyService.SendSystemMessageNotify(ctx, post.UserID, fmt.Sprintf("获得 5 经验值：管理员点赞帖子《%s》", post.Title), url)
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
	// 计算帖子权重
	task.GlobalDispatcher.AddJob(task.Job{Type: task.JOB_TYPE_COMPUTE_POST_WEIGHT, Payload: task.ComputePostWeightPayload{PostID: postID}})

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
	// 使用新的消息服务发送通知
	var url string
	if post.Type == "diary" {
		url = fmt.Sprintf("/diary/%d", post.ID)
	} else {
		url = fmt.Sprintf("/learn/%d", post.ID)
	}
	// 使用新的通知服务（站内消息 + 邮件通知）
	notifyService.SendLikeNotify(
		ctx,
		post.UserID,
		operatorID,
		fmt.Sprintf("赞了你的帖子 《%s》", post.Title),
		url,
	)
	return
}

func (l *PostLogic) GetLikePost(ctx context.Context, req types.GetLikePostReq) (resp types.GetLikePostResp, err error) {
	defer utils.CtxRecordTime(ctx, time.Now())()
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
	defer utils.CtxRecordTime(ctx, time.Now())()
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
	// 根据用户角色判断评论内容长度限制
	maxLength := getCommentMaxLength(req.UserRole)
	contentLength := utf8.RuneCountInString(req.Content)
	zlog.CtxInfof(ctx, "评论内容长度: %d, 用户角色: %d, 最大长度限制: %d", contentLength, req.UserRole, maxLength)
	if contentLength > maxLength {
		zlog.CtxErrorf(ctx, "评论内容不能超过 %d 个字符: %v", maxLength, err)
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
	var url string
	if post.Type == "diary" {
		url = fmt.Sprintf("/diary/%d", post.ID)
	} else {
		url = fmt.Sprintf("/learn/%d", post.ID)
	}
	// 判断是几级评论，给出对应的提示
	if fatherID == 0 {
		// 一级评论：只给帖子作者发送通知
		content := fmt.Sprintf("在你的帖子 《%s》 评论了: [%s]", post.Title, contentShort)
		notifyService.SendReplyNotify(ctx, post.UserID, userID, content, url)
	} else {
		// 子评论：需要给帖主和评论作者都发送通知
		var fatherComment model.Comment
		fatherComment, err = repo.NewPostRepo(global.DB).GetCommentDetail(fatherID)
		if err != nil {
			zlog.CtxErrorf(ctx, "查询父评论详情失败: %v", err)
			return resp, response.ErrResp(err, response.DATABASE_ERROR)
		}

		fatherContentShort := fatherComment.Content
		fatherContentShort = strings.ReplaceAll(fatherContentShort, "\n", " ")
		fatherContentShort = utils.TruncateString(fatherContentShort, 20)

		// 给评论作者发送通知：你的评论被回复了
		commentContent := fmt.Sprintf("在你的评论 [%s] 回复了: [%s]", fatherContentShort, contentShort)
		notifyService.SendReplyNotify(ctx, fatherComment.UserID, userID, commentContent, url)

		// 给帖子作者发送通知：你的帖子有新回复（如果不是同一个人）
		if post.UserID != fatherComment.UserID {
			postContent := fmt.Sprintf("在你的帖子 《%s》 发布子评论: [%s]", post.Title, contentShort)
			notifyService.SendReplyNotify(ctx, post.UserID, userID, postContent, url)
		}
	}
	task.GlobalDispatcher.AddJob(task.Job{Type: task.JOB_TYPE_COMPUTE_POST_WEIGHT, Payload: task.ComputePostWeightPayload{PostID: postID}})
	return
}

func (l *PostLogic) DeleteComment(ctx context.Context, req types.DeleteCommentReq) (resp types.DeleteCommentResp, err error) {
	defer utils.CtxRecordTime(ctx, time.Now())()
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
	// 查询评论详情
	var comment model.Comment
	comment, err = repo.NewPostRepo(global.DB).GetCommentDetail(commentID)
	if err != nil {
		zlog.CtxErrorf(ctx, "查询评论详情失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	// 判断是否有权限删除
	if comment.UserID != operatorID && !(req.OperatorRole >= global.ROLE_ADMIN) {
		zlog.CtxErrorf(ctx, "无权限删除评论")
		return resp, response.ErrResp(err, response.PERMISSION_DENIED)
	}
	// 删除点赞记录
	err = repo.NewPostRepo(global.DB).DeleteCommentLikeByCommentID(commentID)
	if err != nil {
		zlog.CtxErrorf(ctx, "删除点赞记录失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	// 获取全部子评论
	childComments, err := repo.NewPostRepo(global.DB).GetAllChildCommentsByCommentID(commentID)
	if err != nil {
		zlog.CtxErrorf(ctx, "查询子评论失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	for _, childComment := range childComments {
		// 删除子评论点赞记录
		err = repo.NewPostRepo(global.DB).DeleteCommentLikeByCommentID(childComment.ID)
		if err != nil {
			zlog.CtxErrorf(ctx, "删除子评论点赞记录失败: %v", err)
			return
		}
		// 删除子评论
		err = repo.NewPostRepo(global.DB).DeleteCommentByCommentID(childComment.ID, comment.PostID)
		if err != nil {
			zlog.CtxErrorf(ctx, "删除子评论失败: %v", err)
			return
		}
	}
	// 删除评论
	err = repo.NewPostRepo(global.DB).DeleteCommentByCommentID(commentID, comment.PostID)
	if err != nil {
		zlog.CtxErrorf(ctx, "删除评论失败: %v", err)
		return resp, response.ErrResp(err, response.DATABASE_ERROR)
	}
	// 计算帖子权重
	task.GlobalDispatcher.AddJob(task.Job{Type: task.JOB_TYPE_COMPUTE_POST_WEIGHT, Payload: task.ComputePostWeightPayload{PostID: comment.PostID}})
	return
}

func (l *PostLogic) GetMoreComments(ctx context.Context, req types.GetMoreCommentsReq) (resp types.GetMoreCommentsResp, err error) {
	defer utils.CtxRecordTime(ctx, time.Now())()
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
	defer utils.CtxRecordTime(ctx, time.Now())()
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
				// 判断增加经验条件 (帖子为求助帖，且评论为一级评论)
				if post.Type == "help" && comment.FatherID == 0 {
					err = repo.NewUserRepo(global.DB).AddUserXp(comment.UserID, 4)
					if err != nil {
						zlog.CtxErrorf(ctx, "增加经验失败: %v", err)
						return resp, response.ErrResp(err, response.DATABASE_ERROR)
					}
					// 发送经验增加通知
					messageService.SendSystemMessage(comment.UserID, fmt.Sprintf("获得 4 经验值：管理员将你的评论标记为优质解答 \"%s\" ", comment.Content), fmt.Sprintf("/learn/%d", post.ID))
				}
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
	// 简化评论内容 (去掉换行符)
	contentShort := comment.Content
	contentShort = strings.ReplaceAll(contentShort, "\n", " ")
	contentShort = utils.TruncateString(contentShort, 20)
	// 发送通知
	var url string
	if post.Type == "diary" {
		url = fmt.Sprintf("/diary/%d", post.ID)
	} else {
		url = fmt.Sprintf("/learn/%d", post.ID)
	}
	// 使用新的通知服务发送通知（站内消息 + 邮件通知）
	notifyService.SendLikeNotify(
		ctx,
		comment.UserID,
		operatorID,
		fmt.Sprintf("赞了你的评论 [ %s ]", contentShort),
		url,
	)
	return
}

func (l *PostLogic) GetLikeComment(ctx context.Context, req types.GetLikeCommentReq) (resp types.GetLikeCommentResp, err error) {
	defer utils.CtxRecordTime(ctx, time.Now())()
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
	defer utils.CtxRecordTime(ctx, time.Now())()
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
	defer utils.CtxRecordTime(ctx, time.Now())()
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
	} else if req.By == "user" {
		// 查看个人
		var userID int64
		userID, err = strconv.ParseInt(req.UserID, 10, 64)
		if err != nil {
			zlog.CtxErrorf(ctx, "%v 转换 int64 错误: %v", req.UserID, err)
			return resp, response.ErrResp(err, response.PARAM_NOT_VALID)
		}
		posts, resp.PageTotal, err = repo.NewPostRepo(global.DB).GetPagePostByUser(req.Type, userID, req.Page, req.Count)
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
	defer utils.CtxRecordTime(ctx, time.Now())()
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

	// 发送经验增加通知
	var url string
	if post.Type == "diary" {
		url = fmt.Sprintf("/diary/%d", postID)
	} else {
		url = fmt.Sprintf("/learn/%d", postID)
	}
	messageService.SendSystemMessage(post.UserID, fmt.Sprintf("获得 20 经验值：帖子《%s》被设为精华", post.Title), url)
	// 计算帖子权重
	task.GlobalDispatcher.AddJob(task.Job{Type: task.JOB_TYPE_COMPUTE_POST_WEIGHT, Payload: task.ComputePostWeightPayload{PostID: postID}})
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
	defer utils.CtxRecordTime(ctx, time.Now())()
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
	defer utils.CtxRecordTime(ctx, time.Now())()
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
